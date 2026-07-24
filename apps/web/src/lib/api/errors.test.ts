import { describe, expect, it } from 'vitest';

import { ApiError } from './client';
import { safeErrorKey, safeErrorKeyFromCode } from './errors';

describe('safe API errors', () => {
  it('maps known server codes without exposing server messages', () => {
    expect(safeErrorKey(new ApiError('ROOM_FULL', 409))).toBe('error.ROOM_FULL');
  });

  it('uses a generic message for unknown and malformed errors', () => {
    expect(safeErrorKeyFromCode('DATABASE_CONNECTION_STRING')).toBe('error.default');
    expect(safeErrorKey(new Error('private detail'))).toBe('error.default');
  });
});
