"use client";

import * as React from "react";

import type { ToastItem } from "@/components/ui/toast";

export function useToast() {
  const [items, setItems] = React.useState<ToastItem[]>([]);

  const push = React.useCallback((item: Omit<ToastItem, "id">) => {
    const next: ToastItem = {
      id: crypto.randomUUID(),
      ...item,
    };

    setItems((current) => [...current, next]);
    setTimeout(() => {
      setItems((current) => current.filter((value) => value.id !== next.id));
    }, 3500);
  }, []);

  return {
    items,
    toast: push,
  };
}
