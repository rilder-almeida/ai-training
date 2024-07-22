package ai

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/ardanlabs/ai-training/foundation/mongodb"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms/ollama"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Board represents connect 4 board information.
type Board struct {
	ID        string    `bson:"board_id"`
	Board     string    `bson:"board"`
	Text      string    `bson:"text"`
	Embedding []float32 `bson:"embedding"`
}

// AI provides support to process connect 4 boards.
type AI struct {
	filePath string
	client   *mongo.Client
	col      *mongo.Collection
	llm      *ollama.LLM
}

// New construct the AI api for use.
func New(client *mongo.Client, llm *ollama.LLM) (*AI, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// -------------------------------------------------------------------------
	// Create database and collection

	const dbName = "connect4"
	const collectionName = "boards"

	db := client.Database(dbName)

	col, err := mongodb.CreateCollection(ctx, db, collectionName)
	if err != nil {
		return nil, fmt.Errorf("createCollection: %w", err)
	}

	// -------------------------------------------------------------------------
	// Create indexes

	unique := true
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "board_id", Value: 1}},
		Options: &options.IndexOptions{Unique: &unique},
	}
	col.Indexes().CreateOne(ctx, indexModel)

	const indexName = "vector_index"
	settings := mongodb.VectorIndexSettings{
		NumDimensions: 4,
		Path:          "embedding",
		Similarity:    "cosine",
	}

	if err := mongodb.CreateVectorIndex(ctx, col, indexName, settings); err != nil {
		return nil, fmt.Errorf("createVectorIndex: %w", err)
	}

	// -------------------------------------------------------------------------
	// Return the api

	ai := AI{
		filePath: "cmd/connect/board-files/",
		client:   client,
		col:      col,
		llm:      llm,
	}

	return &ai, nil
}

// SaveBoardData knows how to write a board file with the following information.
func (ai *AI) SaveBoardData(boardData string, blue int, red int, gameOver bool, lastWinner string) string {

	// -------------------------------------------------------------------------
	// Check if we have captured this board alread.

	var foundMatch bool

	fsys := os.DirFS(ai.filePath)

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		if foundMatch {
			return errors.New("found match")
		}

		if err != nil {
			return fmt.Errorf("walkdir failure: %w", err)
		}

		file, err := fsys.Open(fileName)
		if err != nil {
			return fmt.Errorf("opening key file: %w", err)
		}
		defer file.Close()

		var board strings.Builder
		var lineCount int

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			board.WriteString(scanner.Text())
			board.WriteString("\n")

			lineCount++
			if lineCount == 6 {
				break
			}
		}

		if strings.Compare(boardData, board.String()) == 0 {
			foundMatch = true
		}

		return nil
	}

	fs.WalkDir(fsys, ".", fn)

	if foundMatch {
		return "** BOARD FOUND **"
	}

	// -------------------------------------------------------------------------
	// Save a copy of this board and extra information.

	f, _ := os.Create("cmd/connect/board/board-files/" + uuid.NewString() + ".txt")
	defer f.Close()

	f.WriteString(boardData)
	f.WriteString("\n")

	switch {
	case blue == 1 && (red == 0 || red > 1):
		fmt.Fprintf(f, "There is %d space occupied by a Blue marker and %d spaces occupied by Red markers on the game board.\n\n", blue, red)
	case red == 1 && (blue == 0 || blue > 1):
		fmt.Fprintf(f, "There are %d spaces occupied by Blue markers and %d space occupied by a Red marker on the game board.\n\n", blue, red)
	case blue == 1 && red == 1:
		fmt.Fprintf(f, "There is %d space occupied by a Blue marker and %d space occupied by a Red marker on the game board.\n\n", blue, red)
	default:
		fmt.Fprintf(f, "There are %d spaces occupied by Blue markers and %d spaces occupied by Red markers on the game board.\n\n", blue, red)
	}

	switch gameOver {
	case true:
		if lastWinner == "Tie Game" {
			f.WriteString("The game is over and Red and Blue have tied the game.\n")
		} else {
			fmt.Fprintf(f, "The game is over and %s has won the game.\n", lastWinner)
		}
	default:
		switch {
		case blue > red:
			f.WriteString("The Red player goes next and they should choose one of the following columns from the specified list:\n")
		case red > blue:
			f.WriteString("The Blue player goes next and they should choose one of the following columns from the specified list:\n")
		case red == blue:
			f.WriteString("If the Blue player goes next they should choose one of the following columns from the specified list:\n\n")
			f.WriteString("If the Red player goes next they should choose one of the following columns from the specified list:\n")
		}
	}

	return "** BOARD SAVED **"
}

