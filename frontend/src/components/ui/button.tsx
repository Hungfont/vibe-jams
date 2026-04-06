import * as React from "react";

import { cn } from "@/lib/utils";

type ButtonVariant = "default" | "secondary" | "outline" | "ghost" | "destructive";
type ButtonSize = "default" | "sm" | "lg";

const buttonVariantClasses: Record<ButtonVariant, string> = {
  default: "bg-emerald-500 text-black hover:bg-emerald-400",
  secondary: "bg-zinc-800 text-zinc-100 hover:bg-zinc-700",
  outline: "border border-zinc-700 bg-transparent text-zinc-200 hover:bg-zinc-900",
  ghost: "bg-transparent text-zinc-200 hover:bg-zinc-900",
  destructive: "bg-rose-600 text-white hover:bg-rose-500",
};

const buttonSizeClasses: Record<ButtonSize, string> = {
  default: "h-10 px-4 py-2 text-sm",
  sm: "h-8 px-3 py-1.5 text-xs",
  lg: "h-11 px-6 py-2.5 text-base",
};

export function buttonVariants(variant: ButtonVariant, size: ButtonSize, className?: string): string {
  return cn(
    "inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500/60 disabled:pointer-events-none disabled:opacity-50",
    buttonVariantClasses[variant],
    buttonSizeClasses[size],
    className,
  );
}

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
}

export function Button({
  className,
  variant = "default",
  size = "default",
  ...props
}: ButtonProps) {
  return <button className={buttonVariants(variant, size, className)} {...props} />;
}
