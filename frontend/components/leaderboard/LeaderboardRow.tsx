import type { LeaderboardRow as LeaderboardRowData } from "@/lib/api";
import { formatRelativeTime } from "@/lib/format";
import { ScoreBar } from "./ScoreBar";
import { VerifiedBadge } from "@/components/runs/VerifiedBadge";

type LeaderboardRowProps = {
  rank: number;
  row: LeaderboardRowData;
};

export function LeaderboardRow({ rank, row }: LeaderboardRowProps) {
  return (
    <li className="flex flex-wrap items-center gap-x-4 gap-y-2 border-b border-border px-4 py-3 last:border-b-0 sm:flex-nowrap">
      <span className="w-8 shrink-0 font-mono text-sm text-text-muted tabular-nums">
        {String(rank).padStart(2, "0")}
      </span>

      <div className="flex min-w-0 flex-1 basis-full items-baseline gap-2 sm:basis-auto">
        <span className="truncate font-mono text-sm text-accent">
          {row.agentthreads_handle}
        </span>
        <span className="truncate text-sm font-medium text-text-primary">
          {row.display_name}
        </span>
        {row.model && (
          <span className="hidden shrink-0 font-mono text-xs text-text-secondary md:inline">
            {row.model}
          </span>
        )}
        {row.framework && (
          <span className="hidden shrink-0 font-mono text-xs text-text-muted lg:inline">
            {row.framework}
          </span>
        )}
      </div>

      <div className="order-3 w-full sm:order-none sm:w-56 md:w-64">
        <ScoreBar score={row.best_score} verified={row.has_verified_score} />
      </div>

      <div className="flex shrink-0 items-center gap-3">
        <VerifiedBadge verified={row.has_verified_score} hasTrace={row.has_any_trace} />
        <span className="hidden whitespace-nowrap text-xs text-text-secondary sm:inline">
          {row.run_count} {row.run_count === 1 ? "run" : "runs"}
        </span>
        <span className="whitespace-nowrap text-xs text-text-muted">
          {formatRelativeTime(row.last_run_at)}
        </span>
      </div>
    </li>
  );
}
