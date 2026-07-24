import type { components } from '../../../../../contracts/generated/typescript/http/index';

export type RoomResponse = components['schemas']['RoomResponse'];
export type RoomPreview = components['schemas']['RoomPreviewResponse'];
export type Room = components['schemas']['Room'];
export type RoomParticipant = components['schemas']['RoomParticipant'];
export type Difficulty = components['schemas']['RoomDifficulty'];
export type ReplayCapabilityResponse = components['schemas']['ReplayCapabilityResponse'];

type ErrorEnvelope = components['schemas']['ErrorEnvelope'];

export class ApiError extends Error {
  constructor(
    public readonly code: string,
    public readonly status: number,
    public readonly details: Record<string, string | boolean | number | null> = {},
  ) {
    super(code);
  }
}

export function createRequestId(now = Date.now()): string {
  const bytes = crypto.getRandomValues(new Uint8Array(16));
  let timestamp = BigInt(now);
  for (let index = 5; index >= 0; index--) {
    bytes[index] = Number(timestamp & 0xffn);
    timestamp >>= 8n;
  }
  bytes[6] = 0x70 | ((bytes[6] ?? 0) & 0x0f);
  bytes[8] = 0x80 | ((bytes[8] ?? 0) & 0x3f);
  const hex = [...bytes].map((value) => value.toString(16).padStart(2, '0')).join('');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

export function normalizeRoomCode(value: string): string {
  return value.trim().toUpperCase().replaceAll(/\s+/g, '');
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set('Accept', 'application/json');
  if (init.body !== undefined) headers.set('Content-Type', 'application/json');
  if (init.method && init.method !== 'GET') headers.set('Idempotency-Key', createRequestId());

  const response = await fetch(`/api/v1${path}`, {
    ...init,
    headers,
    credentials: 'include',
  });
  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as ErrorEnvelope | null;
    throw new ApiError(body?.error.code ?? 'UNKNOWN', response.status, body?.error.details ?? {});
  }
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}

export function createRoom(displayName: string, difficulty: Difficulty): Promise<RoomResponse> {
  return request('/rooms', {
    method: 'POST',
    body: JSON.stringify({ displayName, difficulty, mode: 'Coop' }),
  });
}

export function previewRoom(code: string, signal?: AbortSignal): Promise<RoomPreview> {
  return request(`/rooms/${encodeURIComponent(normalizeRoomCode(code))}`, { signal });
}

export function joinRoom(
  code: string,
  displayName: string,
  role: 'Player' | 'Spectator' = 'Player',
): Promise<RoomResponse> {
  return request(`/rooms/${encodeURIComponent(normalizeRoomCode(code))}/join`, {
    method: 'POST',
    body: JSON.stringify({ displayName, role }),
  });
}

export function resumeRoom(code: string): Promise<RoomResponse> {
  return request(`/rooms/${encodeURIComponent(normalizeRoomCode(code))}/resume`, {
    method: 'POST',
  });
}

export function leaveRoom(code: string): Promise<void> {
  return request(`/rooms/${encodeURIComponent(normalizeRoomCode(code))}/leave`, {
    method: 'POST',
    body: JSON.stringify({ intent: 'leave_lobby' }),
  });
}

export function createReplayCapability(matchId: string): Promise<ReplayCapabilityResponse> {
  return request(`/replays/${encodeURIComponent(matchId)}/capabilities`, {
    method: 'POST',
  });
}
