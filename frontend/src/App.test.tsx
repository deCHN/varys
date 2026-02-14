import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import App from './App';
import '@testing-library/jest-dom';

const appMocks = vi.hoisted(() => ({
    SubmitTask: vi.fn(() => Promise.resolve("Mock Response")),
    GetConfig: vi.fn(() => Promise.resolve({ vault_path: '', model_path: '', llm_model: '', context_size: 8192 })),
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
    GetAppVersion: vi.fn(() => Promise.resolve("v0.3.4")),
    ReadClipboardText: vi.fn(() => Promise.resolve("sk-test-clipboard-key-12345678")),
    CancelTask: vi.fn(() => Promise.resolve()),
    OpenOllamaModelLibrary: vi.fn(() => Promise.resolve("ok")),
}));

// Mock the Wails JS backend call
vi.mock('../wailsjs/go/main/App', () => ({
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
}));

// Mock Wails Runtime (for EventsOn)
vi.mock('../wailsjs/runtime', () => ({
    EventsOn: vi.fn(() => () => {}),
    WindowSetTitle: vi.fn(),
}));

describe('App Component', () => {
    it('calls UpdateVaultPath and UpdateModelPath after browse actions in dependency health', async () => {
        appMocks.GetStartupDiagnostics.mockResolvedValue({
            generated_at: '',
            provider: 'ollama',
            blockers: ['vault_path', 'model_path'],
            ready: false,
            items: [
                {
                    id: 'vault_path',
                    name: 'Vault Path',
                    status: 'misconfigured',
                    required_for: ['export'],
                    detected_path: '',
                    fix_suggestion: 'fix',
                    fix_commands: [],
                    can_auto_fix: false,
                    is_blocker: true
                },
                {
                    id: 'model_path',
                    name: 'Whisper Model Path',
                    status: 'misconfigured',
                    required_for: ['transcribe'],
                    detected_path: '',
                    fix_suggestion: 'fix',
                    fix_commands: [],
                    can_auto_fix: false,
                    is_blocker: true
                }
            ]
        } as any);
        appMocks.SelectVaultPath.mockResolvedValue('/tmp/new-vault');
        appMocks.SelectModelPath.mockResolvedValue('/tmp/new-model.bin');

        render(<App />);

        const vaultButton = await screen.findByText('Browse Vault Path');
        const modelButton = await screen.findByText('Browse Whisper Model');

        fireEvent.click(vaultButton);
        fireEvent.click(modelButton);

        await waitFor(() => {
            expect(appMocks.UpdateVaultPath).toHaveBeenCalledWith('/tmp/new-vault');
            expect(appMocks.UpdateModelPath).toHaveBeenCalledWith('/tmp/new-model.bin');
        });
    });

    it('renders the correct title and removes logo', async () => {
        render(<App />);

        // Use findBy to wait for initial render and background updates to settle
        const tabElement = await screen.findByText('Task');
        expect(tabElement).toBeInTheDocument();

        // Check Logo is gone (by alt text)
        const logo = screen.queryByAltText('logo');
        expect(logo).not.toBeInTheDocument();
    });

    it('renders the input and process button correctly', async () => {
        render(<App />);

        // Wait for the input to appear, which also allows background effects to run
        const inputElement = await screen.findByPlaceholderText('Enter YouTube/Bilibili URL');
        expect(inputElement).toBeInTheDocument();

        // Check Button exists (by title or role)
        const buttonElement = screen.getByTitle(/Start Processing/i);
        expect(buttonElement).toBeInTheDocument();
        
        // Ensure it contains an SVG icon
        const svg = buttonElement.querySelector('svg');
        expect(svg).toBeInTheDocument();
    });
});
