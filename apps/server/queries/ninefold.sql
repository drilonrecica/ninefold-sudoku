-- name: CreatePuzzle :exec
INSERT INTO puzzles (
    id, revision, state, difficulty, hardest_technique, quality_score,
    multiplayer_approved, generator_version, solver_version, canonical_fingerprint,
    clues, solution, created_at_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetPuzzleByID :one
SELECT * FROM puzzles
WHERE id = ? AND revision = ?;

-- name: GetPuzzleByFingerprint :one
SELECT * FROM puzzles
WHERE canonical_fingerprint = ?;

-- name: CountPuzzlesByDifficulty :one
SELECT COUNT(*) FROM puzzles
WHERE state = 'Active' AND difficulty = ?;

-- name: ListActivePuzzlesByDifficulty :many
SELECT id, revision, difficulty, hardest_technique, quality_score, multiplayer_approved,
       generator_version, solver_version, canonical_fingerprint, clues, created_at_ms
FROM puzzles
WHERE state = 'Active' AND difficulty = ?
ORDER BY id, revision;

-- name: CountActivePuzzlesByDifficultyAndMultiplayer :one
SELECT COUNT(*) FROM puzzles
WHERE state = 'Active' AND difficulty = ? AND multiplayer_approved = ?;

-- name: ListActivePuzzlesByDifficultyAndMultiplayer :many
SELECT id, revision, difficulty, hardest_technique, quality_score, multiplayer_approved,
       generator_version, solver_version, canonical_fingerprint, clues, created_at_ms
FROM puzzles
WHERE state = 'Active' AND difficulty = ? AND multiplayer_approved = ?
ORDER BY id, revision;

-- name: CreateRoom :exec
INSERT INTO rooms (
    id, code, state, version, mode, difficulty, error_preset, hints_enabled,
    shared_notes, auto_remove_notes, spectators_allowed, host_participant_id,
    current_match_id, created_at_ms, last_activity_at_ms, expires_at_ms, rematch_number
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateRoom :exec
UPDATE rooms SET
    code = ?, state = ?, version = ?, mode = ?, difficulty = ?, error_preset = ?,
    hints_enabled = ?, shared_notes = ?, auto_remove_notes = ?, spectators_allowed = ?,
    host_participant_id = ?, current_match_id = ?, last_activity_at_ms = ?, expires_at_ms = ?,
    rematch_number = ?
WHERE id = ? AND version = ?;

-- name: GetRoomByID :one
SELECT * FROM rooms WHERE id = ?;

-- name: GetRoomByCode :one
SELECT * FROM rooms WHERE code = ?;

-- name: CreateRoomParticipant :exec
INSERT INTO room_participants (
    id, room_id, display_name, role, is_host, is_ready, joined_at_ms,
    left_at_ms, removed_at_ms, removed_reason
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateRoomParticipant :exec
UPDATE room_participants SET
    display_name = ?, role = ?, is_host = ?, is_ready = ?, left_at_ms = ?,
    removed_at_ms = ?, removed_reason = ?
WHERE id = ?;

-- name: UpsertRoomParticipant :exec
INSERT INTO room_participants (
    id, room_id, display_name, role, is_host, is_ready, joined_at_ms,
    left_at_ms, removed_at_ms, removed_reason
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    room_id = excluded.room_id,
    display_name = excluded.display_name,
    role = excluded.role,
    is_host = excluded.is_host,
    is_ready = excluded.is_ready,
    joined_at_ms = excluded.joined_at_ms,
    left_at_ms = excluded.left_at_ms,
    removed_at_ms = excluded.removed_at_ms,
    removed_reason = excluded.removed_reason;

-- name: GetRoomParticipantByID :one
SELECT * FROM room_participants WHERE id = ?;

-- name: ListActiveRoomParticipants :many
SELECT * FROM room_participants
WHERE room_id = ? AND left_at_ms IS NULL AND removed_at_ms IS NULL
ORDER BY joined_at_ms;

-- name: CountActiveRoomParticipants :one
SELECT COUNT(*) FROM room_participants
WHERE room_id = ? AND left_at_ms IS NULL AND removed_at_ms IS NULL;

-- name: CountActiveRoomPlayers :one
SELECT COUNT(*) FROM room_participants
WHERE room_id = ? AND left_at_ms IS NULL AND removed_at_ms IS NULL AND role = 'Player';

-- name: CountActiveRoomSpectators :one
SELECT COUNT(*) FROM room_participants
WHERE room_id = ? AND left_at_ms IS NULL AND removed_at_ms IS NULL AND role = 'Spectator';

-- name: GetRoomBlock :one
SELECT * FROM room_blocks WHERE room_id = ? AND participant_id = ?;

-- name: CreateRoomBlock :exec
INSERT INTO room_blocks (room_id, participant_id, blocked_at_ms, reason)
VALUES (?, ?, ?, ?);

-- name: DeleteRoomBlock :exec
DELETE FROM room_blocks WHERE room_id = ? AND participant_id = ?;

-- name: CreateRoomSession :exec
INSERT INTO room_sessions (
    token_hash, room_id, participant_id, created_at_ms, expires_at_ms, revoked_at_ms
) VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateRoomSession :exec
UPDATE room_sessions SET
    expires_at_ms = ?, revoked_at_ms = ?
WHERE token_hash = ?;

-- name: GetRoomSessionByHash :one
SELECT * FROM room_sessions WHERE token_hash = ?;

-- name: ListActiveRoomSessionsByParticipant :many
SELECT * FROM room_sessions
WHERE participant_id = ? AND revoked_at_ms IS NULL AND expires_at_ms > ?;

-- name: DeleteExpiredRoomSessions :exec
DELETE FROM room_sessions WHERE expires_at_ms <= ?;

-- name: CreateMatch :exec
INSERT INTO matches (
    id, room_id, state, version, mode, difficulty, error_preset, hints_enabled,
    auto_remove_notes, rule_version, puzzle_id, puzzle_revision, transformation_seed,
    puzzle_difficulty, generator_version, solver_version, started_at_ms,
    completed_at_ms, result_reason, elapsed_ms, assisted, created_at_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateMatch :exec
UPDATE matches SET
    room_id = ?, state = ?, version = ?, mode = ?, difficulty = ?, error_preset = ?,
    hints_enabled = ?, auto_remove_notes = ?, rule_version = ?, puzzle_id = ?,
    puzzle_revision = ?, transformation_seed = ?, puzzle_difficulty = ?,
    generator_version = ?, solver_version = ?, started_at_ms = ?, completed_at_ms = ?,
    result_reason = ?, elapsed_ms = ?, assisted = ?
WHERE id = ? AND version = ?;

-- name: GetMatchByID :one
SELECT * FROM matches WHERE id = ?;

-- name: GetMatchByRoomID :many
SELECT * FROM matches WHERE room_id = ? ORDER BY created_at_ms DESC;

-- name: ListRecentPuzzleIDsByRoom :many
SELECT puzzle_id FROM matches WHERE room_id = ? ORDER BY created_at_ms DESC LIMIT 20;

-- name: CreateMatchParticipant :exec
INSERT INTO match_participants (
    match_id, participant_id, connected, mistakes, hints_used
) VALUES (?, ?, ?, ?, ?);

-- name: UpdateMatchParticipant :exec
UPDATE match_participants SET
    connected = ?, mistakes = ?, hints_used = ?
WHERE match_id = ? AND participant_id = ?;

-- name: GetMatchParticipant :one
SELECT * FROM match_participants WHERE match_id = ? AND participant_id = ?;

-- name: ListMatchParticipants :many
SELECT * FROM match_participants WHERE match_id = ?;

-- name: CreateMatchEvent :exec
INSERT INTO match_events (
    match_id, event_number, aggregate_version, public_event_type, public_actor_id,
    request_id, occurred_at_ms, public_payload_json, private_payload_blob,
    private_payload_salt, private_payload_digest, previous_hash, event_hash
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetMatchEvents :many
SELECT * FROM match_events WHERE match_id = ? ORDER BY event_number;

-- name: GetMatchEventByNumber :one
SELECT * FROM match_events WHERE match_id = ? AND event_number = ?;

-- name: GetMaxMatchEventNumber :one
SELECT COALESCE(MAX(event_number), 0) FROM match_events WHERE match_id = ?;

-- name: CreateMatchSnapshot :exec
INSERT INTO match_snapshots (
    match_id, event_number, aggregate_version, state_format, state_blob, integrity_hash, created_at_ms
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetLatestMatchSnapshot :one
SELECT * FROM match_snapshots
WHERE match_id = ?
ORDER BY event_number DESC
LIMIT 1;

-- name: GetMatchSnapshotByEventNumber :one
SELECT * FROM match_snapshots WHERE match_id = ? AND event_number = ?;

-- name: CreateMatchResult :exec
INSERT INTO match_results (
    match_id, result_reason, elapsed_ms, assisted, created_at_ms
) VALUES (?, ?, ?, ?, ?);

-- name: CreateMatchResultPlayer :exec
INSERT INTO match_result_players (
    match_id, participant_id, display_name, mistakes, hints_used, score
) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetMatchResult :one
SELECT * FROM match_results WHERE match_id = ?;

-- name: GetMatchResultPlayers :many
SELECT * FROM match_result_players WHERE match_id = ? ORDER BY participant_id;

-- name: CreateMatchTombstone :exec
INSERT INTO match_tombstones (
    match_id, mode, difficulty, result_reason, started_at_ms, ended_at_ms,
    schema_version, proof_version, replay_deleted_at_ms, replay_expired_at_ms, created_at_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetMatchTombstone :one
SELECT * FROM match_tombstones WHERE match_id = ?;

-- name: CreateCommandReceipt :exec
INSERT INTO command_receipts (
    request_id, authenticated_scope_hash, command_type, request_fingerprint,
    terminal_status, safe_response_json, created_at_ms, expires_at_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetCommandReceipt :one
SELECT * FROM command_receipts WHERE request_id = ?;

-- name: DeleteExpiredCommandReceipts :exec
DELETE FROM command_receipts WHERE expires_at_ms <= ?;

-- name: CreateReplayCapability :exec
INSERT INTO replay_capabilities (
    token_hash, replay_id, match_id, created_at_ms, expires_at_ms, revoked_at_ms
) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetReplayCapabilityByHash :one
SELECT * FROM replay_capabilities WHERE token_hash = ?;

-- name: GetReplayCapabilityByReplayID :one
SELECT * FROM replay_capabilities WHERE replay_id = ?;

-- name: RevokeReplayCapability :exec
UPDATE replay_capabilities SET revoked_at_ms = ? WHERE token_hash = ?;

-- name: RevokeReplayCapabilitiesByMatch :exec
UPDATE replay_capabilities SET revoked_at_ms = ?
WHERE match_id = ? AND revoked_at_ms IS NULL;

-- name: ListReplayCapabilitiesByMatch :many
SELECT * FROM replay_capabilities WHERE match_id = ?;

-- name: DeleteExpiredReplayCapabilities :exec
DELETE FROM replay_capabilities WHERE expires_at_ms <= ?;

-- name: CreateReplaySeal :exec
INSERT INTO replay_seals (
    match_id, final_event_number, final_event_hash, terminal_at_ms,
    signing_key_id, signature, proof_version, created_at_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetReplaySeal :one
SELECT * FROM replay_seals WHERE match_id = ?;

-- name: CreateAdminAuditLog :exec
INSERT INTO admin_audit_log (action, actor, target, details, created_at_ms)
VALUES (?, ?, ?, ?, ?);

-- name: ListAdminAuditLog :many
SELECT * FROM admin_audit_log ORDER BY created_at_ms DESC LIMIT ?;

-- name: DeleteExpiredAdminAuditLog :exec
DELETE FROM admin_audit_log WHERE created_at_ms <= ?;

-- name: ListNonTerminalMatches :many
SELECT * FROM matches WHERE state NOT IN ('Completed', 'Cancelled', 'Abandoned', 'Invalidated');

-- name: ListExpiredRooms :many
SELECT * FROM rooms WHERE expires_at_ms <= ? AND state NOT IN ('Expired', 'Cancelled', 'TerminatedByAdmin');

-- name: ExpireRoom :exec
UPDATE rooms SET state = 'Expired', expires_at_ms = ? WHERE id = ?;

-- name: ScrubParticipantLinkedMatchData :exec
UPDATE match_events SET
    public_actor_id = NULL,
    public_payload_json = '',
    private_payload_blob = NULL,
    private_payload_salt = NULL,
    private_payload_digest = NULL,
    previous_hash = NULL,
    event_hash = NULL,
    request_id = ''
WHERE match_id = ?;

-- name: DeleteMatchSnapshots :exec
DELETE FROM match_snapshots WHERE match_id = ?;

-- name: DeleteReplayCapabilitiesByMatch :exec
DELETE FROM replay_capabilities WHERE match_id = ?;

-- name: DeleteReplaySealByMatch :exec
DELETE FROM replay_seals WHERE match_id = ?;

-- name: DeleteMatchResultPlayersByMatch :exec
DELETE FROM match_result_players WHERE match_id = ?;

-- name: DeleteMatchResultByMatch :exec
DELETE FROM match_results WHERE match_id = ?;

-- name: DeleteMatchParticipantsByMatch :exec
DELETE FROM match_participants WHERE match_id = ?;

-- name: DeleteMatchEventsByMatch :exec
DELETE FROM match_events WHERE match_id = ?;

-- name: DeleteMatchByID :exec
DELETE FROM matches WHERE id = ?;

-- name: AnonymizeRoomParticipants :exec
UPDATE room_participants SET display_name = '' WHERE room_id = ?;

-- name: DeleteRoomSessionsByRoom :exec
DELETE FROM room_sessions WHERE room_id = ?;

-- name: DeleteRoomBlocksByRoom :exec
DELETE FROM room_blocks WHERE room_id = ?;

-- name: CountRoomsByCode :one
SELECT COUNT(*) FROM rooms WHERE code = ?;

-- name: CountActiveRoomsByParticipant :one
SELECT COUNT(*) FROM room_sessions rs
JOIN room_participants rp ON rp.id = rs.participant_id
JOIN rooms r ON r.id = rs.room_id
WHERE rs.token_hash = ? AND rs.revoked_at_ms IS NULL AND rs.expires_at_ms > ?;

-- name: ListReplayCapabilitiesExpired :many
SELECT * FROM replay_capabilities WHERE expires_at_ms <= ?;

-- name: ListCommandReceiptsExpired :many
SELECT * FROM command_receipts WHERE expires_at_ms <= ?;

-- name: ListRoomSessionsExpired :many
SELECT * FROM room_sessions WHERE expires_at_ms <= ?;

-- name: ListMatchTombstonesExpired :many
SELECT * FROM match_tombstones WHERE ended_at_ms <= ?;

-- name: ListRoomsExpired :many
SELECT * FROM rooms WHERE expires_at_ms <= ?;
