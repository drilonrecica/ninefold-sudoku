package ops

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type adminIdentityKey struct{}
type clientAddressKey struct{}

const (
	ProxySecretHeader   = "X-Ninefold-Proxy-Secret"
	ClientAddressHeader = "X-Ninefold-Client-IP"
)

func AdminIdentity(ctx context.Context) string {
	identity, _ := ctx.Value(adminIdentityKey{}).(string)
	return identity
}

// ClientAddress returns the gateway-authenticated client IP, or the socket peer
// when the request did not carry a valid gateway identity.
func ClientAddress(r *http.Request) string {
	if address, ok := r.Context().Value(clientAddressKey{}).(string); ok {
		return address
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// ResolveClientAddress accepts the normalized address only from a gateway that
// proves knowledge of the deployment proxy secret.
func ResolveClientAddress(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		presented, err := base64.StdEncoding.DecodeString(r.Header.Get(ProxySecretHeader))
		address, addressErr := netip.ParseAddr(strings.TrimSpace(r.Header.Get(ClientAddressHeader)))
		if err == nil && addressErr == nil && len(presented) == len(secret) &&
			subtle.ConstantTimeCompare(presented, secret) == 1 {
			ctx := context.WithValue(r.Context(), clientAddressKey{}, address.Unmap().String())
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func AdminOnly(header string, trusted []netip.Prefix, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		address, parseErr := netip.ParseAddr(host)
		identity := strings.TrimSpace(r.Header.Get(header))
		allowed := err == nil && parseErr == nil && identity != "" && len(identity) <= 64
		if allowed {
			allowed = false
			for _, prefix := range trusted {
				if prefix.Contains(address.Unmap()) {
					allowed = true
					break
				}
			}
		}
		if !allowed {
			http.Error(w, `{"error":{"code":"ADMIN_ACCESS_DENIED"}}`, http.StatusNotFound)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), adminIdentityKey{}, identity)))
	})
}

func SecurityHeaders(production bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set("Content-Security-Policy", "default-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'; object-src 'none'")
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("X-Frame-Options", "DENY")
		headers.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws") ||
			strings.HasPrefix(r.URL.Path, "/internal/") {
			headers.Set("Referrer-Policy", "no-referrer")
			headers.Set("Cache-Control", "no-store")
		} else {
			headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		}
		if production {
			headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

func BodyLimit(maxBytes int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxBytes {
			http.Error(w, `{"error":{"code":"REQUEST_TOO_LARGE"}}`, http.StatusRequestEntityTooLarge)
			return
		}
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(w.ResponseWriter).Hijack()
}

func (w *statusRecorder) Flush() {
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func RequestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			var value [12]byte
			if _, err := rand.Read(value[:]); err == nil {
				requestID = hex.EncodeToString(value[:])
			}
		}
		w.Header().Set("X-Request-ID", requestID)
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unmatched"
		}
		logger.Info("http request", "requestID", requestID, "method", r.Method, "route", route,
			"status", recorder.status, "latencyMs", time.Since(started).Milliseconds())
	})
}
