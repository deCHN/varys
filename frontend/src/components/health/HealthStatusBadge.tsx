interface HealthStatusBadgeProps {
    status: string;
    isBlocker?: boolean;
    compact?: boolean;
    onClick?: (e: React.MouseEvent) => void;
}

function badgeClass(status: string, interactive: boolean): string {
    if (status === "ok") {
        return interactive 
            ? "text-emerald-300 bg-emerald-500/10 border-emerald-500/30 hover:bg-red-500/20 hover:text-red-300 hover:border-red-500/30" 
            : "text-emerald-300 bg-emerald-500/10 border-emerald-500/30";
    }
    if (status === "missing") {
        return "text-red-300 bg-red-500/10 border-red-500/30";
    }
    return interactive
        ? "text-amber-300 bg-amber-500/10 border-amber-500/30 hover:bg-emerald-500/20 hover:text-emerald-300 hover:border-emerald-500/30"
        : "text-amber-300 bg-amber-500/10 border-amber-500/30";
}

export default function HealthStatusBadge(props: HealthStatusBadgeProps) {
    const { status, isBlocker, compact, onClick } = props;
    const isInteractive = !!onClick;

    return (
        <span 
            className={`${compact ? "text-[11px]" : "text-xs"} px-2 py-1 border rounded-md transition-all ${badgeClass(status, isInteractive)} ${isInteractive ? 'cursor-pointer active:scale-95 select-none' : ''}`}
            onClick={onClick}
        >
            {status.toUpperCase()}
            {isBlocker ? " · BLOCKER" : ""}
        </span>
    );
}
