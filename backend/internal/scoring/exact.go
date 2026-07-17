package scoring

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
)

type regexExpectation struct {
	Regex string `json:"regex"`
}

// ScoreExact returns 1.0 or 0.0.
func ScoreExact(expectedOutput, actualOutput []byte) (float64, error) {
	var re regexExpectation
	if err := json.Unmarshal(expectedOutput, &re); err == nil && re.Regex != "" {
		var actual string
		if err := json.Unmarshal(actualOutput, &actual); err != nil {
			return 0.0, nil
		}
		matched, err := regexp.MatchString(re.Regex, actual)
		if err != nil {
			return 0, fmt.Errorf("scoring: invalid regex in expected_output: %w", err)
		}
		if matched {
			return 1.0, nil
		}
		return 0.0, nil
	}

	var expected, actual any
	if err := json.Unmarshal(expectedOutput, &expected); err != nil {
		return 0, fmt.Errorf("scoring: invalid expected_output: %w", err)
	}
	if err := json.Unmarshal(actualOutput, &actual); err != nil {
		return 0.0, nil
	}
	if reflect.DeepEqual(expected, actual) {
		return 1.0, nil
	}
	return 0.0, nil
}
