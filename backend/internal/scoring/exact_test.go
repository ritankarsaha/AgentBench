package scoring

import "testing"

func TestScoreExactStringMatch(t *testing.T) {
	score, err := ScoreExact([]byte(`"341"`), []byte(`"341"`))
	if err != nil {
		t.Fatal(err)
	}
	if score != 1.0 {
		t.Errorf("got %v, want 1.0", score)
	}

	score, err = ScoreExact([]byte(`"341"`), []byte(`"342"`))
	if err != nil {
		t.Fatal(err)
	}
	if score != 0.0 {
		t.Errorf("got %v, want 0.0", score)
	}
}

func TestScoreExactRegex(t *testing.T) {
	score, err := ScoreExact([]byte(`{"regex": "^\\d+$"}`), []byte(`"9930"`))
	if err != nil {
		t.Fatal(err)
	}
	if score != 1.0 {
		t.Errorf("got %v, want 1.0", score)
	}

	score, err = ScoreExact([]byte(`{"regex": "^\\d+$"}`), []byte(`"abc"`))
	if err != nil {
		t.Fatal(err)
	}
	if score != 0.0 {
		t.Errorf("got %v, want 0.0", score)
	}
}

func TestScoreExactMalformedActualScoresZero(t *testing.T) {
	score, err := ScoreExact([]byte(`"341"`), []byte(`not valid json`))
	if err != nil {
		t.Fatal(err)
	}
	if score != 0.0 {
		t.Errorf("got %v, want 0.0", score)
	}
}
