import {useState, useEffect, useRef} from 'react';
import {SubmitTask, CancelTask} from "../wailsjs/go/main/App";
import {EventsOn} from "../wailsjs/runtime";

export default function Dashboard() {
    const [resultText, setResultText] = useState("");
    const [url, setUrl] = useState('');
    const [downloadVideo, setDownloadVideo] = useState(false);
    const [isProcessing, setIsProcessing] = useState(false);
    const [logs, setLogs] = useState<string[]>([]);
    const [analysisStream, setAnalysisStream] = useState("");
    const [progress, setProgress] = useState(0);
    const inputRef = useRef<HTMLInputElement>(null);
    const logEndRef = useRef<HTMLDivElement>(null);
    const streamEndRef = useRef<HTMLDivElement>(null);

    const updateUrl = (e: any) => setUrl(e.target.value);

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
            handleProcessToggle();
        }
    };

    function addLog(msg: string) {
        const time = new Date().toLocaleTimeString([], { hour12: false });
        setLogs(prev => [...prev, `[${time}] ${msg}`]);
    }

    useEffect(() => {
        // Auto-focus input on mount
        if (inputRef.current) {
            inputRef.current.focus();
        }

        const unsubLog = EventsOn("task:log", (msg: string) => addLog(msg));
        const unsubAnalysis = EventsOn("task:analysis", (chunk: string) => {
            setAnalysisStream(prev => prev + chunk);
        });
        const unsubProgress = EventsOn("task:progress", (p: number) => {
            setProgress(p);
        });

        return () => {
            unsubLog();
            unsubAnalysis();
            unsubProgress();
        };
    }, []);

    useEffect(() => {
        logEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [logs]);

    useEffect(() => {
        streamEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [analysisStream]);

    function handleProcessToggle() {
        if (isProcessing) {
            // Stop/Cancel
            CancelTask().then(() => {
                addLog("User requested cancellation...");
                setIsProcessing(false);
                setResultText("Cancelled");
                setProgress(0);
            });
        } else {
            // Start
            processUrl();
        }
    }

    function processUrl() {
        if (!url) return;
        setLogs([]);
        setAnalysisStream("");
        setProgress(0);
        setIsProcessing(true);
        setResultText("Processing...");
        const audioOnly = !downloadVideo;
        addLog(`Processing URL: ${url} (AudioOnly: ${audioOnly})`);

        SubmitTask(url, audioOnly).then((response: string) => {
             addLog(`Backend Response: ${response}`);
             setResultText("Task completed");
             setProgress(0); // Reset after done
             setIsProcessing(false);
        }).catch((err: any) => {
             addLog(`Error: ${err}`);
             setResultText("Task failed");
             setProgress(0);
             setIsProcessing(false);
        });
    }

    function copyLogs() {
        if (logs.length === 0) return;
        navigator.clipboard.writeText(logs.join('\n')).then(() => {
            // Optional: Visual feedback could be added here
        }).catch(err => {
            console.error('Failed to copy logs:', err);
        });
    }

    return (
        <div className="flex flex-col h-full max-w-5xl mx-auto p-6 w-full">
            {/* Input Section */}
            <div className="flex gap-3 mb-6 items-center">
                <div className="flex-1 relative">
                    <input
                        ref={inputRef}
                        className="w-full bg-slate-800 border border-slate-700 text-slate-100 pl-4 pr-32 py-3 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-500 transition-all"
                        value={url}
                        onChange={updateUrl}
                        onKeyDown={handleKeyDown}
                        placeholder="Enter YouTube/Bilibili URL"
                        disabled={isProcessing}
                    />

                    {/* Integrated Toggle Switch */}
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
                {/* System Logs */}
                <div className="flex-1 flex flex-col bg-slate-800/50 border border-slate-800 rounded-xl overflow-hidden">
                    <div className="px-4 py-3 border-b border-slate-800 bg-slate-800/80 font-medium text-slate-400 text-xs uppercase tracking-wider flex justify-between items-center">
                        <span>System Logs</span>
                        <button
                            onClick={copyLogs}
                            className="text-slate-500 hover:text-slate-300 transition-colors p-1 rounded hover:bg-slate-700/50"
                            title="Copy to Clipboard"
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                            </svg>
                        </button>
                    </div>
                    <div className="flex-1 overflow-y-auto p-4 font-mono text-xs space-y-1 text-left select-text">
                        {logs.length === 0 && <div className="text-slate-600 italic">Logs will appear here...</div>}
                        {logs.map((log, index) => (
                            <div key={index} className="text-slate-300 break-all border-l-2 border-transparent hover:border-slate-600 pl-2 -ml-2 py-0.5">
                                {log}
                            </div>
                        ))}
                        <div ref={logEndRef} />
                    </div>
                </div>

                {/* Live Analysis */}
                {analysisStream && (
                    <div className="flex-1 flex flex-col bg-slate-800/50 border border-slate-800 rounded-xl overflow-hidden animate-in fade-in slide-in-from-bottom-4 duration-500">
                        <div className="px-4 py-3 border-b border-slate-800 bg-slate-800/80 font-medium text-blue-400 text-xs uppercase tracking-wider flex justify-between items-center">
                            <span>Live Analysis</span>
                            <span className="flex h-2 w-2 relative">
                                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                                <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
                            </span>
                        </div>
                        <div className="flex-1 overflow-y-auto p-5 text-sm leading-relaxed text-slate-200">
                            <div className="whitespace-pre-wrap markdown-body">
                                {analysisStream}
                            </div>
                            <div ref={streamEndRef} />
                        </div>
                    </div>
                )}
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
    )
}