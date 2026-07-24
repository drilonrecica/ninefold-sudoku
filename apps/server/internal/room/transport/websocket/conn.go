package websocket

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"nhooyr.io/websocket"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	matchdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/match/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/actor"
	roomdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/domain"
	"github.com/drilonrecica/ninefold-sudoku/contracts/generated/go/realtime"
)

// connection represents one browser tab attached to a Room actor.
type connection struct {
	ws            *websocket.Conn
	actor         *actor.Actor
	registry      *actor.Registry
	roomID        shared.RoomID
	participantID shared.ParticipantID
	connID        shared.ConnectionID
	sessionHash   []byte
	sendCh        chan []byte
	logger        *slog.Logger
	rateWindows   map[string]rateWindow
}

type rateWindow struct {
	start time.Time
	count int
}

const outboundQueueCapacity = 128

const (
	heartbeatInterval = 20 * time.Second
	heartbeatTimeout  = 10 * time.Second
)

func newConnection(ws *websocket.Conn, roomActor *actor.Actor, registry *actor.Registry, roomID shared.RoomID, participantID shared.ParticipantID, connID shared.ConnectionID, sessionHash []byte, logger *slog.Logger) *connection {
	return &connection{
		ws:            ws,
		actor:         roomActor,
		registry:      registry,
		roomID:        roomID,
		participantID: participantID,
		connID:        connID,
		sessionHash:   append([]byte(nil), sessionHash...),
		sendCh:        make(chan []byte, outboundQueueCapacity),
		logger:        logger,
		rateWindows:   make(map[string]rateWindow),
	}
}

func (c *connection) run(ctx context.Context) {
	go c.writePump(ctx)

	if _, err := c.actor.Subscribe(ctx, c.connID, c.participantID, c.sendCh, func() { _ = c.ws.CloseNow() }); err != nil {
		c.logger.Warn("subscribe failed", "error", err, "connID", c.connID)
		c.send([]byte(serverMessage("connection.rejected", 0, 0, map[string]any{"code": "SERVER_BUSY"})))
		c.close(websocket.StatusInternalError, fmt.Sprintf("subscribe failed: %v", err))
		return
	}

	c.readPump(ctx)
}

func (c *connection) readPump(ctx context.Context) {
	defer c.unregister(ctx)
	for {
		_, data, err := c.ws.Read(ctx)
		if err != nil {
			status := websocket.CloseStatus(err)
			if status == websocket.StatusGoingAway || status == websocket.StatusAbnormalClosure || status == websocket.StatusNormalClosure {
				c.logger.Debug("websocket read ended", "connID", c.connID)
			}
			return
		}
		if len(data) > maxMessageBytes {
			c.sendRejection("", "RATE_LIMITED")
			continue
		}
		var msg realtime.ClientMessage
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&msg); err != nil {
			c.sendRejection("", "INVALID_VALUE")
			continue
		}
		if err := c.handleMessage(ctx, msg); err != nil {
			c.sendRejection(string(msg.RequestId), errorCode(err))
		}
	}
}

func (c *connection) writePump(ctx context.Context) {
	defer c.close(websocket.StatusNormalClosure, "writer done")
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			pingCtx, cancel := context.WithTimeout(ctx, heartbeatTimeout)
			err := c.ws.Ping(pingCtx)
			cancel()
			if err != nil {
				c.logger.Debug("websocket heartbeat failed", "error", err, "connID", c.connID)
				return
			}
		case msg, ok := <-c.sendCh:
			if !ok {
				return
			}
			if err := c.wsWrite(msg); err != nil {
				c.logger.Debug("websocket write failed", "error", err, "connID", c.connID)
				return
			}
		}
	}
}

func (c *connection) send(msg []byte) bool {
	select {
	case c.sendCh <- msg:
		return true
	default:
		c.close(websocket.StatusInternalError, "send queue full")
		return false
	}
}

func (c *connection) wsWrite(msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.ws.Write(ctx, websocket.MessageText, msg)
}

