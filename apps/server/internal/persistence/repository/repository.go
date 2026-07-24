package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
)

// Repository exposes aggregate-oriented persistence on top of SQLC generated queries.
// All mutations are performed inside a transaction supplied by the caller; the repository
// never starts its own transaction, so application code can bundle room, match, events,
// snapshots, results, and command receipts in one atomic SQLite commit.
type Repository struct {
	db *sqlite.DB
	q  *gen.Queries
}

// New builds a repository that uses the writer pool for explicit transactions and the
// reader pool for ordinary reads.
func New(db *sqlite.DB) *Repository {
	return &Repository{db: db, q: gen.New(db.Readers())}
}

// WithTx returns a repository bound to the supplied transaction. Use this for writes.
func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{db: r.db, q: r.q.WithTx(tx)}
}

// BeginTx starts a new transaction on the single writer pool.
func (r *Repository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, *Repository, error) {
	tx, err := r.db.Writer().BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	return tx, r.WithTx(tx), nil
}

// DB exposes the underlying database wrapper for health checks and direct use.
func (r *Repository) DB() *sqlite.DB { return r.db }

// --- Puzzles ---

// CreatePuzzle inserts a fully validated puzzle record.
func (r *Repository) CreatePuzzle(ctx context.Context, p gen.Puzzle) error {
	return r.q.CreatePuzzle(ctx, gen.CreatePuzzleParams{
		ID:                   p.ID,
		Revision:             p.Revision,
		State:                p.State,
		Difficulty:           p.Difficulty,
		HardestTechnique:     p.HardestTechnique,
		QualityScore:         p.QualityScore,
		MultiplayerApproved:  p.MultiplayerApproved,
		GeneratorVersion:     p.GeneratorVersion,
		SolverVersion:        p.SolverVersion,
		CanonicalFingerprint: p.CanonicalFingerprint,
		Clues:                p.Clues,
		Solution:             p.Solution,
		CreatedAtMs:          p.CreatedAtMs,
	})
}

// GetPuzzle returns a puzzle including the server-only solution. Use this only for
// internal assignment and validation; do not expose the solution over transport.
func (r *Repository) GetPuzzle(ctx context.Context, id string, revision int64) (gen.Puzzle, error) {
	return r.q.GetPuzzleByID(ctx, gen.GetPuzzleByIDParams{ID: id, Revision: revision})
}

type candidatePuzzle struct {
	id       string
	revision int64
}

// SelectPuzzleForAssignment picks a random active puzzle matching the difficulty and
// optional multiplayer approval, excluding the provided puzzle IDs.
func (r *Repository) SelectPuzzleForAssignment(ctx context.Context, difficulty string, excludeIDs []string, multiplayer bool) (gen.Puzzle, error) {
	exclude := make(map[string]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		exclude[id] = struct{}{}
	}

	var candidates []candidatePuzzle
	if multiplayer {
		rows, err := r.q.ListActivePuzzlesByDifficultyAndMultiplayer(ctx, gen.ListActivePuzzlesByDifficultyAndMultiplayerParams{
			Difficulty:          difficulty,
			MultiplayerApproved: 1,
		})
		if err != nil {
			return gen.Puzzle{}, err
		}
		for _, row := range rows {
			if _, ok := exclude[row.ID]; !ok {
				candidates = append(candidates, candidatePuzzle{id: row.ID, revision: row.Revision})
			}
		}
	} else {
		rows, err := r.q.ListActivePuzzlesByDifficulty(ctx, difficulty)
		if err != nil {
			return gen.Puzzle{}, err
		}
		for _, row := range rows {
			if _, ok := exclude[row.ID]; !ok {
				candidates = append(candidates, candidatePuzzle{id: row.ID, revision: row.Revision})
			}
		}
	}
	if len(candidates) == 0 {
		return gen.Puzzle{}, errors.New("no puzzle available")
	}
	selected := candidates[rand.IntN(len(candidates))]
	return r.q.GetPuzzleByID(ctx, gen.GetPuzzleByIDParams{ID: selected.id, Revision: selected.revision})
}

// --- Rooms ---

