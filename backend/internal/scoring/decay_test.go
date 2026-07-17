package scoring

import "testing"

func TestDecayFactorForDays(t *testing.T) {
	cases := []struct {
		days     float64
		expected float64
	}{
		{0, 1.0},
		{30, 0.3},
		{60, 0.3},
		{-5, 1.0},
		{15, 1.0 - (15.0/30.0)*0.7},
	}

	const epsilon = 1e-9
	for _, c := range cases {
		got := decayFactorForDays(c.days)
		diff := got - c.expected
		if diff < -epsilon || diff > epsilon {
			t.Errorf("decayFactorForDays(%v) = %v, want %v", c.days, got, c.expected)
		}
	}
}
