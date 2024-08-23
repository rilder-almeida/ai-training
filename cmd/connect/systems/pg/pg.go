// Package pg provides an implementation for using prediction guard.
package pg

import (
	"context"
	"fmt"

	"github.com/predictionguard/go-client"
)

// Embedder implements the Embedder interface.
type Embedder struct {
	client *client.Client
}

// NewEmbedder constructs Prediction Guard support for embedding.
func NewEmbedder(apiKey string) *Embedder {
	host := "https://api.predictionguard.com"

	logger := func(ctx context.Context, msg string, v ...any) {}
	embedder := Embedder{
		client: client.New(logger, host, apiKey),
	}

	return &embedder
}

// CreateEmbedding implements the Embedder interface.
func (emb *Embedder) CreateEmbedding(ctx context.Context, image []byte, text string) ([]float64, error) {
	inp := []client.EmbeddingInput{
		{
			Image: newRawImage(image),
		},
	}

	result, err := emb.client.Embedding(ctx, inp)
	if err != nil {
		return nil, fmt.Errorf("create embedding: %w", err)
	}

	return result.Data[0].Embedding, nil
}
