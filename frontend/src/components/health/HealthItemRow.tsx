import { app } from "../../../wailsjs/go/models";
import HealthStatusBadge from "./HealthStatusBadge";
import { useState } from "react";

interface HealthItemRowProps {
    item: app.DiagnosticItem;
    onFix?: (id: string) => Promise<void>;
}

export default function HealthItemRow(props: HealthItemRowProps) {
    const { item, onFix } = props;
    const [fixing, setFixing] = useState(false);

    const handleToggle = async () => {
        if (!onFix) return;
        setFixing(true);
        try {
            await onFix(item.id);
        } finally {
            setFixing(false);
        }
    };

    const isOllama = item.id === 'ollama';
    const isRunning = item.status === 'ok';

    return (
        <div className="bg-slate-800/50 border border-slate-800 p-3 rounded-lg flex items-center justify-between gap-3">
            <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                    <div className="text-slate-300 truncate font-medium">{item.name}</div>
                    {!isOllama && <HealthStatusBadge status={item.status} isBlocker={item.is_blocker} compact />}
                </div>
                <div className="text-[10px] text-slate-500 truncate mt-0.5 uppercase tracking-wider font-semibold opacity-70">
                    ID: {item.id}
                    {item.is_blocker ? " · blocker" : ""}
                </div>
                {item.status !== "ok" && item.fix_suggestion && (
                    <div className="text-[10px] text-slate-400 mt-1 italic leading-tight">
                        {item.fix_suggestion}
                    </div>
                )}
            </div>
            
            <div className="flex flex-col items-end gap-2">
                {isOllama ? (
                    <div className="flex items-center gap-3">
                        <span className={`text-[10px] font-bold uppercase tracking-widest ${isRunning ? 'text-emerald-400' : 'text-slate-500'}`}>
                            {fixing ? 'Processing...' : (isRunning ? 'Running' : 'Stopped')}
                        </span>
                        <button
                            onClick={handleToggle}
                            disabled={fixing}
                            className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus:outline-none ${
                                isRunning ? 'bg-emerald-600' : 'bg-slate-700'
                            } ${fixing ? 'opacity-50 cursor-wait' : 'cursor-pointer'}`}
                        >
                            <span
                                className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform ${
                                    isRunning ? 'translate-x-5' : 'translate-x-1'
                                }`}
                            />
                        </button>
                    </div>
                ) : (
                    <HealthStatusBadge 
                        status={fixing ? "fixing" : item.status} 
                        isBlocker={item.is_blocker} 
                        compact 
                        onClick={item.can_auto_fix ? (e) => { e.stopPropagation(); handleToggle(); } : undefined}
                    />
                )}
            </div>
        </div>
    );
}
