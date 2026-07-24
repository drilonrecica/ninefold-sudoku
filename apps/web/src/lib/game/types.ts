export type Digit = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9;
export type MatchState =
  'Prepared' | 'Countdown' | 'Active' | 'RecoveryPending' | 'Completed' | 'Cancelled';
export type ErrorPreset = 'Casual' | 'Challenge' | 'Blind' | 'Clean';

export interface MatchRules {
  mode: 'Coop';
  difficulty: 'Easy' | 'Medium' | 'Hard' | 'Expert';
  errorPreset: ErrorPreset;
  hintsEnabled: boolean;
  autoRemoveNotes: boolean;
  ruleVersion: number;
}

export interface MatchCell {
  index: number;
  isClue: boolean;
  value?: Digit;
  notes: Digit[];
  attribution?: string;
  correct?: boolean | null;
}

export interface MatchSnapshot {
  id: string;
  state: MatchState;
  version: number;
  startedAt?: number;
  completedAt?: number;
  penaltiesMs: number;
  pausedMs?: number;
  recoveryStartedAt?: number;
  recoveryGeneration?: number;
  hintsUsed: number;
  assisted: boolean;
  rules: MatchRules;
  cells: MatchCell[];
  mistakes: Record<string, number>;
  contributions: Record<string, number>;
  result?: MatchResult;
}

export interface MatchResult {
  reason: string;
  completedAt: number;
  elapsedMs: number;
  penaltyMs: number;
  assisted: boolean;
  mistakesByPlayer: Record<string, number>;
  contributionsByPlayer: Record<string, number>;
  disconnectsByPlayer: Record<string, number>;
  hintCount: number;
  contributionCount: number;
  replayAvailable: boolean;
}

export interface MatchEventEnvelope {
  schemaVersion: number;
  eventNumber: number;
  aggregateVersion: number;
  serverTimestamp: number;
  event: {
    type: string;
    payload: Record<string, unknown>;
  };
}

export type PendingKind = 'place' | 'erase' | 'add_note' | 'remove_note' | 'hint';

export interface PendingCommand {
  requestId: string;
  kind: PendingKind;
  cell?: number;
  digit?: Digit;
  createdAt: number;
  uncertain: boolean;
}

export interface MatchClientState {
  confirmed: MatchSnapshot | null;
  lastEventNumber: number;
  pending: Map<string, PendingCommand>;
  recoveryRequired: boolean;
  announcement: string;
}
