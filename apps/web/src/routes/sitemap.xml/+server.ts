import type { RequestHandler } from './$types';

const pages = ['', '/how-to-play', '/privacy', '/accessibility'];

export const GET: RequestHandler = () => {
  const urls = pages
    .map((path) => `<url><loc>https://ninefold.recica.dev${path || '/'}</loc></url>`)
    .join('');
  return new Response(
    `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">${urls}</urlset>`,
    {
      headers: { 'Content-Type': 'application/xml; charset=utf-8' },
    },
  );
};
