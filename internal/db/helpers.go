package db

import (
	"github.com/bikky/oc-monitor/internal/registry"
)

func joinModel(provider, model string) string {
	switch {
	case provider == "" && model == "":
		return ""
	case provider == "":
		return model
	case model == "":
		return provider
	default:
		return provider + "/" + model
	}
}

// SetRegistry attaches a registry client for dynamic model token lookups.
func (d *DB) SetRegistry(r *registry.Client) {
	d.reg = r
}

// Registry returns the attached registry client (may be nil).
func (d *DB) Registry() *registry.Client {
	return d.reg
}

// MaxInputTokens returns the max input token count for a model.
// It delegates to the registry client if available, otherwise returns 0.
func (d *DB) MaxInputTokens(modelID string) int {
	if d.reg != nil {
		return d.reg.MaxInputTokens(modelID)
	}
	return 0
}
