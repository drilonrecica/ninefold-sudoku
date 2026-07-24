package config

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	valid := validEnvironment(t)
	tests := []struct {
		name    string
		mutate  func(map[string]string)
		wantErr string
	}{
		{name: "development"},
		{name: "test", mutate: func(env map[string]string) { env["NINEFOLD_ENVIRONMENT"] = "test" }},
		{
			name: "production",
			mutate: func(env map[string]string) {
				env["NINEFOLD_ENVIRONMENT"] = "production"
				env["NINEFOLD_PUBLIC_URL"] = "https://ninefold.example"
				env["NINEFOLD_ALLOWED_ORIGINS"] = "https://ninefold.example"
				env["NINEFOLD_COOKIE_SECRET"] = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("s", 32)))
			},
		},
		{
			name: "missing",
			mutate: func(env map[string]string) {
				delete(env, "NINEFOLD_PUBLIC_URL")
			},
			wantErr: "NINEFOLD_PUBLIC_URL is required",
		},
		{
			name: "malformed URL",
			mutate: func(env map[string]string) {
				env["NINEFOLD_PUBLIC_URL"] = "localhost"
			},
			wantErr: "absolute HTTP(S) URL",
		},
		{
			name: "production requires HTTPS",
			mutate: func(env map[string]string) {
				env["NINEFOLD_ENVIRONMENT"] = "production"
			},
			wantErr: "must use HTTPS",
		},
		{
			name: "production rejects placeholder secret",
			mutate: func(env map[string]string) {
				env["NINEFOLD_ENVIRONMENT"] = "production"
				env["NINEFOLD_PUBLIC_URL"] = "https://ninefold.example"
				env["NINEFOLD_ALLOWED_ORIGINS"] = "https://ninefold.example"
				env["NINEFOLD_COOKIE_SECRET"] = base64.StdEncoding.EncodeToString(make([]byte, 32))
			},
			wantErr: "production placeholder",
		},
		{
			name: "short cookie secret",
			mutate: func(env map[string]string) {
				env["NINEFOLD_COOKIE_SECRET"] = base64.StdEncoding.EncodeToString([]byte("short"))
			},
			wantErr: "at least 32 bytes",
		},
		{
			name: "malformed signing key",
			mutate: func(env map[string]string) {
				env["NINEFOLD_REPLAY_SIGNING_KEY"] = base64.StdEncoding.EncodeToString([]byte("not-pkcs8"))
			},
			wantErr: "PKCS#8",
		},
		{
			name: "unsafe key ID",
			mutate: func(env map[string]string) {
				env["NINEFOLD_REPLAY_SIGNING_KEY_ID"] = "../key"
			},
			wantErr: "safe token",
		},
		{
			name: "retention too long",
			mutate: func(env map[string]string) {
				env["NINEFOLD_REPLAY_RETENTION"] = "169h"
			},
			wantErr: "no greater than 168h0m0s",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			env := clone(valid)
			if test.mutate != nil {
				test.mutate(env)
			}
			cfg, err := Parse(mapLookup(env))
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("Parse() error = %v", err)
				}
				if cfg.Sanitized()["configured"] != "true" {
					t.Fatal("sanitized configuration should report configured")
				}
				for _, value := range cfg.Sanitized() {
					if strings.Contains(value, env["NINEFOLD_COOKIE_SECRET"]) ||
						strings.Contains(value, env["NINEFOLD_REPLAY_SIGNING_KEY"]) {
						t.Fatal("sanitized configuration exposed a secret")
					}
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Parse() error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}

func validEnvironment(t *testing.T) map[string]string {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]string{
		"NINEFOLD_ENVIRONMENT":               "development",
		"NINEFOLD_PUBLIC_URL":                "http://localhost:5173",
		"NINEFOLD_HTTP_ADDRESS":              "127.0.0.1:0",
		"NINEFOLD_DATABASE_PATH":             t.TempDir() + "/ninefold.db",
		"NINEFOLD_ALLOWED_ORIGINS":           "http://localhost:5173",
		"NINEFOLD_COOKIE_SECRET":             base64.StdEncoding.EncodeToString([]byte(strings.Repeat("c", 32))),
		"NINEFOLD_REPLAY_SIGNING_KEY":        base64.StdEncoding.EncodeToString(der),
		"NINEFOLD_REPLAY_SIGNING_KEY_ID":     "test-1",
		"NINEFOLD_ADMIN_PROXY_HEADER":        "X-Ninefold-Admin",
		"NINEFOLD_LOG_LEVEL":                 "debug",
		"NINEFOLD_REPLAY_RETENTION":          "168h",
		"NINEFOLD_MATCH_TOMBSTONE_RETENTION": "720h",
		"NINEFOLD_COMMAND_RECEIPT_RETENTION": "24h",
		"NINEFOLD_SHUTDOWN_TIMEOUT":          "5s",
	}
}

func mapLookup(env map[string]string) LookupFunc {
	return func(name string) (string, bool) {
		value, ok := env[name]
		return value, ok
	}
}

func clone(source map[string]string) map[string]string {
	target := make(map[string]string, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}
