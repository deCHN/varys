import { useState, useEffect, useRef } from 'react';
import { GetConfig, UpdateConfig, SelectVaultPath, SelectModelPath, CheckDependencies, GetAIModels, GetConfigPath, GetAppVersion } from "../wailsjs/go/main/App";

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
    const [deps, setDeps] = useState<any>({});
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

    const refreshSystemCheck = () => {
        CheckDependencies().then(setDeps);
    };

    useEffect(() => {
        GetConfig().then((c: any) => setCfg(c));
        refreshSystemCheck();
        GetConfigPath().then(setConfigPath);
        GetAppVersion().then(setVersion);
    }, []);

    useEffect(() => {
        if (props.isActive) {
            refreshSystemCheck();
        }
    }, [props.isActive]);

    useEffect(() => {
        const onOpenSystemCheck = () => {
            refreshSystemCheck();
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
            setTimeout(() => setStatus({msg: '', type: ''}), 5000);
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

    const StatusIcon = ({ ok }: { ok: boolean }) => (
        <span className={`ml-2 text-xs font-bold ${ok ? 'text-green-400' : 'text-red-400'}`}>
            {ok ? "OK" : "MISSING"}
        </span>
    );

    return (
        <div className="max-w-2xl mx-auto p-8 w-full">

            <div className="mb-10">
                <h3 className="text-lg font-semibold text-slate-200 mb-6 border-b border-slate-800 pb-2">Configuration</h3>

                <div className="space-y-6">
                    <div>
                        <label className="block text-sm font-medium text-slate-400 mb-2">Obsidian Vault</label>
                        <div className="flex gap-2">
                            <input
                                className="flex-1 bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500"
                                value={cfg.vault_path}
                                readOnly
                            />
                            <button
                                className="bg-slate-700 hover:bg-slate-600 text-slate-200 px-4 py-2 rounded-lg text-sm transition-colors border border-slate-600"
                                onClick={selectVault}
                            >
                                Browse
                            </button>
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-400 mb-2">Whisper Model (.bin)</label>
                        <div className="flex gap-2">
                            <input
                                className="flex-1 bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500"
                                value={cfg.model_path}
                                readOnly
                            />
                            <button
                                className="bg-slate-700 hover:bg-slate-600 text-slate-200 px-4 py-2 rounded-lg text-sm transition-colors border border-slate-600"
                                onClick={selectModel}
                            >
                                Browse
                            </button>
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-400 mb-2">AI Provider</label>
                        <select
                            className="w-full bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500 appearance-none"
                            value={cfg.ai_provider || 'ollama'}
                            onChange={e => setCfg({...cfg, ai_provider: e.target.value})}
                        >
                            <option value="ollama">Ollama (Local)</option>
                            <option value="openai">OpenAI (Cloud)</option>
                        </select>
                    </div>

                    {cfg.ai_provider === 'openai' && (
                        <div>
                            <label className="block text-sm font-medium text-slate-400 mb-2 flex justify-between">
                                <span>API Key</span>
                                <span className={`text-xs ${cfg.openai_key && cfg.openai_key.length > 20 ? 'text-green-400' : 'text-slate-500'}`}>
                                    {cfg.openai_key ? `Length: ${cfg.openai_key.length}` : 'Not Set'}
                                </span>
                            </label>
                            <input
                                type="password"
                                className="w-full bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500 font-mono"
                                value={cfg.openai_key || ''}
                                onChange={e => setCfg({...cfg, openai_key: e.target.value})}
                                placeholder="sk-..."
                            />
                            <p className="mt-1 text-xs text-slate-500">
                                Your API Key is stored locally in <code>config.json</code>.
                            </p>
                        </div>
                    )}

                    <div>
                        <label className="block text-sm font-medium text-slate-400 mb-2">
                            {cfg.ai_provider === 'openai' ? 'OpenAI Model' : 'Ollama Model'}
                        </label>
                        {aiModels.length > 0 ? (
                            <select
                                className="w-full bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500 appearance-none"
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
                                className="w-full bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500"
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
                        <label className="block text-sm font-medium text-slate-400 mb-2">Target Language</label>
                        <select
                            className="w-full bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500 appearance-none"
                            value={cfg.target_language || "Simplified Chinese"}
                            onChange={e => setCfg({...cfg, target_language: e.target.value})}
                        >
                            {languages.map(lang => <option key={lang} value={lang}>{lang}</option>)}
                        </select>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-400 mb-2">Context Size (Tokens)</label>
                        <select 
                            className="w-full bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500 appearance-none" 
                            value={cfg.context_size || 8192} 
                            onChange={e => setCfg({...cfg, context_size: parseInt(e.target.value)})} 
                        >
                            <option value={4096}>4k (Low Memory)</option>
                            <option value={8192}>8k (Default)</option>
                            <option value={16384}>16k (High)</option>
                            <option value={32768}>32k (Max)</option>
                        </select>
                        <p className="mt-1 text-xs text-slate-500">
                            Higher values allow longer videos but require more RAM. 8k is safe for 16GB RAM.
                        </p>
                    </div>

                    <div>
                        <div className="flex justify-between items-center mb-2">
                             <label className="block text-sm font-medium text-slate-400">Custom Analysis Prompt</label>
                             <button onClick={resetPrompt} className="text-xs text-blue-400 hover:text-blue-300 hover:underline">Reset to Default</button>
                        </div>
                        <textarea 
                            className="w-full bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500 font-mono"
                            rows={8}
                            placeholder={defaultPrompt}
                            value={cfg.custom_prompt || ""}
                            onChange={e => setCfg({...cfg, custom_prompt: e.target.value})}
                        />
                        <p className="mt-1 text-xs text-slate-500">
                            Leave empty to use the default smart prompt. If set, this will replace the system instruction. 
                            Ensure you request JSON format compatible with the app.
                        </p>
                    </div>
                </div>
            </div>

            <div className="mb-10" ref={systemCheckRef}>
                <div className="flex items-center justify-between mb-6 border-b border-slate-800 pb-2">
                    <h3 className="text-lg font-semibold text-slate-200">System Check</h3>
                    <button
                        className="px-3 py-1.5 rounded-md bg-slate-700 hover:bg-slate-600 text-xs text-slate-100 border border-slate-600"
                        onClick={async () => {
                            refreshSystemCheck();
                            await props.onOpenDependencyHealth?.();
                        }}
                    >
                        Re-check
                    </button>
                </div>
                <div className="grid grid-cols-2 gap-4 text-sm">
                    <div className="bg-slate-800/50 border border-slate-800 p-3 rounded-lg flex justify-between items-center">
                        <span className="text-slate-400">yt-dlp</span> <StatusIcon ok={deps.yt_dlp} />
                    </div>
                    <div className="bg-slate-800/50 border border-slate-800 p-3 rounded-lg flex justify-between items-center">
                        <span className="text-slate-400">ffmpeg</span> <StatusIcon ok={deps.ffmpeg} />
                    </div>
                    <div className="bg-slate-800/50 border border-slate-800 p-3 rounded-lg flex justify-between items-center">
                        <span className="text-slate-400">whisper-cpp</span> <StatusIcon ok={deps.whisper} />
                    </div>
                    <div className="bg-slate-800/50 border border-slate-800 p-3 rounded-lg flex justify-between items-center">
                        <span className="text-slate-400">ollama</span> <StatusIcon ok={deps.ollama} />
                    </div>
                </div>
            </div>

            <div className="flex flex-col items-end gap-3 pt-6 border-t border-slate-800">
                <button
                    className="bg-blue-600 hover:bg-blue-500 text-white px-6 py-2.5 rounded-lg font-medium transition-colors shadow-lg shadow-blue-900/20 active:scale-95"
                    onClick={save}
                >
                    Save Changes
                </button>
                {status.msg && (
                    <div className={`text-xs ${status.type === 'error' ? 'text-red-400' : 'text-green-400'} text-right max-w-full break-all`}>
                        {status.msg}
                    </div>
                )}
            </div>

            <div className="mt-8 text-center text-xs text-slate-600">
                Varys {version}
            </div>
        </div>
    );
}
