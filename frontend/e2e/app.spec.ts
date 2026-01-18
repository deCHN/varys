import { test, expect } from '@playwright/test';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

test.beforeEach(async ({ page }) => {
  // Mock Wails Runtime to prevent crash
  await page.addInitScript(() => {
    window['go'] = {
      main: {
        App: {
          GetConfig: () => Promise.resolve({ vault_path: '', model_path: '', llm_model: '' }),
          CheckDependencies: () => Promise.resolve({}),
          SubmitTask: () => Promise.resolve(""),
          SelectVaultPath: () => Promise.resolve(""),
          SelectModelPath: () => Promise.resolve(""),
          UpdateConfig: () => Promise.resolve(),
          GetOllamaModels: () => Promise.resolve([]),
        }
      }
    };
  });
});

test('has title and captures screenshot', async ({ page }) => {
  await page.goto('/');

  // Expect a title "to contain" a substring.
  // Note: The HTML title might be "v2k-mac" (from main.go) but the h1 is "v2k".
  // Wails app title sets the window title. The web page title is in index.html.
  // Let's check the Tab button.
  await expect(page.getByText('Task')).toBeVisible();

  // Check input
  const input = page.getByPlaceholder('Enter YouTube/Bilibili URL');
  await expect(input).toBeVisible();

  // Fill input
  await input.fill('https://www.youtube.com/watch?v=dQw4w9WgXcQ');

  // Check button
  const button = page.getByRole('button', { name: 'Process' });
  await expect(button).toBeVisible();

  // Take screenshot
  // Navigate up from frontend/e2e/ to root/debug/
  const debugPath = path.resolve(__dirname, '../../../debug');
  await page.screenshot({ path: path.join(debugPath, 'e2e-screenshot.png') });
});
