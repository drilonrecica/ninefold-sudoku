import type { ReplayDocument, ReplayEvent } from './reducer';

export type VerificationStatus =
  'verified' | 'legacy' | 'corrupted' | 'unknown-key' | 'unsupported';

export interface VerificationResult {
  status: VerificationStatus;
  hiddenCommitments: number;
}

const genesisHash: Uint8Array<ArrayBuffer> = new Uint8Array(32);

export async function verifyReplay(
  replay: ReplayDocument,
  trustedKeys: Readonly<Record<string, string>>,
): Promise<VerificationResult> {
  if (replay.schemaVersion !== 1 || replay.proof?.proofVersion !== 1) {
    return { status: 'legacy', hiddenCommitments: 0 };
  }
  if (!globalThis.crypto?.subtle) return { status: 'unsupported', hiddenCommitments: 0 };
  const publicKey = trustedKeys[replay.proof.keyId];
  if (!publicKey) return { status: 'unknown-key', hiddenCommitments: 0 };

  try {
    let previous = genesisHash;
    let hiddenCommitments = 0;
    for (let index = 0; index < replay.events.length; index++) {
      const event = replay.events[index]!;
      if (event.eventNumber !== index + 1 || event.proofVersion !== 1) {
        return { status: 'corrupted', hiddenCommitments };
      }
      if (!equalBytes(fromBase64(event.previousEventHash), previous)) {
        return { status: 'corrupted', hiddenCommitments };
      }
      if (event.privatePayloadDigest) hiddenCommitments++;
      const digest = new Uint8Array(
        await crypto.subtle.digest(
          'SHA-256',
          new TextEncoder().encode(canonicalize(envelope(replay.matchId, event))),
        ),
      );
      if (toHex(digest) !== event.eventHash) {
        return { status: 'corrupted', hiddenCommitments };
      }
      previous = digest;
    }
    if (
      replay.events.length === 0 ||
      replay.proof.matchId !== replay.matchId ||
      replay.proof.finalEventNumber !== replay.events.length ||
      replay.proof.finalEventHash !== toHex(previous)
    ) {
      return { status: 'corrupted', hiddenCommitments };
    }
    const key = await crypto.subtle.importKey(
      'raw',
      fromBase64(publicKey),
      { name: 'Ed25519' },
      false,
      ['verify'],
    );
    const valid = await crypto.subtle.verify(
      { name: 'Ed25519' },
      key,
      fromBase64(replay.proof.signature),
      previous,
    );
    return { status: valid ? 'verified' : 'corrupted', hiddenCommitments };
  } catch (error) {
    if (error instanceof DOMException && error.name === 'NotSupportedError') {
      return { status: 'unsupported', hiddenCommitments: 0 };
    }
    return { status: 'corrupted', hiddenCommitments: 0 };
  }
}

function envelope(matchId: string, event: ReplayEvent): Record<string, unknown> {
  return {
    proofVersion: event.proofVersion,
    matchId,
    eventNumber: event.eventNumber,
    aggregateVersion: event.aggregateVersion,
    publicEventType: event.publicEventType,
    publicActorId: event.publicActorId,
    occurredAtMs: event.occurredAtMs,
    publicPayload: event.publicPayload,
    privatePayloadDigest: event.privatePayloadDigest,
    previousEventHash: event.previousEventHash,
  };
}

export function canonicalize(value: unknown): string {
  if (value === null || typeof value === 'boolean' || typeof value === 'string') {
    return JSON.stringify(value);
  }
  if (typeof value === 'number') {
    if (!Number.isFinite(value)) throw new TypeError('non-finite JSON number');
    return JSON.stringify(value);
  }
  if (Array.isArray(value)) return `[${value.map(canonicalize).join(',')}]`;
  if (typeof value === 'object') {
    const record = value as Record<string, unknown>;
    return `{${Object.keys(record)
      .sort()
      .map((key) => `${JSON.stringify(key)}:${canonicalize(record[key])}`)
      .join(',')}}`;
  }
  throw new TypeError('unsupported JSON value');
}

function fromBase64(value: string): Uint8Array<ArrayBuffer> {
  const binary = atob(value);
  const bytes = new Uint8Array(binary.length);
  for (let index = 0; index < binary.length; index++) bytes[index] = binary.charCodeAt(index);
  return bytes;
}

function toHex(value: Uint8Array): string {
  return [...value].map((byte) => byte.toString(16).padStart(2, '0')).join('');
}

function equalBytes(left: Uint8Array, right: Uint8Array): boolean {
  return left.length === right.length && left.every((value, index) => value === right[index]);
}
