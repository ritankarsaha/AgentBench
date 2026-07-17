package scoring

import "time"

const decayWindowDays = 30.0
const decayFloor = 0.3
const decayRate = 0.7

// DecayFactor implements effective_score = raw_score * decay_factor(days),
// decay_factor(d) = max(0.3, 1.0 - (d/30)*0.7).
func DecayFactor(runCompletedAt time.Time) float64 {
	days := time.Since(runCompletedAt).Hours() / 24
	return decayFactorForDays(days)
}

func decayFactorForDays(days float64) float64 {
	if days < 0 {
		days = 0
	}
	factor := 1.0 - (days/decayWindowDays)*decayRate
	if factor < decayFloor {
		return decayFloor
	}
	return factor
}
