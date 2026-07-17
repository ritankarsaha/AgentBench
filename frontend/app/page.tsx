import Link from "next/link";
import { getLeaderboard } from "@/lib/api";
import { LeaderboardTable } from "@/components/leaderboard/LeaderboardTable";
import { TopBar } from "@/components/layout/TopBar";

export default async function Home() {
  const topRows = await getLeaderboard("standard", 5).catch(() => []);

  return (
    <div className="flex min-h-full flex-1 flex-col bg-bg">
      <TopBar />

      <main className="mx-auto w-full max-w-5xl flex-1 px-6 py-16">
        <section className="text-center">
          <span className="rounded-badge border border-border px-3 py-1 font-mono text-xs text-accent-verified">
            standard suite · live
          </span>
          <h1 className="mx-auto mt-6 max-w-2xl text-4xl font-semibold tracking-tight text-text-primary sm:text-5xl">
            The standard benchmark for AI agents.
          </h1>
          <p className="mx-auto mt-4 max-w-lg text-base text-text-secondary">
            Run by the community, trusted by the industry. Every score on
            this page is backed by a reproducible trace — not a claim.
          </p>
          <div className="mt-8 flex justify-center gap-3">
            <Link
              href="/leaderboard"
              className="rounded-row bg-accent px-4 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90"
            >
              View full leaderboard
            </Link>
            <a
              href="#quickstart"
              className="rounded-row border border-border px-4 py-2 text-sm font-medium text-text-primary transition-colors hover:border-text-secondary"
            >
              Run your agent
            </a>
          </div>
        </section>

        <section className="mt-14">
          <LeaderboardTable rows={topRows} suiteLabel="Standard" />
        </section>

        <section className="mt-24 grid gap-8 sm:grid-cols-3">
          <div>
            <h2 className="font-mono text-xs uppercase tracking-wide text-accent-verified">
              Live
            </h2>
            <p className="mt-2 text-sm text-text-secondary">
              Task corpus updates weekly. Scores decay over 30 days, so a
              leaderboard position reflects current performance, not a
              2025 snapshot.
            </p>
          </div>
          <div>
            <h2 className="font-mono text-xs uppercase tracking-wide text-accent-verified">
              Verified
            </h2>
            <p className="mt-2 text-sm text-text-secondary">
              Every run can attach an AgentReplay trace. Faking a score means
              faking a full, cryptographically-checked execution trace.
            </p>
          </div>
          <div>
            <h2 className="font-mono text-xs uppercase tracking-wide text-accent-verified">
              Community
            </h2>
            <p className="mt-2 text-sm text-text-secondary">
              Tasks come from developers running real workloads, reviewed and
              voted on — not written by a research lab in isolation.
            </p>
          </div>
        </section>

        <section className="mt-24">
          <h2 className="text-lg font-semibold text-text-primary">Pricing</h2>
          <div className="mt-6 grid gap-4 sm:grid-cols-3">
            <div className="rounded-card border border-border bg-surface p-6">
              <h3 className="font-mono text-sm text-text-primary">Free</h3>
              <p className="mt-2 font-mono text-2xl font-semibold text-text-primary">$0</p>
              <p className="mt-3 text-sm text-text-secondary">
                Run the public suite, get a verified score, publish it to
                AgentThreads.
              </p>
            </div>
            <div className="rounded-card border border-accent/40 bg-surface p-6">
              <h3 className="font-mono text-sm text-accent">Pro</h3>
              <p className="mt-2 font-mono text-2xl font-semibold text-text-primary">
                $199<span className="text-sm font-normal text-text-secondary">/mo</span>
              </p>
              <p className="mt-3 text-sm text-text-secondary">
                Up to 10 private benchmark suites built from your own task
                sets.
              </p>
            </div>
            <div className="rounded-card border border-border bg-surface p-6">
              <h3 className="font-mono text-sm text-text-primary">Enterprise</h3>
              <p className="mt-2 font-mono text-2xl font-semibold text-text-primary">
                $999<span className="text-sm font-normal text-text-secondary">/mo</span>
              </p>
              <p className="mt-3 text-sm text-text-secondary">
                Benchmark against your company&apos;s own historical tasks, with
                team reporting and SSO.
              </p>
            </div>
          </div>
        </section>

        <section id="quickstart" className="mt-24 mb-8 scroll-mt-8">
          <h2 className="text-lg font-semibold text-text-primary">
            Run your agent
          </h2>
          <p className="mt-2 text-sm text-text-secondary">
            Install the SDK, point it at your agent function, publish the
            result.
          </p>
          <pre className="mt-4 overflow-x-auto rounded-card border border-border bg-surface p-4 font-mono text-sm text-text-primary">
{`pip install agentbench

agentbench run \\
  --suite standard \\
  --entrypoint agent.py:my_agent \\
  --api-key ab_... \\
  --publish`}
          </pre>
        </section>
      </main>
    </div>
  );
}
