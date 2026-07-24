package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"
)

const CookieName = "ninefold_room_session"

const tokenBytes = 32

// Token is an opaque room session credential. The browser holds the token in
// an HTTP-only cookie; the server stores only its hash.
type Token struct {
	Value string
	Hash  []byte
}

// Generate creates a new opaque session token and its SHA-256 hash.
func Generate() (Token, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return Token{}, err
	}
	value := base64.RawURLEncoding.EncodeToString(b)
	return Token{Value: value, Hash: Hash(value)}, nil
}

// Hash returns the SHA-256 hash of a token value.
func Hash(value string) []byte {
	h := sha256.Sum256([]byte(value))
	return h[:]
}

// Cookie builds the canonical ninefold_room_session cookie.
func Cookie(value string, secure bool, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/api/v1",
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// ClearCookie returns a cookie that deletes the session in the browser.
func ClearCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/api/v1",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// Read extracts the raw token value from a request cookie.
func Read(r *http.Request) string {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