func (c *connection) unregister(ctx context.Context) {
	_ = c.actor.Unsubscribe(ctx, c.connID)
	c.registry.Release(c.roomID)
}

func (c *connection) close(code websocket.StatusCode, reason string) {
	_ = c.ws.Close(code, reason)
}

func (c *connection) handleMessage(ctx context.Context, msg realtime.ClientMessage) error {
	if version, ok := msg.SchemaVersion.(float64); !ok || version != 1 {
		return shared.DomainError{Code: shared.ErrMatchCommandInvalid}
	}
	if !c.allowMessage(msg.Type, time.Now()) {
		return shared.DomainError{Code: shared.ErrRateLimited}
	}
	requestID, err := shared.ParseRequestID(string(msg.RequestId))
	if err != nil {
		return shared.DomainError{Code: shared.ErrInvalidValue}
	}
	seq, err := shared.NewClientSequence(uint64(msg.ClientSequence))
	if err != nil {
		return shared.DomainError{Code: shared.ErrInvalidValue}
	}

	meta := shared.CommandMetadata{
		RequestID:                  requestID,
		AuthenticatedParticipantID: c.participantID,
		ClientSequence:             seq,
		Target:                     aggregateTarget(msg.Target),
	}
	if msg.Target != nil {
		meta.ExpectedVersion = uint64(msg.Target.ExpectedVersion)
	}

	switch msg.Type {
	case realtime.ClientMessageTypeConnectionInitialize:
		// Explicit resync after reconnect or missed events.
		return c.actor.Sync(ctx, c.connID, c.participantID, c.sendCh)
	case realtime.ClientMessageTypeConnectionHeartbeat:
		return nil
	case realtime.ClientMessageTypeConnectionRequestControl:
		return c.submit(ctx, requestID, "connection.request_control", actor.ControlCommand{Meta: meta})
	case realtime.ClientMessageTypeCommandStatus:
		return c.submit(ctx, requestID, "command.status", actor.ControlCommand{Meta: meta})
	case realtime.ClientMessageTypeRoomSetReady:
		cmd, err := c.buildSetReady(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "room.set_ready", cmd)
	case realtime.ClientMessageTypeRoomChangeSettings:
		cmd, err := c.buildChangeSettings(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "room.change_settings", cmd)
	case realtime.ClientMessageTypeRoomStartCountdown:
		cmd, err := c.buildStartCountdown(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "room.start_countdown", cmd)
	case realtime.ClientMessageTypeRoomCancelCountdown:
		cmd, err := c.buildCancelCountdown(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "room.cancel_countdown", cmd)
	case realtime.ClientMessageTypeRoomLeave:
		cmd, err := c.buildLeave(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "room.leave", cmd)
	case realtime.ClientMessageTypeRoomTransferHost:
		cmd, err := c.buildTransferHost(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "room.transfer_host", cmd)
	case realtime.ClientMessageTypeMatchPlaceValue:
		cmd, err := c.buildPlaceValue(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "match.place_value", cmd)
	case realtime.ClientMessageTypeMatchEraseValue:
		cmd, err := c.buildEraseValue(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "match.erase_value", cmd)
	case realtime.ClientMessageTypeMatchAddNote:
		cmd, err := c.buildAddNote(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "match.add_note", cmd)
	case realtime.ClientMessageTypeMatchRemoveNote:
		cmd, err := c.buildRemoveNote(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "match.remove_note", cmd)
	case realtime.ClientMessageTypeMatchUseHint:
		cmd, err := c.buildUseHint(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "match.use_hint", cmd)
	case realtime.ClientMessageTypeMatchPing:
		cmd, err := c.buildPing(meta, msg.Payload)
		if err != nil {
			return err
		}
		return c.submit(ctx, requestID, "match.ping", cmd)
	case realtime.ClientMessageTypeMatchReaction:
		return c.submitReaction(msg.Payload)
	case realtime.ClientMessageTypeMatchFocusCell:
		return c.publishFocus(msg.Payload, true)
	case realtime.ClientMessageTypeMatchReleaseFocus:
		return c.publishFocus(msg.Payload, false)
	default:
		return shared.DomainError{Code: shared.ErrMatchCommandInvalid}
	}
}

