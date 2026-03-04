import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import App from './App';
import '@testing-library/jest-dom';

const appMocks = vi.hoisted(() => ({
    SubmitTask: vi.fn(() => Promise.resolve("Mock Response")),
    GetConfig: vi.fn(() => Promise.resolve({ 
        vault_path: '', 
        model_path: '', 
        llm_model: 'qwen3:8b', 
        translation_model: 'qwen3:0.6b',
        target_language: 'English',
        context_size: 8192,
        custom_prompt: '',
        ai_provider: 'ollama',
        openai_model: 'gpt-4o',
        openai_key: ''
    })),
    CheckDependencies: vi.fn(() => Promise.resolve({ yt_dlp: true, ffmpeg: true, whisper: true, ollama: true })),
    GetStartupDiagnostics: vi.fn(() => Promise.resolve({ generated_at: '', provider: 'ollama', blockers: [] as string[], ready: true, items: [] as any[] })),
    SelectVaultPath: vi.fn(() => Promise.resolve("/tmp/vault")),
    SelectModelPath: vi.fn(() => Promise.resolve("/tmp/model.bin")),
    StartOllamaService: vi.fn(() => Promise.resolve("ok")),
    StopOllamaService: vi.fn(() => Promise.resolve("ok")),
    UpdateVaultPath: vi.fn(() => Promise.resolve()),
    UpdateModelPath: vi.fn(() => Promise.resolve()),
    UpdateConfig: vi.fn(() => Promise.resolve()),
    GetAIModels: vi.fn(() => Promise.resolve(["qwen3:8b"])),
    GetConfigPath: vi.fn(() => Promise.resolve("/tmp/config.json")),
    GetAppVersion: vi.fn(() => Promise.resolve("v0.4.3")),
    ReadClipboardText: vi.fn(() => Promise.resolve("sk-test-clipboard-key-12345678")),
    CancelTask: vi.fn(() => Promise.resolve()),
    OpenOllamaModelLibrary: vi.fn(() => Promise.resolve("ok")),
    GetDefaultPrompt: vi.fn(() => Promise.resolve("Mock Default Prompt")),
    LocateConfigFile: vi.fn(() => Promise.resolve()),
}));

// Mock the Wails JS backend call
vi.mock('../wailsjs/go/app/App', () => ({
    SubmitTask: appMocks.SubmitTask,
    GetConfig: appMocks.GetConfig,
    CheckDependencies: appMocks.CheckDependencies,
    GetStartupDiagnostics: appMocks.GetStartupDiagnostics,
    SelectVaultPath: appMocks.SelectVaultPath,
    SelectModelPath: appMocks.SelectModelPath,
    StartOllamaService: appMocks.StartOllamaService,
    StopOllamaService: appMocks.StopOllamaService,
    UpdateVaultPath: appMocks.UpdateVaultPath,
    UpdateModelPath: appMocks.UpdateModelPath,
    UpdateConfig: appMocks.UpdateConfig,
    GetAIModels: appMocks.GetAIModels,
    GetConfigPath: appMocks.GetConfigPath,
    GetAppVersion: appMocks.GetAppVersion,
    ReadClipboardText: appMocks.ReadClipboardText,
    CancelTask: appMocks.CancelTask,
    OpenOllamaModelLibrary: appMocks.OpenOllamaModelLibrary,
    GetDefaultPrompt: appMocks.GetDefaultPrompt,
    LocateConfigFile: appMocks.LocateConfigFile,
}));

// Mock Wails Runtime (for EventsOn)
vi.mock('../wailsjs/runtime', () => ({
    EventsOn: vi.fn(() => () => {}),
    WindowSetTitle: vi.fn(),
}));

describe('App Component', () => {
    it('renders the correct title and navigation', async () => {
        render(<App />);

        // Wait for rendering to settle
        const settingsButton = await screen.findByTitle(/Settings/i);
        expect(settingsButton).toBeInTheDocument();

        // Check Logo is present in the header (by alt text)
        const logo = screen.getByAltText('Logo');
        expect(logo).toBeInTheDocument();
    });

    it('renders the input and process button correctly', async () => {
        render(<App />);

        // Wait for the input to appear
        const inputElement = await screen.findByPlaceholderText(/Enter YouTube\/Bilibili URL/i);
        expect(inputElement).toBeInTheDocument();

        // Check Button exists
        const buttonElement = screen.getByTitle(/Start Processing/i);
        expect(buttonElement).toBeInTheDocument();
    });
});
