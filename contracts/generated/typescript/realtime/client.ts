// Code generated from contracts. DO NOT EDIT.

/**
 * This interface was referenced by `ClientMessage`'s JSON-Schema
 * via the `definition` "uuidv7".
 */
export type Uuidv7 = string;
/**
 * This interface was referenced by `ClientMessage`'s JSON-Schema
 * via the `definition` "safePositiveInteger".
 */
export type SafePositiveInteger = number;
/**
 * This interface was referenced by `ClientMessage`'s JSON-Schema
 * via the `definition` "safeInteger".
 */
export type SafeInteger = number;

export interface ClientMessage {
  schemaVersion: 1;
  requestId: Uuidv7;
  clientSequence: SafePositiveInteger;
  target?: {
    kind: "Room" | "Match";
    id: Uuidv7;
    expectedVersion: SafeInteger;
  };
  type:
    | "connection.initialize"
    | "connection.request_control"
    | "connection.heartbeat"
    | "room.set_ready"
    | "room.change_settings"
    | "room.start_countdown"
    | "room.cancel_countdown"
    | "room.leave"
    | "room.transfer_host"
    | "match.place_value"
    | "match.erase_value"
    | "match.add_note"
    | "match.remove_note"
    | "match.use_hint"
    | "match.ping"
    | "match.reaction"
    | "match.focus_cell"
    | "match.release_focus"
    | "command.status";
  payload: {
    roomCode?: string;
    lastRoomVersion?: SafeInteger;
    lastMatchId?: Uuidv7;
    lastMatchEventNumber?: SafeInteger;
    ready?: boolean;
    displayName?: string;
    participantId?: Uuidv7;
    settings?: {
      difficulty?: string;
      errorPreset?: string;
      hintsEnabled?: boolean;
      autoRemoveNotes?: boolean;
      locked?: boolean;
    };
    cell?: number;
    value?: number;
    digits?: number[];
    level?: "Nudge" | "Reveal";
    intent?: string;
    reaction?: "agree" | "nice_move";
    targetCell?: number;
  };
}