func (c *connection) allowMessage(messageType realtime.ClientMessageType, now time.Time) bool {
	var category string
	var limit int
	var duration time.Duration
	switch messageType {
	case realtime.ClientMessageTypeMatchPlaceValue, realtime.ClientMessageTypeMatchEraseValue:
		category, limit, duration = "value", 10, time.Second
	case realtime.ClientMessageTypeMatchAddNote, realtime.ClientMessageTypeMatchRemoveNote:
		category, limit, duration = "note", 30, time.Second
	case realtime.ClientMessageTypeMatchPing, realtime.ClientMessageTypeMatchReaction:
		category, limit, duration = "social", 5, 10*time.Second
	case realtime.ClientMessageTypeMatchFocusCell, realtime.ClientMessageTypeMatchReleaseFocus:
		category, limit, duration = "focus", 10, time.Second
	case realtime.ClientMessageTypeRoomChangeSettings:
		category, limit, duration = "settings", 5, 10*time.Second
	default:
		return true
	}
	window := c.rateWindows[category]
	if window.start.IsZero() || now.Sub(window.start) >= duration {
		c.rateWindows[category] = rateWindow{start: now, count: 1}
		return true
	}
	if window.count >= limit {
		return false
	}
	window.count++
	c.rateWindows[category] = window
	return true
}

func aggregateTarget(t *realtime.ClientMessageTarget) shared.AggregateTarget {
	if t == nil {
		return shared.AggregateTarget{}
	}
	switch t.Kind {
	case realtime.ClientMessageTargetKindRoom:
		return shared.NewRoomTarget(shared.RoomID(string(t.Id)))
	case realtime.ClientMessageTargetKindMatch:
		return shared.NewMatchTarget(shared.MatchID(string(t.Id)))
	}
	return shared.AggregateTarget{}
}

func (c *connection) submit(ctx context.Context, requestID shared.RequestID, commandType string, cmd any) error {
	canonicalCommand, err := json.Marshal(cmd)
	if err != nil {
		return shared.DomainError{Code: shared.ErrInvalidValue}
	}
	fingerprint := fmt.Sprintf("%x", sha256.Sum256(canonicalCommand))
	result, err := c.actor.Submit(ctx, actor.Envelope{
		RequestID:    requestID,
		CommandType:  commandType,
		ScopeHash:    c.sessionHash,
		Fingerprint:  fingerprint,
		Command:      cmd,
		ConnectionID: c.connID,
	})
	if err != nil {
		return err
	}
	if commandType == "command.status" {
		var payload map[string]any
		if err := json.Unmarshal(result.Response, &payload); err != nil {
			return shared.DomainError{Code: shared.ErrPersistenceFailed}
		}
		c.send(serverMessage("command.status", 0, 0, payload))
		return nil
	}
	c.sendAck(requestID, result.Response)
	for _, msg := range result.OriginBroadcasts {
		if !c.send(msg) {
			break
		}
	}
	return nil
}

func (c *connection) submitReaction(payload realtime.ClientMessagePayload) error {
	if payload.Reaction == nil {
		return shared.DomainError{Code: shared.ErrInvalidValue}
	}
	msg := serverMessage("ephemeral.reaction", 0, 0, map[string]any{
		"participantId": c.participantID.String(),
		"reaction":      string(*payload.Reaction),
	})
	c.actor.PublishEphemeral(msg)
	return nil
}

func (c *connection) publishFocus(payload realtime.ClientMessagePayload, focused bool) error {
	if payload.Cell == nil {
		return shared.DomainError{Code: shared.ErrInvalidValue}
	}
	if _, err := shared.NewCellIndex(int(*payload.Cell)); err != nil {
		return err
	}
	c.actor.PublishEphemeral(serverMessage("ephemeral.focus", 0, 0, map[string]any{
		"participantId": c.participantID.String(),
		"cell":          *payload.Cell,
		"focused":       focused,
	}))
	return nil
}

