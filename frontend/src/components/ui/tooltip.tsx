import * as React from "react";

import { cn } from "@/lib/utils";

interface TooltipProps {
  content: string;
  children: React.ReactNode;
  className?: string;
}

export function Tooltip({ content, children, className }: TooltipProps) {
  return (
    <span className={cn("group relative inline-flex", className)}>
      {children}
      <span className="pointer-events-none absolute -top-8 left-1/2 hidden -translate-x-1/2 rounded bg-zinc-800 px-2 py-1 text-[10px] text-zinc-100 group-hover:block">
        {content}
      </span>
    </span>
  );
}
