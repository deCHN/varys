import { useState, useEffect } from 'react';
import { GetConfig, UpdateConfig, SelectVaultPath, SelectModelPath, CheckDependencies, GetOllamaModels, GetConfigPath } from "../wailsjs/go/main/App";

interface Config {
    vault_path: string;
    model_path: string;
    llm_model: string;
}

export default function Settings() {
    const [cfg, setCfg] = useState<Config>({ vault_path: '', model_path: '', llm_model: '' });
    const [deps, setDeps] = useState<any>({});
    const [ollamaModels, setOllamaModels] = useState<string[]>([]);
    const [configPath, setConfigPath] = useState<string>('');
    const [status, setStatus] = useState<{msg: string, type: 'success' | 'error' | ''}>({msg: '', type: ''});

    useEffect(() => {
        GetConfig().then((c: any) => setCfg(c));
        CheckDependencies().then(setDeps);
        GetOllamaModels().then(setOllamaModels).catch(err => console.error("Failed to fetch models", err));
        GetConfigPath().then(setConfigPath);
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
        <span style={{ color: ok ? '#188038' : '#d93025', marginLeft: '8px' }}>
            {ok ? "✓" : "⚠️"}
        </span>
    );

    return (
        <div style={{ padding: '24px', maxWidth: '700px', margin: '0 auto' }}>
            
            <div className="section">
                <h3>Paths</h3>
                
                <div style={{ marginBottom: '20px' }}>
                    <label style={{ display: 'block', marginBottom: '6px', fontSize: '13px', fontWeight: 500 }}>Obsidian Vault</label>
                    <div style={{ display: 'flex', gap: '8px' }}>
                        <input className="input" value={cfg.vault_path} readOnly style={{ flex: 1, background: '#f8f9fa' }} />
                        <button className="btn" style={{ background: '#fff', color: '#1a73e8', border: '1px solid #dadce0' }} onClick={selectVault}>Browse</button>
                    </div>
                </div>

                <div style={{ marginBottom: '20px' }}>
                    <label style={{ display: 'block', marginBottom: '6px', fontSize: '13px', fontWeight: 500 }}>Whisper Model (.bin)</label>
                    <div style={{ display: 'flex', gap: '8px' }}>
                        <input className="input" value={cfg.model_path} readOnly style={{ flex: 1, background: '#f8f9fa' }} />
                        <button className="btn" style={{ background: '#fff', color: '#1a73e8', border: '1px solid #dadce0' }} onClick={selectModel}>Browse</button>
                    </div>
                </div>
                
                <div style={{ marginBottom: '20px' }}>
                    <label style={{ display: 'block', marginBottom: '6px', fontSize: '13px', fontWeight: 500 }}>Ollama Model</label>
                    {ollamaModels.length > 0 ? (
                        <select 
                            className="input" 
                            value={cfg.llm_model} 
                            onChange={e => setCfg({...cfg, llm_model: e.target.value})} 
                            style={{ width: '100%' }}
                        >
                            <option value="" disabled>Select a model...</option>
                            {ollamaModels.map(m => <option key={m} value={m}>{m}</option>)}
                        </select>
                    ) : (
                        <input 
                            className="input" 
                            value={cfg.llm_model} 
                            onChange={e => setCfg({...cfg, llm_model: e.target.value})} 
                            placeholder="Type model name (e.g. qwen2.5:7b)"
                            style={{ width: '100%' }} 
                        />
                    )}
                </div>
            </div>

            <div className="section">
                <h3>System Check</h3>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', fontSize: '13px' }}>
                    <div style={{ padding: '12px', border: '1px solid #dadce0', borderRadius: '4px', display: 'flex', justifyContent: 'space-between' }}>
                        <span>yt-dlp</span> <StatusIcon ok={deps.yt_dlp} />
                    </div>
                    <div style={{ padding: '12px', border: '1px solid #dadce0', borderRadius: '4px', display: 'flex', justifyContent: 'space-between' }}>
                        <span>ffmpeg</span> <StatusIcon ok={deps.ffmpeg} />
                    </div>
                    <div style={{ padding: '12px', border: '1px solid #dadce0', borderRadius: '4px', display: 'flex', justifyContent: 'space-between' }}>
                        <span>whisper-cpp</span> <StatusIcon ok={deps.whisper} />
                    </div>
                    <div style={{ padding: '12px', border: '1px solid #dadce0', borderRadius: '4px', display: 'flex', justifyContent: 'space-between' }}>
                        <span>ollama</span> <StatusIcon ok={deps.ollama} />
                    </div>
                </div>
            </div>

            <div style={{ marginTop: '32px', display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '8px' }}>
                <button className="btn" onClick={save}>Save Changes</button>
                {status.msg && (
                    <div style={{ 
                        fontSize: '12px', 
                        color: status.type === 'error' ? '#d93025' : '#188038',
                        maxWidth: '100%',
                        textAlign: 'right'
                    }}>
                        {status.msg}
                    </div>
                )}
            </div>
        </div>
    );
}
