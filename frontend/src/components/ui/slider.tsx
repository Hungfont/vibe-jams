import * as React from "react";

import { cn } from "@/lib/utils";

export type SliderProps = React.InputHTMLAttributes<HTMLInputElement>;

export function Slider({ className, ...props }: SliderProps) {
  return (
    <input
      type="range"
      className={cn("h-2 w-full cursor-pointer appearance-none rounded-lg bg-zinc-800", className)}
      {...props}
    />
  );
}
