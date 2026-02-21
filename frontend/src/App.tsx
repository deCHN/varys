import { useCallback, useEffect, useState } from 'react';
import { GetStartupDiagnostics, OpenOllamaModelLibrary, ReadClipboardText, SelectModelPath, SelectVaultPath, StartOllamaService, StopOllamaService, UpdateModelPath, UpdateVaultPath } from '../wailsjs/go/main/App';
import { UpdateOpenAIKey } from '../wailsjs/go/main/App';
import { main } from '../wailsjs/go/models';
import Dashboard from './Dashboard';
import Settings from './Settings';
import StartupHealthWizard from './components/StartupHealthWizard';
import UpdateNotifier from './components/UpdateNotifier';
import logo from './assets/images/varys_logo.png';
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
        <div className="flex flex-col h-screen bg-varys-bg text-slate-100 font-sans selection:bg-varys-primary/30">
            <UpdateNotifier />
            {/* Top Navigation Bar */}
            <header className="flex items-center justify-between px-8 py-4 border-b border-varys-border/10 bg-varys-bg/80 backdrop-blur-md sticky top-0 z-10">
                <div className="flex-1">
                    {/* Left side empty for macOS traffic lights */}
                </div>
                
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => setView(view === 'dashboard' ? 'settings' : 'dashboard')}
                        className={`p-1.5 rounded-xl transition-all group relative flex items-center justify-center ${
                            view === 'settings' 
                            ? 'bg-varys-primary/20 ring-1 ring-varys-primary/50' 
                            : 'hover:bg-white/5'
                        }`}
                        title="Settings"
                    >
                        {/* The Gear Icon - Custom size to fit logo in center */}
                        <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className={`text-slate-500 group-hover:text-varys-primary group-hover:rotate-90 transition-all duration-700 ${view === 'settings' ? 'rotate-90 text-varys-primary' : ''}`}>
                            <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/>
                        </svg>
                        
                        {/* The Small Logo inside the Gear */}
                        <img 
                            src={logo} 
                            alt="Logo" 
                            className={`absolute w-3.5 h-3.5 rounded-full shadow-sm transition-transform duration-700 ${view === 'settings' ? 'scale-110' : 'group-hover:scale-110'}`} 
                        />
                    </button>
                </div>
            </header>

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
                        onSaved={() => setView('dashboard')}
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
