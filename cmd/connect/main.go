package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ardanlabs/ai-training/cmd/connect/ai"
	"github.com/ardanlabs/ai-training/cmd/connect/board"
	"github.com/ardanlabs/ai-training/foundation/mongodb"
	"github.com/tmc/langchaingo/llms/ollama"
)

var train bool

func init() {
	flag.BoolVar(&train, "train", false, "process training data")

	flag.Parse()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// -------------------------------------------------------------------------
	// Connect to mongo.

	fmt.Println("Connecting to MongoDB ...")

	client, err := mongodb.Connect(ctx, "mongodb://localhost:27017", "ardan", "ardan")
	if err != nil {
		return fmt.Errorf("mongo connect: %w", err)
	}
	defer client.Disconnect(ctx)

	// -------------------------------------------------------------------------
	// Open a connection with ollama to access the model.

	fmt.Println("Connected to Ollama ...")

	llm, err := ollama.New(ollama.WithModel("mxbai-embed-large"))
	if err != nil {
		return fmt.Errorf("ollama: %w", err)
	}

	// -------------------------------------------------------------------------
	// Construct the AI api.

	fmt.Println("Establish AI support ...")

	ai, err := ai.New(client, llm)
	if err != nil {
		return fmt.Errorf("new ai: %w", err)
	}

	// -------------------------------------------------------------------------
	// Train or play the game.

	if train {
		return ai.ProcessBoardFiles()
	}

	return game(ai)
}

// =============================================================================

func game(ai *ai.AI) error {

	// -------------------------------------------------------------------------
	// Create the board and initialize the display.

	board, err := board.New(ai)
	if err != nil {
		return fmt.Errorf("new board: %w", err)
	}
	defer board.Shutdown()

	// -------------------------------------------------------------------------
	// Start handling board input.

	<-board.Run()

	return nil
}
