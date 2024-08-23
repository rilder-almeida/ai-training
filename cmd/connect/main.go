package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ardanlabs/ai-training/cmd/connect/ai"
	"github.com/ardanlabs/ai-training/cmd/connect/board"
	"github.com/ardanlabs/ai-training/cmd/connect/systems/ollama"
	"github.com/ardanlabs/ai-training/cmd/connect/systems/pg"
	"github.com/ardanlabs/ai-training/foundation/mongodb"
)

const (
	SystemOllama = "ollama"
	SystemPG     = "pg"
)

var (
	train      bool
	debug      bool
	embSystem  string
	embModel   string
	chatSystem string
	chatModel  string
)

func init() {
	flag.BoolVar(&train, "train", false, "process training data")
	flag.BoolVar(&debug, "debug", true, "log debug information")

	embSystem = SystemOllama
	embModel = "mxbai-embed-large" // Needed for Ollama but not PG

	chatSystem = SystemOllama
	chatModel = "gemma2:27b" // llama3.1

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
	// Connect to mongo

	fmt.Println("Connecting to MongoDB ...")

	client, err := mongodb.Connect(ctx, "mongodb://localhost:27017", "ardan", "ardan")
	if err != nil {
		return fmt.Errorf("mongo connect: %w", err)
	}
	defer client.Disconnect(ctx)

	// -------------------------------------------------------------------------
	// Construct the AI support

	fmt.Println("Establish AI support ...")

	var embedder ai.Embedder
	var vecDimension int

	switch embSystem {
	case SystemOllama:
		embedder, err = ollama.NewEmbedder(embModel)
		if err != nil {
			return fmt.Errorf("ollama embedder: %w", err)
		}
		vecDimension = 1024

	case SystemPG:
		apiKey := os.Getenv("PGKEY")
		if apiKey == "" {
			return errors.New("missing PG api key")
		}
		embedder = pg.NewEmbedder(apiKey)
		vecDimension = 512
	}

	var chat ai.Chatter

	switch chatSystem {
	case SystemOllama:
		chat, err = ollama.NewChatter(chatModel)
		if err != nil {
			return fmt.Errorf("ollama chatter: %w", err)
		}
	}

	ai, err := ai.New(client, embedder, chat, vecDimension, debug)
	if err != nil {
		return fmt.Errorf("new ai: %w", err)
	}

	// -------------------------------------------------------------------------
	// Train or play the game

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
