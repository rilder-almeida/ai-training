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

			keyType := ev.Key()

			// Process the specified keys.
			switch keyType {
			case tcell.KeyEscape:
				if b.modalUp {
					b.closeModal()
				}

			case tcell.KeyRune:
				switch ev.Rune() {
				case rune('q'):
					close(quit)
					return

				case rune('n'):
					b.newGame()

				case rune(' '):
					b.dropPiece(true)
					b.runAISupport()
				}

			case tcell.KeyLeft:
				b.movePlayerPiece(dirLeft)

			case tcell.KeyRight:
				b.movePlayerPiece(dirRight)

			case tcell.KeyEnter, tcell.KeyDown:
				b.dropPiece(true)
				b.runAISupport()
			}
		}
	}()

	return quit
}
