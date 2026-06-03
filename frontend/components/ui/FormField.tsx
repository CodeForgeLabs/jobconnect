import { InputHTMLAttributes, ReactNode, TextareaHTMLAttributes } from "react";
import { classNames } from "@/lib/classNames";

interface FieldBaseProps {
  label: string;
  hint?: string;
  error?: string | null;
  required?: boolean;
  children: ReactNode;
}

function FieldFrame({ label, hint, error, required, children }: FieldBaseProps) {
  return (
    <label className="block">
      <span className="mb-1.5 block text-sm font-medium text-[var(--jc-ink)]">
        {label}
        {required ? <span className="ml-1 text-[var(--jc-accent)]">*</span> : null}
      </span>
      {children}
      {error ? (
        <span className="mt-1.5 block text-xs text-[var(--jc-danger)]">{error}</span>
      ) : hint ? (
        <span className="mt-1.5 block text-xs text-[var(--jc-ink-muted)]">{hint}</span>
      ) : null}
    </label>
  );
}

export interface InputFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  hint?: string;
  error?: string | null;
}

export function InputField({ label, hint, error, className, ...props }: InputFieldProps) {
  return (
    <FieldFrame label={label} hint={hint} error={error} required={props.required}>
      <input
        {...props}
        className={classNames(
          "h-12 w-full rounded-xl border border-[var(--jc-border)] bg-[var(--jc-surface)] px-4 text-[var(--jc-ink)] outline-none transition placeholder:text-[var(--jc-ink-muted)] focus:border-[var(--jc-accent)] focus:ring-2 focus:ring-[var(--jc-accent-soft)]",
          "disabled:cursor-not-allowed disabled:border-[var(--jc-border)] disabled:bg-[var(--jc-surface-alt)] disabled:text-[var(--jc-ink-muted)] disabled:opacity-80",
          error ? "border-[var(--jc-danger)] focus:ring-[rgba(193,63,63,0.16)]" : "",
          className
        )}
      />
    </FieldFrame>
  );
}

export interface TextAreaFieldProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label: string;
  hint?: string;
  error?: string | null;
}

export function TextAreaField({ label, hint, error, className, ...props }: TextAreaFieldProps) {
  return (
    <FieldFrame label={label} hint={hint} error={error} required={props.required}>
      <textarea
        {...props}
        className={classNames(
          "w-full rounded-xl border border-[var(--jc-border)] bg-[var(--jc-surface)] px-4 py-3 text-[var(--jc-ink)] outline-none transition placeholder:text-[var(--jc-ink-muted)] focus:border-[var(--jc-accent)] focus:ring-2 focus:ring-[var(--jc-accent-soft)]",
          "disabled:cursor-not-allowed disabled:border-[var(--jc-border)] disabled:bg-[var(--jc-surface-alt)] disabled:text-[var(--jc-ink-muted)] disabled:opacity-80",
          error ? "border-[var(--jc-danger)] focus:ring-[rgba(193,63,63,0.16)]" : "",
          className
        )}
      />
    </FieldFrame>
  );
}

