import { redirect } from '@sveltejs/kit';

import { normalizeRoomCode } from '$lib/api/client';

import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ url }) => {
  const code = normalizeRoomCode(url.searchParams.get('code') ?? '');
  if (/^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{6}$/.test(code)) {
    redirect(303, `/join/${code}`);
  }
  return { invalidCode: code.length > 0 };
};
