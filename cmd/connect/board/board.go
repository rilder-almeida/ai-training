// Package board handles the game board and all interactions.
package board

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ardanlabs/ai-training/cmd/connect/ai"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

const (
	rows        = 6
	cols        = 7
	cellWidth   = 5
	cellHeight  = 2
	boardWidth  = cols*cellWidth + 1
	boardHeight = rows * cellHeight
	padTop      = 4
	padLeft     = 1
)

const (
	hozTopRune = '━'
	hozBotRune = '▅'
	verRune    = '┃'
	space      = 32
)

const (
	colorBlue = "Blue"
	colorRed  = "Red"
)

const (
	dirLeft  = "left"
	dirRight = "right"
)

type cell struct {
	hasPiece bool
	color    string
}

// Board represents the game board and all its state.
type Board struct {
	ai            *ai.AI
	screen        tcell.Screen
	style         tcell.Style
	cells         [cols][rows]cell
	inputCol      int
	currentTurn   string
	lastWinner    string
	lastWinnerMsg string
	lastAIMsg     string
	gameOver      bool
	modalUp       bool
}

// New contructs a game board and renders the board.
func New(ai *ai.AI) (*Board, error) {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("new screen: %w", err)
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("screen init: %w", err)
	}

	style := tcell.StyleDefault
	style = style.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)

	currentTurn := colorBlue
	nBig, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		return nil, fmt.Errorf("random number: %w", err)
	}

	if n := nBig.Int64(); n%2 == 0 {
		currentTurn = colorRed
	}

	board := Board{
		ai:          ai,
		screen:      screen,
		style:       style,
		inputCol:    4,
		currentTurn: currentTurn,
	}

	board.drawInit()

	return &board, nil
}

// Shutdown tears down the game board.
func (b *Board) Shutdown() {
	b.screen.Fini()
}

// Run starts a goroutine to handle terminal events. This is a
// blocking call.
func (b *Board) Run() chan struct{} {
	return b.pollEvents()
}

func (b *Board) newGame() {
	b.inputCol = 4
	b.cells = [cols][rows]cell{}
	b.gameOver = false
	b.lastAIMsg = ""

	if b.lastWinner != "" {
		b.currentTurn = b.lastWinner
	}

	b.drawInit()
}

func (b *Board) drawInit() {
	b.drawEmptyGameBoard()
	b.appyBoardState()
}

func (b *Board) drawEmptyGameBoard() {
	b.screen.Clear()

	width := boardWidth
	height := boardHeight

	style := b.style
	style = style.Background(tcell.ColorBlack).Foreground(tcell.ColorGrey)

	for h := 0; h <= height; h++ {
		for w := 0; w < width; w++ {

			// Clear the entire line.
			b.screen.SetContent(w+padLeft, h+padTop, space, nil, style)

			if h == 0 || h%cellHeight == 0 {

				// These are the '━' characters creating each row.
				b.screen.SetContent(w+padLeft, h+padTop, hozTopRune, nil, style)

				if h == height {

					// These are the '▅' characters creating the bottom row.
					b.screen.SetContent(w+padLeft, h+padTop, hozBotRune, nil, style)
				}
			}

			if w == 0 || w%cellWidth == 0 {

				// These are the '┃' characters creating each column.
				b.screen.SetContent(w+padLeft, h+padTop, verRune, nil, style)
			}
		}
	}

	b.print(10, 1, "Connect 4 AI Version")
	b.print(0, boardHeight+padTop+1, "   ①    ②    ③    ④    ⑤    ⑥    ⑦")

	b.print(boardWidth+3, padTop-1, "<n> new game      <q> quit game")
	b.print(boardWidth+3, padTop+1, "Last Winner:                   ")

	screenWidth, _ := b.screen.Size()

	b.drawBox(boardWidth+3, padTop+3, boardWidth+(screenWidth-boardWidth-2), padTop+3+10)
	b.print(boardWidth+4, padTop+3, " AI PLAYER ")
}

func (b *Board) appyBoardState() {

	// Need the cells to be empty to use the dropPiece function.
	oldCells := b.cells
	b.cells = [cols][rows]cell{}

	// Just drop the pieces again, but without animation.
	for col := range oldCells {
		for row := rows - 1; row >= 0; row-- {
			cell := oldCells[col][row]
			if !cell.hasPiece {
				continue
			}

			b.inputCol = col + 1
			b.currentTurn = cell.color
			b.dropPiece(false)
		}
	}

	b.print(boardWidth+3, padTop+1, "Last Winner: "+b.lastWinnerMsg)
	b.printAI()

	if !b.gameOver {
		var whichColor string
		switch b.gameOver {
		case true:
			whichColor = b.lastWinner
		default:
			whichColor = b.currentTurn
		}

		switch whichColor {
		case colorBlue:
			b.print(padLeft+2+(cellWidth*(b.inputCol-1)), padTop-1, "🔵")
		default:
			b.print(padLeft+2+(cellWidth*(b.inputCol-1)), padTop-1, "🔴")
		}
	}

	b.screen.Show()
}

