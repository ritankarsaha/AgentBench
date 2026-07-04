from __future__ import annotations

from typing import Callable

from ._client import AgentBenchClient
from ._runner import run as _run
from .types import AgentFn, RunResult, TaskInput, TaskResult

__all__ = ["init", "run", "RunResult", "TaskInput", "TaskResult", "AgentFn"]

__version__ = "0.1.0"

_state: dict[str, object] = {}


def init(
    api_key: str,
    api_url: str = "https://api.agentbench.space",
    agentreplay_api_key: str | None = None,
    redact: Callable[[str], str] | None = None,
) -> None:
    """Initialize the AgentBench SDK. Must be called before agentbench.run()."""
    _state["client"] = AgentBenchClient(api_key=api_key, base_url=api_url)
    _state["agentreplay_api_key"] = agentreplay_api_key
    _state["redact"] = redact or (lambda x: x)


def run(
    suite: str,
    agent: AgentFn,
    publish: bool = False,
    agentthreads_handle: str | None = None,
    dry_run: bool = False,
) -> RunResult:
    client = _state.get("client")
    if client is None:
        raise RuntimeError(
            "agentbench.init() must be called before agentbench.run()"
        )
    return _run(
        suite=suite,
        agent=agent,
        client=client,  # type: ignore[arg-type]
        publish=publish,
        agentthreads_handle=agentthreads_handle,
        agentreplay_api_key=_state.get("agentreplay_api_key"),  # type: ignore[arg-type]
        dry_run=dry_run,
    )
