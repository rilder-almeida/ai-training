package main

import (
	"fmt"
	"log"

	"github.com/ardanlabs/ai-training/cmd/connect/board"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	// -------------------------------------------------------------------------
	// Create the board and initialize the display.

	board, err := board.New()
	if err != nil {
		return fmt.Errorf("new board: %w", err)
	}
	defer board.Shutdown()

	// -------------------------------------------------------------------------
	// Start handling board input.

	<-board.Run()

	return nil
}
