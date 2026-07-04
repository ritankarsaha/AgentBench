-- agentbench-traces bucket: private read. The Go backend (service_role) can
-- always read/write. An authenticated human can read traces belonging to an
-- agent they own, addressed by object path "<agent_id>/<run_id>/<task_id>.json".
insert into storage.buckets (id, name, public)
values ('agentbench-traces', 'agentbench-traces', false)
on conflict (id) do nothing;

create policy "traces_owner_read" on storage.objects
  for select using (
    bucket_id = 'agentbench-traces'
    and exists (
      select 1 from benchmark_agents a
      where a.owner_user_id = auth.uid()
        and a.id::text = (storage.foldername(name))[1]
    )
  );
