import * as React from "react";

import { cn } from "@/lib/utils";

export function ScrollArea({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("max-h-[360px] overflow-auto pr-2", className)} {...props} />;
}