// CreateRoomTx creates a room, its participants, and its host session atomically.
func (r *Repository) CreateRoomTx(ctx context.Context, tx *sql.Tx, room gen.Room, participants []gen.RoomParticipant, session gen.RoomSession) error {
	q := r.q.WithTx(tx)
	if err := q.CreateRoom(ctx, roomToParams(room)); err != nil {
		return err
	}
	for _, p := range participants {
		if err := q.CreateRoomParticipant(ctx, participantToParams(p)); err != nil {
			return err
		}
	}
	if err := q.CreateRoomSession(ctx, roomSessionToParams(session)); err != nil {
		return err
	}
	return nil
}

// UpdateRoomTx updates the room projection and, if provided, its current session.
// expectedVersion is the previously committed aggregate version used for the optimistic guard.
// Participants are upserted individually; callers that mutate participant state should pass all
// current active participants.
func (r *Repository) UpdateRoomTx(ctx context.Context, tx *sql.Tx, room gen.Room, expectedVersion int64, participants []gen.RoomParticipant, session *gen.RoomSession) error {
	q := r.q.WithTx(tx)
	if err := q.UpdateRoom(ctx, gen.UpdateRoomParams{
		Code:              room.Code,
		State:             room.State,
		Version:           room.Version,
		Mode:              room.Mode,
		Difficulty:        room.Difficulty,
		ErrorPreset:       room.ErrorPreset,
		HintsEnabled:      room.HintsEnabled,
		SharedNotes:       room.SharedNotes,
		AutoRemoveNotes:   room.AutoRemoveNotes,
		SpectatorsAllowed: room.SpectatorsAllowed,
		HostParticipantID: room.HostParticipantID,
		CurrentMatchID:    room.CurrentMatchID,
		LastActivityAtMs:  room.LastActivityAtMs,
		ExpiresAtMs:       room.ExpiresAtMs,
		ID:                room.ID,
		Version_2:         expectedVersion,
	}); err != nil {
		return err
	}
	if affected, err := changes(ctx, tx); err != nil || affected == 0 {
		if err != nil {
			return err
		}
		return ErrConflict
	}
	for _, p := range participants {
		if err := q.UpsertRoomParticipant(ctx, upsertParticipantToParams(p)); err != nil {
			return err
		}
	}
	if session != nil {
		if err := q.UpdateRoomSession(ctx, gen.UpdateRoomSessionParams{
			ExpiresAtMs: session.ExpiresAtMs,
			RevokedAtMs: session.RevokedAtMs,
			TokenHash:   session.TokenHash,
		}); err != nil {
			return err
		}
	}
	return nil
}

// GetRoomByID loads the room projection.
func (r *Repository) GetRoomByID(ctx context.Context, id string) (gen.Room, error) {
	return r.q.GetRoomByID(ctx, id)
}

// GetRoomByCode loads a room by its short invitation code.
func (r *Repository) GetRoomByCode(ctx context.Context, code string) (gen.Room, error) {
	return r.q.GetRoomByCode(ctx, code)
}

// ListActiveRoomParticipants returns participants who have not left or been removed.
func (r *Repository) ListActiveRoomParticipants(ctx context.Context, roomID string) ([]gen.RoomParticipant, error) {
	return r.q.ListActiveRoomParticipants(ctx, roomID)
}

// GetRoomParticipantByID loads a single participant.
func (r *Repository) GetRoomParticipantByID(ctx context.Context, id string) (gen.RoomParticipant, error) {
	return r.q.GetRoomParticipantByID(ctx, id)
}

// --- Sessions ---

// GetRoomSessionByHash loads a session by the hashed bearer token.
func (r *Repository) GetRoomSessionByHash(ctx context.Context, hash []byte) (gen.RoomSession, error) {
	return r.q.GetRoomSessionByHash(ctx, hash)
}

// --- Matches ---

// CreateMatchTx creates a prepared/active match, its participants, and optional initial events/snapshot atomically.
func (r *Repository) CreateMatchTx(ctx context.Context, tx *sql.Tx, match gen.Match, participants []gen.MatchParticipant, events []gen.MatchEvent, snapshot *gen.MatchSnapshot) error {
	q := r.q.WithTx(tx)
	if err := q.CreateMatch(ctx, matchToParams(match)); err != nil {
		return err
	}
	for _, p := range participants {
		if err := q.CreateMatchParticipant(ctx, matchParticipantToParams(p)); err != nil {
			return err
		}
	}
	for _, e := range events {
		if err := q.CreateMatchEvent(ctx, matchEventToParams(e)); err != nil {
			return err
		}
	}
	if snapshot != nil {
		if err := q.CreateMatchSnapshot(ctx, matchSnapshotToParams(*snapshot)); err != nil {
			return err
		}
	}
	return nil
}