func (b *Board) movePlayerPiece(direction string) {
	if b.gameOver {
		return
	}

	if direction == dirLeft && b.inputCol == 1 {
		return
	}

	if direction == dirRight && b.inputCol == cols {
		return
	}

	// Clear the current marker location.
	column := padLeft + 2
	if b.inputCol > 1 {
		column = column + (cellWidth * (b.inputCol - 1))
	}
	b.print(column, padTop-1, " ")

	// Move the marker column location.
	switch direction {
	case dirLeft:
		b.inputCol--
	case dirRight:
		b.inputCol++
	}

	// Display the marker again in the new location.
	column = padLeft + 2
	if b.inputCol > 1 {
		column = column + (cellWidth * (b.inputCol - 1))
	}

	switch b.currentTurn {
	case colorBlue:
		b.print(column, padTop-1, "🔵")
	case colorRed:
		b.print(column, padTop-1, "🔴")
	}
}

func (b *Board) dropPiece(animate bool) bool {

	// Identify where the input marker is located.
	column := padLeft + 2
	if b.inputCol > 1 {
		column = column + (cellWidth * (b.inputCol - 1))
	}
	stopRow := padTop + 1

	// Calculate what row to drop the marker in.
	row := -1
	for i := rows - 1; i >= 0; i-- {
		cell := b.cells[b.inputCol-1][i]
		if !cell.hasPiece {
			row = i
			break
		}
	}

	// Is the column full.
	if row == -1 {
		return false
	}

	// Set this piece in the cells.
	b.cells[b.inputCol-1][row].hasPiece = true
	b.cells[b.inputCol-1][row].color = b.currentTurn

	// We don't use index 0 for the display, so we need to adjust.
	row++

	// Clear the marker.
	b.print(column, padTop-1, " ")

	// Drop the marker into that row.
	for r := 1; r <= row; r++ {
		switch b.currentTurn {
		case colorBlue:
			b.print(column, stopRow, "🔵")
		case colorRed:
			b.print(column, stopRow, "🔴")
		}

		if r < row {
			if animate {
				time.Sleep(250 * time.Millisecond)
			}
			b.print(column, stopRow, " ")
			stopRow = stopRow + cellHeight
		}
	}

	if animate {
		// Check for winner based on the marker being placed
		// in this location.
		if isWinner := b.checkForWinner(b.inputCol-1, row-1); isWinner {
			return true
		}

		// Set the next input marker.
		b.inputCol = 4
		switch b.currentTurn {
		case colorBlue:
			b.currentTurn = colorRed
			b.print(padLeft+2+(cellWidth*(b.inputCol-1)), padTop-1, "🔴")
		case colorRed:
			b.currentTurn = colorBlue
			b.print(padLeft+2+(cellWidth*(b.inputCol-1)), padTop-1, "🔵")
		}
	}

	return false
}

func (b *Board) checkForWinner(col int, row int) bool {

	// -------------------------------------------------------------------------
	// Is there a winner in the specified row.

	var red int
	var blue int

	for col := 0; col < cols; col++ {
		if !b.cells[col][row].hasPiece {
			red = 0
			blue = 0
			continue
		}

		switch b.cells[col][row].color {
		case colorBlue:
			blue++
			red = 0
		case colorRed:
			red++
			blue = 0
		}

		switch {
		case red == 4:
			b.showWinner(colorRed)
			return true
		case blue == 4:
			b.showWinner(colorBlue)
			return true
		}
	}

	// -------------------------------------------------------------------------
	// Is there a winner in the specified column.

	red = 0
	blue = 0

	for row := 0; row < rows; row++ {
		if !b.cells[col][row].hasPiece {
			red = 0
			blue = 0
			continue
		}

		switch b.cells[col][row].color {
		case colorBlue:
			blue++
			red = 0
		case colorRed:
			red++
			blue = 0
		}

		switch {
		case red == 4:
			b.showWinner(colorRed)
			return true
		case blue == 4:
			b.showWinner(colorBlue)
			return true
		}
	}

	// -------------------------------------------------------------------------
	// Is there a winner in the NW to SE line.

	red = 0
	blue = 0

	// Walk up in a diagonal until we hit column 0.
	useRow := row
	useCol := col
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

		switch b.cells[useCol][useRow].color {
		case colorBlue:
			blue++
			red = 0
		case colorRed:
			red++
			blue = 0
		}

		switch {
		case red == 4:
			b.showWinner(colorRed)
			return true
		case blue == 4:
			b.showWinner(colorBlue)
			return true
		}

		useCol++
		useRow++
	}

	// -------------------------------------------------------------------------
	// Is there a winner in the SW to NE line.

	red = 0
	blue = 0

	// Walk up in a diagonal until we hit column 0.
	useRow = row
	useCol = col
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

		switch b.cells[useCol][useRow].color {
		case colorBlue:
			blue++
			red = 0
		case colorRed:
			red++
			blue = 0
		}

		switch {
		case red == 4:
			b.showWinner(colorRed)
			return true
		case blue == 4:
			b.showWinner(colorBlue)
			return true
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
		b.showWinner("Tie Game")
	}

	return false
}

