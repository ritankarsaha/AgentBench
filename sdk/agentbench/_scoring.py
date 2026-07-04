from __future__ import annotations

from typing import Any


def validate_output_format(task_type: str, output: Any) -> str | None:
    """Local pre-validation before submitting to the API.

    Returns an error message if the output is obviously malformed, else None.
    Server-side scoring is authoritative — this only catches cheap mistakes
    (wrong type, empty output) before spending a network round-trip.
    """
    if output is None:
        return "agent returned None"
    if task_type == "functional" and not isinstance(output, str):
        return "functional tasks expect a string (code) output"
    return None
