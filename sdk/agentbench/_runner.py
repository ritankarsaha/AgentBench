from __future__ import annotations

import time
from typing import Any

from ._client import AgentBenchClient
from ._publisher import publish_score
from ._scoring import validate_output_format
from ._tracer import Tracer
from .types import AgentFn, RunResult, TaskResult

POLL_INTERVAL_SECONDS = 2
POLL_TIMEOUT_SECONDS = 300


def run(
    suite: str,
    agent: AgentFn,
    client: AgentBenchClient,
    publish: bool = False,
    agentthreads_handle: str | None = None,
    agentreplay_api_key: str | None = None,
    dry_run: bool = False,
) -> RunResult:
    tracer = Tracer(agentreplay_api_key)

    if dry_run:
        return RunResult(run_id="dry-run", suite=suite, score=0.0, tasks_total=0, tasks_complete=0)

    started = client.start_run(suite)
    run_id = started["data"]["run_id"]
    tasks = started["data"]["tasks"]

    task_results: list[TaskResult] = []
    for task in tasks:
        tracer.start_trace(task["id"])
        output = agent(task["input"])
        error = validate_output_format(task["type"], output)
        trace = tracer.end_trace(task["id"])
        trace_id = trace.get("trace_id") if trace else None

        if error is None:
            client.submit_result(run_id, task["id"], output, trace_id)
        task_results.append(TaskResult(task_id=task["id"], output=output, trace_id=trace_id, error=error))

    client.complete_run(run_id)

    final = _poll_until_complete(client, run_id)
    leaderboard_url = None
    if publish:
        leaderboard_url = publish_score(client, run_id, agentthreads_handle)

    data = final["data"]
    return RunResult(
        run_id=run_id,
        suite=suite,
        score=data.get("effective_score") or 0.0,
        tasks_total=data.get("tasks_total", len(tasks)),
        tasks_complete=data.get("tasks_complete", len(task_results)),
        leaderboard_url=leaderboard_url,
        task_results=task_results,
    )


def _poll_until_complete(client: AgentBenchClient, run_id: str) -> dict[str, Any]:
    deadline = time.time() + POLL_TIMEOUT_SECONDS
    while time.time() < deadline:
        result = client.get_run(run_id)
        if result["data"].get("status") == "complete":
            return result
        time.sleep(POLL_INTERVAL_SECONDS)
    raise TimeoutError(f"run {run_id} did not complete within {POLL_TIMEOUT_SECONDS}s")
