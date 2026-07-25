package deploymentconfig

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:[-+][A-Za-z0-9.-]+)?$`)

type Options struct {
	PublicURL string
	Version   string
	Output    string
	Random    io.Reader
}

// Write creates a production environment file atomically with respect to
// pre-existing paths. It never returns generated secret values.
func Write(options Options) error {
	if options.Output == "" {
		return errors.New("output path is required")
	}
	publicURL, err := url.ParseRequestURI(strings.TrimSpace(options.PublicURL))
	if err != nil || publicURL.Scheme != "https" || publicURL.Hostname() == "" ||
		publicURL.User != nil || publicURL.RawQuery != "" || publicURL.Fragment != "" ||
		(publicURL.Path != "" && publicURL.Path != "/") {
		return errors.New("public URL must be an HTTPS origin")
	}
	if !versionPattern.MatchString(options.Version) {
		return errors.New("version must be a semantic version without a v prefix")
	}
	random := options.Random
	if random == nil {
		random = rand.Reader
	}

	publicKey, privateKey, err := ed25519.GenerateKey(random)
	if err != nil {
		return fmt.Errorf("generate replay signing key: %w", err)
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("encode replay signing key: %w", err)
	}
	cookieSecret, err := randomBytes(random, 32)
	if err != nil {
		return fmt.Errorf("generate cookie secret: %w", err)
	}
	proxySecret, err := randomBytes(random, 32)
	if err != nil {
		return fmt.Errorf("generate proxy secret: %w", err)
	}
	keyDigest := sha256.Sum256(publicKey)
	keyID := "replay-" + hex.EncodeToString(keyDigest[:8])
	origin := strings.TrimSuffix(publicURL.String(), "/")

	values := [][2]string{
		{"NINEFOLD_VERSION", options.Version},
		{"NINEFOLD_DOMAIN", publicURL.Hostname()},
		{"NINEFOLD_ENVIRONMENT", "production"},
		{"NINEFOLD_PUBLIC_URL", origin},
		{"NINEFOLD_HTTP_ADDRESS", "0.0.0.0:8080"},
		{"NINEFOLD_DATABASE_PATH", "/app/data/ninefold.db"},
		{"NINEFOLD_ALLOWED_ORIGINS", origin},
		{"NINEFOLD_COOKIE_SECRET", base64.StdEncoding.EncodeToString(cookieSecret)},
		{"NINEFOLD_REPLAY_SIGNING_KEY", base64.StdEncoding.EncodeToString(privateDER)},
		{"NINEFOLD_REPLAY_SIGNING_KEY_ID", keyID},
		{"NINEFOLD_REPLAY_PUBLIC_KEY", base64.StdEncoding.EncodeToString(publicKey)},
		{"NINEFOLD_PROXY_SECRET", base64.StdEncoding.EncodeToString(proxySecret)},
		{"NINEFOLD_ADMIN_PROXY_HEADER", "X-Ninefold-Admin"},
		{"NINEFOLD_ADMIN_TRUSTED_PROXIES", "127.0.0.0/8,::1/128"},
		{"NINEFOLD_LOG_LEVEL", "info"},
		{"NINEFOLD_REPLAY_RETENTION", "168h"},
		{"NINEFOLD_MATCH_TOMBSTONE_RETENTION", "720h"},
		{"NINEFOLD_COMMAND_RECEIPT_RETENTION", "24h"},
		{"NINEFOLD_SHUTDOWN_TIMEOUT", "60s"},
	}

	file, err := os.OpenFile(options.Output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("refusing to overwrite %s", options.Output)
		}
		return fmt.Errorf("create output: %w", err)
	}
	ok := false
	defer func() {
		_ = file.Close()
		if !ok {
			_ = os.Remove(options.Output)
		}
	}()
	for _, entry := range values {
		if _, err := fmt.Fprintf(file, "%s=%s\n", entry[0], entry[1]); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync output: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close output: %w", err)
	}
	ok = true
	return nil
}

func randomBytes(random io.Reader, size int) ([]byte, error) {
	value := make([]byte, size)
	if _, err := io.ReadFull(random, value); err != nil {
		return nil, err
	}
	return value, nil
}
