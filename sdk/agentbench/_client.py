from __future__ import annotations

import time
from typing import Any

import requests


class AgentBenchClient:
    """Typed HTTP client for the AgentBench API.

    Retries once on 429 with a 1s backoff, per SDK conventions.
    """

    def __init__(self, api_key: str, base_url: str = "https://api.agentbench.space"):
        self.api_key = api_key
        self.base_url = base_url.rstrip("/")
        self._session = requests.Session()

    def _headers(self) -> dict[str, str]:
        return {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json",
        }

    def _request(self, method: str, path: str, **kwargs: Any) -> dict[str, Any]:
        url = f"{self.base_url}{path}"
        response = self._session.request(method, url, headers=self._headers(), **kwargs)
        if response.status_code == 429:
            time.sleep(1)
            response = self._session.request(method, url, headers=self._headers(), **kwargs)
        response.raise_for_status()
        return response.json()

    def start_run(self, suite: str) -> dict[str, Any]:
        return self._request("POST", "/api/v1/runs", json={"suite": suite})

    def submit_result(self, run_id: str, task_id: str, output: Any, trace_id: str | None = None) -> dict[str, Any]:
        body: dict[str, Any] = {"task_id": task_id, "output": output}
        if trace_id:
            body["trace_id"] = trace_id
        return self._request("POST", f"/api/v1/runs/{run_id}/results", json=body)

    def complete_run(self, run_id: str) -> dict[str, Any]:
        return self._request("POST", f"/api/v1/runs/{run_id}/complete")

    def get_run(self, run_id: str) -> dict[str, Any]:
        return self._request("GET", f"/api/v1/runs/{run_id}")
