import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  // Mock Wails Runtime
  await page.addInitScript(() => {
    window['go'] = {
      main: {
        App: {
          GetConfig: () => Promise.resolve({ vault_path: '/tmp/vault', model_path: '/tmp/model.bin', llm_model: 'qwen' }),
          CheckDependencies: () => Promise.resolve({ yt_dlp: true, ffmpeg: true, whisper: true, ollama: true }),
          SubmitTask: (url) => {
             return new Promise(resolve => setTimeout(() => resolve("Saved to: /tmp/note.md"), 500));
          },
          SelectVaultPath: () => Promise.resolve("/tmp/vault"),
          SelectModelPath: () => Promise.resolve("/tmp/model.bin"),
          UpdateConfig: () => Promise.resolve(),
          GetOllamaModels: () => Promise.resolve(["qwen2.5:7b"]),
        }
      }
    };
  });
});

test('full workflow: settings check and task processing', async ({ page }) => {
  test.setTimeout(30000); // 30s is enough for mocked backend

  // 1. Settings Check
  await page.goto('/');
  await page.getByText('Settings').click();
  
  // Verify mocked dependencies
  await expect(page.locator('.section ul')).toContainText('yt-dlp: ✅');
  await expect(page.locator('.section ul')).toContainText('ffmpeg: ✅');
  
  // 2. Dashboard Processing
  await page.getByText('Dashboard').click();
  
  const input = page.getByPlaceholder('Enter YouTube/Bilibili URL');
  await input.fill('https://www.youtube.com/watch?v=6ESe2jw0fks');
  
  await page.getByRole('button', { name: 'Process' }).click();
  
  // 3. Verify Log Updates
  await expect(page.locator('.console-log')).toContainText('Processing URL');
  await expect(page.locator('.console-log')).toContainText('Backend Response: Saved to: /tmp/note.md');
  
  // Final status
  await expect(page.locator('.result')).toHaveText('Status: Task completed');
});