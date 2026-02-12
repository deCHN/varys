import { useState, useRef, useEffect } from 'react';
import { useTaskRunner } from './hooks/useTaskRunner';
import LogConsole from './components/LogConsole';
import AnalysisViewer from './components/AnalysisViewer';
import { GetStartupDiagnostics } from '../wailsjs/go/main/App';
import { main } from '../wailsjs/go/models';

interface DashboardProps {
    onPreflightFailed?: (diag: main.StartupDiagnostics) => void;
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
                    props.onPreflightFailed?.(diag);
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
        <div className="flex flex-col h-full max-w-5xl mx-auto p-6 w-full">
            {/* Input Section */}
            <div className="flex gap-3 mb-6 items-center">
                <div className="flex-1 relative">
                    <input
                        ref={inputRef}
                        className="w-full bg-slate-800 border border-slate-700 text-slate-100 pl-4 pr-32 py-3 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-500 transition-all"
                        value={url}
                        onChange={(e) => setUrl(e.target.value)}
                        onKeyDown={handleKeyDown}
                        placeholder="Enter YouTube/Bilibili URL"
                        disabled={isProcessing}
                    />

                    <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center bg-slate-700/50 rounded-full px-2 py-1">
                        <label className={`flex items-center cursor-pointer gap-2 select-none ${isProcessing ? 'opacity-50 cursor-not-allowed' : ''}`}>
                            <span className={`text-xs font-medium transition-colors ${downloadVideo ? 'text-blue-400' : 'text-slate-400'}`}>Video</span>
                            <div className="relative">
                                <input
                                    type="checkbox"
                                    className="sr-only peer"
                                    checked={downloadVideo}
                                    onChange={(e) => setDownloadVideo(e.target.checked)}
                                    disabled={isProcessing}
                                />
                                <div className="w-7 h-4 bg-slate-600 rounded-full peer peer-checked:bg-blue-500/80 peer-focus:ring-2 peer-focus:ring-blue-800 transition-colors"></div>
                                <div className="absolute left-[2px] top-[2px] bg-white w-3 h-3 rounded-full transition-transform peer-checked:translate-x-3"></div>
                            </div>
                        </label>
                    </div>
                </div>

                <button
                    className={`px-6 py-3 rounded-lg font-medium transition-all shadow-lg active:scale-95 flex items-center justify-center w-16 ${
                        isProcessing 
                        ? 'bg-red-600 hover:bg-red-500 text-white shadow-red-900/20' 
                        : 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-emerald-900/20'
                    }`}
                    onClick={handleProcessToggle}
                    title={isProcessing ? "Stop & Clear" : "Start Processing"}
                >
                    {isProcessing ? (
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <rect x="6" y="6" width="12" height="12" rx="1" />
                        </svg>
                    ) : (
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <polygon points="5 3 19 12 5 21 5 3" />
                        </svg>
                    )}
                </button>
            </div>

            {/* Progress Bar */}
            {progress > 0 && (
                <div className="w-full bg-slate-800 rounded-full h-1.5 mb-6 overflow-hidden">
                    <div
                        className="bg-emerald-500 h-full transition-all duration-500 ease-out"
                        style={{ width: `${progress}%` }}
                    ></div>
                </div>
            )}

            {/* Split View: Logs & Analysis */}
            <div className="flex gap-6 flex-1 min-h-0">
                <LogConsole logs={logs} />
                <AnalysisViewer content={analysisStream} />
            </div>

            {/* Footer Status */}
            {resultText && (
                <div className={`mt-4 text-center text-sm font-medium py-2 rounded-lg ${
                    resultText.includes("failed")
                    ? 'bg-red-500/10 text-red-400 border border-red-500/20'
                    : 'bg-green-500/10 text-green-400 border border-green-500/20'
                }`}>
                    {resultText}
                </div>
            )}
        </div>
    );
}
