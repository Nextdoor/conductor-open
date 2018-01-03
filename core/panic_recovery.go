package core

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Nextdoor/conductor/shared/logger"
)

// parsePanic should be called with recover() as the parameter.
// Needed because recover() doesn't work if not called directly by the deferred function.
// See https://golang.org/ref/spec#Handling_panics
func parsePanic(panicValue interface{}) (error, string) {
	var err error
	var stack string
	if panicValue != nil {
		switch v := panicValue.(type) {
		case string:
			err = errors.New(v)
		case error:
			err = v
		default:
			err = errors.New("Unknown error")
		}
		stack = string(debug.Stack())
	}
	return err, stack
}

func newPanicRecoveryMiddleware() panicRecoveryMiddleware {
	return panicRecoveryMiddleware{}
}

type panicRecoveryMiddleware struct{}

func (_ panicRecoveryMiddleware) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err, stack := parsePanic(recover())
			if err != nil {
				logger.Error("Panic in request: %v. Stack trace: %v", err, stack)
				errorResponse(
					fmt.Sprintf("Panic: %s. Stack trace: %v", err.Error(), stack),
					http.StatusInternalServerError).Write(w, r)
				return
			}
		}()
		handler.ServeHTTP(w, r)
	})
}
