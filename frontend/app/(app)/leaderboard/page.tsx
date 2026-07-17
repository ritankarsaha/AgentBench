import Link from "next/link";
import { getLeaderboard, SUITES } from "@/lib/api";
import { LeaderboardTable } from "@/components/leaderboard/LeaderboardTable";

export const metadata = {
  title: "Leaderboard — AgentBench",
};

type PageProps = {
  searchParams: Promise<{ suite?: string }>;
};

export default async function LeaderboardPage({ searchParams }: PageProps) {
  const params = await searchParams;
  const activeSuite = SUITES.find((s) => s.slug === params.suite) ?? SUITES[0];

  const rows = await getLeaderboard(activeSuite.slug);

  return (
    <div>
      <h1 className="text-2xl font-semibold tracking-tight text-text-primary">
        Leaderboard
      </h1>
      <p className="mt-1 text-sm text-text-secondary">
        Ranked by effective score — decayed over 30 days, never fully expires.
      </p>

      <div className="mt-6 flex gap-1 overflow-x-auto border-b border-border">
        {SUITES.map((suite) => {
          const active = suite.slug === activeSuite.slug;
          return (
            <Link
              key={suite.slug}
              href={suite.slug === "standard" ? "/leaderboard" : `/leaderboard?suite=${suite.slug}`}
              className={
                "whitespace-nowrap border-b-2 px-3 py-2 text-sm transition-colors " +
                (active
                  ? "border-accent text-text-primary"
                  : "border-transparent text-text-secondary hover:text-text-primary")
              }
              aria-current={active ? "page" : undefined}
            >
              {suite.label}
            </Link>
          );
        })}
      </div>

      <div className="mt-6">
        <LeaderboardTable rows={rows} suiteLabel={activeSuite.label} />
      </div>
    </div>
  );
}
