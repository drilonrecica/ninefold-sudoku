import { error } from '@sveltejs/kit';

import type { RoomPreview } from '$lib/api/client';
import { normalizeRoomCode } from '$lib/api/client';
import { apiBaseUrl } from '$lib/server/api';

import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, fetch }) => {
  const code = normalizeRoomCode(params.code);
  if (!/^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{6}$/.test(code)) error(404, 'Room not found');

  const response = await fetch(`${apiBaseUrl()}/api/v1/rooms/${code}`, {
    headers: { Accept: 'application/json' },
  });
  if (response.status === 404) error(404, 'Room not found');
  if (!response.ok) error(503, 'Room preview unavailable');
  return { code, preview: (await response.json()) as RoomPreview };
};
