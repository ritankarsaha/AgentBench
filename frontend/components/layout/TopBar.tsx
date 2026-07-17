import Link from "next/link";
import { createClient } from "@/lib/supabase/server";

export async function TopBar() {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  return (
    <header className="border-b border-border bg-surface">
      <div className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4">
        <Link href="/" className="font-mono text-sm font-semibold tracking-tight text-text-primary">
          agent<span className="text-accent">bench</span>
        </Link>
        <nav className="flex flex-1 items-center gap-4 text-sm">
          <Link href="/leaderboard" className="text-text-primary">
            Leaderboard
          </Link>
          {user && (
            <Link href="/settings/agents" className="text-text-primary">
              My Agents
            </Link>
          )}
        </nav>
        {user ? (
          <form action="/signout" method="post">
            <button
              type="submit"
              className="text-sm text-text-secondary transition-colors hover:text-text-primary"
            >
              Sign out
            </button>
          </form>
        ) : (
          <Link
            href="/login"
            className="rounded-row border border-border px-3 py-1.5 text-sm font-medium text-text-primary transition-colors hover:border-text-secondary"
          >
            Sign in
          </Link>
        )}
      </div>
    </header>
  );
}
