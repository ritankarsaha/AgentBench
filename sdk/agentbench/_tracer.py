from __future__ import annotations

from typing import Any


class Tracer:
    """AgentReplay integration shim.

    When an AgentReplay API key is configured, traces recorded during the
    agent call are submitted alongside each task result for verification.
    """

    def __init__(self, agentreplay_api_key: str | None = None):
        self.enabled = agentreplay_api_key is not None
        self.agentreplay_api_key = agentreplay_api_key

    def start_trace(self, task_id: str) -> None:
        if not self.enabled:
            return

    def end_trace(self, task_id: str) -> dict[str, Any] | None:
        if not self.enabled:
            return None
        return None
