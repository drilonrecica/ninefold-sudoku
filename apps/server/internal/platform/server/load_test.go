package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
)

const (
	qualificationRooms       = 25
	qualificationConnections = 100
)

// TestQualificationLoad100Connections25Rooms exercises the documented
// single-instance target with real HTTP, cookies, SQLite writes, WebSocket
// upgrades, actor queues, and reconnects. It intentionally avoids gameplay
// payload generation; command correctness is covered by the focused suites.
func TestQualificationLoad100Connections25Rooms(t *testing.T) {
	if testing.Short() {
		t.Skip("qualification load test")
	}

	instance, _, listener := newTestServer(t)
	stopped := make(chan error, 1)
	go func() { stopped <- instance.Serve(listener) }()
	baseURL := "http://" + listener.Addr().String()

	clients := make([]*http.Client, 0, qualificationConnections)
	roomCodes := make([]string, 0, qualificationRooms)
	for roomIndex := 0; roomIndex < qualificationRooms; roomIndex++ {
		host := qualificationClient(t, roomIndex+2)
		code := createQualificationRoom(t, host, baseURL, roomIndex)
		clients = append(clients, host)
		roomCodes = append(roomCodes, code)

		for participant := 0; participant < 3; participant++ {
			source := 2 + qualificationRooms + roomIndex*3 + participant
			guest := qualificationClient(t, source)
			joinQualificationRoom(t, guest, baseURL, code, roomIndex, participant)
			clients = append(clients, guest)
		}
	}

	connect := func() []*websocket.Conn {
		t.Helper()
		connections := make([]*websocket.Conn, len(clients))
		var wg sync.WaitGroup
		errors := make(chan error, len(clients))
		for index := range clients {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				headers := http.Header{"Origin": []string{baseURL}}
				conn, response, err := websocket.Dial(ctx, "ws://"+listener.Addr().String()+"/ws", &websocket.DialOptions{
					HTTPClient: clients[index],
					HTTPHeader: headers,
				})
				if response != nil && response.Body != nil {
					_ = response.Body.Close()
				}
				if err != nil {
					errors <- fmt.Errorf("connection %d: %w", index, err)
					return
				}
				connections[index] = conn
			}(index)
		}
		wg.Wait()
		close(errors)
		for err := range errors {
			t.Error(err)
		}
		if t.Failed() {
			t.FailNow()
		}
		return connections
	}

	connections := connect()
	assertQualificationRuntime(t, instance, qualificationRooms, qualificationConnections)
	closeQualificationConnections(connections)

	// A full reconnect storm must restore the same bounded actor population.
	connections = connect()
	assertQualificationRuntime(t, instance, qualificationRooms, qualificationConnections)
	closeQualificationConnections(connections)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := instance.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}

	_ = roomCodes
}

func qualificationClient(t *testing.T, sourceSuffix int) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		LocalAddr: &net.TCPAddr{IP: net.ParseIP(fmt.Sprintf("127.0.0.%d", sourceSuffix))},
	}
	return &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext:         dialer.DialContext,
			DisableCompression:  true,
			MaxIdleConnsPerHost: 2,
		},
	}
}

func createQualificationRoom(t *testing.T, client *http.Client, baseURL string, roomIndex int) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"displayName": fmt.Sprintf("Host %d", roomIndex),
		"mode":        "Coop",
		"difficulty":  "Easy",
	})
	request := qualificationRequest(t, http.MethodPost, baseURL+"/api/v1/rooms", body)
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create room %d: status=%d", roomIndex, response.StatusCode)
	}
	var result struct {
		Room struct {
			Code string `json:"code"`
		} `json:"room"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result.Room.Code
}

func joinQualificationRoom(t *testing.T, client *http.Client, baseURL, code string, roomIndex, participant int) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"displayName": fmt.Sprintf("Guest %d-%d", roomIndex, participant),
	})
	request := qualificationRequest(t, http.MethodPost, baseURL+"/api/v1/rooms/"+url.PathEscape(code)+"/join", body)
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("join room %d participant %d: status=%d", roomIndex, participant, response.StatusCode)
	}
}

func qualificationRequest(t *testing.T, method, target string, body []byte) *http.Request {
	t.Helper()
	request, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	requestID, err := idgen.Generator{}.RequestID()
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Idempotency-Key", requestID.String())
	return request
}

func assertQualificationRuntime(t *testing.T, instance *Server, wantActors, wantConnections int) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		actors, connections, commandDepth, outboundDepth := instance.registry.OperationalStats()
		if actors == wantActors && connections == wantConnections {
			if commandDepth > wantActors || outboundDepth > wantConnections*2 {
				t.Fatalf("unbounded queues: commands=%d outbound=%d", commandDepth, outboundDepth)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("runtime actors=%d connections=%d, want %d/%d", actors, connections, wantActors, wantConnections)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func closeQualificationConnections(connections []*websocket.Conn) {
	var wg sync.WaitGroup
	for _, connection := range connections {
		if connection == nil {
			continue
		}
		wg.Add(1)
		go func(connection *websocket.Conn) {
			defer wg.Done()
			_ = connection.Close(websocket.StatusNormalClosure, "qualification cycle complete")
		}(connection)
	}
	wg.Wait()
}