// UpdateMatchTx updates the match projection, participants, appends new events, and optionally creates a snapshot and result.
// expectedVersion is the previously committed aggregate version used for the optimistic guard.
func (r *Repository) UpdateMatchTx(ctx context.Context, tx *sql.Tx, match gen.Match, expectedVersion int64, participants []gen.MatchParticipant, events []gen.MatchEvent, snapshot *gen.MatchSnapshot, result *gen.MatchResult) error {
	q := r.q.WithTx(tx)
	if err := q.UpdateMatch(ctx, gen.UpdateMatchParams{
		RoomID:             match.RoomID,
		State:              match.State,
		Version:            match.Version,
		Mode:               match.Mode,
		Difficulty:         match.Difficulty,
		ErrorPreset:        match.ErrorPreset,
		HintsEnabled:       match.HintsEnabled,
		AutoRemoveNotes:    match.AutoRemoveNotes,
		RuleVersion:        match.RuleVersion,
		PuzzleID:           match.PuzzleID,
		PuzzleRevision:     match.PuzzleRevision,
		TransformationSeed: match.TransformationSeed,
		PuzzleDifficulty:   match.PuzzleDifficulty,
		GeneratorVersion:   match.GeneratorVersion,
		SolverVersion:      match.SolverVersion,
		StartedAtMs:        match.StartedAtMs,
		CompletedAtMs:      match.CompletedAtMs,
		ResultReason:       match.ResultReason,
		ElapsedMs:          match.ElapsedMs,
		Assisted:           match.Assisted,
		ID:                 match.ID,
		Version_2:          expectedVersion,
	}); err != nil {
		return err
	}
	if affected, err := changes(ctx, tx); err != nil || affected == 0 {
		if err != nil {
			return err
		}
		return ErrConflict
	}
	for _, p := range participants {
		if err := q.UpdateMatchParticipant(ctx, updateMatchParticipantToParams(p)); err != nil {
			return err
		}
	}
	for _, e := range events {
		if err := q.CreateMatchEvent(ctx, matchEventToParams(e)); err != nil {
			return err
		}
	}
	if snapshot != nil {
		if err := q.CreateMatchSnapshot(ctx, matchSnapshotToParams(*snapshot)); err != nil {
			return err
		}
	}
	if result != nil {
		if err := q.CreateMatchResult(ctx, gen.CreateMatchResultParams{
			MatchID:      result.MatchID,
			ResultReason: result.ResultReason,
			ElapsedMs:    result.ElapsedMs,
			Assisted:     result.Assisted,
			CreatedAtMs:  result.CreatedAtMs,
		}); err != nil {
			return err
		}
	}
	return nil
}

// GetMatchByID loads a match projection.
func (r *Repository) GetMatchByID(ctx context.Context, id string) (gen.Match, error) {
	return r.q.GetMatchByID(ctx, id)
}

// ListMatchParticipants returns the participants of a match.
func (r *Repository) ListMatchParticipants(ctx context.Context, matchID string) ([]gen.MatchParticipant, error) {
	return r.q.ListMatchParticipants(ctx, matchID)
}

// GetMatchEvents returns ordered durable events for a match.
func (r *Repository) GetMatchEvents(ctx context.Context, matchID string) ([]gen.MatchEvent, error) {
	return r.q.GetMatchEvents(ctx, matchID)
}

// GetLatestMatchSnapshot returns the most recent snapshot for reconstruction.
func (r *Repository) GetLatestMatchSnapshot(ctx context.Context, matchID string) (gen.MatchSnapshot, error) {
	return r.q.GetLatestMatchSnapshot(ctx, matchID)
}

// --- Command receipts ---

// CreateCommandReceipt stores a durable terminal outcome for idempotency.
func (r *Repository) CreateCommandReceipt(ctx context.Context, tx *sql.Tx, receipt gen.CommandReceipt) error {
	q := r.q.WithTx(tx)
	return q.CreateCommandReceipt(ctx, gen.CreateCommandReceiptParams{
		RequestID:              receipt.RequestID,
		AuthenticatedScopeHash: receipt.AuthenticatedScopeHash,
		CommandType:            receipt.CommandType,
		RequestFingerprint:     receipt.RequestFingerprint,
		TerminalStatus:         receipt.TerminalStatus,
		SafeResponseJson:       receipt.SafeResponseJson,
		CreatedAtMs:            receipt.CreatedAtMs,
		ExpiresAtMs:            receipt.ExpiresAtMs,
	})
}

