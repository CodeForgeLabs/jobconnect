import { UploadCloud } from "lucide-react";
import { ChangeEvent } from "react";
import { classNames } from "@/lib/classNames";

interface UploadDropzoneProps {
  label: string;
  helper: string;
  accept: string;
  fileName?: string;
  onChange: (event: ChangeEvent<HTMLInputElement>) => void;
  className?: string;
  disabled?: boolean;
}

export default function UploadDropzone({
  label,
  helper,
  accept,
  fileName,
  onChange,
  className,
  disabled = false,
}: UploadDropzoneProps) {
  function handleInputChange(event: ChangeEvent<HTMLInputElement>) {
    if (disabled) return;
    onChange(event);
    // Allow selecting the same file again (e.g. after delete/reset).
    event.currentTarget.value = "";
  }

  return (
    <label
      className={classNames(
        "flex cursor-pointer flex-col items-center justify-center rounded-2xl border border-dashed border-[var(--jc-border-strong)] bg-[var(--jc-surface-alt)] px-6 py-10 text-center transition hover:border-[var(--jc-accent)] hover:bg-[var(--jc-surface-raised)]",
        disabled ? "cursor-not-allowed border-[var(--jc-border)] opacity-70 hover:border-[var(--jc-border)] hover:bg-[var(--jc-surface-alt)]" : "",
        className
      )}
    >
      <input type="file" accept={accept} className="hidden" onChange={handleInputChange} disabled={disabled} />
      <UploadCloud className="h-8 w-8 text-[var(--jc-accent)]" />
      <p className="mt-3 text-base font-semibold text-[var(--jc-ink)]">{label}</p>
      <p className="mt-1 text-xs text-[var(--jc-ink-muted)]">{helper}</p>
      {fileName ? <p className="mt-4 text-sm font-medium text-[var(--jc-ink)]">{fileName}</p> : null}
    </label>
  );
}

