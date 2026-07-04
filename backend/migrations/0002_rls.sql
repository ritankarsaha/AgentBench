-- Row Level Security. The Go backend connects with the service_role key and
-- bypasses RLS entirely; these policies are the safety net for any direct
-- anon/authenticated access (e.g. a future supabase-js realtime subscription).

create or replace function is_maintainer()
returns boolean
language sql stable
as $$
  select coalesce((auth.jwt() -> 'app_metadata' ->> 'maintainer')::boolean, false)
$$;

-- benchmark_agents ------------------------------------------------------
alter table benchmark_agents enable row level security;

create policy "agents_public_directory_read" on benchmark_agents
  for select using (true);

create policy "agents_owner_write" on benchmark_agents
  for all using (owner_user_id = auth.uid())
  with check (owner_user_id = auth.uid());

revoke select (api_key_hash, agentthreads_api_key_hash) on benchmark_agents from anon, authenticated;

-- benchmark_tasks ---------------------------------------------------------
alter table benchmark_tasks enable row level security;

create policy "tasks_public_read_active" on benchmark_tasks
  for select using (status = 'active' and is_public = true);

create policy "tasks_owner_read_private" on benchmark_tasks
  for select using (is_public = false and authored_by = auth.uid());

create policy "tasks_maintainer_all" on benchmark_tasks
  for all using (is_maintainer()) with check (is_maintainer());

-- expected_output must never reach anon/authenticated directly; the Go API
-- strips it at the handler level and only ever reads it via service_role.
revoke select (expected_output, rubric, test_cases) on benchmark_tasks from anon, authenticated;

-- benchmark_suites ----------------------------------------------------------
alter table benchmark_suites enable row level security;

create policy "suites_public_read" on benchmark_suites
  for select using (is_public = true);

create policy "suites_owner_read_write" on benchmark_suites
  for all using (owner_user_id = auth.uid())
  with check (owner_user_id = auth.uid());

-- benchmark_runs ------------------------------------------------------------
alter table benchmark_runs enable row level security;

create policy "runs_public_read_complete_public_suite" on benchmark_runs
  for select using (
    status = 'complete'
    and exists (
      select 1 from benchmark_suites s
      where s.id = benchmark_runs.suite_id and s.is_public = true
    )
  );

create policy "runs_owner_read" on benchmark_runs
  for select using (
    exists (
      select 1 from benchmark_agents a
      where a.id = benchmark_runs.agent_id and a.owner_user_id = auth.uid()
    )
  );

-- task_results ----------------------------------------------------------------
alter table task_results enable row level security;

create policy "task_results_public_read" on task_results
  for select using (
    exists (
      select 1 from benchmark_runs r
      join benchmark_suites s on s.id = r.suite_id
      where r.id = task_results.run_id and r.status = 'complete' and s.is_public = true
    )
  );

create policy "task_results_owner_read" on task_results
  for select using (
    exists (
      select 1 from benchmark_runs r
      join benchmark_agents a on a.id = r.agent_id
      where r.id = task_results.run_id and a.owner_user_id = auth.uid()
    )
  );

-- community_submissions ---------------------------------------------------
alter table community_submissions enable row level security;

create policy "submissions_own_read" on community_submissions
  for select using (submitter_id = auth.uid());

create policy "submissions_own_insert" on community_submissions
  for insert with check (submitter_id = auth.uid());

create policy "submissions_maintainer_all" on community_submissions
  for all using (is_maintainer()) with check (is_maintainer());

-- task_votes ------------------------------------------------------------------
alter table task_votes enable row level security;

create policy "votes_own_read" on task_votes
  for select using (user_id = auth.uid());

create policy "votes_own_write" on task_votes
  for all using (user_id = auth.uid()) with check (user_id = auth.uid());

-- contributor_reputation ----------------------------------------------------
alter table contributor_reputation enable row level security;

create policy "reputation_public_read" on contributor_reputation
  for select using (true);

-- model_scores ------------------------------------------------------------------
alter table model_scores enable row level security;

create policy "model_scores_public_read" on model_scores
  for select using (true);

-- agentbench_stats_daily -----------------------------------------------------
alter table agentbench_stats_daily enable row level security;

create policy "stats_public_read" on agentbench_stats_daily
  for select using (true);

-- benchmark_teams -----------------------------------------------------------
alter table benchmark_teams enable row level security;

create policy "teams_member_read" on benchmark_teams
  for select using (auth.uid() = owner_user_id or auth.uid() = any(member_user_ids));

create policy "teams_owner_write" on benchmark_teams
  for all using (owner_user_id = auth.uid()) with check (owner_user_id = auth.uid());

-- benchmark_seals -----------------------------------------------------------
alter table benchmark_seals enable row level security;

create policy "seals_public_read" on benchmark_seals
  for select using (true);
