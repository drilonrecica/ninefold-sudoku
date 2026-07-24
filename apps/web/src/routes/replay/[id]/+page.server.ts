import { error } from '@sveltejs/kit';

import type { PageServerLoad } from './$types';

const uuidV7 = /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

export const load: PageServerLoad = ({ params }) => {
  if (!uuidV7.test(params.id)) error(404, 'Replay unavailable');
  return { replayId: params.id };
};
