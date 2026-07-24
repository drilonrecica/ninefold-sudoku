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
  target: {
    kind: "Room" | "Match";
    id: Uuidv7;
    expectedVersion: SafeInteger;
  };
  type: "sync.request";
  payload: {};
}
