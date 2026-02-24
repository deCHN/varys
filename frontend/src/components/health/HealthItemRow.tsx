import { main } from "../../../wailsjs/go/models";
import HealthStatusBadge from "./HealthStatusBadge";
import { useState } from "react";

interface HealthItemRowProps {
    item: main.DiagnosticItem;
    onFix?: (id: string) => Promise<void>;
}

export default function HealthItemRow(props: HealthItemRowProps) {
    const { item, onFix } = props;
    const [fixing, setFixing] = useState(false);

    const handleFix = async () => {
        if (!onFix) return;
        setFixing(true);
        try {
            await onFix(item.id);
        } finally {
            setFixing(false);
        }
    };

    return (
        <div className="bg-slate-800/50 border border-slate-800 p-3 rounded-lg flex items-center justify-between gap-3">
            <div className="min-w-0 flex-1">
                <div className="text-slate-300 truncate font-medium">{item.name}</div>
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
                <HealthStatusBadge status={item.status} isBlocker={item.is_blocker} compact />
                {item.can_auto_fix && item.status !== "ok" && onFix && (
                    <button
                        onClick={handleFix}
                        disabled={fixing}
                        className="bg-emerald-600 hover:bg-emerald-500 disabled:bg-slate-700 text-white text-[10px] font-bold px-2 py-1 rounded-md transition-all shadow-lg active:scale-95 whitespace-nowrap"
                    >
                        {fixing ? "Fixing..." : "Auto Fix"}
                    </button>
                )}
            </div>
        </div>
    );
}
