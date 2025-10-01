package logging

import (
	"io"
	"log"
	"os"
)

// Logger is a thin wrapper around the standard logger, allowing dependency injection.
type Logger struct {
	*log.Logger
}

// New creates a new Logger writing to out with a sane default prefix and flags.
func New(out io.Writer) *Logger {
	if out == nil {
		out = os.Stdout
	}
	l := log.New(out, "imgapi ", log.LstdFlags|log.Lshortfile)
	return &Logger{Logger: l}
}
