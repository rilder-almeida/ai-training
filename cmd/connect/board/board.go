// Package board handles the game board and all interactions.
package board

import (
	"fmt"
	"strings"
	"time"

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
	colorBlue = "blue"
	colorRed  = "red"
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
	screen      tcell.Screen
	style       tcell.Style
	cells       [cols][rows]cell
	inputCol    int
	currentTurn string
	gameOver    bool
	lastWinner  string
	modalUp     bool
	modalMsg    string
}

// New contructs a game board and renders the board.
func New() (*Board, error) {
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

	board := Board{
		screen:      screen,
		style:       style,
		inputCol:    4,
		currentTurn: colorBlue,
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

// Print display a version of the board.
func (b *Board) Print() string {
	// E,E,E,E,E,E,E
	// E,E,E,E,E,E,E
	// E,E,E,X,E,E,E
	// E,E,E,X,E,E,E
	// E,E,O,X,E,E,E
	// E,X,O,O,O,E,

	var bld strings.Builder

	for row := range rows {
		for col := range cols {
			cell := b.cells[col][row]
			switch {
			case !cell.hasPiece:
				bld.WriteString("-  ")
			default:
				switch cell.color {
				case colorBlue:
					bld.WriteString("O  ")
				case colorRed:
					bld.WriteString("X  ")
				}
			}
		}
		bld.WriteString("\n")
	}

	return bld.String()
}

func (b *Board) newGame() {
	b.inputCol = 4
	b.cells = [cols][rows]cell{}
	b.currentTurn = colorBlue
	b.gameOver = false

	b.drawInit()
}

func (b *Board) drawInit() {
	b.screen.Clear()
	b.drawEmptyGameBoard()
	b.appyBoardState()
}

func (b *Board) drawEmptyGameBoard() {
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

	if !b.gameOver {
		b.print(padLeft+2+(cellWidth*(b.inputCol-1)), padTop-1, "🔵")
	}

	b.print(boardWidth+3, padTop+0, "Last Winner:")
	b.print(boardWidth+3, padTop+2, "<n> : new game")
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

	b.print(boardWidth+3, padTop+0, "Last Winner: "+b.lastWinner)

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

func (b *Board) dropPiece(animate bool) {
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
		return
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
		b.checkForWinner(b.inputCol-1, row-1)

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
}

func (b *Board) checkForWinner(col int, row int) {

	// -------------------------------------------------------------------------
	// Is there a winner in the specified row.

	var red int
	var blue int

	for col := 0; col < cols; col++ {
		if !b.cells[col][row].hasPiece {
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
			b.showWinner("🔴")
			return
		case blue == 4:
			b.showWinner("🔵")
			return
		}
	}

	// -------------------------------------------------------------------------
	// Is there a winner in the specified column.

	red = 0
	blue = 0

	for row := 0; row < rows; row++ {
		if !b.cells[col][row].hasPiece {
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
			b.showWinner("🔴")
			return
		case blue == 4:
			b.showWinner("🔵")
			return
		}
	}

	// -------------------------------------------------------------------------
	// Is there a winner in the NW to SE line.

	red = 0
	blue = 0

	// Walk up in a diagonal until we hit column 0.
	useRow := row
	useCol := col
	for r := row - 1; r >= 0; r-- {
		if useCol == 0 || useRow == 0 {
			break
		}
		useRow--
		useCol--
	}

	for useCol != cols && useRow != rows {
		if !b.cells[useCol][useRow].hasPiece {
			useCol++
			useRow++
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
			b.showWinner("🔴")
			return
		case blue == 4:
			b.showWinner("🔵")
			return
		}

		useCol++
		useRow++
	}

	// -------------------------------------------------------------------------
	// Is there a winner in the SW to NE line.

	red = 0
	blue = 0

	// Walk up in a diagonal until we hit column cols-1.
	useRow = row
	useCol = col
	for r := row; r < rows; r++ {
		if useCol == 0 || useRow == rows-1 {
			break
		}
		useRow++
		useCol--
	}

	for useCol != cols && useRow > 0 {
		if !b.cells[useCol][useRow].hasPiece {
			useCol++
			useRow--
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
			b.showWinner("🔴")
			return
		case blue == 4:
			b.showWinner("🔵")
			return
		}

		useCol++
		useRow--
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
		b.showWinner("")
	}
}

// showWinner displays a modal dialog box.
func (b *Board) showWinner(piece string) {
	message := fmt.Sprintf("%s WINS", piece)
	if piece == "" {
		message = "TIE"
	}

	b.gameOver = true
	b.lastWinner = message
	b.modalUp = true
	b.modalMsg = message

	b.screen.HideCursor()
	b.drawBox(5, 8, 33, 13)

	h := 10
	l := len(message)
	x := 19 - (l / 2)
	b.print(x, h, message)
}

// closeModal closes the modal dialog box.
func (b *Board) closeModal() {
	b.modalUp = false
	b.modalMsg = ""

	b.drawInit()
}

// drawBox draws an empty box on the screen.
func (b *Board) drawBox(x int, y int, width int, height int) {
	style := b.style
	style = style.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)

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
