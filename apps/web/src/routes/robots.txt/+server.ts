import type { RequestHandler } from './$types';

export const GET: RequestHandler = () =>
  new Response(
    'User-agent: *\nAllow: /\nDisallow: /create\nDisallow: /join/\nDisallow: /room/\nDisallow: /play/\nDisallow: /solo\nDisallow: /replay/\nDisallow: /settings\nDisallow: /admin/\nSitemap: https://ninefold.recica.dev/sitemap.xml\n',
    {
      headers: { 'Content-Type': 'text/plain; charset=utf-8' },
    },
  );
