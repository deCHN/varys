interface HealthStatusBadgeProps {
    status: string;
    isBlocker?: boolean;
    compact?: boolean;
}

function badgeClass(status: string): string {
    if (status === "ok") {
        return "text-emerald-300 bg-emerald-500/10 border-emerald-500/30";
    }
    if (status === "missing") {
        return "text-red-300 bg-red-500/10 border-red-500/30";
    }
    return "text-amber-300 bg-amber-500/10 border-amber-500/30";
}

export default function HealthStatusBadge(props: HealthStatusBadgeProps) {
    const { status, isBlocker, compact } = props;
    return (
        <span className={`${compact ? "text-[11px]" : "text-xs"} px-2 py-1 border rounded-md ${badgeClass(status)}`}>
            {status.toUpperCase()}
            {isBlocker ? " · BLOCKER" : ""}
        </span>
    );
}
