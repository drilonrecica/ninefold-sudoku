import { expect, test } from 'playwright/test';

test('Solo starts online, persists locally, resumes, and requests a reveal', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('link', { name: 'Play Solo' }).click();
  await expect(page.getByRole('heading', { name: 'Play Solo' })).toBeVisible();

  const guided = page
    .locator('section.panel')
    .filter({ has: page.getByRole('heading', { name: 'Guided' }) });
  await guided.getByRole('button', { name: 'Easy' }).click();
  await expect(page.getByRole('heading', { name: 'Solo Sudoku' })).toBeVisible();

  const editable = page.getByRole('gridcell', { name: /editable, empty/ }).first();
  await editable.focus();
  await page.keyboard.press('1');
  await expect(
    page.getByText(/No incorrect values found|Some values need another look/),
  ).toBeVisible();

  await page.reload();
  await page.getByRole('button', { name: /Continue Easy puzzle/ }).click();
  await expect(page.getByRole('heading', { name: 'Solo Sudoku' })).toBeVisible();
  await page.getByRole('gridcell', { name: /editable, value 1/ }).focus();
  await page.keyboard.press('Delete');
  await page.getByRole('button', { name: 'Reveal' }).click();
  await expect(page.getByText(/Revealed row/)).toBeVisible();

  const registrations = await page.evaluate(async () =>
    navigator.serviceWorker?.getRegistrations(),
  );
  expect(registrations ?? []).toHaveLength(0);
});
