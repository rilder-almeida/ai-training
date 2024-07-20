package board

import (
	"github.com/gdamore/tcell/v2"
)

// pollEvents starts a goroutine to handle terminal events.
func (b *Board) pollEvents() chan struct{} {
	quit := make(chan struct{})

	go func() {
		for {
			event := b.screen.PollEvent()

			// Check if we received a key event.
			ev, isEventKey := event.(*tcell.EventKey)
			if !isEventKey {
				continue
			}

			// Check if the escape key was selected.
			keyType := ev.Key()
			if keyType == tcell.KeyEscape {
				if b.modalUp {
					b.closeModal()
					continue
				}
			}

			// Check if we are asked to quit.
			if keyType == tcell.KeyRune {
				switch ev.Rune() {
				case rune('q'):
					close(quit)
					return

				case rune('n'):
					b.newGame()

				case rune(' '):
					b.dropPiece(true)
					b.saveBoard()
				}
			}

			// Process the specified keys.
			switch keyType {
			case tcell.KeyLeft, tcell.KeyRight, tcell.KeyEnter:
				switch keyType {
				case tcell.KeyLeft:
					b.movePlayerPiece(dirLeft)

				case tcell.KeyRight:
					b.movePlayerPiece(dirRight)

				case tcell.KeyEnter, tcell.KeyDown:
					b.dropPiece(true)
					b.saveBoard()
				}
			}
		}
	}()

	return quit
}
