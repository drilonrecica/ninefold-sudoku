package ops

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
)

type Metrics struct {
	mu                       sync.Mutex
	httpTotals               map[string]uint64
	runtime                  func(context.Context) RuntimeMetrics
	commandCount             uint64
	commandLatencySeconds    float64
	commandRejections        map[string]uint64
	recoverySuccess          uint64
	recoveryFailure          uint64
	sqliteTransactionCount   uint64
	sqliteTransactionSeconds float64
	sqliteBusyEvents         uint64
	puzzleAssignmentFailures uint64
	reconnects               uint64
}

func NewMetrics() *Metrics {
	return &Metrics{
		httpTotals:        make(map[string]uint64),
		commandRejections: make(map[string]uint64),
	}
}

type RuntimeMetrics struct {
	ActiveWebSockets   int
	ActiveRoomActors   int
	ActiveMatches      int
	ActorQueueDepth    int
	OutboundQueueDepth int
}

func (m *Metrics) SetRuntimeProvider(provider func(context.Context) RuntimeMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runtime = provider
}

func (m *Metrics) ObserveCommand(duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commandCount++
	m.commandLatencySeconds += duration.Seconds()
	if err == nil {
		return
	}
	code := "INTERNAL"
	var domainError shared.DomainError
	if errors.As(err, &domainError) {
		code = string(domainError.Code)
	}
	m.commandRejections[code]++
}

func (m *Metrics) ObserveRecovery(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if success {
		m.recoverySuccess++
	} else {
		m.recoveryFailure++
	}
}

func (m *Metrics) ObserveSQLiteTransaction(duration time.Duration, busy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sqliteTransactionCount++
	m.sqliteTransactionSeconds += duration.Seconds()
	if busy {
		m.sqliteBusyEvents++
	}
}

func (m *Metrics) ObservePuzzleAssignment(success bool) {
	if success {
		return
	}
	m.mu.Lock()
	m.puzzleAssignmentFailures++
	m.mu.Unlock()
}

func (m *Metrics) ObserveReconnect() {
	m.mu.Lock()
	m.reconnects++
	m.mu.Unlock()
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unmatched"
		}
		key := fmt.Sprintf("%s|%s|%d", r.Method, route, recorder.status)
		m.mu.Lock()
		m.httpTotals[key]++
		m.mu.Unlock()
	})
}

func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Header().Set("Cache-Control", "no-store")
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	fmt.Fprintf(w, "# TYPE ninefold_process_memory_bytes gauge\nninefold_process_memory_bytes %d\n", memory.Alloc)
	fmt.Fprintf(w, "# TYPE ninefold_goroutines gauge\nninefold_goroutines %d\n", runtime.NumGoroutine())
	m.mu.Lock()
	provider := m.runtime
	totals := make(map[string]uint64, len(m.httpTotals))
	for key, count := range m.httpTotals {
		totals[key] = count
	}
	commandCount := m.commandCount
	commandLatency := m.commandLatencySeconds
	rejections := make(map[string]uint64, len(m.commandRejections))
	for code, count := range m.commandRejections {
		rejections[code] = count
	}
	recoverySuccess, recoveryFailure := m.recoverySuccess, m.recoveryFailure
	sqliteCount, sqliteSeconds := m.sqliteTransactionCount, m.sqliteTransactionSeconds
	sqliteBusy, puzzleFailures := m.sqliteBusyEvents, m.puzzleAssignmentFailures
	reconnects := m.reconnects
	m.mu.Unlock()
	if provider != nil {
		current := provider(r.Context())
		fmt.Fprintf(w, "# TYPE ninefold_active_websocket_connections gauge\nninefold_active_websocket_connections %d\n", current.ActiveWebSockets)
		fmt.Fprintf(w, "# TYPE ninefold_active_room_actors gauge\nninefold_active_room_actors %d\n", current.ActiveRoomActors)
		fmt.Fprintf(w, "# TYPE ninefold_active_matches gauge\nninefold_active_matches %d\n", current.ActiveMatches)
		fmt.Fprintf(w, "# TYPE ninefold_actor_queue_depth gauge\nninefold_actor_queue_depth %d\n", current.ActorQueueDepth)
		fmt.Fprintf(w, "# TYPE ninefold_connection_outbound_queue_depth gauge\nninefold_connection_outbound_queue_depth %d\n", current.OutboundQueueDepth)
	}
	fmt.Fprintln(w, "# TYPE ninefold_command_latency_seconds summary")
	fmt.Fprintf(w, "ninefold_command_latency_seconds_sum %g\n", commandLatency)
	fmt.Fprintf(w, "ninefold_command_latency_seconds_count %d\n", commandCount)
	fmt.Fprintln(w, "# TYPE ninefold_command_rejections_total counter")
	for code, count := range rejections {
		fmt.Fprintf(w, "ninefold_command_rejections_total{code=%q} %d\n", code, count)
	}
	fmt.Fprintln(w, "# TYPE ninefold_recovery_total counter")
	fmt.Fprintf(w, "ninefold_recovery_total{result=\"success\"} %d\n", recoverySuccess)
	fmt.Fprintf(w, "ninefold_recovery_total{result=\"failure\"} %d\n", recoveryFailure)
	fmt.Fprintln(w, "# TYPE ninefold_sqlite_transaction_duration_seconds summary")
	fmt.Fprintf(w, "ninefold_sqlite_transaction_duration_seconds_sum %g\n", sqliteSeconds)
	fmt.Fprintf(w, "ninefold_sqlite_transaction_duration_seconds_count %d\n", sqliteCount)
	fmt.Fprintln(w, "# TYPE ninefold_sqlite_busy_events_total counter")
	fmt.Fprintf(w, "ninefold_sqlite_busy_events_total %d\n", sqliteBusy)
	fmt.Fprintln(w, "# TYPE ninefold_puzzle_assignment_failures_total counter")
	fmt.Fprintf(w, "ninefold_puzzle_assignment_failures_total %d\n", puzzleFailures)
	fmt.Fprintln(w, "# TYPE ninefold_reconnects_total counter")
	fmt.Fprintf(w, "ninefold_reconnects_total %d\n", reconnects)
	fmt.Fprintln(w, "# TYPE ninefold_http_requests_total counter")
	for key, count := range totals {
		parts := strings.Split(key, "|")
		if len(parts) != 3 {
			continue
		}
		fmt.Fprintf(w, "ninefold_http_requests_total{method=%q,route=%q,status=%q} %d\n",
			parts[0], parts[1], parts[2], count)
	}
}
