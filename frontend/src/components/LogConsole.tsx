import { useRef, useEffect } from 'react';

interface LogConsoleProps {
    logs: string[];
    version?: string;
    onAboutClick?: () => void;
}

export default function LogConsole({ logs, version, onAboutClick }: LogConsoleProps) {
    const logEndRef = useRef<HTMLDivElement>(null);
    const displayVersion = version
        ? (version.toLowerCase().startsWith('v') ? version : `v${version}`)
        : '';

    useEffect(() => {
        logEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [logs]);

    const copyLogs = () => {
        if (logs.length === 0) return;
        navigator.clipboard.writeText(logs.join('\n'));
    };

    return (
        <div className="flex-1 flex flex-col bg-slate-800/50 border border-slate-800 rounded-xl overflow-hidden relative group/console">
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
            <div className="flex-1 overflow-y-auto p-4 font-mono text-xs space-y-1 text-left select-text relative">
                {logs.length === 0 && <div className="text-slate-600 italic">Logs will appear here...</div>}
                {logs.map((log, index) => (
                    <div key={index} className="text-slate-300 break-all border-l-2 border-transparent hover:border-slate-600 pl-2 -ml-2 py-0.5">
                        {log}
                    </div>
                ))}
                <div ref={logEndRef} />
            </div>

            {/* Version Badge inside the console box */}
            {version && (
                <button
                    onClick={onAboutClick}
                    className="absolute bottom-2 right-3 text-[9px] font-bold text-slate-600/50 hover:text-varys-primary transition-colors tracking-widest uppercase z-20 bg-slate-900/40 backdrop-blur-sm px-1.5 py-0.5 rounded border border-white/5"
                >
                    {displayVersion}
                </button>
            )}
        </div>
    );
}
