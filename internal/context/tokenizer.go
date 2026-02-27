package context

import "math"

// EstimateTokens estimates the number of tokens in a text string.
// Uses a character-based heuristic: ~1 token per 3.6 characters for English.
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return int(math.Ceil(float64(len(text)) / 3.6))
}
