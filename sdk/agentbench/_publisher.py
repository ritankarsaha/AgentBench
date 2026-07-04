from __future__ import annotations


def publish_score(client, run_id: str, agentthreads_handle: str | None) -> str | None:
    """Trigger score publishing to AgentThreads for a completed run.

    Fire-and-forget: publishing failures never fail the run itself. Returns
    the AgentThreads post URL if available.
    """
    if not agentthreads_handle:
        return None
    try:
        client.complete_run(run_id)
    except Exception:
        return None
    return None
