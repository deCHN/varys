import { useState, useRef, useEffect } from 'react';
import { useTaskRunner } from './hooks/useTaskRunner';
import LogConsole from './components/LogConsole';
import AnalysisViewer from './components/AnalysisViewer';
import { GetStartupDiagnostics } from '../wailsjs/go/app/App';

interface DashboardProps {
    onPreflightFailed?: () => void;
    version?: string;
    onAboutClick?: () => void;
}

export default function Dashboard(props: DashboardProps) {
    const [url, setUrl] = useState('');
    const [downloadVideo, setDownloadVideo] = useState(false);
    const inputRef = useRef<HTMLInputElement>(null);
    
    const {
        isProcessing,
        logs,
        analysisStream,
        progress,
        resultText,
        runTask,
        cancel
    } = useTaskRunner();

    useEffect(() => {
        inputRef.current?.focus();
    }, []);

    const handleProcessToggle = async () => {
        if (isProcessing) {
            cancel();
        } else {
            try {
                const diag = await GetStartupDiagnostics();
                if (!diag.ready) {
                    props.onPreflightFailed?.();
                    return;
                }
            } catch (err) {
                console.error('Failed to run startup diagnostics', err);
            }
            runTask(url, downloadVideo);
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') handleProcessToggle();
    };

    return (
        <div className="flex flex-col h-full max-w-5xl mx-auto p-6 w-full relative">
            {/* Hero Section (Always visible to prevent jumps) */}
            <div className="flex flex-col items-center justify-center py-12 animate-in fade-in duration-700">
                <h2 className="text-3xl font-bold text-white mb-1 tracking-tight">Capture Analyze Transcribe</h2>
                <p className="text-slate-400 text-base text-center font-medium opacity-80">Private, offline intelligence for your second brain.</p>
            </div>

            {/* Input Section */}
            <div className="flex gap-3 mb-6 items-center">
                <div className="flex-1 relative group">
                    <input
                        ref={inputRef}
                        className="w-full bg-varys-surface border border-varys-border/20 text-slate-100 pl-4 pr-32 py-4 rounded-xl focus:outline-none focus:ring-2 focus:ring-varys-primary/50 placeholder-slate-500 transition-all shadow-lg group-hover:border-varys-primary/30"
                        value={url}
                        onChange={(e) => setUrl(e.target.value)}
                        onKeyDown={handleKeyDown}
                        placeholder="Enter YouTube/Bilibili URL"
                        disabled={isProcessing}
                    />

                    <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center bg-black/40 backdrop-blur-sm rounded-full px-2 py-1.5 border border-white/5">
                        <label className={`flex items-center cursor-pointer gap-2 select-none ${isProcessing ? 'opacity-50 cursor-not-allowed' : ''}`}>
                            <span className={`text-xs font-semibold transition-colors ${downloadVideo ? 'text-varys-secondary' : 'text-slate-400'}`}>Video</span>
                            <div className="relative">
                                <input
                                    type="checkbox"
                                    className="sr-only peer"
                                    checked={downloadVideo}
                                    onChange={(e) => setDownloadVideo(e.target.checked)}
                                    disabled={isProcessing}
                                />
                                <div className="w-7 h-4 bg-slate-700 rounded-full peer peer-checked:bg-varys-secondary transition-colors"></div>
                                <div className="absolute left-[2px] top-[2px] bg-white w-3 h-3 rounded-full transition-transform peer-checked:translate-x-3"></div>
                            </div>
                        </label>
                    </div>
                </div>

                <button
                    className={`px-6 py-4 rounded-xl font-bold transition-all shadow-xl active:scale-95 flex items-center justify-center w-16 group ${
                        isProcessing 
                        ? 'bg-red-500 hover:bg-red-400 text-white shadow-red-900/40' 
                        : 'bg-varys-primary hover:bg-varys-primary/80 text-white shadow-varys-primary/30'
                    }`}
                    onClick={handleProcessToggle}
                    title={isProcessing ? "Stop & Clear" : "Start Processing"}
                >
                    {isProcessing ? (
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
                            <rect x="6" y="6" width="12" height="12" rx="1" />
                        </svg>
                    ) : (
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor" className="group-hover:translate-x-0.5 transition-transform">
                            <polygon points="5 3 19 12 5 21 5 3" />
                        </svg>
                    )}
                </button>
            </div>

            {/* Progress Bar */}
            {progress > 0 && (
                <div className="w-full bg-varys-surface/50 rounded-full h-2 mb-6 overflow-hidden border border-white/5">
                    <div
                        className="bg-gradient-to-r from-varys-primary to-varys-secondary h-full transition-all duration-500 ease-out shadow-[0_0_10px_rgba(147,51,234,0.5)]"
                        style={{ width: `${progress}%` }}
                    ></div>
                </div>
            )}

            {/* Split View: Logs & Analysis */}
            <div className="flex gap-6 flex-1 min-h-0">
                <LogConsole logs={logs} version={props.version} onAboutClick={props.onAboutClick} />
                <AnalysisViewer content={analysisStream} />
            </div>

            {/* Footer Status */}
            {resultText && (
                <div className={`mt-4 text-center text-sm font-bold py-3 rounded-xl backdrop-blur-sm animate-in zoom-in-95 duration-300 shadow-lg ${
                    resultText.includes("failed")
                    ? 'bg-red-500/10 text-red-400 border border-red-500/30'
                    : 'bg-varys-secondary/10 text-varys-secondary border border-varys-secondary/30'
                }`}>
                    {resultText}
                </div>
            )}
        </div>
    );
}
