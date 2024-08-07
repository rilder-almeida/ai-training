package game

// Cell represents a cell in the game board.
type Cell struct {
	HasPiece bool   `json:"hasPiece"`
	Player   Player `json:"player"`
}

// LastMove represents the last move in the game.
type LastMove struct {
	Column int    `json:"column"`
	Row    int    `json:"row"`
	Player Player `json:"player"`
}

// BoardState represent the state of the board for any UI to display.
type BoardState struct {
	Cells        [cols][rows]Cell `json:"cells"`
	LastMove     LastMove         `json:"lastMove"`
	AIMessage    string           `json:"aiMessage"`
	GameMessage  string           `json:"GameMessage"`
	DebugMessage string           `json:"DebugMessage"`
	GameOver     bool             `json:"gameOver"`
	Winner       Player           `json:"winner"`
}

// ToBoardState represents what we will get from an API.
func (b *Board) ToBoardState() BoardState {
	var cells [cols][rows]Cell
	for c := range b.cells {
		for r := range b.cells[c] {
			cells[c][r].HasPiece = b.cells[c][r].hasPiece
			cells[c][r].Player = b.cells[c][r].player
		}
	}

	return BoardState{
		Cells: cells,
		LastMove: LastMove{
			Column: b.lastMove.column,
			Row:    b.lastMove.row,
			Player: b.lastMove.player,
		},
		AIMessage:    b.aiMessage,
		GameMessage:  b.gameMessage,
		DebugMessage: b.debugMessage,
		GameOver:     b.gameOver,
		Winner:       b.winner,
	}
}
