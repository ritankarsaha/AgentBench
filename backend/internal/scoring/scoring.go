package scoring

import "fmt"

// Score dispatches to the scorer for taskType.
func Score(taskType string, expectedOutput, actualOutput []byte) (float64, error) {
	switch taskType {
	case "exact":
		return ScoreExact(expectedOutput, actualOutput)
	case "structural":
		return ScoreStructural(expectedOutput, actualOutput)
	default:
		return 0, fmt.Errorf("scoring: type %q is not implemented yet", taskType)
	}
}
