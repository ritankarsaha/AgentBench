-- Realtime: allow the frontend to subscribe to run progress and leaderboard changes.
alter publication supabase_realtime add table benchmark_runs;
alter publication supabase_realtime add table task_results;

-- pg_cron: refresh the leaderboard materialized view every 60s (Phase 1.2 requirement).
create extension if not exists pg_cron;

select cron.schedule(
  'refresh-leaderboard',
  '* * * * *', -- every minute (pg_cron here caps sub-minute schedules at 59s)
  $$refresh materialized view concurrently leaderboard$$
);
