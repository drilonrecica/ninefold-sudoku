package config

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	maxReplayRetention         = 7 * 24 * time.Hour
	maxTombstoneRetention      = 30 * 24 * time.Hour
	maxCommandReceiptRetention = 24 * time.Hour
)

var safeToken = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)

type Environment string

const (
	Development Environment = "development"
	Test        Environment = "test"
	Production  Environment = "production"
)

type Config struct {
	Environment             Environment
	PublicURL               *url.URL
	HTTPAddress             string
	DatabasePath            string
	AllowedOrigins          []string
	CookieSecret            []byte
	ReplaySigningKey        ed25519.PrivateKey
	ReplaySigningKeyID      string
	AdminProxyHeader        string
	LogLevel                string
	ReplayRetention         time.Duration
	MatchTombstoneRetention time.Duration
	CommandReceiptRetention time.Duration
	ShutdownTimeout         time.Duration
}

type LookupFunc func(string) (string, bool)

func Load() (Config, error) {
	return Parse(os.LookupEnv)
}

func Parse(lookup LookupFunc) (Config, error) {
	required := func(name string) (string, error) {
		value, ok := lookup(name)
		if !ok || strings.TrimSpace(value) == "" {
			return "", fmt.Errorf("%s is required", name)
		}
		return strings.TrimSpace(value), nil
	}

	var cfg Config
	environment, err := required("NINEFOLD_ENVIRONMENT")
	if err != nil {
		return cfg, err
	}
	cfg.Environment = Environment(environment)
	if !slices.Contains([]Environment{Development, Test, Production}, cfg.Environment) {
		return cfg, errors.New("NINEFOLD_ENVIRONMENT must be development, test, or production")
	}

	publicURL, err := required("NINEFOLD_PUBLIC_URL")
	if err != nil {
		return cfg, err
	}
	cfg.PublicURL, err = parseHTTPURL("NINEFOLD_PUBLIC_URL", publicURL, cfg.Environment == Production)
	if err != nil {
		return cfg, err
	}

	cfg.HTTPAddress, err = required("NINEFOLD_HTTP_ADDRESS")
	if err != nil {
		return cfg, err
	}
	if _, _, err := net.SplitHostPort(cfg.HTTPAddress); err != nil {
		return cfg, fmt.Errorf("NINEFOLD_HTTP_ADDRESS: %w", err)
	}

	cfg.DatabasePath, err = required("NINEFOLD_DATABASE_PATH")
	if err != nil {
		return cfg, err
	}
	if cfg.DatabasePath == ":memory:" || strings.ContainsRune(cfg.DatabasePath, '\x00') || filepath.Clean(cfg.DatabasePath) == "." {
		return cfg, errors.New("NINEFOLD_DATABASE_PATH must name a database file")
	}

	origins, err := required("NINEFOLD_ALLOWED_ORIGINS")
	if err != nil {
		return cfg, err
	}
	for _, raw := range strings.Split(origins, ",") {
		origin, parseErr := parseHTTPURL("NINEFOLD_ALLOWED_ORIGINS", strings.TrimSpace(raw), cfg.Environment == Production)
		if parseErr != nil {
			return cfg, parseErr
		}
		if origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
			return cfg, errors.New("NINEFOLD_ALLOWED_ORIGINS entries must be origins without path, query, or fragment")
		}
		cfg.AllowedOrigins = append(cfg.AllowedOrigins, origin.String())
	}

	cookieSecret, err := required("NINEFOLD_COOKIE_SECRET")
	if err != nil {
		return cfg, err
	}
	cfg.CookieSecret, err = base64.StdEncoding.DecodeString(cookieSecret)
	if err != nil || len(cfg.CookieSecret) < 32 {
		return cfg, errors.New("NINEFOLD_COOKIE_SECRET must be base64 encoding at least 32 bytes")
	}

	signingKey, err := required("NINEFOLD_REPLAY_SIGNING_KEY")
	if err != nil {
		return cfg, err
	}
	der, err := base64.StdEncoding.DecodeString(signingKey)
	if err != nil {
		return cfg, errors.New("NINEFOLD_REPLAY_SIGNING_KEY must be base64")
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return cfg, errors.New("NINEFOLD_REPLAY_SIGNING_KEY must contain a PKCS#8 private key")
	}
	var ok bool
	cfg.ReplaySigningKey, ok = parsedKey.(ed25519.PrivateKey)
	if !ok {
		return cfg, errors.New("NINEFOLD_REPLAY_SIGNING_KEY must be Ed25519")
	}

	cfg.ReplaySigningKeyID, err = required("NINEFOLD_REPLAY_SIGNING_KEY_ID")
	if err != nil || !safeToken.MatchString(cfg.ReplaySigningKeyID) {
		return cfg, errors.New("NINEFOLD_REPLAY_SIGNING_KEY_ID must be a safe token of 1-64 characters")
	}
	cfg.AdminProxyHeader, err = required("NINEFOLD_ADMIN_PROXY_HEADER")
	if err != nil || !safeToken.MatchString(cfg.AdminProxyHeader) {
		return cfg, errors.New("NINEFOLD_ADMIN_PROXY_HEADER must be a safe HTTP header name of 1-64 characters")
	}
	if http.CanonicalHeaderKey(cfg.AdminProxyHeader) == "" {
		return cfg, errors.New("NINEFOLD_ADMIN_PROXY_HEADER must be a valid HTTP header name")
	}

	cfg.LogLevel, err = required("NINEFOLD_LOG_LEVEL")
	if err != nil {
		return cfg, err
	}
	if !slices.Contains([]string{"debug", "info", "warn", "error"}, cfg.LogLevel) {
		return cfg, errors.New("NINEFOLD_LOG_LEVEL must be debug, info, warn, or error")
	}

	if cfg.ReplayRetention, err = duration(lookup, "NINEFOLD_REPLAY_RETENTION", maxReplayRetention); err != nil {
		return cfg, err
	}
	if cfg.MatchTombstoneRetention, err = duration(lookup, "NINEFOLD_MATCH_TOMBSTONE_RETENTION", maxTombstoneRetention); err != nil {
		return cfg, err
	}
	if cfg.CommandReceiptRetention, err = duration(lookup, "NINEFOLD_COMMAND_RECEIPT_RETENTION", maxCommandReceiptRetention); err != nil {
		return cfg, err
	}
	if cfg.ShutdownTimeout, err = duration(lookup, "NINEFOLD_SHUTDOWN_TIMEOUT", time.Minute); err != nil {
		return cfg, err
	}

	if cfg.Environment == Production {
		zeroSeedKey := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
		if cfg.ReplaySigningKey.Equal(zeroSeedKey) {
			return cfg, errors.New("NINEFOLD_REPLAY_SIGNING_KEY contains the test-only development key")
		}
		for name, value := range map[string]string{
			"NINEFOLD_COOKIE_SECRET":         cookieSecret,
			"NINEFOLD_REPLAY_SIGNING_KEY":    signingKey,
			"NINEFOLD_REPLAY_SIGNING_KEY_ID": cfg.ReplaySigningKeyID,
			"NINEFOLD_ADMIN_PROXY_HEADER":    cfg.AdminProxyHeader,
		} {
			lower := strings.ToLower(value)
			if strings.Contains(lower, "placeholder") || strings.Contains(lower, "change-me") ||
				strings.Trim(value, "A=") == "" {
				return cfg, fmt.Errorf("%s contains a production placeholder", name)
			}
		}
	}

	return cfg, nil
}

