import { ReactNode } from "react";
import { classNames } from "@/lib/classNames";

interface SectionCardProps {
  title?: string;
  subtitle?: string;
  action?: ReactNode;
  children: ReactNode;
  className?: string;
}

export default function SectionCard({ title, subtitle, action, children, className }: SectionCardProps) {
  return (
    <section className={classNames("rounded-3xl border border-[var(--jc-border)] bg-[var(--jc-surface)] p-6 md:p-8", className)}>
      {(title || subtitle || action) && (
        <header className="mb-6 flex flex-wrap items-start justify-between gap-4">
          <div>
            {title ? <h2 className="text-2xl font-semibold text-[var(--jc-ink)]">{title}</h2> : null}
            {subtitle ? <p className="mt-1 text-sm text-[var(--jc-ink-muted)]">{subtitle}</p> : null}
          </div>
          {action}
        </header>
      )}
      {children}
    </section>
  );
}

