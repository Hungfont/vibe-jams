"use client";

import * as React from "react";

import { cn } from "@/lib/utils";

interface DialogContextValue {
  open: boolean;
  setOpen: (open: boolean) => void;
}

const DialogContext = React.createContext<DialogContextValue | null>(null);

interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
}

export function Dialog({ open, onOpenChange, children }: DialogProps) {
  return <DialogContext.Provider value={{ open, setOpen: onOpenChange }}>{children}</DialogContext.Provider>;
}

interface DialogTriggerProps {
  children: React.ReactNode;
  asChild?: boolean;
}

export function DialogTrigger({ children, asChild = false }: DialogTriggerProps) {
  const context = React.useContext(DialogContext);
  if (!context) {
    return null;
  }

  if (asChild && React.isValidElement(children)) {
    const child = children as React.ReactElement<{ onClick?: () => void }>;
    return React.cloneElement(child, {
      onClick: () => {
        child.props.onClick?.();
        context.setOpen(true);
      },
    });
  }

  return (
    <button type="button" onClick={() => context.setOpen(true)}>
      {children}
    </button>
  );
}

export function DialogContent({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  const context = React.useContext(DialogContext);
  if (!context || !context.open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-black/70 p-4">
      <div className={cn("w-full max-w-md rounded-lg border border-zinc-700 bg-zinc-950 p-4", className)} {...props} />
    </div>
  );
}

export function DialogHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("mb-3", className)} {...props} />;
}

export function DialogTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h3 className={cn("text-base font-semibold text-zinc-100", className)} {...props} />;
}

export function DialogDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn("mt-1 text-sm text-zinc-400", className)} {...props} />;
}

export function DialogFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("mt-4 flex items-center justify-end gap-2", className)} {...props} />;
}

interface DialogCloseProps {
  children: React.ReactNode;
  asChild?: boolean;
}

export function DialogClose({ children, asChild = false }: DialogCloseProps) {
  const context = React.useContext(DialogContext);
  if (!context) {
    return null;
  }

  if (asChild && React.isValidElement(children)) {
    const child = children as React.ReactElement<{ onClick?: () => void }>;
    return React.cloneElement(child, {
      onClick: () => {
        child.props.onClick?.();
        context.setOpen(false);
      },
    });
  }

  return (
    <button type="button" onClick={() => context.setOpen(false)}>
      {children}
    </button>
  );
}