// GetCommandReceipt retrieves a stored terminal outcome.
func (r *Repository) GetCommandReceipt(ctx context.Context, requestID string) (gen.CommandReceipt, error) {
	return r.q.GetCommandReceipt(ctx, requestID)
}

// RevokeRoomSession marks a session as revoked.
func (r *Repository) RevokeRoomSession(ctx context.Context, tx *sql.Tx, hash []byte, revokedAtMs int64) error {
	return r.q.WithTx(tx).UpdateRoomSession(ctx, gen.UpdateRoomSessionParams{
		TokenHash:   hash,
		RevokedAtMs: sql.NullInt64{Int64: revokedAtMs, Valid: true},
	})
}

// CreateRoomSessionTx inserts a new room session inside a transaction.
func (r *Repository) CreateRoomSessionTx(ctx context.Context, tx *sql.Tx, s gen.RoomSession) error {
	return r.q.WithTx(tx).CreateRoomSession(ctx, gen.CreateRoomSessionParams{
		TokenHash:     s.TokenHash,
		RoomID:        s.RoomID,
		ParticipantID: s.ParticipantID,
		CreatedAtMs:   s.CreatedAtMs,
		ExpiresAtMs:   s.ExpiresAtMs,
		RevokedAtMs:   s.RevokedAtMs,
	})
}

// --- Replay ---

// CreateReplayCapability stores a capability token hash.
func (r *Repository) CreateReplayCapability(ctx context.Context, tx *sql.Tx, rc gen.ReplayCapability) error {
	return r.q.WithTx(tx).CreateReplayCapability(ctx, gen.CreateReplayCapabilityParams{
		TokenHash:   rc.TokenHash,
		ReplayID:    rc.ReplayID,
		MatchID:     rc.MatchID,
		CreatedAtMs: rc.CreatedAtMs,
		ExpiresAtMs: rc.ExpiresAtMs,
		RevokedAtMs: rc.RevokedAtMs,
	})
}

// GetReplayCapabilityByHash loads a capability by its token hash.
func (r *Repository) GetReplayCapabilityByHash(ctx context.Context, hash []byte) (gen.ReplayCapability, error) {
	return r.q.GetReplayCapabilityByHash(ctx, hash)
}

// CreateReplaySeal stores a terminal replay integrity signature.
func (r *Repository) CreateReplaySeal(ctx context.Context, tx *sql.Tx, rs gen.ReplaySeal) error {
	return r.q.WithTx(tx).CreateReplaySeal(ctx, gen.CreateReplaySealParams{
		MatchID:          rs.MatchID,
		FinalEventNumber: rs.FinalEventNumber,
		FinalEventHash:   rs.FinalEventHash,
		TerminalAtMs:     rs.TerminalAtMs,
		SigningKeyID:     rs.SigningKeyID,
		Signature:        rs.Signature,
		ProofVersion:     rs.ProofVersion,
		CreatedAtMs:      rs.CreatedAtMs,
	})
}

// GetReplaySeal loads a replay seal for a match.
func (r *Repository) GetReplaySeal(ctx context.Context, matchID string) (gen.ReplaySeal, error) {
	return r.q.GetReplaySeal(ctx, matchID)
}

// --- Audit ---

// CreateAdminAuditLog inserts a protected audit record.
func (r *Repository) CreateAdminAuditLog(ctx context.Context, tx *sql.Tx, log gen.AdminAuditLog) error {
	return r.q.WithTx(tx).CreateAdminAuditLog(ctx, gen.CreateAdminAuditLogParams{
		Action:      log.Action,
		Actor:       log.Actor,
		Target:      log.Target,
		Details:     log.Details,
		CreatedAtMs: log.CreatedAtMs,
	})
}

// --- Retention helpers ---

// DeleteExpiredSessions removes sessions whose expiration has passed.
func (r *Repository) DeleteExpiredSessions(ctx context.Context, beforeMs int64) error {
	return r.q.DeleteExpiredRoomSessions(ctx, beforeMs)
}

