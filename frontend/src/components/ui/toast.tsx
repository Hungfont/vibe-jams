import * as React from "react";

import { cn } from "@/lib/utils";

export interface ToastItem {
  id: string;
  title: string;
  description?: string;
  variant?: "default" | "error";
}

interface ToastStackProps {
  items: ToastItem[];
}

export function ToastStack({ items }: ToastStackProps) {
  if (items.length === 0) {
    return null;
  }

  return (
    <div className="fixed bottom-4 right-4 z-50 flex w-80 flex-col gap-2">
      {items.map((item) => (
        <div
          key={item.id}
          className={cn(
            "rounded-md border px-3 py-2 text-sm",
            item.variant === "error"
              ? "border-rose-500/40 bg-rose-950/60 text-rose-100"
              : "border-zinc-700 bg-zinc-900 text-zinc-100",
          )}
        >
          <p className="font-semibold">{item.title}</p>
          {item.description ? <p className="text-xs text-zinc-300">{item.description}</p> : null}
        </div>
      ))}
    </div>
  );
}
