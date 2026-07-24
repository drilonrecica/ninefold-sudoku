// Code generated from contracts. DO NOT EDIT.

/**
 * This interface was referenced by `ReplayDocument`'s JSON-Schema
 * via the `definition` "uuidv7".
 */
export type Uuidv7 = string;
/**
 * This interface was referenced by `ReplayDocument`'s JSON-Schema
 * via the `definition` "safePositiveInteger".
 */
export type SafePositiveInteger = number;

export interface ReplayDocument {
  schemaVersion: 1;
  replayId: Uuidv7;
  matchId: Uuidv7;
  /**
   * @maxItems 10000
   */
  events: {
    eventNumber: SafePositiveInteger;
    aggregateVersion: SafePositiveInteger;
    serverTimestamp: SafePositiveInteger;
    type: string;
    payload: {
      [k: string]: unknown;
    };
  }[];
}