// ProcessBoardFiles reads the board-files directory and saved the AI data needed
// for the AI to play connect 4.
func (ai *AI) ProcessBoardFiles() error {
	files, err := os.ReadDir(ai.filePath)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}

	var count int
	total := len(files)

	fmt.Printf("Found %d documents to process\n", total)

	for _, file := range files {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		fmt.Println("-------------------------------------------")

		count++
		fmt.Printf("Processing file %d of %d\n", count, total)

		boardID := strings.TrimSuffix(file.Name(), ".txt")

		fmt.Printf("Checking if file exists in DB: %s\n", boardID)

		if _, err := ai.findBoard(ctx, boardID); err == nil {
			fmt.Printf("Checking is file exists in DB: %s, EXISTS\n", boardID)
			continue
		}

		fmt.Printf("Creating board data: %s\n", boardID)

		board, err := ai.newBoard(ctx, boardID)
		if err != nil {
			return fmt.Errorf("new board: %s: %w", boardID, err)
		}

		fmt.Printf("Saving board data: %s\n", boardID)

		if err := ai.saveBoard(ctx, board); err != nil {
			return fmt.Errorf("saving board: %s: %w", boardID, err)
		}
	}

	return nil
}

func (ai *AI) findBoard(ctx context.Context, boardID string) (Board, error) {
	filter := bson.D{{Key: "board_id", Value: boardID}}
	res := ai.col.FindOne(ctx, filter)
	if res.Err() != nil {
		return Board{}, res.Err()
	}

	var b Board
	if err := res.Decode(&b); err != nil {
		return Board{}, fmt.Errorf("decode: %w", err)
	}

	return b, nil
}

func (ai *AI) newBoard(ctx context.Context, boardID string) (Board, error) {
	fileName := fmt.Sprintf("%s%s.txt", ai.filePath, boardID)

	f, err := os.Open(fileName)
	if err != nil {
		return Board{}, fmt.Errorf("open file: %s: %w", boardID, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return Board{}, fmt.Errorf("read file: %s: %w", boardID, err)
	}

	var board strings.Builder
	var text strings.Builder
	var lineCount int

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		lineCount++

		// Capture the game board.
		if lineCount <= 6 {
			board.WriteString(scanner.Text())
			board.WriteString("\n")
			continue
		}

		// Capture the game board content.
		text.WriteString(scanner.Text())
		text.WriteString("\n")
	}

	boardData := board.String()

	embedding, err := ai.calculateEmbedding(ctx, boardData)
	if err != nil {
		return Board{}, fmt.Errorf("calculate embedding: %s: %w", boardID, err)
	}

	b := Board{
		ID:        boardID,
		Board:     boardData,
		Text:      text.String(),
		Embedding: embedding,
	}

	return b, nil
}

func (ai *AI) calculateEmbedding(ctx context.Context, boardData string) ([]float32, error) {
	embedding, err := ai.llm.CreateEmbedding(ctx, []string{boardData})
	if err != nil {
		return nil, fmt.Errorf("create embedding: %w", err)
	}

	return embedding[0], nil
}

func (ai *AI) saveBoard(ctx context.Context, board Board) error {
	d := Board{
		ID:        board.ID,
		Board:     board.Board,
		Text:      board.Text,
		Embedding: board.Embedding,
	}

	if _, err := ai.col.InsertOne(ctx, d); err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}
