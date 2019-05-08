package datadog

import (
	"fmt"
	"os"

	"github.com/DataDog/datadog-go/statsd"

	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/logger"
)

var enableDatadog = flags.EnvBool("ENABLE_DATADOG", true)

func newStatsdClient() *statsd.Client {
	if !enableDatadog {
		return nil
	}
	// All metrics with be prefixed with "conductor."
	c, err := statsd.New(fmt.Sprintf("%v:%v", os.Getenv("STATSD_HOST"), 8125),
		statsd.WithNamespace("conductor."))
	if err != nil {
		panic(fmt.Sprintf("Could not create statsd client: %v", err))
	}
	return c
}

var c = newStatsdClient()

func Client() *statsd.Client {
	return c
}

func Incr(name string, tags []string) {
	if !enableDatadog {
		return
	}
	err := c.Incr(name, tags, 1)
	if err != nil {
		logger.Error("Error sending %s metric: %v", name, err)
	}
}

func Count(name string, count int, tags []string) {
	if !enableDatadog {
		return
	}
	err := c.Count(name, int64(count), tags, 1)
	if err != nil {
		logger.Error("Error sending %s metric: %v", name, err)
	}
}

func Gauge(name string, value float64, tags []string) {
	if !enableDatadog {
		return
	}
	err := c.Gauge(name, value, tags, 1)
	if err != nil {
		logger.Error("Error sending %s metric: %v", name, err)
	}
}

// log logs an event to stdout, and also sends it to datadog.
func log(alertType statsd.EventAlertType, format string, args ...interface{}) {
	if enableDatadog {
		e := statsd.NewEvent("conductor", fmt.Sprintf(format, args...))
		e.AlertType = alertType
		err := c.Event(e)
		if err != nil {
			logger.Error("Could not create datadog event: %v", err)
		}
	}
	switch alertType {
	case statsd.Info:
		logger.Info(format, args...)
	default:
		logger.Error(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	log(statsd.Info, format, args...)
}

func Error(format string, args ...interface{}) {
	log(statsd.Error, format, args...)
}
