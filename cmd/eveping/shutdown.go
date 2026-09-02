package main

import (
	"log"
	"os"

	"github.com/Sut103/discord-evenotify/internal/scheduler"
)

// waitAndStop blocks until a shutdown signal arrives on sigCh, then stops
// sched so Start() returns and main can run its deferred cleanup (closing
// the Discord session). A running batch (blocked on network calls) doesn't
// interrupt when sched.Stop() is called, so a second signal calls forceExit
// to give an operator a way to end the process immediately rather than
// waiting out the orchestrator's SIGKILL grace period. sigCh and forceExit
// are injected so this is testable without sending a real OS signal or
// actually exiting the process.
func waitAndStop(sched *scheduler.Scheduler, sigCh <-chan os.Signal, logger *log.Logger, forceExit func()) {
	sig := <-sigCh
	logger.Printf("received signal %s, shutting down", sig)
	sched.Stop()

	sig = <-sigCh
	logger.Printf("received second signal %s, forcing exit", sig)
	forceExit()
}