// showWinner displays a modal dialog box.
func (b *Board) showWinner(color string) {
	switch color {
	case colorBlue:
		b.lastWinner = color
		b.lastWinnerMsg = "Blue (🔵)"
	case colorRed:
		b.lastWinner = color
		b.lastWinnerMsg = "Red (🔴)"
	default:
		b.lastWinnerMsg = "Tie Game"
	}

	b.gameOver = true
	b.modalUp = true

	b.screen.HideCursor()
	b.drawBox(5, 8, 33, 13)

	h := 10
	l := len(b.lastWinnerMsg)
	x := 19 - (l / 2)
	b.print(x, h, b.lastWinnerMsg)
}

// closeModal closes the modal dialog box.
func (b *Board) closeModal() {
	b.modalUp = false

	b.drawInit()
}

// drawBox draws an empty box on the screen.
func (b *Board) drawBox(x int, y int, width int, height int) {
	style := b.style
	style = style.Background(tcell.ColorBlack).Foreground(tcell.ColorGray)

	for h := y; h < height; h++ {
		for w := x; w < width; w++ {
			b.screen.SetContent(w, h, ' ', nil, b.style)
		}
	}

	for h := y; h < height; h++ {
		for w := x; w < width; w++ {
			if h == y {
				b.screen.SetContent(w, h, '▀', nil, style)
			}
			if h == height-1 {
				b.screen.SetContent(w, h, '▄', nil, style)
			}
			if w == x || w == width-1 {
				b.screen.SetContent(w, h, '█', nil, style)
			}
		}
	}

	b.screen.Show()
}

func (b *Board) print(x, y int, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		b.screen.SetContent(x, y, c, comb, b.style)
		x += w
	}
	b.screen.Show()
}

func (b *Board) printAI() {
	screenWidth, _ := b.screen.Size()
	actWidth := (screenWidth - boardWidth - 9)

	row := boardWidth + 5
	col := padTop + 4

	for range 8 {
		for range actWidth {
			b.print(row, col, " ")
			row++
		}
		row = boardWidth + 5
		col++
	}

	row = boardWidth + 5
	col = padTop + 4

	scanner := bufio.NewScanner(bytes.NewReader([]byte(b.lastAIMsg)))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		if word == "CRLF" {
			col++
			row = boardWidth + 5
			continue
		}

		b.print(row, col, word)

		row += len(word) + 1
		if row >= boardWidth+actWidth-4 {
			col++
			row = boardWidth + 5
		}
	}
}

func (b *Board) runAISupport() {
	b.lastAIMsg = ""
	b.printAI()

	// -------------------------------------------------------------------------
	// Create a copy of the board.

	var boardData strings.Builder

	var blue int
	var red int

	for row := range rows {
		boardData.WriteString("|")
		for col := range cols {
			cell := b.cells[col][row]
			switch {
			case !cell.hasPiece:
				boardData.WriteString("🟢|")
			default:
				switch cell.color {
				case colorBlue:
					boardData.WriteString("🔵|")
					blue++
				case colorRed:
					boardData.WriteString("🔴|")
					red++
				}
			}
		}
		boardData.WriteString("\n")
	}

	// -------------------------------------------------------------------------
	// Save the board data.

	data := boardData.String()

	display := b.ai.SaveBoardData(data, blue, red, b.gameOver, b.lastWinner)

	if display != "" {
		b.lastAIMsg = fmt.Sprintf("- %s CRLF", display)
		b.printAI()
	}

	// If blue goes there is nothing more for the AI to do.
	if b.currentTurn == colorBlue {
		return
	}

	// -------------------------------------------------------------------------
	// Show AI information

	if display == "" {
		b.lastAIMsg = "- RUNNING AI"
	} else {
		b.lastAIMsg = fmt.Sprintf("- %s CRLF - RUNNING AI", display)
	}

	b.printAI()

	board, err := b.ai.FindSimilarBoard(data)
	if err != nil {
		b.lastAIMsg = err.Error()
		b.printAI()
		return
	}

	b.lastAIMsg = fmt.Sprintf("SCORE: %.2f%% CRLF %s", board.Score*100, board.Text)
	b.printAI()

	// -------------------------------------------------------------------------
	// Move AI piece.

	b.movePlayerPiece(dirLeft)
}
