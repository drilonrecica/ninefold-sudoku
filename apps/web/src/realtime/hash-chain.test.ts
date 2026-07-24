import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { describe, expect, it } from 'vitest';

function canonicalize(value: unknown): string {
  if (value === null || typeof value !== 'object') return JSON.stringify(value);
  if (Array.isArray(value)) return `[${value.map(canonicalize).join(',')}]`;
  const record = value as Record<string, unknown>;
  return `{${Object.keys(record)
    .sort()
    .map((key) => `${JSON.stringify(key)}:${canonicalize(record[key])}`)
    .join(',')}}`;
}

describe('replay hash chain fixture', () => {
  it('matches the Go RFC 8785 hash', () => {
    const fixture = JSON.parse(
      readFileSync(
        resolve(process.cwd(), '../../contracts/fixtures/replay-hash-chain.json'),
        'utf8',
      ),
    ) as { envelope: unknown; eventHash: string };
    const hash = createHash('sha256').update(canonicalize(fixture.envelope)).digest('hex');
    expect(hash).toBe(fixture.eventHash);
  });
});
