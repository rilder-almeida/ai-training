package game

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/ardanlabs/ai-training/cmd/connect/ai"
)

const (
	rows = 6
	cols = 7
)

type cell struct {
	hasPiece bool
	player   Player
}

type lastMove struct {
	column int
	row    int
	player Player
}

// Board represents the game board and all its state.
type Board struct {
	ai           *ai.AI
	cells        [cols][rows]cell
	lastMove     lastMove
	aiMessage    string
	gameMessage  string
	debugMessage string
	gameOver     bool
	winner       Player
}

// New contructs a game board.
func New(ai *ai.AI) (*Board, error) {
	goingFirst := Players.Blue

	nBig, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		return nil, fmt.Errorf("random number: %w", err)
	}

	if n := nBig.Int64(); n%2 == 0 {
		goingFirst = Players.Red
	}

	board := Board{
		ai: ai,
		lastMove: lastMove{
			column: 4,
			row:    0,
			player: goingFirst,
		},
	}

	return &board, nil
}

// AITurn plays for the AI.
func (b *Board) AITurn() BoardState {
	b.gameMessage = ""
	b.aiMessage = ""
	b.debugMessage = ""

	if b.gameOver {
		b.gameMessage = "game is over"
		return b.ToBoardState()
	}

	// -------------------------------------------------------------------------
	// Find a similar boards from the training data

	boardData, blueMarkers, redMarkers := b.BoardData()

	board, err := b.ai.FindSimilarBoard(boardData)
	if err != nil {
		b.gameMessage = err.Error()
		return b.ToBoardState()
	}

	// -------------------------------------------------------------------------
	// Use the LLM to Pick

	pick, err := b.ai.LLMPick(boardData, board)
	if err != nil {
		b.gameMessage = err.Error()
		return b.ToBoardState()
	}

	choice := -1

	// Does that column have an open space?
	if !b.cells[pick.Column-1][0].hasPiece {
		choice = pick.Column
	}

	// If we didn't find a valid column, find an open one.
	if choice == -1 {
		for i := range 6 {
			if !b.cells[i][0].hasPiece {
				choice = i + 1
				break
			}
		}
	}

	// Calculate what row (6 - 1) to drop the marker in.
	row := -1
	for i := rows - 1; i >= 0; i-- {
		cell := b.cells[choice-1][i]
		if !cell.hasPiece {
			row = i
			break
		}
	}

	if row == -1 {
		b.gameMessage = fmt.Sprintf("column is full: %d", choice)
		return b.ToBoardState()
	}

	// Set this piece in the cells.
	b.cells[choice-1][row].hasPiece = true
	b.cells[choice-1][row].player = Players.Red

	// Mark this last move.
	b.lastMove.player = Players.Red
	b.lastMove.column = choice
	b.lastMove.row = row + 1

	// Check if this move allowed the AI player to win the game.
	b.checkForWinner(choice, row+1)

	// Capture a response by the AI.
	var response string
	switch {
	case b.gameOver:
		if b.winner == Players.Red {
			response, err = b.ai.CreateAIResponse("Won-Game", blueMarkers, redMarkers, choice)
		} else {
			response, err = b.ai.CreateAIResponse("Lost-Game", blueMarkers, redMarkers, choice)
		}

		if err != nil {
			b.gameMessage = err.Error()
		}

	default:
		response, err = b.ai.CreateAIResponse(board.MetaData.Feedback, blueMarkers, redMarkers, choice)
		if err != nil {
			b.gameMessage = err.Error()
		}
	}

	b.aiMessage = response

	// Provide final state for display.
	b.debugMessage = fmt.Sprintf("BOARD: %s CRLF CHOICE: %d - OPTIONS: %v - ATTEMPTS: %d CRLF SCORE: %.2f%% CRLF %s", board.ID, choice, board.MetaData.Moves, pick.Attmepts, board.Score*100, pick.Reason)

	return b.ToBoardState()
}

// UserTurn plays the user's choice.
func (b *Board) UserTurn(column int) BoardState {
	b.gameMessage = ""
	b.aiMessage = ""
	b.debugMessage = ""

	if b.gameOver {
		b.gameMessage = "game is over"
		return b.ToBoardState()
	}

	// -------------------------------------------------------------------------
	// Check if we have a new game board

	boardData, blueMarkers, redMarkers := b.BoardData()
	b.ai.SaveBoardData(boardData, b.winner.name, redMarkers, column, b.gameOver)

	defer func() {
		boardData, _, redMarkers := b.BoardData()
		b.ai.SaveBoardData(boardData, b.winner.name, redMarkers, column, b.gameOver)
	}()

	// -------------------------------------------------------------------------
	// Apply the user's column choice

	column--

	// Calculate what row (6 - 1) to drop the marker in.
	row := -1
	for i := rows - 1; i >= 0; i-- {
		cell := b.cells[column][i]
		if !cell.hasPiece {
			row = i
			break
		}
	}

	if row == -1 {
		b.gameMessage = fmt.Sprintf("column is full: %d", column)
		return b.ToBoardState()
	}

	// Set this piece in the cells.
	b.cells[column][row].hasPiece = true
	b.cells[column][row].player = Players.Blue

	// Mark this last move.
	b.lastMove.player = Players.Blue
	b.lastMove.column = column + 1
	b.lastMove.row = row + 1

	// Check if this move allowed the player to win the game.
	b.checkForWinner(column+1, row+1)

	// Capture a response by the AI.
	if b.gameOver {
		var response string
		var err error

		if b.winner == Players.Red {
			response, err = b.ai.CreateAIResponse("Won-Game", blueMarkers, redMarkers, b.lastMove.column)
		} else {
			response, err = b.ai.CreateAIResponse("Lost-Game", blueMarkers, redMarkers, b.lastMove.column)
		}

		b.aiMessage = response
		if err != nil {
			b.gameMessage = err.Error()
		}
	}

	return b.ToBoardState()
}