// DeleteExpiredCommandReceipts removes expired idempotency receipts.
func (r *Repository) DeleteExpiredCommandReceipts(ctx context.Context, beforeMs int64) error {
	return r.q.DeleteExpiredCommandReceipts(ctx, beforeMs)
}

// DeleteExpiredReplayCapabilities removes expired replay capabilities.
func (r *Repository) DeleteExpiredReplayCapabilities(ctx context.Context, beforeMs int64) error {
	return r.q.DeleteExpiredReplayCapabilities(ctx, beforeMs)
}

// --- Mapping helpers ---

func roomToParams(r gen.Room) gen.CreateRoomParams {
	return gen.CreateRoomParams{
		ID:                r.ID,
		Code:              r.Code,
		State:             r.State,
		Version:           r.Version,
		Mode:              r.Mode,
		Difficulty:        r.Difficulty,
		ErrorPreset:       r.ErrorPreset,
		HintsEnabled:      r.HintsEnabled,
		SharedNotes:       r.SharedNotes,
		AutoRemoveNotes:   r.AutoRemoveNotes,
		SpectatorsAllowed: r.SpectatorsAllowed,
		HostParticipantID: r.HostParticipantID,
		CurrentMatchID:    r.CurrentMatchID,
		CreatedAtMs:       r.CreatedAtMs,
		LastActivityAtMs:  r.LastActivityAtMs,
		ExpiresAtMs:       r.ExpiresAtMs,
	}
}

func participantToParams(p gen.RoomParticipant) gen.CreateRoomParticipantParams {
	return gen.CreateRoomParticipantParams{
		ID:            p.ID,
		RoomID:        p.RoomID,
		DisplayName:   p.DisplayName,
		Role:          p.Role,
		IsHost:        p.IsHost,
		IsReady:       p.IsReady,
		JoinedAtMs:    p.JoinedAtMs,
		LeftAtMs:      p.LeftAtMs,
		RemovedAtMs:   p.RemovedAtMs,
		RemovedReason: p.RemovedReason,
	}
}

func updateParticipantToParams(p gen.RoomParticipant) gen.UpdateRoomParticipantParams {
	return gen.UpdateRoomParticipantParams{
		ID:            p.ID,
		DisplayName:   p.DisplayName,
		Role:          p.Role,
		IsHost:        p.IsHost,
		IsReady:       p.IsReady,
		LeftAtMs:      p.LeftAtMs,
		RemovedAtMs:   p.RemovedAtMs,
		RemovedReason: p.RemovedReason,
	}
}

func upsertParticipantToParams(p gen.RoomParticipant) gen.UpsertRoomParticipantParams {
	return gen.UpsertRoomParticipantParams{
		ID:            p.ID,
		RoomID:        p.RoomID,
		DisplayName:   p.DisplayName,
		Role:          p.Role,
		IsHost:        p.IsHost,
		IsReady:       p.IsReady,
		JoinedAtMs:    p.JoinedAtMs,
		LeftAtMs:      p.LeftAtMs,
		RemovedAtMs:   p.RemovedAtMs,
		RemovedReason: p.RemovedReason,
	}
}

func roomSessionToParams(s gen.RoomSession) gen.CreateRoomSessionParams {
	return gen.CreateRoomSessionParams{
		TokenHash:     s.TokenHash,
		RoomID:        s.RoomID,
		ParticipantID: s.ParticipantID,
		CreatedAtMs:   s.CreatedAtMs,
		ExpiresAtMs:   s.ExpiresAtMs,
		RevokedAtMs:   s.RevokedAtMs,
	}
}

func matchToParams(m gen.Match) gen.CreateMatchParams {
	return gen.CreateMatchParams{
		ID:                 m.ID,
		RoomID:             m.RoomID,
		State:              m.State,
		Version:            m.Version,
		Mode:               m.Mode,
		Difficulty:         m.Difficulty,
		ErrorPreset:        m.ErrorPreset,
		HintsEnabled:       m.HintsEnabled,
		AutoRemoveNotes:    m.AutoRemoveNotes,
		RuleVersion:        m.RuleVersion,
		PuzzleID:           m.PuzzleID,
		PuzzleRevision:     m.PuzzleRevision,
		TransformationSeed: m.TransformationSeed,
		PuzzleDifficulty:   m.PuzzleDifficulty,
		GeneratorVersion:   m.GeneratorVersion,
		SolverVersion:      m.SolverVersion,
		StartedAtMs:        m.StartedAtMs,
		CompletedAtMs:      m.CompletedAtMs,
		ResultReason:       m.ResultReason,
		ElapsedMs:          m.ElapsedMs,
		Assisted:           m.Assisted,
		CreatedAtMs:        m.CreatedAtMs,
	}
}

