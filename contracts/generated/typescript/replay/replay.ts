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
  expiresAt: SafePositiveInteger;
  clues: string;
  rules: {
    mode: "Coop";
    difficulty: "Easy" | "Medium" | "Hard" | "Expert";
    errorPreset: "Casual" | "Challenge" | "Blind" | "Clean";
    hintsEnabled: boolean;
    autoRemoveNotes: boolean;
    ruleVersion: SafePositiveInteger;
  };
  /**
   * @maxItems 8
   */
  participants:
    | []
    | [
        {
          id: Uuidv7;
          name: string;
        }
      ]
    | [
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        }
      ]
    | [
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        }
      ]
    | [
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        }
      ]
    | [
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        }
      ]
    | [
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        }
      ]
    | [
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        }
      ]
    | [
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        },
        {
          id: Uuidv7;
          name: string;
        }
      ];
  /**
   * @maxItems 10000
   */
  events: {
    proofVersion: 1;
    eventNumber: SafePositiveInteger;
    aggregateVersion: SafePositiveInteger;
    publicEventType: string;
    publicActorId: string;
    occurredAtMs: SafePositiveInteger;
    publicPayload: {
      [k: string]: unknown;
    };
    privatePayloadDigest: "" | string;
    previousEventHash: string;
    eventHash: string;
  }[];
  proof: {
    proofVersion: 1;
    matchId: Uuidv7;
    finalEventNumber: SafePositiveInteger;
    finalEventHash: string;
    terminalAtMs: SafePositiveInteger;
    keyId: string;
    signature: string;
  };
}
