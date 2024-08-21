// Package ollama provides an implementation that can be used by the system.
package ollama

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms/ollama"
)

// Ollama implements the Embedder and LLM interfaces.
type Ollama struct {
	*ollama.LLM
}

// New constructs an Ollama value for use by the ai package.
func New(model string) (*Ollama, error) {
	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		return nil, fmt.Errorf("new: %w", err)
	}

	ollama := Ollama{
		llm,
	}

	return &ollama, nil
}

// CreateEmbedding implements the Embedder interface.
func (oll *Ollama) CreateEmbedding(ctx context.Context, input []byte) ([]float32, error) {
	results, err := oll.LLM.CreateEmbedding(ctx, []string{string(input)})
	if err != nil {
		return nil, fmt.Errorf("create embedding: %w", err)
	}

	return results[0], nil
}
