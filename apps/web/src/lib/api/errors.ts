import type { MessageKey } from '$lib/i18n';

import { ApiError } from './client';

const safeErrorKeys: Record<string, MessageKey> = {
  NAME_INVALID: 'error.NAME_INVALID',
  NAME_ALREADY_USED: 'error.NAME_ALREADY_USED',
  ROOM_NOT_FOUND: 'error.ROOM_NOT_FOUND',
  ROOM_LOCKED: 'error.ROOM_LOCKED',
  ROOM_FULL: 'error.ROOM_FULL',
  ROOM_EXPIRED: 'error.ROOM_EXPIRED',
  ACTIVE_ROOM_SESSION_EXISTS: 'error.ACTIVE_ROOM_SESSION_EXISTS',
  RATE_LIMITED: 'error.RATE_LIMITED',
  SERVER_BUSY: 'error.SERVER_BUSY',
};

export function safeErrorKey(error: unknown): MessageKey {
  return error instanceof ApiError ? safeErrorKeyFromCode(error.code) : 'error.default';
}

export function safeErrorKeyFromCode(code: string | undefined): MessageKey {
  return code ? (safeErrorKeys[code] ?? 'error.default') : 'error.default';
}
