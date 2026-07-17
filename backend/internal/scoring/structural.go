package scoring

import (
	"encoding/json"
	"fmt"
)

type toolCallSequence struct {
	ToolCalls []string `json:"tool_calls"`
}

func ScoreStructural(expectedOutput, actualOutput []byte) (float64, error) {
	var expected toolCallSequence
	if err := json.Unmarshal(expectedOutput, &expected); err != nil {
		return 0, fmt.Errorf("scoring: invalid expected_output: %w", err)
	}
	if len(expected.ToolCalls) == 0 {
		return 0, fmt.Errorf("scoring: expected_output has no tool_calls")
	}

	var actual toolCallSequence
	if err := json.Unmarshal(actualOutput, &actual); err != nil {
		return 0.0, nil
	}

	correct := 0
	for i, want := range expected.ToolCalls {
		if i < len(actual.ToolCalls) && actual.ToolCalls[i] == want {
			correct++
		}
	}
	return float64(correct) / float64(len(expected.ToolCalls)), nil
}
