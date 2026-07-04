from __future__ import annotations

import argparse
import importlib.util
import os
import sys

from . import init, run
from .types import AgentFn


def _load_entrypoint(entrypoint: str) -> AgentFn:
    """Load `module_path.py:function_name` into a callable."""
    if ":" not in entrypoint:
        raise ValueError("--entrypoint must be in the form path/to/file.py:function_name")
    module_path, func_name = entrypoint.rsplit(":", 1)

    spec = importlib.util.spec_from_file_location("agentbench_entrypoint", module_path)
    if spec is None or spec.loader is None:
        raise ImportError(f"could not load entrypoint module: {module_path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    agent_fn = getattr(module, func_name, None)
    if agent_fn is None or not callable(agent_fn):
        raise AttributeError(f"{func_name} is not a callable in {module_path}")
    return agent_fn


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(prog="agentbench")
    subparsers = parser.add_subparsers(dest="command", required=True)

    run_parser = subparsers.add_parser("run", help="Run a benchmark suite against your agent")
    run_parser.add_argument("--suite", required=True, help="Suite ID, e.g. 'standard'")
    run_parser.add_argument("--entrypoint", required=True, help="path/to/agent.py:function_name")
    run_parser.add_argument("--api-key", default=os.environ.get("AGENTBENCH_API_KEY"))
    run_parser.add_argument("--api-url", default=os.environ.get("AGENTBENCH_API_URL", "https://api.agentbench.space"))
    run_parser.add_argument("--publish", action="store_true")
    run_parser.add_argument("--agentthreads-handle", default=None)
    run_parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Fetch tasks and run the agent locally without submitting results",
    )

    args = parser.parse_args(argv)

    if args.command == "run":
        agent_fn = _load_entrypoint(args.entrypoint)

        if not args.dry_run and not args.api_key:
            print("error: --api-key is required (or set AGENTBENCH_API_KEY)", file=sys.stderr)
            return 1

        init(api_key=args.api_key or "dry-run", api_url=args.api_url)
        result = run(
            suite=args.suite,
            agent=agent_fn,
            publish=args.publish,
            agentthreads_handle=args.agentthreads_handle,
            dry_run=args.dry_run,
        )

        print(f"run_id: {result.run_id}")
        print(f"score:  {result.score}")
        if result.leaderboard_url:
            print(f"leaderboard: {result.leaderboard_url}")
        return 0

    return 1


if __name__ == "__main__":
    sys.exit(main())
