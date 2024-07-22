package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ardanlabs/ai-training/cmd/connect/board"
	"github.com/ardanlabs/ai-training/foundation/mongodb"
	"github.com/tmc/langchaingo/llms/ollama"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var train bool

func init() {
	flag.BoolVar(&train, "train", false, "process training data")

	flag.Parse()
}

func main() {
	if train {
		if err := saveTrainingData(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := game(); err != nil {
		log.Fatal(err)
	}
}

// =============================================================================

func game() error {

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

// =============================================================================

func saveTrainingData() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// -------------------------------------------------------------------------
	// Connect to mongo

	fmt.Println("Connecting to MongoDB ...")

	client, err := mongodb.Connect(ctx, "mongodb://localhost:27017", "ardan", "ardan")
	if err != nil {
		return fmt.Errorf("connectToMongo: %w", err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Connected to MongoDB")

	// -------------------------------------------------------------------------
	// Create database and collection

	const dbName = "connect4"
	const collectionName = "boards"

	db := client.Database(dbName)

	col, err := mongodb.CreateCollection(ctx, db, collectionName)
	if err != nil {
		return fmt.Errorf("createCollection: %w", err)
	}

	fmt.Println("Created Collection")

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
		return fmt.Errorf("createVectorIndex: %w", err)
	}

	fmt.Println("Created Indexes")

	// -------------------------------------------------------------------------
	// Open a connection with ollama to access the model.

	llm, err := ollama.New(ollama.WithModel("mxbai-embed-large"))
	if err != nil {
		return fmt.Errorf("ollama: %w", err)
	}

	// -------------------------------------------------------------------------
	// Process all the files in the board-files directory.

	return processFiles(col, llm)
}

func processFiles(col *mongo.Collection, llm *ollama.LLM) error {
	const filePath = "cmd/connect/board/board-files/"

	files, err := os.ReadDir(filePath)
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

		filter := bson.D{{Key: "board_id", Value: boardID}}
		res := col.FindOne(ctx, filter)
		if res == nil {
			fmt.Printf("Checking is file exists in DB: %s, EXISTS\n", boardID)
			continue
		}

		if !errors.Is(res.Err(), mongo.ErrNoDocuments) {
			fmt.Printf("Checking is file exists in DB: %s, EXISTS\n", boardID)
			continue
		}

		f, err := os.Open(filePath + file.Name())
		if err != nil {
			return fmt.Errorf("open file: %s: %w", boardID, err)
		}

		err = func() error {
			defer f.Close()

			data, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("read file: %s: %w", boardID, err)
			}

			if err := saveFile(ctx, col, llm, boardID, data); err != nil {
				return fmt.Errorf("create document: %s: %w", file.Name(), err)
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func saveFile(ctx context.Context, col *mongo.Collection, llm *ollama.LLM, boardID string, data []byte) error {
	var board strings.Builder
	var text strings.Builder
	var lineCount int

	fmt.Println("Scanning file")

	// Separate the board from the board content.
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		lineCount++

		if lineCount <= 6 {
			board.WriteString(scanner.Text())
			board.WriteString("\n")
			continue
		}

		text.WriteString(scanner.Text())
		text.WriteString("\n")
	}

	fmt.Println("Creating embedding")

	// Get the vector embedding for this board.
	s := board.String()
	s = s[:len(s)-1]
	embedding, err := llm.CreateEmbedding(ctx, []string{s})
	if err != nil {
		return fmt.Errorf("create embedding: %w", err)
	}

	d := struct {
		BoardID   string    `bson:"board_id"`
		Text      string    `bson:"text"`
		Embedding []float32 `bson:"embedding"`
	}{
		BoardID:   boardID,
		Text:      text.String(),
		Embedding: embedding[0],
	}

	fmt.Println("Saving file")

	res, err := col.InsertOne(ctx, d)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	fmt.Printf("File Saved: %v\n", res.InsertedID)

	return nil
}
