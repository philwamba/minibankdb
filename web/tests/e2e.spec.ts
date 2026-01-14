import { test, expect } from '@playwright/test';

test('has title', async ({ page }) => {
  await page.goto('http://localhost:3000');

  await expect(page).toHaveTitle(/MiniBankDB/);
});

test('can run query', async ({ page }) => {
  await page.goto('http://localhost:3000');

  await page.getByRole('button', { name: 'Run Query' }).click();

  await expect(page.getByText('Query Results')).toBeVisible();
});
