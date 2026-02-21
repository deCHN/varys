import { useState, useEffect, useRef } from 'react';
import { GetConfig, UpdateConfig, SelectVaultPath, SelectModelPath, GetAIModels, GetConfigPath, GetAppVersion, GetStartupDiagnostics } from "../wailsjs/go/main/App";
import { main } from "../wailsjs/go/models";
import HealthStatusBadge from "./components/health/HealthStatusBadge";
import HealthItemRow from "./components/health/HealthItemRow";

interface Config {
    vault_path: string;
    model_path: string;
    llm_model: string;
    translation_model: string;
    target_language: string;
    context_size: number;
    custom_prompt: string;
    ai_provider: string;
    openai_model: string;
    openai_key: string;
}

interface SettingsProps {
    isActive?: boolean;
    onSaved?: () => void;
    onOpenDependencyHealth?: () => Promise<void> | void;
}

export default function Settings(props: SettingsProps) {
    const [cfg, setCfg] = useState<Config>({ 
        vault_path: '', 
        model_path: '', 
        llm_model: '', 
        translation_model: 'qwen3:0.6b',
        target_language: '', 
        context_size: 8192,
        custom_prompt: '',
        ai_provider: 'ollama',
        openai_model: 'gpt-4o',
        openai_key: ''
    });
    const [diagnostics, setDiagnostics] = useState<main.StartupDiagnostics | null>(null);
    const [aiModels, setAIModels] = useState<string[]>([]);
    const [configPath, setConfigPath] = useState<string>('');
    const [version, setVersion] = useState<string>('');
    const [status, setStatus] = useState<{msg: string, type: 'success' | 'error' | ''}>({msg: '', type: ''});
    const systemCheckRef = useRef<HTMLDivElement>(null);

    const defaultPrompt = `You are an expert content analyst.
Task: Analyze the following text and provide a structured analysis in [Target Language].

Rules:
1. OUTPUT MUST BE IN [Target Language].
2. If the input text is in English, TRANSLATE your analysis to [Target Language].
3. Tags must be single words or hyphenated (no spaces).

Format: Return ONLY a valid JSON object with the following structure:
{
  "summary": "Concise summary of the content",
  "key_points": ["Point 1", "Point 2", "Point 3"],
  "tags": ["Tag1", "Tag2", "Tag3"],
  "assessment": {
    "authenticity": "Rating/Comment",
    "effectiveness": "Rating/Comment",
    "timeliness": "Rating/Comment",
    "alternatives": "Rating/Comment"
  }
}`;

    const languages = [
        "Simplified Chinese",
        "Traditional Chinese",
        "English",
        "Japanese",
        "Spanish",
        "French",
        "German",
        "Korean",
        "Russian"
    ];

    const refreshDiagnostics = () => {
        GetStartupDiagnostics()
            .then(setDiagnostics)
            .catch(err => {
                console.error("Failed to load startup diagnostics", err);
            });
    };

    useEffect(() => {
        GetConfig().then((c: any) => setCfg(c));
        refreshDiagnostics();
        GetConfigPath().then(setConfigPath);
        GetAppVersion().then(setVersion);
    }, []);

    useEffect(() => {
        if (props.isActive) {
            refreshDiagnostics();
        }
    }, [props.isActive]);

    useEffect(() => {
        const onOpenSystemCheck = () => {
            refreshDiagnostics();
            setTimeout(() => {
                systemCheckRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' });
            }, 50);
        };

        window.addEventListener('open-system-check', onOpenSystemCheck as EventListener);
        return () => {
            window.removeEventListener('open-system-check', onOpenSystemCheck as EventListener);
        };
    }, []);

    useEffect(() => {
        if (cfg.ai_provider) {
            // Only fetch if key is present for openai
            if (cfg.ai_provider === 'openai' && !cfg.openai_key) {
                setAIModels([]);
                return;
            }
            GetAIModels(cfg.ai_provider, cfg.openai_key)
                .then(setAIModels)
                .catch(err => {
                    console.error("Failed to fetch models", err);
                    setAIModels([]);
                });
        }
    }, [cfg.ai_provider, cfg.openai_key]);

    const save = () => {
        setStatus({msg: 'Saving...', type: ''});
        UpdateConfig(cfg as any).then(() => {
            setStatus({msg: `Saved to: ${configPath}`, type: 'success'});
            setTimeout(() => {
                setStatus({msg: '', type: ''});
                props.onSaved?.();
            }, 1000);
        }).catch(err => {
            setStatus({msg: `Error: ${err}`, type: 'error'});
        });
    };

    const selectVault = () => {
        SelectVaultPath().then(path => {
            if(path) setCfg({...cfg, vault_path: path});
        });
    };

    const selectModel = () => {
        SelectModelPath().then(path => {
            if(path) setCfg({...cfg, model_path: path});
        });
    };

    const resetPrompt = () => {
        setCfg({...cfg, custom_prompt: ''});
    };

    return (
        <div className="max-w-2xl mx-auto p-8 w-full">

            <div className="mb-10">
                <h3 className="text-lg font-bold text-white mb-6 border-b border-varys-border/20 pb-2">Configuration</h3>

                <div className="space-y-6">
                    <div>
                        <label className="block text-sm font-semibold text-slate-400 mb-2">Obsidian Vault</label>
                        <div className="flex gap-2">
                            <input
                                className="flex-1 bg-varys-surface border border-varys-border/20 text-slate-300 px-3 py-2.5 rounded-lg text-sm focus:outline-none focus:border-varys-primary/50 shadow-inner"
                                value={cfg.vault_path}
                                readOnly
                            />
                            <button
                                className="bg-varys-muted hover:bg-varys-muted/80 text-slate-200 px-4 py-2 rounded-lg text-sm transition-colors border border-varys-border/10 shadow-lg"
                                onClick={selectVault}
                            >
                                Browse
                            </button>
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-semibold text-slate-400 mb-2">Whisper Model (.bin)</label>
                        <div className="flex gap-2">
                            <input
                                className="flex-1 bg-varys-surface border border-varys-border/20 text-slate-300 px-3 py-2.5 rounded-lg text-sm focus:outline-none focus:border-varys-primary/50 shadow-inner"
                                value={cfg.model_path}
                                readOnly
                            />
                            <button
                                className="bg-varys-muted hover:bg-varys-muted/80 text-slate-200 px-4 py-2 rounded-lg text-sm transition-colors border border-varys-border/10 shadow-lg"
                                onClick={selectModel}
                            >
                                Browse
                            </button>
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-semibold text-slate-400 mb-2">AI Provider</label>
                        <select
                            className="w-full bg-varys-surface border border-varys-border/20 text-slate-300 px-3 py-2.5 rounded-lg text-sm focus:outline-none focus:border-varys-primary/50 appearance-none shadow-inner"
                            value={cfg.ai_provider || 'ollama'}
                            onChange={e => setCfg({...cfg, ai_provider: e.target.value})}
                        >
                            <option value="ollama">Ollama (Local)</option>
                            <option value="openai">OpenAI (Cloud)</option>
                        </select>
                    </div>

                    {cfg.ai_provider === 'openai' && (
                        <div>
                            <label className="block text-sm font-semibold text-slate-400 mb-2 flex justify-between">
                                <span>API Key</span>
                                <span className={`text-xs ${cfg.openai_key && cfg.openai_key.length > 20 ? 'text-varys-secondary' : 'text-slate-500'}`}>
                                    {cfg.openai_key ? `Length: ${cfg.openai_key.length}` : 'Not Set'}
                                </span>
                            </label>
                            <input
                                type="password"
                                className="w-full bg-varys-surface border border-varys-border/20 text-slate-300 px-3 py-2.5 rounded-lg text-sm focus:outline-none focus:border-varys-primary/50 font-mono shadow-inner"
                                value={cfg.openai_key || ''}
                                onChange={e => setCfg({...cfg, openai_key: e.target.value})}
                                placeholder="sk-..."
                            />
                            <p className="mt-1 text-[10px] text-slate-500 italic">
                                Your API Key is stored locally in <code>config.json</code>.
                            </p>
                        </div>
                    )}

                    <div>
                        <label className="block text-sm font-semibold text-slate-400 mb-2">
                            {cfg.ai_provider === 'openai' ? 'OpenAI Model' : 'Ollama Model'}
                        </label>
                        {aiModels.length > 0 ? (
                            <select
                                className="w-full bg-varys-surface border border-varys-border/20 text-slate-300 px-3 py-2.5 rounded-lg text-sm focus:outline-none focus:border-varys-primary/50 appearance-none shadow-inner"
                                value={cfg.ai_provider === 'openai' ? cfg.openai_model : cfg.llm_model}
                                onChange={e => {
                                    if (cfg.ai_provider === 'openai') {
                                        setCfg({...cfg, openai_model: e.target.value});
                                    } else {
                                        setCfg({...cfg, llm_model: e.target.value});
                                    }
                                }}
                            >
                                <option value="" disabled>Select a model...</option>
                                {aiModels.map(m => <option key={m} value={m}>{m}</option>)}
                            </select>
                        ) : (
                            <input
                                className="w-full bg-varys-surface border border-varys-border/20 text-slate-300 px-3 py-2.5 rounded-lg text-sm focus:outline-none focus:border-varys-primary/50 shadow-inner"
                                value={cfg.ai_provider === 'openai' ? cfg.openai_model : cfg.llm_model}
                                onChange={e => {
                                    if (cfg.ai_provider === 'openai') {
                                        setCfg({...cfg, openai_model: e.target.value});
                                    } else {
                                        setCfg({...cfg, llm_model: e.target.value});
                                    }
                                }}
                                placeholder={cfg.ai_provider === 'openai' ? "e.g. gpt-4o" : "e.g. qwen3:8b"}
                            />
                        )}
                    </div>

                    <div>
                        <label className="block text-sm font-semibold text-slate-400 mb-2">Target Language</label>
                        <select
                            className="w-full bg-varys-surface border border-varys-border/20 text-slate-300 px-3 py-2.5 rounded-lg text-sm focus:outline-none focus:border-varys-primary/50 appearance-none shadow-inner"
                            value={cfg.target_language || "Simplified Chinese"}
                            onChange={e => setCfg({...cfg, target_language: e.target.value})}
                        >
                            {languages.map(lang => <option key={lang} value={lang}>{lang}</option>)}
                        </select>
                    </div>

                    <div>
                        <label className="block text-sm font-semibold text-slate-400 mb-2">Context Size (Tokens)</label>
                        <select 
                            className="w-full bg-varys-surface border border-varys-border/20 text-slate-300 px-3 py-2.5 rounded-lg text-sm focus:outline-none focus:border-varys-primary/50 appearance-none shadow-inner" 
                            value={cfg.context_size || 8192} 
                            onChange={e => setCfg({...cfg, context_size: parseInt(e.target.value)})} 
                        >
                            <option value={4096}>4k (Low Memory)</option>
                            <option value={8192}>8k (Default)</option>
                            <option value={16384}>16k (High)</option>
                            <option value={32768}>32k (Max)</option>
                        </select>
                        <p className="mt-1 text-[10px] text-slate-500 italic">
                            Higher values require more RAM. 8k is recommended for most setups.
                        </p>
                    </div>

                    <div>
                        <div className="flex justify-between items-center mb-2">
                             <label className="block text-sm font-semibold text-slate-400">Custom Analysis Prompt</label>
                             <button onClick={resetPrompt} className="text-xs text-varys-secondary hover:underline">Reset to Default</button>
                        </div>
                        <textarea 
                            className="w-full bg-varys-surface border border-varys-border/20 text-slate-300 px-3 py-2.5 rounded-lg text-sm focus:outline-none focus:border-varys-primary/50 font-mono shadow-inner"
                            rows={6}
                            placeholder={defaultPrompt}
                            value={cfg.custom_prompt || ""}
                            onChange={e => setCfg({...cfg, custom_prompt: e.target.value})}
                        />
                    </div>
                </div>
            </div>

            <div className="mb-10" ref={systemCheckRef}>
                <div className="flex items-center justify-between mb-6 border-b border-varys-border/20 pb-2">
                    <h3 className="text-lg font-bold text-white">Dependency Health</h3>
                    <div className="flex items-center gap-2">
                        <button
                            className="px-3 py-1.5 rounded-md bg-varys-muted hover:bg-varys-muted/80 text-xs text-slate-100 border border-varys-border/10 transition-colors shadow-lg"
                            onClick={refreshDiagnostics}
                        >
                            Re-check
                        </button>
                    </div>
                </div>
                <div className="bg-varys-surface/40 border border-varys-border/10 rounded-xl p-5 mb-4 shadow-xl">
                    <div className="flex items-center justify-between">
                        <span className="text-sm font-medium text-slate-300">Overall Status</span>
                        <HealthStatusBadge
                            status={diagnostics?.ready ? "ok" : "misconfigured"}
                            isBlocker={!diagnostics?.ready}
                        />
                    </div>
                    <div className="mt-2 text-[10px] text-slate-500 uppercase tracking-wider font-bold">
                        Last checked: {diagnostics?.generated_at || 'Not checked yet'}
                    </div>
                </div>

                <div className="grid grid-cols-1 gap-3 text-sm">
                    {(diagnostics?.items || []).map((item) => <HealthItemRow key={item.id} item={item} />)}
                </div>
            </div>

            <div className="flex flex-col items-end gap-3 pt-8 border-t border-varys-border/20">
                <button
                    className="bg-varys-primary hover:bg-varys-primary/80 text-white px-8 py-3 rounded-xl font-bold transition-all shadow-xl shadow-varys-primary/20 active:scale-95"
                    onClick={save}
                >
                    Save Changes
                </button>
                {status.msg && (
                    <div className={`text-xs font-semibold ${status.type === 'error' ? 'text-red-400' : 'text-varys-secondary'} text-right max-w-full break-all`}>
                        {status.msg}
                    </div>
                )}
            </div>

            <div className="mt-8 text-center text-[10px] text-slate-600 font-bold uppercase tracking-widest">
                Varys {version}
            </div>
        </div>
    );
}
