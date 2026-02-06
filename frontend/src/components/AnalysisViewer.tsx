import { useRef, useEffect } from 'react';

interface AnalysisViewerProps {
    content: string;
}

export default function AnalysisViewer({ content }: AnalysisViewerProps) {
    const streamEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        streamEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [content]);

    if (!content) return null;

    return (
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
                    {content}
                </div>
                <div ref={streamEndRef} />
            </div>
        </div>
    );
}
