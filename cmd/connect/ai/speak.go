package ai

import (
	"os"

	htgotts "github.com/hegedustibor/htgo-tts"
	handlers "github.com/hegedustibor/htgo-tts/handlers"
	voices "github.com/hegedustibor/htgo-tts/voices"
)

// TurnSoundOnOff turns the sound for speaking on or off.
func (ai *AI) TurnSoundOnOff() bool {
	ai.sound = !ai.sound
	return ai.sound
}

// Speak will use the Mplayer to speak the specified message.
func (ai *AI) Speak(msg string) {
	if !ai.sound {
		return
	}

	go func() {
		speech := htgotts.Speech{Folder: "audio", Language: voices.English, Handler: &handlers.MPlayer{}}

		os.Remove("audio/speech.mp3")

		fileName, err := speech.CreateSpeechFile(msg, "speech")
		if err != nil {
			ai.writeLogf("create speech file: %s", err)
			return
		}

		defer os.Remove("audio/speech.mp3")

		if err := speech.PlaySpeechFile(fileName); err != nil {
			ai.writeLogf("play speech file: %s", err)
			return
		}
	}()
}
