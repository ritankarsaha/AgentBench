import { formatScore } from "@/lib/format";

type ScoreBarProps = {
  score: number;
  verified: boolean;
};

export function ScoreBar({ score, verified }: ScoreBarProps) {
  const pct = Math.max(0, Math.min(1, score)) * 100;

  return (
    <div className="flex min-w-0 flex-1 items-center gap-3">
      <div
        className="relative h-2.5 flex-1 overflow-hidden rounded-row bg-border"
        role="meter"
        aria-valuenow={Math.round(score * 100)}
        aria-valuemin={0}
        aria-valuemax={100}
      >
        <div
          className="h-full rounded-row border-l border-white/25"
          style={{
            width: `${pct}%`,
            background: verified
              ? "linear-gradient(90deg, var(--accent), var(--accent-verified))"
              : "linear-gradient(90deg, var(--accent), var(--text-muted))",
          }}
        />
      </div>
      <span className="w-14 shrink-0 text-right font-mono text-sm font-semibold tabular-nums text-text-primary">
        {formatScore(score)}
      </span>
    </div>
  );
}
