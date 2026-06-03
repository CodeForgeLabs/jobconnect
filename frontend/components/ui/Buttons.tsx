import { ButtonHTMLAttributes } from "react";
import { classNames } from "@/lib/classNames";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  wide?: boolean;
}

export function PrimaryButton({ className, wide, ...props }: ButtonProps) {
  return (
    <button
      {...props}
      className={classNames(
        "inline-flex h-12 items-center justify-center rounded-full bg-[var(--jc-accent)] px-6 text-sm font-semibold text-white transition hover:bg-[var(--jc-accent-strong)] disabled:cursor-not-allowed disabled:opacity-55",
        wide ? "w-full" : "",
        className
      )}
    />
  );
}

export function SecondaryButton({ className, wide, ...props }: ButtonProps) {
  return (
    <button
      {...props}
      className={classNames(
        "inline-flex h-12 items-center justify-center rounded-full border border-[var(--jc-border-strong)] bg-[var(--jc-surface)] px-6 text-sm font-semibold text-[var(--jc-ink)] transition hover:bg-[var(--jc-surface-raised)] disabled:cursor-not-allowed disabled:opacity-55",
        wide ? "w-full" : "",
        className
      )}
    />
  );
}

