// Package core contains the core logic of conductor.
// It is critical that the app itself is entirely stateless.
// All state is stored in the services themselves.
// The services may store their data externally depending on implementation.
package core

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/Nextdoor/conductor/services/auth"
	"github.com/Nextdoor/conductor/services/code"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/services/messaging"
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/services/ticket"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/types"
)

// Preload attempts to load all services and start background tasks.
// Intended to be run once, at boot time.
// Blocking and parallel.
func Preload() {
	logger.Info("Preloading services in the background.")

	var waitGroup sync.WaitGroup
	waitGroup.Add(6) // One for each service

	funcs := []func(){
		func() {
			auth.GetService()
		},
		func() {
			code.GetService()
		},
		func() {
			data.NewClient()
		},
		func() {
			messaging.GetService()
		},
		func() {
			phase.GetService()
		},
		func() {
			ticket.GetService()
		},
	}
	for i := range funcs {
		f := funcs[i]
		go func() {
			defer waitGroup.Done()
			f()
		}()
	}

	waitGroup.Wait()

	go backgroundTaskLoop()
}

func healthz(_ *http.Request) response {
	return emptyResponse()
}

func coreEndpoints() []endpoint {
	return []endpoint{
		newOpenEp("/healthz", get, healthz),
		newEp("/api/config", get, fetchConfig),
		newEp("/api/mode", get, fetchMode),
		newAdminEp("/api/mode", post, setMode),
		newEp("/api/options", get, fetchOptions),
		newAdminEp("/api/options", post, setOptions),
	}
}

func fetchConfig(_ *http.Request) response {
	dataClient := data.NewClient()
	config, err := dataClient.Config()
	if err != nil {
		return errorResponse(err.Error(), http.StatusInternalServerError)
	}
	return dataResponse(config)
}

func fetchMode(_ *http.Request) response {
	dataClient := data.NewClient()
	mode, err := dataClient.Mode()
	if err != nil {
		return errorResponse(err.Error(), http.StatusInternalServerError)
	}
	return dataResponse(mode.String())
}

func setMode(r *http.Request) response {
	err := r.ParseForm()
	if err != nil {
		return errorResponse("Error parsing POST form", http.StatusBadRequest)
	}
	mode, err := types.ModeFromString(r.PostFormValue("mode"))
	if err != nil {
		return errorResponse(err.Error(), http.StatusBadRequest)
	}

	dataClient := data.NewClient()
	err = dataClient.SetMode(mode)
	if err != nil {
		return errorResponse(
			fmt.Sprintf(
				"Error setting mode: %v",
				err),
			http.StatusInternalServerError)
	}

	return dataResponse(mode.String())
}

func fetchOptions(_ *http.Request) response {
	dataClient := data.NewClient()
	options, err := dataClient.Options()
	if err != nil {
		return errorResponse(err.Error(), http.StatusInternalServerError)
	}
	return dataResponse(options)
}

func setOptions(r *http.Request) response {
	err := r.ParseForm()
	if err != nil {
		return errorResponse("Error parsing POST form", http.StatusBadRequest)
	}

	options := &types.Options{}
	err = options.FromString(r.PostFormValue("options"))
	if err != nil {
		return errorResponse(err.Error(), http.StatusBadRequest)
	}

	dataClient := data.NewClient()
	err = dataClient.SetOptions(options)
	if err != nil {
		return errorResponse(
			fmt.Sprintf(
				"Error setting options: %v",
				err),
			http.StatusInternalServerError)
	}

	return dataResponse(options)
}
