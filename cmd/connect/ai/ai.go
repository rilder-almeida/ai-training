package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ardanlabs/ai-training/foundation/mongodb"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	trainingDataPath = "cmd/connect/training-data/"
	logFile          = "log.txt"
	changeLogFile    = "change_log.txt"
)

// AI provides support to process connect 4 boards.
type AI struct {
	client *mongo.Client
	col    *mongo.Collection
	embed  *ollama.LLM
	chat   *ollama.LLM
}

// New construct the AI api for use.
func New(client *mongo.Client, embed *ollama.LLM, chat *ollama.LLM) (*AI, error) {
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
		NumDimensions: 1024,
		Path:          "embedding",
		Similarity:    "cosine",
	}

	if err := mongodb.CreateVectorIndex(ctx, col, indexName, settings); err != nil {
		return nil, fmt.Errorf("createVectorIndex: %w", err)
	}

	// -------------------------------------------------------------------------
	// Return the api

	ai := AI{
		client: client,
		col:    col,
		embed:  embed,
		chat:   chat,
	}

	return &ai, nil
}

// CalculateEmbedding takes board data and produces a vector embedding.
func (ai *AI) CalculateEmbedding(boardData string) ([]float32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	image, err := generateImage(boardData)
	embData := base64.StdEncoding.EncodeToString(image)

	embedding, err := ai.embed.CreateEmbedding(ctx, []string{embData})
	if err != nil {
		return nil, fmt.Errorf("create embedding: %w", err)
	}

	return embedding[0], nil
}

// LLMPick perform a review of the game board and makes a choice.
func (ai *AI) LLMPick(boardData string, board SimilarBoard) (PickResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// We need the board to look as the LLM expects it to look.
	// It knows about Red and Yellow disks, so Blue will be Yellow.
	// | . | . | Y | Y | . | . | . |
	// | . | . | . | R | . | . | . |
	// | . | . | . | . | . | . | . |

	boardData = strings.ReplaceAll(boardData, "🟢", " . ")
	boardData = strings.ReplaceAll(boardData, "🔵", " Y ")
	boardData = strings.ReplaceAll(boardData, "🔴", " R ")

	// We have to reverse the board so the rows are flipped.
	rows := strings.Split(boardData, "\n")
	var grid string
	for i := 5; i >= 0; i-- {
		grid = fmt.Sprintf("%s%s\n", grid, rows[i])
	}

	// Format the score to a percentage.
	score := fmt.Sprintf("%.2f", board.Score*100)

	// Check if Red is starting the game.
	// We need the AI to randomaly pick a column and it's not good at that.
	// So we will help it by lowering the score.
	if len(board.MetaData.Moves) == 7 {
		score = "25.00"
	}

	// We need the list of possible moves.
	m := make([]string, len(board.MetaData.Moves))
	for i, v := range board.MetaData.Moves {
		m[i] = fmt.Sprintf("%d", v)
	}

	// Generate the prompt to use to ask the LLM to pick a column.
	prompt := fmt.Sprintf(promptPick, strings.Join(m, ","), score, grid)

	var pick PickResponse

	// The LLM sometimes doesn't pick a column from the list, so we may
	// need to tell the LLM it didn't listen and try again.
	attempts := 1
	for ; attempts <= 2; attempts++ {
		writeLog(prompt)

		// Ask the LLM to choose a column from the training data.
		response, err := ai.chat.Call(ctx, prompt, llms.WithMaxTokens(5000), llms.WithTemperature(0.8))
		if err != nil {
			return PickResponse{}, fmt.Errorf("call: %w", err)
		}

		writeLog("Response:")
		writeLog(response)

		// I had a situation where the response was marked as ``` and ```json.
		response = strings.Trim(response, "`")
		response = strings.TrimPrefix(response, "json")

		if err := json.Unmarshal([]byte(response), &pick); err != nil {
			return PickResponse{}, fmt.Errorf("unmarshal: %w", err)
		}

		// Did the LLM listen and pick a column from the redMoves list?
		var found bool
		for _, v := range board.MetaData.Moves {
			if v == pick.Column {
				found = true
				break
			}
		}

		if found {
			break
		}

		// Tell the LLM they didn't listen and try again.
		prompt = fmt.Sprintf(promptPickAgain, prompt, response)
	}

	writeLogf("\nAttempts: %d", attempts)
	writeLog("------------------")

	pick.Attmepts = attempts

	return pick, nil
}

