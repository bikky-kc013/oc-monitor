package registry

import "strings"

// MaxInputTokens returns the max input token count for modelID.
// modelID is expected in "provider/model" format (as produced by joinModel).
//
// Matching strategy:
//  1. Exact match on provider + model ID
//  2. Case-insensitive match on the model ID alone (ignoring provider)
//  3. Substring match on the model ID (case-insensitive) against known models
//  4. Returns 0 if nothing matches
func (c *Client) MaxInputTokens(modelID string) int {
	if modelID == "" {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.entries) == 0 {
		return 0
	}

	provider, model := splitModelID(modelID)

	// 1. Exact match: provider + model
	if provider != "" {
		for _, e := range c.entries {
			if e.Provider == provider && e.ModelID == model {
				return e.maxInput()
			}
		}
	}

	// 2. Case-insensitive model-only match (prefer shortest name = most specific)
	lower := strings.ToLower(model)
	best := 0
	bestLen := 0
	for _, e := range c.entries {
		if strings.ToLower(e.ModelID) == lower {
			n := e.maxInput()
			if bestLen == 0 || len(e.ModelID) < bestLen {
				best = n
				bestLen = len(e.ModelID)
			}
		}
	}
	if best > 0 {
		return best
	}

	// 3. Substring match: modelID contains a known model name (case-insensitive)
	// Use the longest matching substring as best match.
	best = 0
	bestLen = 0
	for _, e := range c.entries {
		eName := strings.ToLower(e.ModelID)
		if strings.Contains(lower, eName) || strings.Contains(eName, lower) {
			n := e.maxInput()
			if len(e.ModelID) > bestLen {
				best = n
				bestLen = len(e.ModelID)
			}
		}
	}

	return best
}

// splitModelID splits "provider/model" into its parts.
// If there is no "/" it returns ("", modelID).
func splitModelID(modelID string) (provider, model string) {
	i := strings.LastIndex(modelID, "/")
	if i < 0 {
		return "", modelID
	}
	return modelID[:i], modelID[i+1:]
}
