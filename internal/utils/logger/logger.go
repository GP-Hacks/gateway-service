package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogger(IsProduction bool, vectorUrl string) {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	multi := zerolog.MultiLevelWriter(consoleWriter)
	if IsProduction {
		httpWriter := NewHTTPWriter(vectorUrl)
		multi = zerolog.MultiLevelWriter(httpWriter, consoleWriter)
	}

	log.Logger = zerolog.New(multi).
		With().
		Timestamp().
		Caller().
		Str("service", "gateway").
		Logger()
}
