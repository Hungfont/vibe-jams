"use client";

import * as React from "react";
import { useRouter } from "next/navigation";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { ToastStack } from "@/components/ui/toast";
import { useToast } from "@/components/ui/use-toast";
import { loginWithPassword } from "@/lib/auth/client";
import { loginRequestSchema } from "@/lib/auth/schemas";

interface FieldErrors {
  identity?: string;
  password?: string;
}

export function LoginForm() {
  const router = useRouter();
  const { items, toast } = useToast();
  const [identity, setIdentity] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [fieldErrors, setFieldErrors] = React.useState<FieldErrors>({});
  const [formError, setFormError] = React.useState("");
  const [isSubmitting, setIsSubmitting] = React.useState(false);

  const onSubmit = React.useCallback(async () => {
    setIsSubmitting(true);
    setFormError("");
    setFieldErrors({});

    const parsed = loginRequestSchema.safeParse({ identity, password });
    if (!parsed.success) {
      const flattened = parsed.error.flatten().fieldErrors;
      setFieldErrors({
        identity: flattened.identity?.[0],
        password: flattened.password?.[0],
      });
      setIsSubmitting(false);
      return;
    }

    const result = await loginWithPassword(parsed.data);
    if (!result.success || !result.data) {
      const message = result.error?.message ?? "Unable to sign in";
      setFormError(message);
      toast({ title: "Sign in failed", description: message, variant: "error" });
      setIsSubmitting(false);
      return;
    }

    toast({ title: "Signed in", description: `Welcome back, ${result.data.claims.userId}` });
    router.push("/");
  }, [identity, password, router, toast]);

  return (
    <>
      <Card className="w-full max-w-md overflow-hidden border-zinc-800/80 bg-zinc-950/95 shadow-[0_20px_80px_rgba(0,0,0,0.45)]">
        <CardHeader className="space-y-3 px-7 pt-8">
          <div className="inline-flex h-11 w-11 items-center justify-center rounded-full bg-emerald-500 text-base font-black text-black">
            S
          </div>
          <CardTitle className="font-[family-name:var(--font-login)] text-2xl tracking-tight text-zinc-50">
            Log in to your music
          </CardTitle>
          <CardDescription className="text-sm text-zinc-400">
            Pick up where you left off and jump back into your jam sessions.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4 px-7 pb-7">
          {formError ? (
            <Alert className="border-rose-500/50 bg-rose-950/40">
              <AlertTitle>Sign in blocked</AlertTitle>
              <AlertDescription>{formError}</AlertDescription>
            </Alert>
          ) : null}

          <div className="space-y-2">
            <label htmlFor="identity" className="text-xs font-semibold uppercase tracking-[0.18em] text-zinc-400">
              Email
            </label>
            <Input
              id="identity"
              type="email"
              autoComplete="email"
              value={identity}
              onChange={(event) => setIdentity(event.target.value)}
              placeholder="you@example.com"
              aria-invalid={Boolean(fieldErrors.identity)}
            />
            {fieldErrors.identity ? <p className="text-xs text-rose-300">{fieldErrors.identity}</p> : null}
          </div>

          <div className="space-y-2">
            <label htmlFor="password" className="text-xs font-semibold uppercase tracking-[0.18em] text-zinc-400">
              Password
            </label>
            <Input
              id="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="Your password"
              aria-invalid={Boolean(fieldErrors.password)}
            />
            {fieldErrors.password ? <p className="text-xs text-rose-300">{fieldErrors.password}</p> : null}
          </div>

          <Button className="mt-1 w-full rounded-full text-sm font-bold uppercase tracking-[0.16em]" disabled={isSubmitting} onClick={() => void onSubmit()}>
            {isSubmitting ? "Signing in..." : "Sign in"}
          </Button>

          <div className="flex items-center gap-3 py-1">
            <Separator className="bg-zinc-800" />
            <span className="text-[10px] font-semibold uppercase tracking-[0.24em] text-zinc-500">or</span>
            <Separator className="bg-zinc-800" />
          </div>

          <Button variant="outline" className="w-full border-zinc-700 text-zinc-200" disabled>
            Continue with Google
          </Button>
          <Button variant="outline" className="w-full border-zinc-700 text-zinc-200" disabled>
            Continue with Apple
          </Button>
        </CardContent>
      </Card>

      <ToastStack items={items} />
    </>
  );
}