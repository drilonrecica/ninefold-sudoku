import { describe, expect, it } from 'vitest';

import { createRequestId, normalizeRoomCode } from './client';

describe('HTTP client helpers', () => {
  it('creates UUIDv7 request identifiers', () => {
    expect(createRequestId(1_767_225_600_000)).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
    );
  });

  it('normalizes room codes for paste and lowercase input', () => {
    expect(normalizeRoomCode(' 7k mp4r ')).toBe('7KMP4R');
  });
});
