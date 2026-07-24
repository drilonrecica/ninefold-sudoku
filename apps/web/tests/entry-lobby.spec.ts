import AxeBuilder from '@axe-core/playwright';
import { expect, test } from 'playwright/test';

test('SSR home is useful, keyboard reachable, private, and accessible', async ({ browser }) => {
  const context = await browser.newContext({ javaScriptEnabled: false });
  const page = await context.newPage();
  const externalRequests: string[] = [];
  page.on('request', (request) => {
    const requestURL = new URL(request.url());
    if (requestURL.hostname !== '127.0.0.1') externalRequests.push(request.url());
  });

  await page.goto('/');
  await expect(
    page.getByRole('heading', { level: 1, name: 'Sudoku is better together.' }),
  ).toBeVisible();
  await expect(page.getByRole('link', { name: 'Create a room' })).toBeVisible();
  await expect(page.getByText('No accounts. No ads. No tracking.', { exact: true })).toBeVisible();
  await expect(page.getByText(/Play Solo/i)).toHaveCount(0);
  expect(externalRequests).toEqual([]);
  await context.close();
});

test('two isolated sessions create, join, and converge in the lobby', async ({ browser }) => {
  const hostContext = await browser.newContext();
  const guestContext = await browser.newContext();
  const host = await hostContext.newPage();
  const guest = await guestContext.newPage();

  await host.goto('/create');
  await host.getByLabel('Temporary display name').fill('Host Player');
  await host.getByLabel('Hard').check();
  await host.getByRole('button', { name: 'Create Room' }).click();
  await expect(host).toHaveURL(/\/room\/[A-Z2-9]{6}$/);
  const roomCode = host.url().split('/').pop()!;
  await expect(host.getByText('Host Player joined the room.')).toBeVisible();

  await guest.goto(`/join/${roomCode}`);
  await expect(guest.getByLabel('Room preview')).not.toContainText('Host Player');
  await guest.getByLabel('Temporary display name').fill('Guest Player');
  await guest.getByRole('button', { name: 'Join Room' }).click();
  await expect(guest).toHaveURL(`/room/${roomCode}`);

  await expect(host.getByRole('list').getByText('Guest Player', { exact: true })).toBeVisible();
  await expect(guest.getByRole('list').getByText('Host Player', { exact: true })).toBeVisible();
  await expect(host.getByRole('status').filter({ hasText: 'Connected' })).toBeVisible();

  await guest.getByRole('button', { name: 'Ready' }).click();
  await expect(
    host.getByRole('listitem').filter({ hasText: 'Guest Player' }).first(),
  ).toContainText('Ready');
  await host.getByRole('button', { name: 'Ready' }).click();
  await expect(host.getByRole('button', { name: 'Start Match' })).toBeEnabled();

  const accessibility = await new AxeBuilder({ page: host }).analyze();
  expect(accessibility.violations).toEqual([]);

  await host.goto('/create');
  await host.getByLabel('Temporary display name').fill('Replacement Host');
  await host.getByRole('button', { name: 'Create Room' }).click();
  await expect(host.getByRole('alert')).toContainText('Leave your current room');
  await expect(host.getByRole('link', { name: 'Return to current room' })).toHaveAttribute(
    'href',
    `/room/${roomCode}`,
  );
  await expect(host.getByRole('button', { name: 'Leave and create new room' })).toBeVisible();

  await hostContext.close();
  await guestContext.close();
});

test('compact, zoomed, dark, and reduced-motion lobby remains usable', async ({ browser }) => {
  const context = await browser.newContext({
    viewport: { width: 320, height: 720 },
    colorScheme: 'dark',
    reducedMotion: 'reduce',
  });
  const page = await context.newPage();
  await page.goto('/create');

  await page.keyboard.press('Tab');
  await expect(page.getByRole('link', { name: 'Skip to main content' })).toBeFocused();
  await page.getByLabel('Temporary display name').fill('Compact Host');
  await page.getByRole('button', { name: 'Create Room' }).click();
  await expect(page).toHaveURL(/\/room\/[A-Z2-9]{6}$/);

  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(
    true,
  );
  await page.setViewportSize({ width: 640, height: 720 });
  await page.evaluate(() => {
    document.documentElement.style.fontSize = '200%';
  });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(
    true,
  );
  await expect(page.getByRole('button', { name: 'Copy invitation link' })).toBeVisible();

  await context.close();
});
