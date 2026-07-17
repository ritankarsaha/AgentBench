alter table task_results add constraint task_results_run_task_unique unique (run_id, task_id);

insert into benchmark_suites (name, slug, description, task_ids, is_public, scoring_mode, decay_enabled)
select 'Standard Benchmark', 'standard',
       'Deterministically-scored tasks (exact and structural) across all AgentBench categories.',
       array_agg(id order by created_at), true, 'weighted_avg', true
from benchmark_tasks
where type in ('exact', 'structural') and status = 'active';
