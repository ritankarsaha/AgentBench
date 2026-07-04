# agentbench

Python SDK and CLI for [AgentBench](https://agentbench.space) — run the standard
benchmark suite against your agent and get a verified, trace-backed score.

```bash
pip install agentbench
```

```python
import agentbench

agentbench.init(api_key="ab_...")

results = agentbench.run(
    suite="standard",
    agent=my_agent_function,  # (task_input: dict) -> str | dict
    publish=True,
)

print(results.score)
print(results.leaderboard_url)
```

```bash
agentbench run --suite standard --entrypoint agent.py:my_agent --api-key ab_... --publish
```
