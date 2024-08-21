// Package ollama provides an implementation for using ollama.
package ollama

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// Embedder implements the Embedder interface.
type Embedder struct {
	llm *ollama.LLM
}

// NewEmbedder constructs Ollama support for embedding.
func NewEmbedder(model string) (*Embedder, error) {
	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		return nil, fmt.Errorf("embedding: %w", err)
	}

	embedder := Embedder{
		llm: llm,
	}

	return &embedder, nil
}

// CreateEmbedding implements the Embedder interface.
func (emb *Embedder) CreateEmbedding(ctx context.Context, input []byte) ([]float64, error) {
	results, err := emb.llm.CreateEmbedding(ctx, []string{string(input)})
	if err != nil {
		return nil, fmt.Errorf("create embedding: %w", err)
	}

	final := make([]float64, len(results[0]))
	for i := range results[0] {
		final[i] = float64(results[0][i])
	}

	return final, nil
}

// =============================================================================

// Chatter implements the Chatter interface.
type Chatter struct {
	llm *ollama.LLM
}

// NewEmbedder constructs Ollama support for chatting.
func NewChatter(model string) (*Chatter, error) {
	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		return nil, fmt.Errorf("embedding: %w", err)
	}

	chatter := Chatter{
		llm: llm,
	}

	return &chatter, nil
}

// Chat implements the Chatter inteface.
func (cht *Chatter) Chat(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return cht.llm.Call(ctx, prompt, options...)
}
