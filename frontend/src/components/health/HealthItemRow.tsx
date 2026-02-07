import { main } from "../../../wailsjs/go/models";
import HealthStatusBadge from "./HealthStatusBadge";

interface HealthItemRowProps {
    item: main.DiagnosticItem;
}

export default function HealthItemRow(props: HealthItemRowProps) {
    const { item } = props;
    return (
        <div className="bg-slate-800/50 border border-slate-800 p-3 rounded-lg flex items-center justify-between gap-3">
            <div className="min-w-0">
                <div className="text-slate-300 truncate">{item.name}</div>
                <div className="text-xs text-slate-500 truncate">
                    ID: {item.id}
                    {item.is_blocker ? " · blocker" : ""}
                </div>
            </div>
            <HealthStatusBadge status={item.status} isBlocker={item.is_blocker} compact />
        </div>
    );
}
