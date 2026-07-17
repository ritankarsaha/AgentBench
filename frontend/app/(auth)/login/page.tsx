"use client";

import { createClient } from "@/lib/supabase/client";

export default function LoginPage() {
  async function signInWithGoogle() {
    const supabase = createClient();
    const params = new URLSearchParams(window.location.search);
    const next = params.get("next") ?? "/leaderboard";

    await supabase.auth.signInWithOAuth({
      provider: "google",
      options: {
        redirectTo: `${window.location.origin}/callback?next=${encodeURIComponent(next)}`,
      },
    });
  }

  return (
    <div className="flex min-h-full flex-1 flex-col items-center justify-center bg-bg px-6">
      <div className="w-full max-w-sm text-center">
        <h1 className="text-xl font-semibold text-text-primary">
          Sign in to AgentBench
        </h1>
        <p className="mt-2 text-sm text-text-secondary">
          Register benchmark agents and manage your API keys.
        </p>
        <button
          onClick={signInWithGoogle}
          className="mt-6 w-full rounded-row border border-border bg-surface px-4 py-2.5 text-sm font-medium text-text-primary transition-colors hover:border-text-secondary"
        >
          Continue with Google
        </button>
      </div>
    </div>
  );
}
