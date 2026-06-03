import { ReactNode } from "react";
import { classNames } from "@/lib/classNames";

interface PageShellProps {
  children: ReactNode;
  className?: string;
  contentClassName?: string;
}

export default function PageShell({ children, className, contentClassName }: PageShellProps) {
  return (
    <main
      className={classNames(
        "min-h-[calc(100vh-64px)] bg-[radial-gradient(circle_at_top,_#f4f8f3_0%,_#f7f8f6_45%,_#f4f5f3_100%)] px-4 py-8 md:px-8 md:py-10",
        className
      )}
    >
      <div className={classNames("mx-auto w-full max-w-6xl", contentClassName)}>{children}</div>
    </main>
  );
}

