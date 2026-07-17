import json

import pytest
import responses

import agentbench._runner as runner_module
from agentbench._client import AgentBenchClient
from agentbench._runner import run


@responses.activate
def test_run_dispatches_tasks_and_submits_results_in_order():
    client = AgentBenchClient(api_key="ab_test", base_url="https://api.test")

    responses.add(
        responses.POST,
        "https://api.test/api/v1/runs",
        json={
            "ok": True,
            "data": {
                "run_id": "run_1",
                "tasks": [
                    {"id": "task_1", "type": "exact", "input": {"q": "2+2"}},
                    {"id": "task_2", "type": "exact", "input": {"q": "3+3"}},
                ],
            },
        },
        status=200,
    )
    responses.add(
        responses.POST,
        "https://api.test/api/v1/runs/run_1/results",
        json={"ok": True, "data": {}},
        status=200,
    )
    responses.add(
        responses.POST,
        "https://api.test/api/v1/runs/run_1/complete",
        json={"ok": True, "data": {}},
        status=200,
    )
    responses.add(
        responses.GET,
        "https://api.test/api/v1/runs/run_1",
        json={
            "ok": True,
            "data": {
                "status": "complete",
                "effective_score": 1.0,
                "tasks_total": 2,
                "tasks_complete": 2,
                "results": [],
            },
        },
        status=200,
    )

    calls = []

    def agent(task_input):
        calls.append(task_input)
        return "4" if task_input["q"] == "2+2" else "6"

    result = run(suite="standard", agent=agent, client=client)

    assert calls == [{"q": "2+2"}, {"q": "3+3"}]
    assert result.run_id == "run_1"
    assert result.score == 1.0
    assert result.tasks_complete == 2

    submitted = [
        json.loads(call.request.body)
        for call in responses.calls
        if call.request.url == "https://api.test/api/v1/runs/run_1/results"
    ]
    assert submitted == [
        {"task_id": "task_1", "output": "4"},
        {"task_id": "task_2", "output": "6"},
    ]


def test_dry_run_never_calls_the_api():
    client = AgentBenchClient(api_key="ab_test", base_url="https://api.test")

    def agent(task_input):
        raise AssertionError("agent should not be called during dry-run scaffolding")

    result = run(suite="standard", agent=agent, client=client, dry_run=True)

    assert result.run_id == "dry-run"
    assert result.tasks_total == 0


@responses.activate
def test_retries_once_on_429_then_succeeds():
    client = AgentBenchClient(api_key="ab_test", base_url="https://api.test")

    responses.add(responses.POST, "https://api.test/api/v1/runs", status=429)
    responses.add(
        responses.POST,
        "https://api.test/api/v1/runs",
        json={"ok": True, "data": {"run_id": "run_2", "tasks": []}},
        status=200,
    )

    result = client.start_run("standard")

    assert result["data"]["run_id"] == "run_2"


@responses.activate
def test_run_times_out_if_run_never_completes(monkeypatch):
    monkeypatch.setattr(runner_module, "POLL_TIMEOUT_SECONDS", 0.05)
    monkeypatch.setattr(runner_module, "POLL_INTERVAL_SECONDS", 0.01)

    client = AgentBenchClient(api_key="ab_test", base_url="https://api.test")

    responses.add(
        responses.POST,
        "https://api.test/api/v1/runs",
        json={"ok": True, "data": {"run_id": "run_3", "tasks": []}},
        status=200,
    )
    responses.add(
        responses.POST,
        "https://api.test/api/v1/runs/run_3/complete",
        json={"ok": True, "data": {}},
        status=200,
    )
    responses.add(
        responses.GET,
        "https://api.test/api/v1/runs/run_3",
        json={"ok": True, "data": {"status": "running"}},
        status=200,
    )

    with pytest.raises(TimeoutError):
        run(suite="standard", agent=lambda task_input: task_input, client=client)
