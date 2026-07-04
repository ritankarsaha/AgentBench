-- Minimal seed set validating the schema across all 5 suites and all 5 task
-- types. The full 50-task curated corpus (10 per suite, Phase 1.4) is
-- deliberately NOT included here — that is separate, later work.

insert into benchmark_tasks (suite, category, type, title, description, input, expected_output, rubric, test_cases, difficulty, weight)
values
  ('codearena', 'debugging', 'functional',
   'Fix a broken IBAN validator',
   'The provided Python function validates IBAN numbers but has an off-by-one bug in the checksum calculation. Fix it so all test cases pass.',
   '{"code": "def validate_iban(iban: str) -> bool:\n    # buggy checksum implementation\n    return True"}'::jsonb,
   null,
   null,
   '[{"input": "GB82WEST12345698765432", "expected": true}, {"input": "GB82WEST12345698765431", "expected": false}]'::jsonb,
   'medium', 1.0),

  ('codearena', 'reasoning', 'exact',
   'Compute the 17th Fibonacci number',
   'Return the 17th number in the Fibonacci sequence, 0-indexed with fib(0)=0, fib(1)=1.',
   '{"n": 17}'::jsonb,
   '"1597"'::jsonb,
   null, null,
   'easy', 1.0),

  ('researchbench', 'synthesis', 'semantic',
   'Summarize the largest DeFi hacks of 2025',
   'Research and summarize the three largest DeFi protocol hacks in 2025 by dollar value, including protocol name, amount lost, and root cause.',
   '{"prompt": "List the 3 largest DeFi hacks of 2025 with amount lost and root cause."}'::jsonb,
   null,
   'Award full credit if the response names 3 real, correctly-ordered DeFi incidents with plausible dollar amounts and a specific root cause (not a vague "exploit"). Deduct credit for fabricated incidents or missing root cause.',
   null,
   'hard', 1.0),

  ('tooluse', 'orchestration', 'structural',
   'Summarize a file, then translate the summary to Spanish',
   'Given a text file, call the summarize tool first, then call the translate tool on the summary output with target language "es".',
   '{"file": "quarterly_report.txt", "steps": ["summarize", "translate"]}'::jsonb,
   '{"tool_calls": ["summarize", "translate"]}'::jsonb,
   null, null,
   'medium', 1.0),

  ('reasonbench', 'multi-step', 'exact',
   'Determine the odd one out',
   'Given the numbers [14, 22, 9, 36, 8], identify the only odd number.',
   '{"numbers": [14, 22, 9, 36, 8]}'::jsonb,
   '"9"'::jsonb,
   null, null,
   'easy', 1.0),

  ('agentops', 'workflow', 'multi-turn',
   'Iteratively debug an agent output across 3 rounds of feedback',
   'Given an agent''s buggy output and 3 rounds of human feedback, the agent must converge on a correct final answer by round 3.',
   '{"rounds": 3, "initial_output": "..."}'::jsonb,
   null, null, null,
   'expert', 1.0)
;
