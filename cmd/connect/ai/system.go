package ai

import (
	"fmt"

	"github.com/tmc/langchaingo/llms/ollama"
)

const (
	SystemOllama          = "ollama"
	SystemPredictionGuard = "prediction-guard"
)

// CreateEmbedder can create an implementation of the Embedder interface based
// on well known systems.
func CreateEmbedder(system string, model string) (Embedder, error) {
	switch system {
	case SystemOllama:
		return ollama.New(ollama.WithModel(model))
	}

	return nil, fmt.Errorf("unknown system or model: system %q, model %q", system, model)
}

// CreateLLM can create an implementation of the LLM interface based
// on well known systems.
func CreateLLM(system string, model string) (LLM, error) {
	switch system {
	case SystemOllama:
		return ollama.New(ollama.WithModel(model))
	}

	return nil, fmt.Errorf("unknown system or model: system %q, model %q", system, model)
}
