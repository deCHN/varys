import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import App from './App';
import '@testing-library/jest-dom';

// Mock the Wails JS backend call
vi.mock('../wailsjs/go/main/App', () => ({
    SubmitTask: vi.fn(() => Promise.resolve("Mock Response")),
    GetConfig: vi.fn(() => Promise.resolve({ vault_path: '', model_path: '', llm_model: '', context_size: 8192 })),
    CheckDependencies: vi.fn(() => Promise.resolve({ yt_dlp: true, ffmpeg: true, whisper: true, ollama: true })),
    SelectVaultPath: vi.fn(() => Promise.resolve("/tmp/vault")),
    SelectModelPath: vi.fn(() => Promise.resolve("/tmp/model.bin")),
    UpdateConfig: vi.fn(() => Promise.resolve()),
    GetOllamaModels: vi.fn(() => Promise.resolve(["qwen3:8b"])),
    GetConfigPath: vi.fn(() => Promise.resolve("/tmp/config.json")),
    GetAppVersion: vi.fn(() => Promise.resolve("v0.3.4")),
    CancelTask: vi.fn(() => Promise.resolve()),
}));

// Mock Wails Runtime (for EventsOn)
vi.mock('../wailsjs/runtime', () => ({
    EventsOn: vi.fn(() => () => {}),
    WindowSetTitle: vi.fn(),
}));

describe('App Component', () => {
    it('renders the correct title and removes logo', () => {
        render(<App />);

        // Check Title (Now Tabs)
        const tabElement = screen.getByText('Task');
        expect(tabElement).toBeInTheDocument();

        // Check Logo is gone (by alt text)
        const logo = screen.queryByAltText('logo');
        expect(logo).not.toBeInTheDocument();
    });

    it('renders the input and process button correctly', () => {
        render(<App />);

        // Check Input
        const inputElement = screen.getByPlaceholderText('Enter YouTube/Bilibili URL');
        expect(inputElement).toBeInTheDocument();

        // Check Button exists (by title or role)
        const buttonElement = screen.getByTitle(/Start Processing/i);
        expect(buttonElement).toBeInTheDocument();
        
        // Ensure it contains an SVG icon
        const svg = buttonElement.querySelector('svg');
        expect(svg).toBeInTheDocument();
    });
});
