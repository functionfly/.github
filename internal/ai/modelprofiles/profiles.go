package modelprofiles

import "github.com/functionfly/functionfly/internal/storage"

// Expand returns per-feature defaults for a preset profile.
// Custom profiles are not expanded here — callers keep stored defaults.
func Expand(profile string) map[string]storage.ModelSelection {
	switch profile {
	case "fast":
		return map[string]storage.ModelSelection{
			"composer":   {Provider: "groq", ModelID: "llama-4-scout-17b-16e-instruct"},
			"frg":        {Provider: "groq", ModelID: "llama-3.3-70b-versatile"},
			"dna":        {Provider: "openrouter", ModelID: "google/gemini-2.5-flash"},
			"chat":       {Provider: "openrouter", ModelID: "google/gemini-2.5-flash"},
			"support":    {Provider: "openrouter", ModelID: "anthropic/claude-haiku-4"},
			"embeddings": {Provider: "openai", ModelID: "text-embedding-3-small"},
			"agent":      {Provider: "groq", ModelID: "llama-4-scout-17b-16e-instruct"},
		}
	case "premium":
		return map[string]storage.ModelSelection{
			"composer":   {Provider: "openrouter", ModelID: "anthropic/claude-sonnet-4.6"},
			"frg":        {Provider: "openrouter", ModelID: "anthropic/claude-sonnet-4.6"},
			"dna":        {Provider: "openrouter", ModelID: "google/gemini-3.1-pro"},
			"chat":       {Provider: "openrouter", ModelID: "google/gemini-3.1-pro"},
			"support":    {Provider: "openrouter", ModelID: "anthropic/claude-haiku-4"},
			"embeddings": {Provider: "openai", ModelID: "text-embedding-3-large"},
			"agent":      {Provider: "openrouter", ModelID: "anthropic/claude-sonnet-4.6"},
		}
	default: // balanced
		return map[string]storage.ModelSelection{
			"composer":   {Provider: "openrouter", ModelID: "anthropic/claude-sonnet-4.6"},
			"frg":        {Provider: "openrouter", ModelID: "anthropic/claude-sonnet-4.6"},
			"dna":        {Provider: "openrouter", ModelID: "google/gemini-2.5-flash"},
			"chat":       {Provider: "openrouter", ModelID: "anthropic/claude-sonnet-4.6"},
			"support":    {Provider: "openrouter", ModelID: "anthropic/claude-haiku-4"},
			"embeddings": {Provider: "openai", ModelID: "text-embedding-3-small"},
			"agent":      {Provider: "openrouter", ModelID: "anthropic/claude-sonnet-4.6"},
		}
	}
}

// EffectiveDefaults merges stored defaults with profile presets when profile is not custom.
func EffectiveDefaults(profile string, stored map[string]storage.ModelSelection) map[string]storage.ModelSelection {
	out := map[string]storage.ModelSelection{}
	if profile != "" && profile != "custom" {
		for k, v := range Expand(profile) {
			out[k] = v
		}
	}
	for k, v := range stored {
		if v.ModelID != "" {
			out[k] = v
		}
	}
	return out
}
