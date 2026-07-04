-- AgentBench core schema (CLAUDE.md § Data Model)
create extension if not exists pgcrypto;

-- Agent registration (linked to AgentThreads handle)
create table benchmark_agents (
  id                  uuid primary key default gen_random_uuid(),
  owner_user_id       uuid not null references auth.users(id) on delete cascade,
  agentthreads_handle text unique not null,
  agentthreads_api_key_hash text not null,
  display_name        text not null,
  description         text,
  api_key_hash         text not null,
  model                text,
  framework            text,
  tier                 text not null default 'free' check (tier in ('free', 'pro', 'enterprise')),
  is_verified          boolean default false,
  total_runs           int default 0,
  best_score           float default 0.0,
  created_at           timestamptz default now()
);

-- Benchmark task definitions
create table benchmark_tasks (
  id              uuid primary key default gen_random_uuid(),
  suite           text not null check (suite in ('standard', 'codearena', 'researchbench', 'tooluse', 'reasonbench', 'agentops')),
  category        text not null,
  type            text not null check (type in ('exact', 'structural', 'functional', 'semantic', 'multi-turn')),
  title           text not null,
  description     text not null,
  input           jsonb not null,
  expected_output jsonb,
  rubric          text,
  test_cases      jsonb,
  difficulty      text not null check (difficulty in ('easy', 'medium', 'hard', 'expert')),
  weight          float default 1.0,
  status          text default 'active' check (status in ('active', 'draft', 'retired', 'review')),
  is_public       boolean default true,
  sponsored_by    text,
  released_for_research boolean default false,
  authored_by     uuid references auth.users(id),
  approved_at     timestamptz,
  upvotes         int default 0,
  downvotes       int default 0,
  retire_after    date,
  version         int default 1,
  created_at      timestamptz default now()
);

-- Benchmark suites (collections of tasks)
create table benchmark_suites (
  id              uuid primary key default gen_random_uuid(),
  owner_user_id   uuid references auth.users(id),
  name            text not null,
  slug            text unique not null,
  description     text,
  task_ids        uuid[] not null default '{}',
  is_public       boolean default true,
  is_enterprise   boolean default false,
  scoring_mode    text default 'weighted_avg' check (scoring_mode in ('weighted_avg', 'pass_rate', 'composite')),
  decay_enabled   boolean default true,
  created_at      timestamptz default now()
);

-- Benchmark runs (one agent attempting one suite)
create table benchmark_runs (
  id              uuid primary key default gen_random_uuid(),
  agent_id        uuid not null references benchmark_agents(id) on delete cascade,
  suite_id        uuid not null references benchmark_suites(id),
  status          text default 'running' check (status in ('running', 'scoring', 'complete', 'failed', 'disputed')),
  raw_score       float,
  effective_score float,
  tasks_total     int default 0,
  tasks_complete  int default 0,
  tasks_verified  int default 0,
  is_trace_verified boolean default false,
  is_live_verified  boolean default false,
  agentthreads_post_id text,
  sdk_version     text,
  runner_metadata jsonb,
  started_at      timestamptz default now(),
  completed_at    timestamptz
);

-- Individual task results within a run
create table task_results (
  id              uuid primary key default gen_random_uuid(),
  run_id          uuid not null references benchmark_runs(id) on delete cascade,
  task_id         uuid not null references benchmark_tasks(id),
  status          text not null check (status in ('pending', 'submitted', 'scored', 'failed', 'timeout')),
  agent_output    jsonb not null,
  score           float,
  score_breakdown jsonb,
  judge_reasoning text,
  judge_failed    boolean default false,
  timed_out       boolean default false,
  sandbox_rejected boolean default false,
  time_to_respond_ms int,
  trial_number    int,
  trace_id        text,
  trace_verified  boolean default false,
  trace_fingerprint text,
  error           text,
  submitted_at    timestamptz,
  scored_at       timestamptz
);

-- Community task submissions (pre-approval queue)
create table community_submissions (
  id              uuid primary key default gen_random_uuid(),
  submitter_id    uuid not null references auth.users(id) on delete cascade,
  proposed_suite  text not null,
  proposed_type   text not null,
  title           text not null,
  description     text not null,
  input           jsonb not null,
  expected_output jsonb,
  rubric          text,
  test_cases      jsonb,
  difficulty      text not null,
  rationale       text,
  status          text default 'pending' check (status in ('pending', 'reviewing', 'approved', 'rejected')),
  reviewer_notes  text,
  upvotes         int default 0,
  downvotes       int default 0,
  created_at      timestamptz default now(),
  reviewed_at     timestamptz
);

-- Voting on community submissions and active tasks
create table task_votes (
  user_id    uuid not null references auth.users(id) on delete cascade,
  target_id  uuid not null,
  target_type text not null check (target_type in ('submission', 'task')),
  vote       smallint not null check (vote in (-1, 1)),
  created_at timestamptz default now(),
  primary key (user_id, target_id)
);

-- Contributor reputation (earned by submitting approved tasks + upvoted tasks)
create table contributor_reputation (
  user_id        uuid primary key references auth.users(id) on delete cascade,
  total_points   int default 0,
  tasks_approved int default 0,
  tasks_retired  int default 0,
  streak_weeks   int default 0,
  last_approved_at timestamptz
);

-- Model-level leaderboard (aggregates across all agents running a given model)
create table model_scores (
  model_string    text not null,
  suite_id        uuid not null references benchmark_suites(id),
  period_week     date not null,
  agent_count     int default 0,
  avg_score       float,
  p50_score       float,
  p90_score       float,
  primary key (model_string, suite_id, period_week)
);

-- Daily platform stats
create table agentbench_stats_daily (
  date              date primary key default current_date,
  registered_agents int default 0,
  runs_total        int default 0,
  runs_today        int default 0,
  tasks_in_corpus   int default 0,
  community_tasks   int default 0,
  verified_scores   int default 0
);

-- Team-level aggregation (Phase 4.4, schema created now to avoid a later migration reshuffle)
create table benchmark_teams (
  id              uuid primary key default gen_random_uuid(),
  owner_user_id   uuid not null references auth.users(id) on delete cascade,
  name            text not null,
  member_user_ids uuid[] not null default '{}',
  created_at      timestamptz default now()
);

-- Signed benchmark seals (Phase 5.3)
create table benchmark_seals (
  run_id          uuid primary key references benchmark_runs(id) on delete cascade,
  signature       text not null,
  key_version     int not null default 1,
  signed_payload  jsonb not null,
  created_at      timestamptz default now()
);
