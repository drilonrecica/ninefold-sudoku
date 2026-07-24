// Code generated from contracts. DO NOT EDIT.

/**
 * This interface was referenced by `ServerMessage`'s JSON-Schema
 * via the `definition` "safePositiveInteger".
 */
export type SafePositiveInteger = number;

export interface ServerMessage {
  schemaVersion: 1;
  eventNumber: SafePositiveInteger;
  aggregateVersion: SafePositiveInteger;
  serverTimestamp: number;
  type: "ack" | "rejection" | "snapshot";
  payload: {
    [k: string]: unknown;
  };
}
