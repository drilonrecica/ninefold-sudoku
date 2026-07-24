import type { RequestHandler } from './$types';

export const GET: RequestHandler = () =>
  new Response(
    'User-agent: *\nAllow: /\nDisallow: /room/\nDisallow: /join/\nDisallow: /play/\nDisallow: /replay/\nDisallow: /admin/\nSitemap: https://ninefold.recica.dev/sitemap.xml\n',
    {
      headers: { 'Content-Type': 'text/plain; charset=utf-8' },
    },
  );
