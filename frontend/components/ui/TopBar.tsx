import Link from "next/link";
import { ReactNode } from "react";
import { classNames } from "@/lib/classNames";

interface TopBarProps {
  left?: ReactNode;
  center?: ReactNode;
  right?: ReactNode;
  className?: string;
  transparent?: boolean;
}

export default function TopBar({ left, center, right, className, transparent = false }: TopBarProps) {
  const hasCustomSlots = Boolean(left || center);

  return (
    <header
      className={classNames(
        "border-b border-[var(--jc-border)]",
        transparent ? "bg-transparent" : "bg-[var(--jc-surface)]",
        className
      )}
    >
      {hasCustomSlots ? (
        <div className="mx-auto flex min-h-16 w-full max-w-7xl flex-wrap items-center gap-3 px-4 py-2 md:px-8">
          <div className="flex min-w-0 items-center gap-3">{left}</div>
          <div className="flex min-w-0 flex-1 items-center justify-center">{center}</div>
          <div className="ml-auto flex min-w-0 items-center gap-3 text-sm text-[var(--jc-ink-muted)]">{right}</div>
        </div>
      ) : (
        <div className="mx-auto flex h-16 w-full max-w-7xl items-center justify-between px-4 md:px-8">
          <Link href="/" className="text-[2rem] font-semibold leading-none tracking-tight text-[var(--jc-ink)]">
            jobconnect
          </Link>
          <div className="flex items-center gap-3 text-sm text-[var(--jc-ink-muted)]">{right}</div>
        </div>
      )}
    </header>
  );
}

