// Code generated from contracts. DO NOT EDIT.

/**
 * This interface was referenced by `ServerMessage`'s JSON-Schema
 * via the `definition` "safeInteger".
 */
export type SafeInteger = number;
/**
 * This interface was referenced by `ServerMessage`'s JSON-Schema
 * via the `definition` "uuidv7".
 */
export type Uuidv7 = string;
/**
 * This interface was referenced by `ServerMessage`'s JSON-Schema
 * via the `definition` "safePositiveInteger".
 */
export type SafePositiveInteger = number;

export interface ServerMessage {
  schemaVersion: 1;
  eventNumber: SafeInteger;
  aggregateVersion: SafeInteger;
  serverTimestamp: number;
  type:
    | "connection.accepted"
    | "connection.rejected"
    | "connection.read_only"
    | "connection.controller_revoked"
    | "connection.status"
    | "command.acknowledged"
    | "command.rejected"
    | "command.status"
    | "room.snapshot"
    | "room.event"
    | "match.snapshot"
    | "match.event"
    | "match.completed"
    | "ephemeral.focus"
    | "ephemeral.soft_lock"
    | "ephemeral.reaction";
  payload: {
    requestId?: Uuidv7;
    accepted?: boolean;
    aggregate?: "room" | "match";
    resultingVersion?: SafeInteger;
    code?: string;
    status?: string;
    protocolVersion?: SafePositiveInteger;
    isController?: boolean;
    currentVersion?: SafeInteger;
    currentMatchEventNumber?: SafeInteger;
    details?: {};
    message?: string;
    connectionState?: "connected" | "reconnecting" | "synchronizing" | "read_only" | "recovery_failed";
    controllerGeneration?: SafeInteger;
    room?: {};
    match?: {};
    event?: {};
    eventNumber?: SafePositiveInteger;
    matchVersion?: SafePositiveInteger;
    roomVersion?: SafePositiveInteger;
    participantId?: Uuidv7;
    cell?: number;
    intent?: string;
    reaction?: "agree" | "nice_move";
    focused?: boolean;
    identity?: {
      participantId?: Uuidv7;
      role?: string;
      isHost?: boolean;
    };
  };
}
