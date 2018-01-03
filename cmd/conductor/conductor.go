package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/Nextdoor/conductor/core"
	"github.com/Nextdoor/conductor/shared/logger"
)

func main() {
	flag.Parse()

	core.Preload()

	endpoints := core.Endpoints()
	server := core.NewServer(endpoints)

	address := ":8400"
	logger.Info("Starting the Conductor server on %s ...", address)
	if e := http.ListenAndServe(address, server); e != nil {
		err := fmt.Errorf("Failed to start server on %s: %v", address, e)
		logger.Error("Shutting down: %v", err)
		os.Exit(1)
	}
}
