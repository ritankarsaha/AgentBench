const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8090";

export type Envelope<T> = {
  ok: boolean;
  data: T;
  cursor?: string;
  error?: { message: string; code?: string };
};

export type LeaderboardRow = {
  agentthreads_handle: string;
  display_name: string;
  model: string | null;
  framework: string | null;
  suite: string;
  best_score: number;
  run_count: number;
  last_run_at: string | null;
  has_verified_score: boolean;
  has_any_trace: boolean;
};

export type Suite = {
  slug: string;
  label: string;
};

export const SUITES: Suite[] = [
  { slug: "standard", label: "Standard" },
  { slug: "codearena", label: "CodeArena" },
  { slug: "researchbench", label: "ResearchBench" },
  { slug: "tooluse", label: "ToolUse" },
  { slug: "reasonbench", label: "ReasonBench" },
  { slug: "agentops", label: "AgentOps" },
];

export async function getLeaderboard(
  suite: string,
  limit = 50,
): Promise<LeaderboardRow[]> {
  const url = `${API_URL}/api/v1/leaderboard?suite=${encodeURIComponent(suite)}&limit=${limit}`;
  const res = await fetch(url, { cache: "no-store" });
  const body: Envelope<LeaderboardRow[]> = await res.json();
  if (!body.ok) {
    throw new Error(body.error?.message ?? "failed to load leaderboard");
  }
  return body.data;
}

export type Agent = {
  id: string;
  agentthreads_handle: string;
  display_name: string;
  description: string | null;
  model: string | null;
  framework: string | null;
  tier: string;
  is_verified: boolean;
  total_runs: number;
  best_score: number;
  created_at: string;
};

export type RegisterAgentInput = {
  agentthreads_handle: string;
  display_name: string;
  description?: string;
  model?: string;
  framework?: string;
};

async function authedFetch<T>(
  accessToken: string,
  path: string,
  init?: RequestInit,
): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    cache: "no-store",
    headers: {
      ...init?.headers,
      Authorization: `Bearer ${accessToken}`,
    },
  });
  const body: Envelope<T> = await res.json();
  if (!body.ok) {
    throw new Error(body.error?.message ?? "request failed");
  }
  return body.data;
}

export function listMyAgents(accessToken: string): Promise<Agent[]> {
  return authedFetch<Agent[]>(accessToken, "/api/v1/agents/mine");
}

export function registerAgent(
  accessToken: string,
  input: RegisterAgentInput,
): Promise<{ agent: Agent; api_key: string }> {
  return authedFetch(accessToken, "/api/v1/agents/register", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function syncUser(accessToken: string): Promise<{ synced: boolean }> {
  return authedFetch(accessToken, "/api/v1/users/sync", { method: "POST" });
}
