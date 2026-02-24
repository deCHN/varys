import { useCallback, useEffect, useState } from 'react';
import { GetStartupDiagnostics, OpenOllamaModelLibrary, ReadClipboardText, SelectModelPath, SelectVaultPath, StartOllamaService, StopOllamaService, UpdateModelPath, UpdateVaultPath } from '../wailsjs/go/app/App';
import { UpdateOpenAIKey } from '../wailsjs/go/app/App';
import { app } from '../wailsjs/go/models';
import Dashboard from './Dashboard';
import Settings from './Settings';
import StartupHealthWizard from './components/StartupHealthWizard';
import UpdateNotifier from './components/UpdateNotifier';
import logo from './assets/images/varys_logo.png';
import varysBg from './assets/images/varys.png';
import { GetAppVersion } from '../wailsjs/go/app/App';
import './App.css';

function App() {
    const [view, setView] = useState<'dashboard' | 'settings'>('dashboard');
    const [diagnostics, setDiagnostics] = useState<app.StartupDiagnostics | null>(null);
    const [wizardOpen, setWizardOpen] = useState(false);
    const [version, setVersion] = useState<string>('');
    const [aboutOpen, setAboutOpen] = useState(false);

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
        GetAppVersion().then(setVersion).catch(console.error);
    }, [refreshDiagnostics]);

    return (
        <div className="flex flex-col h-screen bg-varys-bg text-slate-100 font-sans selection:bg-varys-primary/30 relative">
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
                        version={version}
                        onAboutClick={() => setAboutOpen(true)}
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

            {/* About Modal */}
            {aboutOpen && (
                <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 animate-in fade-in duration-300">
                    <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={() => setAboutOpen(false)} />
                    
                    <div className="relative w-[480px] aspect-square bg-varys-bg border border-varys-border/30 rounded-3xl overflow-hidden shadow-[0_0_50px_rgba(0,0,0,0.5)] animate-in zoom-in-95 duration-300">
                        {/* varys.png Background - More obvious */}
                        <div 
                            className="absolute inset-0 opacity-40 pointer-events-none"
                            style={{ 
                                backgroundImage: `url(${varysBg})`,
                                backgroundPosition: 'center',
                                backgroundRepeat: 'no-repeat',
                                backgroundSize: 'cover'
                            }}
                        />

                        {/* Content Overlay to ensure readability */}
                        <div className="relative h-full p-10 flex flex-col items-center justify-center text-center bg-gradient-to-b from-varys-bg/20 via-varys-bg/60 to-varys-bg/90">
                            <button 
                                onClick={() => setAboutOpen(false)}
                                className="absolute top-6 right-6 p-2 text-white/50 hover:text-white transition-colors z-20"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                            </button>

                            <div className="z-10 mt-auto">
                                <h2 className="text-4xl font-black text-white mb-2 tracking-tighter drop-shadow-2xl">Varys</h2>
                                <p className="text-varys-secondary text-sm font-black tracking-[0.2em] uppercase mb-8 drop-shadow-md">Version {version}</p>
                                
                                <div className="space-y-6 max-w-xs">
                                    <p className="text-slate-100 text-sm leading-relaxed font-bold drop-shadow-lg">
                                        Private, offline multimedia intelligence for your second brain.
                                    </p>
                                    
                                    <div className="flex flex-col items-center gap-4">
                                        <a 
                                            href="https://github.com/dechn/varys" 
                                            target="_blank" 
                                            rel="noopener noreferrer"
                                            className="text-xs text-blue-400 hover:text-blue-300 flex items-center gap-2 transition-colors font-black bg-black/40 px-4 py-2 rounded-full border border-white/5"
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"></path></svg>
                                            GITHUB
                                        </a>
                                    </div>
                                </div>
                            </div>

                            <button
                                onClick={() => setAboutOpen(false)}
                                className="mt-auto mb-4 px-10 py-3 bg-varys-primary hover:bg-varys-primary/80 text-white rounded-2xl text-sm font-black shadow-2xl shadow-varys-primary/40 transition-all active:scale-95 z-10 border border-white/10"
                            >
                                CLOSE
                            </button>
                        </div>
                    </div>
                </div>
            )}

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
