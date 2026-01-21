import {useState, useEffect, useRef} from 'react';
import {SubmitTask} from "../wailsjs/go/main/App";
import {EventsOn} from "../wailsjs/runtime";

export default function Dashboard() {
    const [resultText, setResultText] = useState("");
    const [url, setUrl] = useState('');
    const [logs, setLogs] = useState<string[]>([]);
    const [analysisStream, setAnalysisStream] = useState("");
    const inputRef = useRef<HTMLInputElement>(null);
    const logEndRef = useRef<HTMLDivElement>(null);
    const streamEndRef = useRef<HTMLDivElement>(null);

    const updateUrl = (e: any) => setUrl(e.target.value);

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
        return () => {
            unsubLog();
            unsubAnalysis();
        };
    }, []);

    useEffect(() => {
        logEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [logs]);

    useEffect(() => {
        streamEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [analysisStream]);

    function processUrl() {
        if (!url) return;
        setLogs([]);
        setAnalysisStream("");
        setResultText("Processing...");
        addLog(`Processing URL: ${url}`);
        
        SubmitTask(url).then((response: string) => {
             addLog(`Backend Response: ${response}`);
             setResultText("Task completed");
        }).catch((err: any) => {
             addLog(`Error: ${err}`);
             setResultText("Task failed");
        });
    }

    return (
        <div className="flex flex-col h-full max-w-5xl mx-auto p-6 w-full">
            {/* Input Section */}
            <div className="flex gap-3 mb-6">
                <input 
                    ref={inputRef}
                    className="flex-1 bg-slate-800 border border-slate-700 text-slate-100 px-4 py-3 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-500 transition-all"
                    value={url} 
                    onChange={updateUrl} 
                    placeholder="Enter YouTube/Bilibili URL"
                />
                <button 
                    className="bg-blue-600 hover:bg-blue-500 text-white px-6 py-3 rounded-lg font-medium transition-colors shadow-lg shadow-blue-900/20 active:scale-95"
                    onClick={processUrl}
                >
                    Process
                </button>
            </div>

            {/* Split View: Logs & Analysis */}
            <div className="flex gap-6 flex-1 min-h-0">
                {/* System Logs */}
                <div className="flex-1 flex flex-col bg-slate-800/50 border border-slate-800 rounded-xl overflow-hidden">
                    <div className="px-4 py-3 border-b border-slate-800 bg-slate-800/80 font-medium text-slate-400 text-xs uppercase tracking-wider">
                        System Logs
                    </div>
                    <div className="flex-1 overflow-y-auto p-4 font-mono text-xs space-y-1">
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