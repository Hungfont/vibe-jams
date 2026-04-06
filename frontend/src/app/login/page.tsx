import { Space_Grotesk } from "next/font/google";

import { LoginForm } from "@/components/auth/login-form";

const loginFont = Space_Grotesk({
  subsets: ["latin"],
  weight: ["500", "700"],
  variable: "--font-login",
});

export default function LoginPage() {
  return (
    <main className={`${loginFont.variable} relative isolate flex min-h-screen items-center justify-center overflow-hidden bg-[radial-gradient(circle_at_20%_10%,#133b2b_0%,#07120d_28%,#030303_65%)] px-6 py-10`}>
      <div className="pointer-events-none absolute -left-24 top-10 h-72 w-72 rounded-full bg-emerald-500/20 blur-3xl" />
      <div className="pointer-events-none absolute -right-16 bottom-16 h-80 w-80 rounded-full bg-zinc-600/15 blur-3xl" />
      <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(130deg,transparent_0%,rgba(255,255,255,0.02)_42%,transparent_76%)]" />

      <div className="relative w-full max-w-md">
        <p className="mb-4 text-center text-xs font-semibold uppercase tracking-[0.26em] text-zinc-400">Jam Sign In</p>
        <LoginForm />
      </div>
    </main>
  );
}