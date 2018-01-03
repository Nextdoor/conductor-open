package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Nextdoor/conductor/shared/flags"
)

var (
	debugLoggingEnabled  = flags.EnvBool("DEBUG_LOGGING_ENABLED", false)
	useStructuredLogging = flags.EnvBool("USE_STRUCTURED_LOGGING", false)
)

type logMessage struct {
	Service string `json:"service"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

var output io.Writer = os.Stdout

func log(level string, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if useStructuredLogging {
		structuredMessage, err := json.Marshal(logMessage{
			Service: "conductor",
			Level:   level,
			Message: message,
		})
		if err != nil {
			fmt.Fprintf(output, "Error encoding structured log message: %v\n", err)
		} else {
			fmt.Fprintln(output, string(structuredMessage))
		}
	} else {
		fmt.Fprintf(output, "%s: %s\n", level, message)
	}
}

func Debug(format string, args ...interface{}) {
	if debugLoggingEnabled {
		log("DEBUG", format, args...)
	}
}

func Info(format string, args ...interface{}) {
	log("INFO", format, args...)
}

func Error(format string, args ...interface{}) {
	log("ERROR", format, args...)
}
