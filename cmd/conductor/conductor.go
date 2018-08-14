package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/Nextdoor/conductor/core"
	"github.com/Nextdoor/conductor/shared/datadog"
)

func main() {
	// datadog.Initstatsd()

	flag.Parse()

	core.Preload()

	endpoints := core.Endpoints()
	server := core.NewServer(endpoints)

	address := ":8400"
	datadog.Info("Starting the Conductor server on %s ...", address)
	if e := http.ListenAndServe(address, server); e != nil {
		err := fmt.Errorf("Failed to start server on %s: %v", address, e)
		datadog.Error("Shutting down: %v", err)
		os.Exit(1)
	}
}
