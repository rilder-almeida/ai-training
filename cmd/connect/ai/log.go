package ai

import (
	"fmt"
	"os"
)

func writeLog(v string) {
	f, _ := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()

	f.WriteString(v + "\n")
}

func writeLogf(format string, v ...any) {
	f, _ := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()

	fmt.Fprintf(f, format+"\n", v...)
}

func writeChangeLog(boardID string) {
	f, _ := os.OpenFile(changeLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()

	f.WriteString(boardID + "\n")
}

// DeleteChangeLog deletes the change log file.
func (ai *AI) DeleteChangeLog() error {
	return os.Remove(changeLogFile)
}
