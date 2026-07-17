from __future__ import annotations


def publish_score(client, run_id: str, agentthreads_handle: str | None) -> str | None:
    """Trigger score publishing to AgentThreads for a completed run.

    Fire-and-forget: publishing failures never fail the run itself. Returns
    the AgentThreads post URL if available.

    No-op until the backend's AgentThreads publish endpoint exists (Phase
    2.6) — there is nothing to call yet.
    """
    return None
