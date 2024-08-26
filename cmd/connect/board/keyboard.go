package board

import (
	"fmt"
	"runtime/debug"

	"github.com/ardanlabs/ai-training/cmd/connect/game"
	"github.com/gdamore/tcell/v2"
)

// pollEvents starts a goroutine to handle terminal events.
func (b *Board) pollEvents() chan struct{} {
	quit := make(chan struct{})

	boardState := b.gameBoard.ToBoardState()

	if boardState.LastMove.Player == game.Players.Blue {
		boardState = b.aiTurn()
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.screen.Clear()
				fmt.Println(r)
				debug.PrintStack()
			}
		}()

		for {
			if !boardState.GameOver {
				if boardState.LastMove.Player == game.Players.Blue {
					boardState = b.aiTurn()
				}
			}

			event := b.screen.PollEvent()

			// Check if we received a key event.
			ev, isEventKey := event.(*tcell.EventKey)
			if !isEventKey {
				continue
			}

			keyType := ev.Key()

			// Allow the user to quit the game at any time.
			if keyType == tcell.KeyRune {
				if ev.Rune() == rune('q') {
					close(quit)
					return
				}
			}

			// Only the blue player can control the piece.
			if !boardState.GameOver && boardState.LastMove.Player == game.Players.Blue {
				b.screen.Beep()
				continue
			}

			switch keyType {
			case tcell.KeyRune:
				switch ev.Rune() {
				case rune('t'):
					b.trainGame()

				case rune('g'):
					b.gitUpdate()

				case rune('n'):
					boardState = b.newGame()

				case rune('s'):
					b.turnSoundOnOff()

				case rune(' '):
					if !boardState.GameOver {
						boardState = b.userTurn()
					}
				}

			case tcell.KeyLeft:
				if !boardState.GameOver {
					b.movePlayerPiece(boardState, dirLeft)
				}

			case tcell.KeyRight:
				if !boardState.GameOver {
					b.movePlayerPiece(boardState, dirRight)
				}

			case tcell.KeyEnter, tcell.KeyDown:
				if !boardState.GameOver {
					boardState = b.userTurn()
				}
			}
		}
	}()

	return quit
}
