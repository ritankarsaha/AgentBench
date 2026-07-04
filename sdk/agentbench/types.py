from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Callable


AgentFn = Callable[[dict[str, Any]], Any]


@dataclass
class TaskInput:
    id: str
    suite: str
    type: str
    title: str
    input: dict[str, Any]


@dataclass
class TaskResult:
    task_id: str
    output: Any
    score: float | None = None
    trace_id: str | None = None
    error: str | None = None


@dataclass
class RunResult:
    run_id: str
    suite: str
    score: float
    tasks_total: int
    tasks_complete: int
    leaderboard_url: str | None = None
    task_results: list[TaskResult] = field(default_factory=list)