func matchParticipantToParams(p gen.MatchParticipant) gen.CreateMatchParticipantParams {
	return gen.CreateMatchParticipantParams{
		MatchID:       p.MatchID,
		ParticipantID: p.ParticipantID,
		Connected:     p.Connected,
		Mistakes:      p.Mistakes,
		HintsUsed:     p.HintsUsed,
	}
}

func updateMatchParticipantToParams(p gen.MatchParticipant) gen.UpdateMatchParticipantParams {
	return gen.UpdateMatchParticipantParams{
		MatchID:       p.MatchID,
		ParticipantID: p.ParticipantID,
		Connected:     p.Connected,
		Mistakes:      p.Mistakes,
		HintsUsed:     p.HintsUsed,
	}
}

func matchEventToParams(e gen.MatchEvent) gen.CreateMatchEventParams {
	return gen.CreateMatchEventParams{
		MatchID:              e.MatchID,
		EventNumber:          e.EventNumber,
		AggregateVersion:     e.AggregateVersion,
		PublicEventType:      e.PublicEventType,
		PublicActorID:        e.PublicActorID,
		RequestID:            e.RequestID,
		OccurredAtMs:         e.OccurredAtMs,
		PublicPayloadJson:    e.PublicPayloadJson,
		PrivatePayloadBlob:   e.PrivatePayloadBlob,
		PrivatePayloadSalt:   e.PrivatePayloadSalt,
		PrivatePayloadDigest: e.PrivatePayloadDigest,
		PreviousHash:         e.PreviousHash,
		EventHash:            e.EventHash,
	}
}

func matchSnapshotToParams(s gen.MatchSnapshot) gen.CreateMatchSnapshotParams {
	return gen.CreateMatchSnapshotParams{
		MatchID:          s.MatchID,
		EventNumber:      s.EventNumber,
		AggregateVersion: s.AggregateVersion,
		StateFormat:      s.StateFormat,
		StateBlob:        s.StateBlob,
		IntegrityHash:    s.IntegrityHash,
		CreatedAtMs:      s.CreatedAtMs,
	}
}

// NowMs returns the current time as Unix milliseconds.
func NowMs() int64 { return time.Now().UnixMilli() }

// FormatNullString is a small helper for nullable string fields.
func FormatNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// ParseNullString returns the empty string when invalid.
func ParseNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// FormatNullInt64 returns a valid sql.NullInt64.
func FormatNullInt64(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }

// ParseNullInt64 returns 0 when invalid.
func ParseNullInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

// ErrNoRows exposes sql.ErrNoRows from the repository package.
var ErrNoRows = sql.ErrNoRows

// IsNoRows reports whether an error is a missing-row error.
func IsNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }

// NewNullString is a compact alias for sql.NullString with a valid value.
func NewNullString(s string) sql.NullString { return FormatNullString(s) }

// NewNullInt64 is a compact alias for sql.NullInt64 with a valid value.
func NewNullInt64(v int64) sql.NullInt64 { return FormatNullInt64(v) }

// TxRollback rolls back a transaction, ignoring already-committed errors.
func TxRollback(tx *sql.Tx) error {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	return nil
}

// TxCommit commits a transaction.
func TxCommit(tx *sql.Tx) error { return tx.Commit() }

// ErrConflict is returned when an optimistic-version update affected zero rows.
var ErrConflict = errors.New("optimistic version conflict")

// CheckRowsAffected maps an update/exec result to ErrConflict if no rows were changed.
func CheckRowsAffected(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrConflict
	}
	return nil
}

// ContextWithTimeout returns a context with a short timeout for repository operations.
func ContextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// HashBytes returns a placeholder comment.
func changes(ctx context.Context, tx *sql.Tx) (int64, error) {
	var affected int64
	if err := tx.QueryRowContext(ctx, "SELECT changes()").Scan(&affected); err != nil {
		return 0, err
	}
	return affected, nil
}

func HashBytes(b []byte) string {
	return fmt.Sprintf("%x", b)
}
