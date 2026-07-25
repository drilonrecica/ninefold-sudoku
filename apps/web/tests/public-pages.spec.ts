import AxeBuilder from '@axe-core/playwright';
import { expect, test } from 'playwright/test';

const publicPages = [
  ['/', 'Sudoku is better together.', 'https://ninefold.recica.dev/'],
  ['/how-to-play', 'How to Play', 'https://ninefold.recica.dev/how-to-play'],
  ['/privacy', 'Privacy', 'https://ninefold.recica.dev/privacy'],
  ['/accessibility', 'Accessibility', 'https://ninefold.recica.dev/accessibility'],
] as const;

for (const [path, heading, canonical] of publicPages) {
  test(`${path} has useful SSR metadata and no serious accessibility violations`, async ({
    page,
    request,
  }) => {
    const response = await request.get(path);
    expect(await response.text()).toContain(heading);

    await page.goto(path);
    await expect(page.getByRole('heading', { level: 1, name: heading })).toBeVisible();
    await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', canonical);
    await expect(page.locator('meta[property="og:title"]')).toHaveCount(1);

    const accessibility = await new AxeBuilder({ page }).analyze();
    const severe = accessibility.violations.filter(
      ({ impact }) => impact === 'serious' || impact === 'critical',
    );
    expect(severe).toEqual([]);
  });
}

test('private routes are noindex and public discovery contains no private identifiers', async ({
  page,
  request,
}) => {
  for (const path of ['/create', '/join', '/solo', '/settings']) {
    await page.goto(path);
    await expect(page.locator('meta[name="robots"]')).toHaveAttribute(
      'content',
      /noindex, nofollow/,
    );
  }

  const sitemap = await (await request.get('/sitemap.xml')).text();
  expect([...sitemap.matchAll(/<loc>([^<]+)<\/loc>/g)].map((match) => match[1])).toEqual([
    'https://ninefold.recica.dev/',
    'https://ninefold.recica.dev/how-to-play',
    'https://ninefold.recica.dev/privacy',
    'https://ninefold.recica.dev/accessibility',
  ]);

  const robots = await (await request.get('/robots.txt')).text();
  for (const privatePath of [
    '/create',
    '/join',
    '/room',
    '/play',
    '/solo',
    '/replay',
    '/settings',
  ]) {
    expect(robots).toContain(`Disallow: ${privatePath}`);
  }
});

test('settings confirmation defaults to safety and restores keyboard focus', async ({ page }) => {
  await page.goto('/settings');
  const clearButton = page.getByRole('button', { name: 'Clear Local Data' });
  await clearButton.click();
  await expect(page.locator('#keep-local-data')).toBeFocused();
  await page.keyboard.press('Escape');
  await expect(clearButton).toBeFocused();

  const accessibility = await new AxeBuilder({ page }).analyze();
  expect(accessibility.violations).toEqual([]);
});

test('pseudo-localized compact pages reflow without horizontal scrolling', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 720 });
  for (const path of ['/', '/how-to-play', '/privacy', '/accessibility', '/settings']) {
    await page.goto(`${path}?locale=pseudo`);
    await expect(page.locator('html')).toHaveAttribute('data-locale', 'pseudo');
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
      `${path} should reflow at 320 CSS pixels`,
    ).toBe(true);
    await expect(page.locator('body')).toContainText('［');
  }
});
