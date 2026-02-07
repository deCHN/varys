import { useCallback, useEffect, useState } from 'react';
import { GetStartupDiagnostics, OpenOllamaModelLibrary, ReadClipboardText, SelectModelPath, SelectVaultPath, StartOllamaService, StopOllamaService, UpdateModelPath, UpdateVaultPath } from '../wailsjs/go/main/App';
import { UpdateOpenAIKey } from '../wailsjs/go/main/App';
import { main } from '../wailsjs/go/models';
import Dashboard from './Dashboard';
import Settings from './Settings';
import StartupHealthWizard from './components/StartupHealthWizard';
import './App.css';

function App() {
    const [view, setView] = useState<'dashboard' | 'settings'>('dashboard');
    const [diagnostics, setDiagnostics] = useState<main.StartupDiagnostics | null>(null);
    const [wizardOpen, setWizardOpen] = useState(false);

    const refreshDiagnostics = useCallback(async () => {
        try {
            const result = await GetStartupDiagnostics();
            setDiagnostics(result);
            if (!result.ready) {
                setWizardOpen(true);
            }
        } catch (err) {
            console.error('Failed to load startup diagnostics', err);
        }
    }, []);

    useEffect(() => {
        refreshDiagnostics();
    }, [refreshDiagnostics]);

    return (
        <div className="flex flex-col h-screen bg-slate-900 text-slate-100 font-sans">
            {/* Top Navigation Bar */}
            <div className="flex items-center justify-center gap-6 py-4 border-b border-slate-800 bg-slate-900/50 backdrop-blur-sm sticky top-0 z-10">
                <button
                    className={`px-4 py-2 text-sm font-medium rounded-full transition-all duration-200 ${
                        view === 'dashboard'
                        ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20'
                        : 'text-slate-400 hover:text-white hover:bg-slate-800'
                    }`}
                    onClick={() => setView('dashboard')}
                >
                    Task
                </button>
                <button
                    className={`px-4 py-2 text-sm font-medium rounded-full transition-all duration-200 ${
                        view === 'settings'
                        ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20'
                        : 'text-slate-400 hover:text-white hover:bg-slate-800'
                    }`}
                    onClick={() => setView('settings')}
                >
                    Settings
                </button>
            </div>

            {/* Main Content Area */}
            <main className="flex-1 overflow-hidden flex flex-col relative">
                <div className={`absolute inset-0 overflow-y-auto ${view === 'dashboard' ? 'block' : 'hidden'}`}>
                    <Dashboard
                        onPreflightFailed={(diag) => {
                            setDiagnostics(diag);
                            setWizardOpen(true);
                        }}
                    />
                </div>
                <div className={`absolute inset-0 overflow-y-auto ${view === 'settings' ? 'block' : 'hidden'}`}>
                    <Settings
                        isActive={view === 'settings'}
                        onOpenDependencyHealth={async () => {
                            await refreshDiagnostics();
                            setWizardOpen(true);
                        }}
                    />
                </div>
            </main>

            <StartupHealthWizard
                diagnostics={diagnostics}
                open={wizardOpen}
                onClose={() => setWizardOpen(false)}
                onRecheck={refreshDiagnostics}
                onStartOllama={async () => {
                    try {
                        await StartOllamaService();
                    } catch (err) {
                        console.error('Failed to start ollama service', err);
                    }
                    await refreshDiagnostics();
                }}
                onStopOllama={async () => {
                    try {
                        await StopOllamaService();
                    } catch (err) {
                        console.error('Failed to stop ollama service', err);
                    }
                    await refreshDiagnostics();
                }}
                onOpenSettings={() => {
                    setView('settings');
                    setWizardOpen(false);
                    setTimeout(() => {
                        window.dispatchEvent(new Event('open-system-check'));
                    }, 60);
                }}
                onOpenModelLibrary={async () => {
                    try {
                        await OpenOllamaModelLibrary();
                    } catch (err) {
                        console.error('Failed to open ollama model library', err);
                    }
                }}
                onBrowseVaultPath={async () => {
                    try {
                        const selected = await SelectVaultPath();
                        if (!selected) {
                            return;
                        }
                        await UpdateVaultPath(selected);
                    } catch (err) {
                        console.error('Failed to update vault path', err);
                    }
                    await refreshDiagnostics();
                }}
                onBrowseModelPath={async () => {
                    try {
                        const selected = await SelectModelPath();
                        if (!selected) {
                            return;
                        }
                        await UpdateModelPath(selected);
                    } catch (err) {
                        console.error('Failed to update model path', err);
                    }
                    await refreshDiagnostics();
                }}
                onPasteOpenAIKey={async (inputKey?: string) => {
                    try {
                        const key = (inputKey ?? (await ReadClipboardText())).trim();
                        if (!key) {
                            return "";
                        }
                        await UpdateOpenAIKey(key);
                        await refreshDiagnostics();
                        return key;
                    } catch (err) {
                        console.error('Failed to update OpenAI key', err);
                        return "";
                    }
                }}
            />
        </div>
    )
}
export default App;
