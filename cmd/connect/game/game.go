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
	// Perform some defensive training to start

	if err := b.learnDefense(); err != nil {
		b.debugMessage = err.Error()
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

	// -------------------------------------------------------------------------
	// Play the choice on the board

	// Set this piece in the cells.
	b.cells[choice-1][row].hasPiece = true
	b.cells[choice-1][row].player = Players.Red

	// Mark this last move.
	b.lastMove.player = Players.Red
	b.lastMove.column = choice
	b.lastMove.row = row + 1

	// Check if this move allowed the AI player to win the game.
	b.checkForWinner(choice, row+1)

	// -------------------------------------------------------------------------
	// Generate the snarky response

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
	// Capture the current state of the board before the player's choice
	// is applied.

	boardData, blueMarkers, redMarkers := b.BoardData()

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

	// -------------------------------------------------------------------------
	// Generate a win or lost response if applicable

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

	// -------------------------------------------------------------------------
	// Perform some training thanks to the blue player

	if err := b.learnFromBlue(boardData, blueMarkers, column, row); err != nil {
		b.debugMessage = err.Error()
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

func (b *Board) learnFromBlue(boardData string, blueMarkers int, column int, row int) error {

	// We want to see if Blue just blocked Red from winning.
	blocked := b.checkIfPlayerWins(column+1, row+1, Players.Red)

	// Save the current board in reverse to pretend Red just moved.
	column++
	if err := b.ai.SaveBoardData(true, boardData, blueMarkers, column, b.winner.String(), blocked); err != nil {
		return err
	}

	return nil
}

func (b *Board) learnDefense() error {

	// -------------------------------------------------------------------------
	// Before we find a similar board, let's check blue can't win at this point
	// because we need to train the game to play some basic defense. We won't
	// use this in the decision making.

	for choice := 1; choice <= 7; choice++ {

		// Which row in the choice column is empty.
		row := -1
		for i := rows - 1; i >= 0; i-- {
			cell := b.cells[choice-1][i]
			if !cell.hasPiece {
				row = i
				break
			}
		}

		if row != -1 {
			if b.checkIfPlayerWins(choice, row+1, Players.Blue) {
				boardData, _, redMarkers := b.BoardData()
				b.ai.SaveBoardData(false, boardData, redMarkers, choice, Players.Blue.name, false)

				// Let's try to train immediately so it can be used.

				l := func(format string, v ...any) {}
				if err := b.ai.ProcessBoardFiles(l); err != nil {
					return err
				}

				if err := b.ai.DeleteChangeLog(); err != nil {
					return err
				}

				return nil
			}
		}
	}

	return nil
}

// checkForWinner checks the current board to see if any player won and
// updates the game state.
func (b *Board) checkForWinner(colInput int, rowInput int) {
	defer func() {
		if b.gameOver {
			b.gameMessage = fmt.Sprintf("The %s player has won", b.winner)
			if b.winner.IsZero() {
				b.gameMessage = "There was a Tie between the Blue and Red player"
			}
		}
	}()

	yes, player := b.checkForAnyWinner(colInput, rowInput)
	switch {
	case yes && player.IsZero(): // Tie Game
		b.gameOver = true
	case yes:
		b.winner = player
		b.gameOver = true
	}
}

// checkForAnyWinner checks the current board to see if any player won.
func (b *Board) checkForAnyWinner(colInput int, rowInput int) (bool, Player) {
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
			return true, Players.Red
		case blue == 4:
			return true, Players.Blue
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
			return true, Players.Red
		case blue == 4:
			return true, Players.Blue
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
			return true, Players.Red
		case blue == 4:
			return true, Players.Blue
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
			return true, Players.Red
		case blue == 4:
			return true, Players.Blue
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
		return true, Player{}
	}

	return false, Player{}
}

// checkIfPlayerWins makes the specified move and checks if the specified
// player will win. The move is reversed when the function returns.
func (b *Board) checkIfPlayerWins(colInput int, rowInput int, player Player) bool {
	colInput--
	rowInput--

	save := b.cells[colInput][rowInput]

	b.cells[colInput][rowInput].player = player
	b.cells[colInput][rowInput].hasPiece = true
	defer func() {
		b.cells[colInput][rowInput] = save
	}()

	// -------------------------------------------------------------------------
	// Does the player win in the specified row.

	var counter int

	for col := 0; col < cols; col++ {
		if !b.cells[col][rowInput].hasPiece {
			counter = 0
			continue
		}

		if b.cells[col][rowInput].player == player {
			counter++
		} else {
			counter = 0
		}

		if counter == 4 {
			return true
		}
	}

	// -------------------------------------------------------------------------
	// Does the player win in the specified column.

	counter = 0

	for row := 0; row < rows; row++ {
		if !b.cells[colInput][row].hasPiece {
			counter = 0
			continue
		}

		if b.cells[colInput][row].player == player {
			counter++
		} else {
			counter = 0
		}

		if counter == 4 {
			return true
		}
	}

	// -------------------------------------------------------------------------
	// Does the player win in the NW to SE line.

	counter = 0

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
			counter = 0
			continue
		}

		if b.cells[useCol][useRow].player == player {
			counter++
		} else {
			counter = 0
		}

		if counter == 4 {
			return true
		}

		useCol++
		useRow++
	}

	// -------------------------------------------------------------------------
	// Does the player win in the SW to NE line.

	counter = 0

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
			counter = 0
			continue
		}

		if b.cells[useCol][useRow].player == player {
			counter++
		} else {
			counter = 0
		}

		if counter == 4 {
			return true
		}

		useCol--
		useRow++
	}

	return false
}
