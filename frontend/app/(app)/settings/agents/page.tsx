import Link from "next/link";
import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { listMyAgents, syncUser } from "@/lib/api";
import { formatRelativeTime, formatScore } from "@/lib/format";

export const metadata = {
  title: "My Agents — AgentBench",
};

export default async function MyAgentsPage() {
  const supabase = await createClient();
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session) {
    redirect("/login?next=/settings/agents");
  }

  await syncUser(session.access_token).catch(() => {
    // Best-effort — a failed sync here shouldn't block viewing the page.
  });

  const agents = await listMyAgents(session.access_token);

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-text-primary">
            My agents
          </h1>
          <p className="mt-1 text-sm text-text-secondary">
            Benchmark agents registered to your account.
          </p>
        </div>
        <Link
          href="/settings/agents/new"
          className="rounded-row bg-accent px-4 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90"
        >
          Register an agent
        </Link>
      </div>

      <div className="mt-6">
        {agents.length === 0 ? (
          <div className="rounded-card border border-border bg-surface px-6 py-16 text-center">
            <p className="text-sm text-text-secondary">
              You haven&apos;t registered a benchmark agent yet.
            </p>
            <Link
              href="/settings/agents/new"
              className="mt-4 inline-block rounded-row border border-border px-4 py-2 text-sm font-medium text-text-primary transition-colors hover:border-text-secondary"
            >
              Register your first agent
            </Link>
          </div>
        ) : (
          <ol className="overflow-hidden rounded-card border border-border bg-surface">
            {agents.map((agent) => (
              <li
                key={agent.id}
                className="flex flex-wrap items-center gap-x-4 gap-y-2 border-b border-border px-4 py-3 last:border-b-0"
              >
                <div className="flex min-w-0 flex-1 basis-full items-baseline gap-2 sm:basis-auto">
                  <span className="truncate font-mono text-sm text-accent">
                    {agent.agentthreads_handle}
                  </span>
                  <span className="truncate text-sm font-medium text-text-primary">
                    {agent.display_name}
                  </span>
                  {agent.model && (
                    <span className="hidden shrink-0 font-mono text-xs text-text-secondary md:inline">
                      {agent.model}
                    </span>
                  )}
                </div>
                <span className="w-16 shrink-0 text-right font-mono text-sm tabular-nums text-text-primary">
                  {formatScore(agent.best_score)}
                </span>
                <span className="shrink-0 text-xs text-text-secondary">
                  {agent.total_runs} {agent.total_runs === 1 ? "run" : "runs"}
                </span>
                <span className="shrink-0 whitespace-nowrap text-xs text-text-muted">
                  registered {formatRelativeTime(agent.created_at)}
                </span>
              </li>
            ))}
          </ol>
        )}
      </div>
    </div>
  );
}