func (c *connection) sendAck(requestID shared.RequestID, response []byte) {
	var details map[string]any
	_ = json.Unmarshal(response, &details)
	payload := map[string]any{
		"requestId": requestID.String(),
		"accepted":  true,
		"details":   details,
	}
	if room, ok := details["room"].(map[string]any); ok {
		if version, ok := room["version"]; ok {
			payload["roomVersion"] = version
			payload["resultingVersion"] = version
			payload["aggregate"] = "room"
		}
	}
	if match, ok := details["match"].(map[string]any); ok {
		if version, ok := match["version"]; ok {
			payload["matchVersion"] = version
			payload["resultingVersion"] = version
			payload["aggregate"] = "match"
		}
	}
	msg := serverMessage("command.acknowledged", 0, 0, payload)
	c.send(msg)
}

func (c *connection) sendRejection(requestID string, code string) {
	msg := serverMessage("command.rejected", 0, 0, map[string]any{
		"requestId": requestID,
		"accepted":  false,
		"code":      code,
	})
	c.send(msg)
}

func errorCode(err error) string {
	if de, ok := err.(shared.DomainError); ok {
		return string(de.Code)
	}
	return string(shared.ErrServerBusy)
}

// --- command builders ---

func (c *connection) buildSetReady(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (roomdomain.SetReadyCommand, error) {
	if payload.Ready == nil {
		return roomdomain.SetReadyCommand{}, shared.DomainError{Code: shared.ErrInvalidValue}
	}
	return roomdomain.SetReadyCommand{Meta: meta, Ready: *payload.Ready}, nil
}

func (c *connection) buildChangeSettings(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (roomdomain.ChangeRoomSettingsCommand, error) {
	if payload.Settings == nil {
		return roomdomain.ChangeRoomSettingsCommand{}, shared.DomainError{Code: shared.ErrInvalidValue}
	}
	patch := roomdomain.RoomSettingsPatch{}
	if payload.Settings.Difficulty != nil {
		d, err := shared.ParseDifficulty(*payload.Settings.Difficulty)
		if err != nil {
			return roomdomain.ChangeRoomSettingsCommand{}, err
		}
		patch.Difficulty = &d
	}
	if payload.Settings.ErrorPreset != nil {
		ep, err := shared.ParseErrorPreset(*payload.Settings.ErrorPreset)
		if err != nil {
			return roomdomain.ChangeRoomSettingsCommand{}, err
		}
		patch.ErrorPreset = &ep
	}
	if payload.Settings.HintsEnabled != nil {
		patch.HintsEnabled = payload.Settings.HintsEnabled
	}
	if payload.Settings.AutoRemoveNotes != nil {
		patch.AutoRemoveNotes = payload.Settings.AutoRemoveNotes
	}
	return roomdomain.ChangeRoomSettingsCommand{Meta: meta, Settings: patch}, nil
}

func (c *connection) buildStartCountdown(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (roomdomain.StartCountdownCommand, error) {
	return roomdomain.StartCountdownCommand{Meta: meta}, nil
}

func (c *connection) buildCancelCountdown(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (roomdomain.CancelCountdownCommand, error) {
	return roomdomain.CancelCountdownCommand{Meta: meta}, nil
}

func (c *connection) buildLeave(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (roomdomain.LeaveRoomCommand, error) {
	return roomdomain.LeaveRoomCommand{Meta: meta, Intent: ""}, nil
}

func (c *connection) buildTransferHost(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (roomdomain.TransferHostCommand, error) {
	if payload.ParticipantId == nil {
		return roomdomain.TransferHostCommand{}, shared.DomainError{Code: shared.ErrInvalidValue}
	}
	pid, err := shared.ParseParticipantID(string(*payload.ParticipantId))
	if err != nil {
		return roomdomain.TransferHostCommand{}, err
	}
	return roomdomain.TransferHostCommand{Meta: meta, Participant: pid}, nil
}

func (c *connection) buildPlaceValue(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (matchdomain.PlaceValueCommand, error) {
	if payload.Cell == nil || payload.Value == nil {
		return matchdomain.PlaceValueCommand{}, shared.DomainError{Code: shared.ErrInvalidValue}
	}
	cell, err := shared.NewCellIndex(int(*payload.Cell))
	if err != nil {
		return matchdomain.PlaceValueCommand{}, err
	}
	value, err := shared.NewDigit(int(*payload.Value))
	if err != nil {
		return matchdomain.PlaceValueCommand{}, err
	}
	return matchdomain.PlaceValueCommand{Meta: meta, Cell: cell, Digit: value}, nil
}

func (c *connection) buildEraseValue(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (matchdomain.EraseValueCommand, error) {
	if payload.Cell == nil {
		return matchdomain.EraseValueCommand{}, shared.DomainError{Code: shared.ErrInvalidValue}
	}
	cell, err := shared.NewCellIndex(int(*payload.Cell))
	if err != nil {
		return matchdomain.EraseValueCommand{}, err
	}
	return matchdomain.EraseValueCommand{Meta: meta, Cell: cell}, nil
}

func (c *connection) buildAddNote(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (matchdomain.AddNoteCommand, error) {
	if payload.Cell == nil || len(payload.Digits) == 0 {
		return matchdomain.AddNoteCommand{}, shared.DomainError{Code: shared.ErrInvalidValue}
	}
	cell, err := shared.NewCellIndex(int(*payload.Cell))
	if err != nil {
		return matchdomain.AddNoteCommand{}, err
	}
	digits := make([]shared.Digit, len(payload.Digits))
	for i, d := range payload.Digits {
		dg, err := shared.NewDigit(int(d))
		if err != nil {
			return matchdomain.AddNoteCommand{}, err
		}
		digits[i] = dg
	}
	return matchdomain.AddNoteCommand{Meta: meta, Cell: cell, Digits: digits}, nil
}

func (c *connection) buildRemoveNote(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (matchdomain.RemoveNoteCommand, error) {
	if payload.Cell == nil || len(payload.Digits) == 0 {
		return matchdomain.RemoveNoteCommand{}, shared.DomainError{Code: shared.ErrInvalidValue}
	}
	cell, err := shared.NewCellIndex(int(*payload.Cell))
	if err != nil {
		return matchdomain.RemoveNoteCommand{}, err
	}
	digits := make([]shared.Digit, len(payload.Digits))
	for i, d := range payload.Digits {
		dg, err := shared.NewDigit(int(d))
		if err != nil {
			return matchdomain.RemoveNoteCommand{}, err
		}
		digits[i] = dg
	}
	return matchdomain.RemoveNoteCommand{Meta: meta, Cell: cell, Digits: digits}, nil
}

func (c *connection) buildUseHint(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (matchdomain.UseHintCommand, error) {
	if payload.Level == nil {
		return matchdomain.UseHintCommand{}, shared.DomainError{Code: shared.ErrInvalidValue}
	}
	level, err := shared.ParseHintLevel(string(*payload.Level))
	if err != nil {
		return matchdomain.UseHintCommand{}, err
	}
	var target *shared.CellIndex
	if payload.TargetCell != nil {
		t, err := shared.NewCellIndex(int(*payload.TargetCell))
		if err != nil {
			return matchdomain.UseHintCommand{}, err
		}
		target = &t
	}
	return matchdomain.UseHintCommand{Meta: meta, Level: level, Target: target}, nil
}

func (c *connection) buildPing(meta shared.CommandMetadata, payload realtime.ClientMessagePayload) (matchdomain.PingCommand, error) {
	if payload.Cell == nil || payload.Intent == nil || *payload.Intent == "" {
		return matchdomain.PingCommand{}, shared.DomainError{Code: shared.ErrInvalidValue}
	}
	cell, err := shared.NewCellIndex(int(*payload.Cell))
	if err != nil {
		return matchdomain.PingCommand{}, err
	}
	return matchdomain.PingCommand{Meta: meta, Cell: cell, Intent: *payload.Intent}, nil
}
