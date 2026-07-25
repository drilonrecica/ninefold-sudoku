package ops

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestAdminOnlyRejectsSpoofedUntrustedRequest(t *testing.T) {
	t.Parallel()
	handler := AdminOnly("X-Ninefold-Admin", []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if AdminIdentity(r.Context()) != "operator" {
				t.Fatal("missing identity")
			}
			w.WriteHeader(http.StatusNoContent)
		}))

	for _, test := range []struct {
		name       string
		remoteAddr string
		header     string
		want       int
	}{
		{name: "missing identity", remoteAddr: "10.0.0.2:1234", want: http.StatusNotFound},
		{name: "spoofed identity", remoteAddr: "203.0.113.4:1234", header: "operator", want: http.StatusNotFound},
		{name: "trusted proxy", remoteAddr: "10.0.0.2:1234", header: "operator", want: http.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/internal/status", nil)
			request.RemoteAddr = test.remoteAddr
			request.Header.Set("X-Ninefold-Admin", test.header)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status=%d want=%d", response.Code, test.want)
			}
		})
	}
}

func TestLogsAndMetricsUseRouteTemplates(t *testing.T) {
	t.Parallel()
	const secret = "7KMP4R"
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	metrics := NewMetrics()
	metrics.SetRuntimeProvider(func(context.Context) RuntimeMetrics {
		return RuntimeMetrics{ActiveWebSockets: 2, ActiveRoomActors: 1}
	})
	router := chi.NewRouter()
	router.Use(metrics.Middleware)
	router.Use(func(next http.Handler) http.Handler { return RequestLog(logger, next) })
	router.Get("/api/v1/rooms/{code}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/"+secret, nil)
	router.ServeHTTP(httptest.NewRecorder(), request)

	metricsResponse := httptest.NewRecorder()
	metrics.ServeHTTP(metricsResponse, request)
	metricsBody, _ := io.ReadAll(metricsResponse.Result().Body)
	if strings.Contains(logs.String(), secret) || strings.Contains(string(metricsBody), secret) {
		t.Fatal("log or metric exposed a route parameter")
	}
	if !strings.Contains(string(metricsBody), "ninefold_active_websocket_connections 2") {
		t.Fatal("runtime gauges missing")
	}
}

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/SECRET", nil)
	SecurityHeaders(true, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(response, request)
	for _, header := range []string{
		"Content-Security-Policy", "Strict-Transport-Security", "Referrer-Policy",
		"X-Content-Type-Options", "Permissions-Policy", "X-Frame-Options",
	} {
		if response.Header().Get(header) == "" {
			t.Errorf("missing %s", header)
		}
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("private response must not be cached")
	}
}
