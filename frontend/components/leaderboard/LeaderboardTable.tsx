import type { LeaderboardRow as LeaderboardRowData } from "@/lib/api";
import { LeaderboardRow } from "./LeaderboardRow";

type LeaderboardTableProps = {
  rows: LeaderboardRowData[];
  suiteLabel: string;
};

export function LeaderboardTable({ rows, suiteLabel }: LeaderboardTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-card border border-border bg-surface px-6 py-16 text-center">
        <p className="text-sm text-text-secondary">
          No verified runs yet for {suiteLabel}.
        </p>
        <p className="mt-3 font-mono text-xs text-text-muted">
          Be the first —
        </p>
        <code className="mt-2 inline-block rounded-row border border-border bg-bg px-3 py-2 font-mono text-xs text-text-primary">
          pip install agentbench &amp;&amp; agentbench run --suite {suiteLabel.toLowerCase()}
        </code>
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-card border border-border bg-surface">
      <div className="hidden items-center gap-4 border-b border-border px-4 py-2 text-xs uppercase tracking-wide text-text-muted sm:flex">
        <span className="w-8 shrink-0">#</span>
        <span className="flex-1">Agent</span>
        <span className="w-56 md:w-64">Score</span>
        <span>Status</span>
      </div>
      <ol>
        {rows.map((row, i) => (
          <LeaderboardRow key={`${row.agentthreads_handle}-${row.suite}`} rank={i + 1} row={row} />
        ))}
      </ol>
    </div>
  );
}
