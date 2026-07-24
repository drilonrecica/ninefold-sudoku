import { error } from '@sveltejs/kit';

import { normalizeRoomCode } from '$lib/api/client';

import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
  const code = normalizeRoomCode(params.code);
  if (!/^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{6}$/.test(code)) error(404, 'Room not found');
  return { code };
};
