import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { describe, expect, it } from 'vitest';

import type { ReplayDocument } from './reducer';
import { canonicalize, verifyReplay } from './verifier';

const fixture = JSON.parse(
  readFileSync(resolve(process.cwd(), '../../contracts/fixtures/replay-proof-v1.json'), 'utf8'),
) as {
  events: { envelope: Record<string, unknown>; canonical: string; eventHash: string }[];
  signing: { keyId: string; publicKey: string; finalHash: string; signature: string };
};

function document(): ReplayDocument {
  return {
    schemaVersion: 1,
    replayId: '01900000-0000-7000-8000-000000000003',
    matchId: '01900000-0000-7000-8000-000000000001',
    expiresAt: 2_000_000_000_000,
    clues: '0'.repeat(81),
    rules: {
      mode: 'Coop',
      difficulty: 'Easy',
      errorPreset: 'Casual',
      hintsEnabled: true,
      autoRemoveNotes: true,
      ruleVersion: 1,
    },
    participants: [],
    events: fixture.events.map(({ envelope, eventHash }) => ({
      ...(envelope as unknown as ReplayDocument['events'][number]),
      eventHash,
    })),
    proof: {
      proofVersion: 1,
      matchId: '01900000-0000-7000-8000-000000000001',
      finalEventNumber: 2,
      finalEventHash: fixture.signing.finalHash,
      terminalAtMs: 1767225600123,
      keyId: fixture.signing.keyId,
      signature: fixture.signing.signature,
    },
  };
}

describe('replay verifier', () => {
  it('matches the cross-language canonical vectors', () => {
    for (const event of fixture.events) expect(canonicalize(event.envelope)).toBe(event.canonical);
  });

  it('verifies a sealed chain and reports hidden commitments', async () => {
    await expect(
      verifyReplay(document(), { [fixture.signing.keyId]: fixture.signing.publicKey }),
    ).resolves.toEqual({ status: 'verified', hiddenCommitments: 1 });
  });

  it('rejects mutations, gaps, signatures, and unknown keys', async () => {
    const cases = [
      (value: ReplayDocument) => (value.events[0]!.publicPayload = { schemaVersion: 2 }),
      (value: ReplayDocument) => value.events.reverse(),
      (value: ReplayDocument) => value.events[1]!.occurredAtMs++,
      (value: ReplayDocument) => (value.events[1]!.previousEventHash = 'A'.repeat(44)),
      (value: ReplayDocument) => (value.events[1]!.privatePayloadDigest = '0'.repeat(64)),
      (value: ReplayDocument) => value.events.pop(),
      (value: ReplayDocument) => value.events.push(structuredClone(value.events[1]!)),
      (value: ReplayDocument) => (value.proof.signature = `A${value.proof.signature.slice(1)}`),
    ];
    for (const mutate of cases) {
      const replay = document();
      mutate(replay);
      expect(
        (await verifyReplay(replay, { [fixture.signing.keyId]: fixture.signing.publicKey })).status,
      ).toBe('corrupted');
    }
    await expect(verifyReplay(document(), {})).resolves.toMatchObject({ status: 'unknown-key' });
  });
});
