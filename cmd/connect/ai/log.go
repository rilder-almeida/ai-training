package ai

import (
	"fmt"
	"os"
)

func (ai *AI) writeLog(v string) {
	if !ai.debug {
		return
	}

	f, _ := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()

	f.WriteString(v + "\n")
}

func (ai *AI) writeLogf(format string, v ...any) {
	if !ai.debug {
		return
	}

	f, _ := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()

	fmt.Fprintf(f, format+"\n", v...)
}

func writeChangeLog(boardID string) {
	f, _ := os.OpenFile(changeLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()

	f.WriteString(boardID + "\n")
}

// DeleteChangeLog deletes the change log file.
func (*AI) DeleteChangeLog() error {
	f, err := os.Open(changeLogFile)
	if err != nil {
		return nil
	}
	f.Close()

	return os.Remove(changeLogFile)
}
