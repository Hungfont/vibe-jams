"use client";

import * as React from "react";

import { cn } from "@/lib/utils";

interface DropdownMenuContextValue {
  open: boolean;
  setOpen: (open: boolean) => void;
}

const DropdownMenuContext = React.createContext<DropdownMenuContextValue | null>(null);

export function DropdownMenu({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = React.useState(false);
  return (
    <DropdownMenuContext.Provider value={{ open, setOpen }}>
      <div className="relative inline-flex">{children}</div>
    </DropdownMenuContext.Provider>
  );
}

interface DropdownMenuTriggerProps {
  children: React.ReactNode;
  asChild?: boolean;
}

export function DropdownMenuTrigger({ children, asChild = false }: DropdownMenuTriggerProps) {
  const context = React.useContext(DropdownMenuContext);
  if (!context) {
    return null;
  }

  const handleClick = () => {
    context.setOpen(!context.open);
  };

  if (asChild && React.isValidElement(children)) {
    const child = children as React.ReactElement<{ onClick?: () => void }>;
    return React.cloneElement(child, {
      onClick: () => {
        child.props.onClick?.();
        handleClick();
      },
    });
  }

  return (
    <button type="button" onClick={handleClick}>
      {children}
    </button>
  );
}

export function DropdownMenuContent({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  const context = React.useContext(DropdownMenuContext);
  if (!context || !context.open) {
    return null;
  }

  return (
    <div
      className={cn("absolute right-0 top-8 z-20 min-w-32 rounded-md border border-zinc-700 bg-zinc-950 p-1", className)}
      {...props}
    />
  );
}

export function DropdownMenuItem({ className, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) {
  const context = React.useContext(DropdownMenuContext);
  return (
    <button
      className={cn("flex w-full rounded px-2 py-1.5 text-left text-xs text-zinc-200 hover:bg-zinc-800", className)}
      onClick={(event) => {
        props.onClick?.(event);
        context?.setOpen(false);
      }}
      type="button"
      {...props}
    />
  );
}