// BoardData converts the game board into a text representation.
func (b *Board) BoardData() (boardData string, blue int, red int) {
	var data strings.Builder

	for row := range rows {
		data.WriteString("|")
		for col := range cols {
			cell := b.cells[col][row]
			switch {
			case !cell.hasPiece:
				data.WriteString("🟢|")
			default:
				switch cell.player {
				case Players.Blue:
					data.WriteString("🔵|")
					blue++
				case Players.Red:
					data.WriteString("🔴|")
					red++
				}
			}
		}
		data.WriteString("\n")
	}

	return data.String(), blue, red
}

// =============================================================================

func (b *Board) checkForWinner(colInput int, rowInput int) {
	defer func() {
		if b.gameOver {
			b.gameMessage = fmt.Sprintf("The %s player has won", b.winner)
			if b.winner.IsZero() {
				b.gameMessage = "There was a Tie between the Blue and Red player"
			}
		}
	}()

	colInput--
	rowInput--

	// -------------------------------------------------------------------------
	// Is there a winner in the specified row.

	var red int
	var blue int

	for col := 0; col < cols; col++ {
		if !b.cells[col][rowInput].hasPiece {
			red = 0
			blue = 0
			continue
		}

		switch b.cells[col][rowInput].player {
		case Players.Blue:
			blue++
			red = 0
		case Players.Red:
			red++
			blue = 0
		}

		switch {
		case red == 4:
			b.winner = Players.Red
			b.gameOver = true
			return
		case blue == 4:
			b.winner = Players.Blue
			b.gameOver = true
			return
		}
	}

	// -------------------------------------------------------------------------
	// Is there a winner in the specified column.

	red = 0
	blue = 0

	for row := 0; row < rows; row++ {
		if !b.cells[colInput][row].hasPiece {
			red = 0
			blue = 0
			continue
		}

		switch b.cells[colInput][row].player {
		case Players.Blue:
			blue++
			red = 0
		case Players.Red:
			red++
			blue = 0
		}

		switch {
		case red == 4:
			b.winner = Players.Red
			b.gameOver = true
			return
		case blue == 4:
			b.winner = Players.Blue
			b.gameOver = true
			return
		}
	}

	// -------------------------------------------------------------------------
	// Is there a winner in the NW to SE line.

	red = 0
	blue = 0

	// Walk up in a diagonal until we hit column 0.
	useRow := rowInput
	useCol := colInput
	for useCol != 0 && useRow != 0 {
		useRow--
		useCol--
	}

	for useCol != cols && useRow != rows {
		if !b.cells[useCol][useRow].hasPiece {
			useCol++
			useRow++
			red = 0
			blue = 0
			continue
		}

		switch b.cells[useCol][useRow].player {
		case Players.Blue:
			blue++
			red = 0
		case Players.Red:
			red++
			blue = 0
		}

		switch {
		case red == 4:
			b.winner = Players.Red
			b.gameOver = true
			return
		case blue == 4:
			b.winner = Players.Blue
			b.gameOver = true
			return
		}

		useCol++
		useRow++
	}

	// -------------------------------------------------------------------------
	// Is there a winner in the SW to NE line.

	red = 0
	blue = 0

	// Walk up in a diagonal until we hit column 0.
	useRow = rowInput
	useCol = colInput
	for useCol != cols-1 && useRow != 0 {
		useRow--
		useCol++
	}

	for useCol >= 0 && useRow != rows {
		if !b.cells[useCol][useRow].hasPiece {
			useCol--
			useRow++
			red = 0
			blue = 0
			continue
		}

		switch b.cells[useCol][useRow].player {
		case Players.Blue:
			blue++
			red = 0
		case Players.Red:
			red++
			blue = 0
		}

		switch {
		case red == 4:
			b.winner = Players.Red
			b.gameOver = true
			return
		case blue == 4:
			b.winner = Players.Blue
			b.gameOver = true
			return
		}

		useCol--
		useRow++
	}

	// No winner, but is there a tie?
	tie := true
stop:
	for col := range b.cells {
		for _, cell := range b.cells[col] {
			if !cell.hasPiece {
				tie = false
				break stop
			}
		}
	}

	if tie {
		b.gameOver = true
	}
}
