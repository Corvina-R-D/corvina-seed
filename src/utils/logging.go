package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func setLoggingMode(mode string) {
	if mode == "" {
		mode = "pretty"
	}

	switch mode {
	case "pretty":
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:           os.Stderr,
			TimeFormat:    "2006-01-02 03:04:05PM",
			FormatMessage: formatMessage})
	case "json":
		log.Logger = log.Output(os.Stderr)
		zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999Z07:00"
	}
}

func formatMessage(i interface{}) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%s |", i)
}

func setLoggingLevel(level string) {
	if level == "" {
		level = "WARN"
	}

	level = strings.ToUpper(level)

	switch level {
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "ALL":
	case "TRACE":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
}

func VerboseLog() {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
}

func InitLog() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	setLoggingLevel(os.Getenv("LOG_LEVEL"))
	setLoggingMode(os.Getenv("LOG_MODE"))
}
