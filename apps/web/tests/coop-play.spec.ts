import AxeBuilder from '@axe-core/playwright';
import { expect, test } from 'playwright/test';

test('two players complete the keyboard and touch Co-op flow and converge', async ({ browser }) => {
  const hostContext = await browser.newContext({ viewport: { width: 1280, height: 900 } });
  const guestContext = await browser.newContext({ viewport: { width: 375, height: 812 } });
  const host = await hostContext.newPage();
  const guest = await guestContext.newPage();
  const frames: string[] = [];
  host.on('websocket', (socket) => {
    socket.on('framereceived', (event) => frames.push(String(event.payload)));
  });

  await host.goto('/create');
  await host.getByLabel('Temporary display name').fill('Board Host');
  await host.getByLabel('Easy').check();
  await host.getByRole('button', { name: 'Create Room' }).click();
  await expect(host).toHaveURL(/\/room\/[A-Z2-9]{6}$/);
  const roomCode = host.url().split('/').pop()!;

  await guest.goto(`/join/${roomCode}`);
  await guest.getByLabel('Temporary display name').fill('Board Guest');
  await guest.getByRole('button', { name: 'Join Room' }).click();
  await expect(guest).toHaveURL(`/room/${roomCode}`);
  await expect(host.getByRole('list').getByText('Board Guest', { exact: true })).toBeVisible();

  await guest.getByRole('button', { name: 'Ready' }).click();
  await expect(host.getByRole('listitem').filter({ hasText: 'Board Guest' }).first()).toContainText(
    'Ready',
  );
  await host.getByRole('button', { name: 'Ready' }).click();
  await expect(host.getByRole('button', { name: 'Start Match' })).toBeEnabled();
  await host.getByRole('button', { name: 'Start Match' }).click();

  await expect(host).toHaveURL(/\/play\/[0-9a-f-]+$/, { timeout: 10_000 });
  await expect(guest).toHaveURL(/\/play\/[0-9a-f-]+$/, { timeout: 10_000 });
  await expect(host.getByRole('gridcell')).toHaveCount(81);
  await expect(guest.getByRole('gridcell')).toHaveCount(81);
  expect(frames.join('')).not.toContain('"solution"');

  const hostEditable = host.getByRole('gridcell', { name: /editable, empty/ }).first();
  const cellId = await hostEditable.getAttribute('id');
  expect(cellId).toBeTruthy();
  const guestEditable = guest.locator(`#${cellId}`);
  await hostEditable.click();
  await expect(guestEditable).toHaveAttribute('aria-label', /Board Host is working here/);
  await guestEditable.click();
  const override = guest.getByRole('button', { name: 'Use anyway' });
  await expect(override).toBeVisible();
  await override.click();

  await Promise.all([host.keyboard.press('1'), guest.keyboard.press('2')]);
  await expect(host.locator(`#${cellId}`)).toHaveAttribute('aria-label', /value [12]/);
  await expect(guest.locator(`#${cellId}`)).toHaveAttribute('aria-label', /value [12]/);
  const hostValue = (await host.locator(`#${cellId}`).getAttribute('aria-label'))?.match(
    /value (\d)/,
  )?.[1];
  const guestValue = (await guest.locator(`#${cellId}`).getAttribute('aria-label'))?.match(
    /value (\d)/,
  )?.[1];
  expect(guestValue).toBe(hostValue);

  await host.locator(`#${cellId}`).focus();
  await host.keyboard.press('Delete');
  await expect(host.locator(`#${cellId}`)).toHaveAttribute('aria-label', /editable, empty/);
  await expect(guest.locator(`#${cellId}`)).toHaveAttribute('aria-label', /editable, empty/);

  const noteCell = guest.getByRole('gridcell', { name: /editable, empty/ }).nth(1);
  const noteCellId = await noteCell.getAttribute('id');
  await noteCell.click();
  await guest.getByRole('button', { name: /Notes N/ }).click();
  await guest
    .getByRole('region', { name: 'Number input' })
    .getByRole('button', { name: '3' })
    .click();
  await expect(guest.locator(`#${noteCellId}`)).toHaveAttribute('aria-label', /shared notes 3/);
  await expect(host.locator(`#${noteCellId}`)).toHaveAttribute('aria-label', /shared notes 3/);

  await host.reload();
  await expect(host.getByRole('gridcell')).toHaveCount(81);
  await expect(host.locator(`#${noteCellId}`)).toHaveAttribute('aria-label', /shared notes 3/);

  await hostContext.setOffline(true);
  await expect(host.getByText(/Connection lost\. Reconnecting/)).toBeVisible();
  await expect(host.getByRole('region', { name: 'Number input' })).toBeVisible();
  await hostContext.setOffline(false);
  await expect(host.getByText('Connected and synchronized.')).toBeAttached({ timeout: 10_000 });

  const observer = await hostContext.newPage();
  await observer.goto(host.url());
  await expect(
    observer.getByRole('status').getByText('This Room is active in another tab.'),
  ).toBeVisible();
  await observer.getByRole('button', { name: 'Control from this tab' }).click();
  await expect(
    host.getByRole('status').getByText('This Room is active in another tab.'),
  ).toBeVisible();
  await host.getByRole('button', { name: 'Control from this tab' }).click();
  await expect(
    observer.getByRole('status').getByText('This Room is active in another tab.'),
  ).toBeVisible();
  await observer.close();

  await guest.getByRole('button', { name: 'Look here' }).click();
  await guest.getByRole('button', { name: 'Nice move' }).click();
  await expect(host.getByText('Board Guest: nice move')).toBeVisible();

  const accessibility = await new AxeBuilder({ page: host }).analyze();
  expect(accessibility.violations).toEqual([]);
  expect(
    await guest.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
  ).toBe(true);
  await guest.setViewportSize({ width: 640, height: 812 });
  await guest.evaluate(() => {
    document.documentElement.style.fontSize = '200%';
  });
  expect(
    await guest.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
  ).toBe(true);
  const zoomedCell = await guest.getByRole('gridcell').first().boundingBox();
  expect(zoomedCell?.width).toBeGreaterThanOrEqual(24);
  await guest.evaluate(() => {
    document.documentElement.style.fontSize = '';
  });

  const reveal = host.getByRole('button', { name: 'Reveal' });
  for (let attempt = 0; attempt < 81; attempt++) {
    if (await host.getByText('Puzzle completed.', { exact: true }).count()) break;
    const filledBefore = await host.getByRole('gridcell', { name: /value \d/ }).count();
    await expect(reveal).toBeEnabled();
    await reveal.click();
    await expect
      .poll(() => host.getByRole('gridcell', { name: /value \d/ }).count())
      .toBeGreaterThan(filledBefore);
  }
  await expect(host.getByText('Puzzle completed.', { exact: true })).toBeAttached();
  await expect(guest.getByText('Puzzle completed.', { exact: true })).toBeAttached();

  await hostContext.close();
  await guestContext.close();
});
