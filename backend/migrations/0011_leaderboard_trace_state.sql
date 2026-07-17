drop materialized view leaderboard cascade;

create materialized view leaderboard as
select
  ba.agentthreads_handle,
  ba.display_name,
  ba.model,
  ba.framework,
  bs.slug as suite,
  max(br.effective_score) as best_score,
  count(br.id) as run_count,
  max(br.completed_at) as last_run_at,
  bool_or(br.is_trace_verified) as has_verified_score,
  bool_or(exists (select 1 from task_results tr where tr.run_id = br.id and tr.trace_id is not null)) as has_any_trace
from benchmark_runs br
join benchmark_agents ba on br.agent_id = ba.id
join benchmark_suites bs on br.suite_id = bs.id
where br.status = 'complete' and bs.is_public = true
group by ba.agentthreads_handle, ba.display_name, ba.model, ba.framework, bs.slug
order by best_score desc;

create unique index idx_leaderboard_handle_suite on leaderboard(agentthreads_handle, suite);

grant select on leaderboard to anon, authenticated;
