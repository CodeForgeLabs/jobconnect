import { CheckCircle2, CircleAlert, Info } from "lucide-react";
import { ReactNode } from "react";
import { classNames } from "@/lib/classNames";

type AlertTone = "error" | "success" | "info" | "warning";

interface InlineAlertProps {
  tone: AlertTone;
  children: ReactNode;
  className?: string;
}

const toneClassMap: Record<AlertTone, string> = {
  error: "border-[rgba(193,63,63,0.35)] bg-[rgba(193,63,63,0.08)] text-[var(--jc-danger)]",
  success: "border-[rgba(17,123,34,0.28)] bg-[rgba(17,123,34,0.08)] text-[var(--jc-accent-strong)]",
  info: "border-[rgba(53,88,74,0.28)] bg-[rgba(53,88,74,0.08)] text-[var(--jc-ink)]",
  warning: "border-[rgba(176,120,27,0.35)] bg-[rgba(176,120,27,0.11)] text-[var(--jc-warning)]",
};

const toneIconMap: Record<AlertTone, ReactNode> = {
  error: <CircleAlert className="mt-0.5 h-4 w-4" />,
  success: <CheckCircle2 className="mt-0.5 h-4 w-4" />,
  info: <Info className="mt-0.5 h-4 w-4" />,
  warning: <CircleAlert className="mt-0.5 h-4 w-4" />,
};

export default function InlineAlert({ tone, children, className }: InlineAlertProps) {
  return (
    <div className={classNames("flex items-start gap-2 rounded-xl border px-3 py-2.5 text-sm", toneClassMap[tone], className)}>
      {toneIconMap[tone]}
      <span>{children}</span>
    </div>
  );
}

