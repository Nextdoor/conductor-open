package datadog

import (
	"fmt"
	"os"

	"github.com/DataDog/datadog-go/statsd"

	"github.com/Nextdoor/conductor/shared/logger"
)

var (
	C *statsd.Client
	Err error
)

func Initstatsd() {
	C, Err = statsd.New(fmt.Sprintf("%v:%v", os.Getenv("STATSD_HOST"), 8125))
	if Err != nil {
		// failed to get statsd client
		logger.Error("Could not set up statsd: %v", Err)
		return
	}
	logger.Info("Set up statsd: %v!", C)
	C.Namespace = "Conductor."
}

func log(alertType statsd.EventAlertType, format string, args ...interface{}) {
	// Send event to statsd and log it too!
	if C != nil {
		e := statsd.NewEvent("Conductor", fmt.Sprintf(format, args...))
		e.AlertType = alertType
		Err = C.Event(e)
		if Err != nil {
			logger.Error("Could not create datadog event: %v", Err)
		}
	}
	switch alertType {
	case statsd.Info:
		logger.Info(format, args)
	default:
		logger.Error(format, args)
	}
}

func Info(format string, args ...interface{}) {
	log(statsd.Info, format, args)
}

func Error(format string, args ...interface{}) {
	log(statsd.Error, format, args)
}
