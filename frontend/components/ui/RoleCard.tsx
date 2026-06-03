import { ArrowRight } from "lucide-react";
import { ReactNode } from "react";
import { classNames } from "@/lib/classNames";

interface RoleCardProps {
  title: string;
  description: string;
  icon: ReactNode;
  active?: boolean;
  onClick?: () => void;
}

export default function RoleCard({ title, description, icon, active = false, onClick }: RoleCardProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={classNames(
        "group w-full rounded-2xl border bg-[var(--jc-surface)] p-4 text-left transition",
        active
          ? "border-[var(--jc-accent)] shadow-[0_14px_28px_rgba(19,120,22,0.14)]"
          : "border-[var(--jc-border)] hover:border-[var(--jc-accent)] hover:shadow-[0_10px_20px_rgba(0,0,0,0.06)]"
      )}
    >
      <div className="rounded-xl bg-[linear-gradient(135deg,_#d9efc9_0%,_#bfeac7_100%)] p-5 text-[var(--jc-ink)]">{icon}</div>
      <div className="mt-4 flex items-center justify-between">
        <p className="text-xl font-semibold text-[var(--jc-ink)]">{title}</p>
        <ArrowRight className="h-5 w-5 text-[var(--jc-ink-muted)] transition group-hover:text-[var(--jc-ink)]" />
      </div>
      <p className="mt-1 text-sm text-[var(--jc-ink-muted)]">{description}</p>
    </button>
  );
}

