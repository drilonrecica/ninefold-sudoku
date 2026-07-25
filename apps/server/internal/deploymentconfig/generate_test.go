package deploymentconfig

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteProducesMatchingLockedProductionConfig(t *testing.T) {
	output := filepath.Join(t.TempDir(), ".env.production")
	if err := Write(Options{
		PublicURL: "https://ninefold.recica.dev",
		Version:   "0.3.1",
		Output:    output,
		Random:    strings.NewReader(strings.Repeat("r", 256)),
	}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode=%o want=600", info.Mode().Perm())
	}
	env := readEnvironment(t, output)
	privateDER, err := base64.StdEncoding.DecodeString(env["NINEFOLD_REPLAY_SIGNING_KEY"])
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := x509.ParsePKCS8PrivateKey(privateDER)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		t.Fatal("generated key is not Ed25519")
	}
	publicKey, err := base64.StdEncoding.DecodeString(env["NINEFOLD_REPLAY_PUBLIC_KEY"])
	if err != nil {
		t.Fatal(err)
	}
	if !privateKey.Public().(ed25519.PublicKey).Equal(ed25519.PublicKey(publicKey)) {
		t.Fatal("public and private replay keys do not match")
	}
	for _, name := range []string{"NINEFOLD_COOKIE_SECRET", "NINEFOLD_PROXY_SECRET"} {
		value, err := base64.StdEncoding.DecodeString(env[name])
		if err != nil || len(value) != 32 {
			t.Fatalf("%s is not a 32-byte base64 secret", name)
		}
	}
	if env["NINEFOLD_DOMAIN"] != "ninefold.recica.dev" || env["NINEFOLD_VERSION"] != "0.3.1" {
		t.Fatal("public deployment metadata missing")
	}
	if err := Write(Options{PublicURL: "https://example.com", Version: "0.3.1", Output: output}); err == nil ||
		!strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("second Write() error=%v", err)
	}
}

func TestWriteRejectsInvalidInputsWithoutCreatingFile(t *testing.T) {
	for _, test := range []struct {
		name      string
		publicURL string
		version   string
	}{
		{name: "HTTP URL", publicURL: "http://example.com", version: "0.3.1"},
		{name: "URL path", publicURL: "https://example.com/path", version: "0.3.1"},
		{name: "version prefix", publicURL: "https://example.com", version: "v0.3.1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), ".env.production")
			if err := Write(Options{PublicURL: test.publicURL, Version: test.version, Output: output}); err == nil {
				t.Fatal("expected validation error")
			}
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				t.Fatalf("unexpected output file: %v", err)
			}
		})
	}
}

func readEnvironment(t *testing.T, path string) map[string]string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	values := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		name, value, ok := strings.Cut(line, "=")
		if !ok {
			t.Fatalf("malformed line %q", line)
		}
		values[name] = value
	}
	return values
}
