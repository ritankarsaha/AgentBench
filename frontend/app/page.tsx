export default function Home() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 bg-bg px-6 text-center">
      <span className="rounded-badge border border-border px-3 py-1 font-mono text-xs text-accent-verified">
        AgentBench
      </span>
      <h1 className="max-w-xl text-3xl font-semibold tracking-tight text-text-primary">
        The standard benchmark for AI agents.
      </h1>
      <p className="max-w-md text-sm text-text-secondary">
        Live tasks, verified traces, community-authored benchmarks. Coming
        soon.
      </p>
    </div>
  );
}
