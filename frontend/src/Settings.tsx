import { useState, useEffect } from 'react';
import { GetConfig, UpdateConfig, SelectVaultPath, SelectModelPath, CheckDependencies, GetOllamaModels, GetConfigPath, GetAppVersion } from "../wailsjs/go/main/App";

interface Config {
    vault_path: string;
    model_path: string;
    llm_model: string;
    translation_model: string;
    target_language: string;
    context_size: number;
}

export default function Settings() {
    const [cfg, setCfg] = useState<Config>({ 
        vault_path: '', 
        model_path: '', 
        llm_model: '', 
        translation_model: 'qwen3:0.6b',
        target_language: '', 
        context_size: 8192 
    });
    const [deps, setDeps] = useState<any>({});
    const [ollamaModels, setOllamaModels] = useState<string[]>([]);
    const [configPath, setConfigPath] = useState<string>('');
    const [version, setVersion] = useState<string>('');
    const [status, setStatus] = useState<{msg: string, type: 'success' | 'error' | ''}>({msg: '', type: ''});

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

    useEffect(() => {
        GetConfig().then((c: any) => setCfg(c));
        CheckDependencies().then(setDeps);
        GetOllamaModels().then(setOllamaModels).catch(err => console.error("Failed to fetch models", err));
        GetConfigPath().then(setConfigPath);
        GetAppVersion().then(setVersion);
    }, []);

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
                        <label className="block text-sm font-medium text-slate-400 mb-2">Ollama Model</label>
                        {ollamaModels.length > 0 ? (
                            <select
                                className="w-full bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500 appearance-none"
                                value={cfg.llm_model}
                                onChange={e => setCfg({...cfg, llm_model: e.target.value})}
                            >
                                <option value="" disabled>Select a model...</option>
                                {ollamaModels.map(m => <option key={m} value={m}>{m}</option>)}
                            </select>
                        ) : (
                            <input
                                className="w-full bg-slate-800 border border-slate-700 text-slate-300 px-3 py-2 rounded-lg text-sm focus:outline-none focus:border-blue-500"
                                value={cfg.llm_model}
                                onChange={e => setCfg({...cfg, llm_model: e.target.value})}
                                placeholder="Type model name (e.g. qwen2.5:7b)"
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
                </div>
            </div>

            <div className="mb-10">
                <h3 className="text-lg font-semibold text-slate-200 mb-6 border-b border-slate-800 pb-2">System Check</h3>
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
