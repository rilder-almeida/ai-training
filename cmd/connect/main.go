package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ardanlabs/ai-training/cmd/connect/ai"
	"github.com/ardanlabs/ai-training/cmd/connect/board"
	"github.com/ardanlabs/ai-training/foundation/mongodb"
	"github.com/tmc/langchaingo/llms/ollama"
)

/*
	- Produce images for vectorizing data
	- Try larger Llama3.1 model: llama3.1:405b
	- Try Gemma model: gemma2:27b
	- Build out all boards with
*/

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

	embed, err := ollama.New(ollama.WithModel("mxbai-embed-large"))
	if err != nil {
		return fmt.Errorf("ollama: %w", err)
	}

	chat, err := ollama.New(ollama.WithModel("llama3.1"))
	if err != nil {
		return fmt.Errorf("ollama: %w", err)
	}

	// -------------------------------------------------------------------------
	// Construct the AI api.

	fmt.Println("Establish AI support ...")

	ai, err := ai.New(client, embed, chat)
	if err != nil {
		return fmt.Errorf("new ai: %w", err)
	}

	// -------------------------------------------------------------------------
	// Train or play the game.

	switch {
	case train:
		return training(ai)

	default:
		return gaming(ai)
	}
}

// =============================================================================

func training(ai *ai.AI) error {

	// -------------------------------------------------------------------------
	// Process any new boards or changes

	l := func(format string, v ...any) {
		fmt.Printf(format, v...)
	}

	err := ai.ProcessBoardFiles(l)

	// -------------------------------------------------------------------------
	// Ask to delete the change file

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nDo you want to delete the change file? (y/n) ")

	question, _ := reader.ReadString('\n')
	if question[:1] != "y" {
		return err
	}

	if err := ai.DeleteChangeLog(); err != nil {
		return err
	}

	fmt.Println("deleted")

	return err
}

func gaming(ai *ai.AI) error {

	// -------------------------------------------------------------------------
	// Create the board and initialize the display

	board, err := board.New(ai)
	if err != nil {
		return fmt.Errorf("new board: %w", err)
	}
	defer board.Shutdown()

	// -------------------------------------------------------------------------
	// Start handling board input

	<-board.Run()

	return nil
}
