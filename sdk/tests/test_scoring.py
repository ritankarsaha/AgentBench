from agentbench._scoring import validate_output_format


def test_none_output_is_rejected():
    assert validate_output_format("exact", None) is not None


def test_functional_requires_string_output():
    assert validate_output_format("functional", {"not": "a string"}) is not None
    assert validate_output_format("functional", "def solve(): pass") is None


def test_exact_accepts_any_non_none_output():
    assert validate_output_format("exact", 42) is None
    assert validate_output_format("exact", "42") is None
