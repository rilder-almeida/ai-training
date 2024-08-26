package game

import (
	"os"

	htgotts "github.com/hegedustibor/htgo-tts"
	handlers "github.com/hegedustibor/htgo-tts/handlers"
	voices "github.com/hegedustibor/htgo-tts/voices"
)

func (b *Board) speak(msg string) {
	go func() {
		speech := htgotts.Speech{Folder: "audio", Language: voices.English, Handler: &handlers.MPlayer{}}

		os.Remove("audio/speech.mp3")

		fileName, err := speech.CreateSpeechFile(msg, "speech")
		if err != nil {
			b.debugMessage = err.Error()
			return
		}

		if err := speech.PlaySpeechFile(fileName); err != nil {
			b.debugMessage = err.Error()
			return
		}

		os.Remove("audio/speech.mp3")
	}()
}
