import { describe, expect, it } from 'vitest';

import { sanitizeCheckpoint } from './checkpoint';

describe('connection checkpoint', () => {
  it('retains only bounded UI-safe recovery metadata', () => {
    const requestId = '01900000-0000-7000-8000-000000000003';
    expect(
      sanitizeCheckpoint({
        roomCode: ' abcd1234-secret ',
        roomVersion: 4,
        matchId: '01900000-0000-7000-8000-000000000001',
        matchVersion: 8,
        matchEventNumber: 11,
        pendingRequestIds: [requestId, requestId, 'not-a-token'],
        updatedAt: 123,
      }),
    ).toEqual({
      roomCode: 'ABCD1234',
      roomVersion: 4,
      matchId: '01900000-0000-7000-8000-000000000001',
      matchVersion: 8,
      matchEventNumber: 11,
      pendingRequestIds: [requestId],
      updatedAt: 123,
    });
  });
});
