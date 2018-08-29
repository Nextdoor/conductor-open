package core

import (
	"time"

	"github.com/Nextdoor/conductor/services/code"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/services/messaging"
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/services/ticket"
	"github.com/Nextdoor/conductor/shared/datadog"
	"github.com/Nextdoor/conductor/shared/flags"
)

const SyncTicketsInterval = time.Second * 10
const CheckJobsInterval = time.Second * 5
const CheckTrainLockInterval = time.Second * 5

// How long to wait until starting background tasks after boot, in seconds.
// This is useful when upgrading Conductor, to avoid race conditions when two instances are polling at once.
var backgroundTaskStartDelay = flags.EnvInt("BACKGROUND_TASK_START_DELAY", 0)

func backgroundTaskLoop() {
	datadog.Info("Waiting %d seconds to start background jobs.", backgroundTaskStartDelay)
	time.Sleep(time.Second * time.Duration(backgroundTaskStartDelay))
	datadog.Info("Starting background jobs.")

	// This loop handles restarting the background task loop if it ever panics.
	killed := make(chan bool)
	for {
		go func() {
			dataClient := data.NewClient()
			codeService := code.GetService()
			messagingService := messaging.GetService()
			phaseService := phase.GetService()
			ticketService := ticket.GetService()

			syncTicketsTicker := time.NewTicker(SyncTicketsInterval)
			checkJobsTicker := time.NewTicker(CheckJobsInterval)
			checkTrainLockTicker := time.NewTicker(CheckTrainLockInterval)
			defer func() {
				err, stack := parsePanic(recover())
				if err != nil {
					datadog.Error("Panic in background task: %v. Stack trace: %v", err, stack)
				}
				killed <- true
			}()

			for {
				select {
				case <-syncTicketsTicker.C:
					syncTickets(dataClient, codeService, messagingService, phaseService, ticketService)
				case <-checkJobsTicker.C:
					checkJobs(dataClient)
				case <-checkTrainLockTicker.C:
					checkTrainLock(dataClient, codeService, messagingService, phaseService, ticketService)
				}
			}
		}()
		<-killed
	}
}