func parseHTTPURL(name, raw string, httpsRequired bool) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("%s must be an absolute HTTP(S) URL", name)
	}
	if httpsRequired && parsed.Scheme != "https" {
		return nil, fmt.Errorf("%s must use HTTPS in production", name)
	}
	return parsed, nil
}

func duration(lookup LookupFunc, name string, maximum time.Duration) (time.Duration, error) {
	raw, ok := lookup(name)
	if !ok || strings.TrimSpace(raw) == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil || value <= 0 || value > maximum {
		return 0, fmt.Errorf("%s must be a positive duration no greater than %s", name, maximum)
	}
	return value, nil
}

func (c Config) Sanitized() map[string]string {
	return map[string]string{
		"environment":               string(c.Environment),
		"public_url":                c.PublicURL.String(),
		"http_address":              c.HTTPAddress,
		"database_path":             c.DatabasePath,
		"allowed_origins":           strings.Join(c.AllowedOrigins, ","),
		"replay_signing_key_id":     c.ReplaySigningKeyID,
		"admin_proxy_header":        c.AdminProxyHeader,
		"log_level":                 c.LogLevel,
		"replay_retention":          c.ReplayRetention.String(),
		"tombstone_retention":       c.MatchTombstoneRetention.String(),
		"command_receipt_retention": c.CommandReceiptRetention.String(),
		"shutdown_timeout":          c.ShutdownTimeout.String(),
		"configured":                strconv.FormatBool(len(c.CookieSecret) >= 32 && len(c.ReplaySigningKey) > 0),
	}
}