// FindSimilarBoard performs a vector search to find the most similar board.
func (ai *AI) FindSimilarBoard(boardData string) (SimilarBoard, error) {
	embedding, err := ai.CalculateEmbedding(boardData)
	if err != nil {
		return SimilarBoard{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// We want to find the nearest neighbors from the question vector embedding.
	pipeline := mongo.Pipeline{
		{{
			Key: "$vectorSearch",
			Value: bson.M{
				"index":       "vector_index",
				"exact":       true,
				"path":        "embedding",
				"queryVector": embedding,
				"limit":       1,
			}},
		},
		{{
			Key: "$project",
			Value: bson.M{
				"board_id": 1,
				"board":    1,
				"meta_data": bson.M{
					"markers":  1,
					"moves":    1,
					"feedback": 1,
				},
				"embedding": 1,
				"score": bson.M{
					"$meta": "vectorSearchScore",
				},
			}},
		},
	}

	cur, err := ai.col.Aggregate(ctx, pipeline)
	if err != nil {
		return SimilarBoard{}, fmt.Errorf("aggregate: %w", err)
	}
	defer cur.Close(ctx)

	var boards []SimilarBoard
	if err := cur.All(ctx, &boards); err != nil {
		return SimilarBoard{}, fmt.Errorf("all: %w", err)
	}

	return boards[0], nil
}

// CreateAIResponse produces a game response based on a similar board.
func (ai *AI) CreateAIResponse(prompt string, blueMarkerCount int, redMarkerCounted int, lastMove int) (string, error) {
	switch prompt {
	case "Normal-GamePlay":
		prompt = promptNormalGamePlay

	case "Will-Win":
		prompt = promptWonGame

	case "Blocked-Win":
		prompt = promptBlockedWin

	case "Won-Game":
		prompt = promptWonGame

	case "Lost-Game":
		prompt = promptLostGame

	default:
		prompt = promptNormalGamePlay
	}

	prompt = fmt.Sprintf(prompt, blueMarkerCount, redMarkerCounted, lastMove)

	writeLog(prompt)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	response, err := ai.chat.Call(ctx, prompt, llms.WithMaxTokens(5000), llms.WithTemperature(0.8))
	if err != nil {
		return "", fmt.Errorf("call: %w", err)
	}

	writeLog("Response:")
	writeLog(response)
	writeLog("\n------------------")

	return response, nil
}

// SaveBoardData knows how to write a board file with the following information.
func (ai *AI) SaveBoardData(reverse bool, boardData string, markers int, lastMove int, winner string, blocked bool) error {

	// In case this will be a new board, this will represent the possible
	// moves people have made so far when seeing this board configuration.
	moves := []int{lastMove}

	// -------------------------------------------------------------------------
	// Check if we have captured this board already so we can update the
	// moves on the board.

	var foundMatch string

	// Reverse the board data because we want to pretend the blue player
	// is playing as red.
	if reverse {
		boardData = strings.ReplaceAll(boardData, "🔵", "R")
		boardData = strings.ReplaceAll(boardData, "🔴", "🔵")
		boardData = strings.ReplaceAll(boardData, "R", "🔴")
	}

	// Iterate over every file until we find a match or we looked at
	// everything.
	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		if foundMatch != "" {
			return errors.New("found match")
		}

		if err != nil {
			return fmt.Errorf("walkdir failure: %w", err)
		}

		if fileName == "." {
			return nil
		}

		boardID := strings.TrimSuffix(fileName, ".txt")
		board, err := ai.readBoardFromDisk(boardID)
		if err != nil {
			return fmt.Errorf("reading key file: %w", err)
		}

		if strings.Compare(boardData, board.Board) == 0 {
			moves = board.MetaData.Moves
			foundMatch = boardID

			var foundMove bool
			for _, v := range board.MetaData.Moves {
				if v == lastMove {
					foundMove = true
					break
				}
			}

			if !foundMove {
				moves = append([]int{lastMove}, board.MetaData.Moves...)
			}

			return errors.New("found match")
		}

		return nil
	}

	fsys := os.DirFS(trainingDataPath)
	fs.WalkDir(fsys, ".", fn)

	// -------------------------------------------------------------------------
	// Save or update this board.

	fileID := foundMatch
	if foundMatch == "" {
		fileID = uuid.NewString()
	}

	// TODO: NEED TO DEAL WITH PLAYER RED/BLUE

	feedback := "Normal-GamePlay"
	switch {
	case reverse && winner != "" && winner == "Blue":
		feedback = "Will-Win"
	case !reverse && winner != "" && winner == "Blue":
		feedback = "Blocked-Win"
	case blocked:
		feedback = "Blocked-Win"
	}

	m := make([]string, len(moves))
	for i, v := range moves {
		m[i] = fmt.Sprintf("%d", v)
	}

	f, err := os.Create(trainingDataPath + fileID + ".txt")
	if err != nil {
		return fmt.Errorf("create training data: %w", err)
	}
	defer f.Close()

	template := `%s
{
    "markers": %d,
    "moves": [%s],
    "feedback": %q
}
`

	_, err = fmt.Fprintf(f,
		template,
		boardData,
		markers,
		strings.Join(m, ","),
		feedback)

	if err != nil {
		return err
	}

	writeChangeLog(fileID)

	return nil
}

// ProcessBoardFiles reads the training data and creates all the vector
// embeddings, storing that inside the vector database.
func (ai *AI) ProcessBoardFiles(l func(format string, v ...any)) error {
	log := func(format string, v ...any) {
		writeLogf(format, v...)
		l(format, v...)
	}

	files, err := os.ReadDir(trainingDataPath)
	if err != nil {
		return fmt.Errorf("read training data directory: %w", err)
	}

	var changes []string
	changeFile, err := os.ReadFile(changeLogFile)
	if err == nil {
		changes = strings.Split(string(changeFile), "\n")
	}

	var count int
	total := len(files)

	log("Found %d documents to process\n", total)

	for _, file := range files {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		count++

		boardID := strings.TrimSuffix(file.Name(), ".txt")

		if _, err := ai.findBoardInDB(ctx, boardID); err == nil {
			if changeFile == nil {
				continue
			}

			// Does this board show up in the change file?
			var found bool
			for _, id := range changes {
				if id == boardID {

					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		log("-------------------------------------------\n")
		log("Processing file %d of %d\n", count, total)

		log("Creating board: %s\n", boardID)

		board, err := ai.readBoardFromDisk(boardID)
		if err != nil {
			return fmt.Errorf("new board: %s: %w", boardID, err)
		}

		log("Create board embedding: %s\n", boardID)

		board, err = ai.createEmbedding(board)
		if err != nil {
			return fmt.Errorf("create embedding: %s: %w", boardID, err)
		}

		log("Saving board data: %s\n", boardID)

		if err := ai.saveBoard(ctx, board); err != nil {
			return fmt.Errorf("save board: %s: %w", boardID, err)
		}
	}

	log("Processing file %d of %d\n", count, total)

	return nil
}

// GitUpdate takes new files and changes and pushes them to git.
func (ai *AI) GitUpdate(l func(format string, v ...any)) error {
	log := func(format string, v ...any) {
		writeLogf(format, v...)
		l(format, v...)
	}

	out, err := exec.Command("git", "add", "-A").CombinedOutput()
	if err != nil {
		return err
	}

	log(string(out))

	out, err = exec.Command("git", "commit", "-m", "save game data").CombinedOutput()
	if err != nil {
		return err
	}

	log(string(out))

	out, err = exec.Command("git", "push").CombinedOutput()
	if err != nil {
		return err
	}

	log(string(out))

	return nil
}

// =============================================================================

func (ai *AI) findBoardInDB(ctx context.Context, boardID string) (Board, error) {
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

func (ai *AI) readBoardFromDisk(boardID string) (Board, error) {
	fileName := fmt.Sprintf("%s%s.txt", trainingDataPath, boardID)

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
	var md strings.Builder
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

		// Remove the extra CRLF.
		if lineCount == 7 {
			continue
		}

		// Capture the game board metadata.
		md.WriteString(scanner.Text())
		md.WriteString("\n")
	}

	boardData := board.String()

	var metaData MetaData
	if err := json.Unmarshal([]byte(md.String()), &metaData); err != nil {
		return Board{}, err
	}

	b := Board{
		ID:       boardID,
		Board:    boardData,
		MetaData: metaData,
	}

	return b, nil
}

func (ai *AI) createEmbedding(board Board) (Board, error) {
	embedding, err := ai.CalculateEmbedding(board.Board)
	if err != nil {
		return Board{}, fmt.Errorf("calculate embedding: %s: %w", board.ID, err)
	}

	board.Embedding = embedding

	return board, nil
}

func (ai *AI) saveBoard(ctx context.Context, board Board) error {
	filter := bson.D{{Key: "board_id", Value: board.ID}}

	d := Board{
		ID:        board.ID,
		Board:     board.Board,
		MetaData:  board.MetaData,
		Embedding: board.Embedding,
	}

	ai.col.DeleteOne(ctx, filter)

	if _, err := ai.col.InsertOne(ctx, d); err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	// var uo options.UpdateOptions
	// if _, err := ai.col.UpdateOne(ctx, filter, d, uo.SetUpsert(true)); err != nil {
	// 	return fmt.Errorf("insert: %w", err)
	// }

	return nil
}
