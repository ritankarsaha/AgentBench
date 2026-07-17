package scoring

import "testing"

func TestScoreStructuralFullMatch(t *testing.T) {
	score, err := ScoreStructural(
		[]byte(`{"tool_calls": ["summarize", "translate"]}`),
		[]byte(`{"tool_calls": ["summarize", "translate"]}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	if score != 1.0 {
		t.Errorf("got %v, want 1.0", score)
	}
}

func TestScoreStructuralPartialCredit(t *testing.T) {
	score, err := ScoreStructural(
		[]byte(`{"tool_calls": ["a", "b", "c", "d"]}`),
		[]byte(`{"tool_calls": ["a", "x", "c"]}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	if score != 0.5 {
		t.Errorf("got %v, want 0.5", score)
	}
}

func TestScoreStructuralEmptyExpectedIsError(t *testing.T) {
	_, err := ScoreStructural([]byte(`{"tool_calls": []}`), []byte(`{"tool_calls": []}`))
	if err == nil {
		t.Error("expected an error for empty expected tool_calls")
	}
}
