import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    let ollamaStatus = "misconfigured";
    window['go'] = {
      app: {
        App: {
          GetConfig: () => Promise.resolve({ ai_provider: 'ollama' }),
          GetConfigPath: () => Promise.resolve('/fake/path/config.json'),
          GetAppVersion: () => Promise.resolve('v0.3.8'),
          GetStartupDiagnostics: () => Promise.resolve({
            ready: false,
            provider: 'ollama',
            items: [
              {
                id: 'ollama',
                name: 'ollama',
                status: ollamaStatus,
                can_auto_fix: true,
                fix_suggestion: 'Ollama is stopped'
              }
            ]
          }),
          StartOllamaService: () => {
            ollamaStatus = "ok";
            return Promise.resolve("started");
          },
          StopOllamaService: () => {
            ollamaStatus = "misconfigured";
            return Promise.resolve("stopped");
          },
          GetAIModels: () => Promise.resolve(['qwen3:8b']),
        }
      }
    };
  });
});

test('Ollama switcher toggles service state', async ({ page }) => {
  await page.goto('/');

  // Navigate to Settings
  await page.getByTitle('Settings').click();

  // Verify Ollama item exists and shows Stopped
  const ollamaRow = page.locator('div').filter({ hasText: /^ollama$/ }).first();
  await expect(ollamaRow).toBeVisible();
  await expect(page.getByText('STOPPED')).toBeVisible();

  // Click Toggle
  const toggle = page.getByRole('button').filter({ has: page.locator('span.bg-white') });
  await toggle.click();

  // Verify UI shows Running
  // Note: The UI updates based on diagnostics refresh
  await expect(page.getByText('RUNNING')).toBeVisible();

  // Toggle off
  await toggle.click();

  // Verify UI shows Stopped again
  await expect(page.getByText('STOPPED')).toBeVisible();
});
