package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func New() zerolog.Logger {
	// Set global time format to RFC3339
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	return logger.Level(zerolog.DebugLevel) // TODO: change to InfoLevel in production
}
