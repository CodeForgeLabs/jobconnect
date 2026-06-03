import { classNames } from "@/lib/classNames";

type StatusTone = "neutral" | "success" | "warning" | "danger";

interface StatusBadgeProps {
  label: string;
  tone?: StatusTone;
  className?: string;
}

const toneClassMap: Record<StatusTone, string> = {
  neutral: "bg-[var(--jc-surface-alt)] text-[var(--jc-ink-muted)]",
  success: "bg-[rgba(17,123,34,0.12)] text-[var(--jc-accent-strong)]",
  warning: "bg-[rgba(176,120,27,0.12)] text-[var(--jc-warning)]",
  danger: "bg-[rgba(193,63,63,0.12)] text-[var(--jc-danger)]",
};

export default function StatusBadge({ label, tone = "neutral", className }: StatusBadgeProps) {
  return (
    <span
      className={classNames(
        "inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide",
        toneClassMap[tone],
        className
      )}
    >
      {label}
    </span>
  );
}

