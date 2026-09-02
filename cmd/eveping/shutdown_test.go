package main

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Sut103/discord-evenotify/internal/scheduler"
)

func TestWaitAndStop_SignalStopsScheduler(t *testing.T) {
	sched := scheduler.New(time.Hour, func() {})
	sigCh := make(chan os.Signal, 1)
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	forceExitCalled := make(chan struct{}, 1)

	done := make(chan struct{})
	go func() {
		sched.Start()
		close(done)
	}()

	go waitAndStop(sched, sigCh, logger, func() { forceExitCalled <- struct{}{} })
	sigCh <- os.Interrupt

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop after signal")
	}

	select {
	case <-forceExitCalled:
		t.Fatal("forceExit was called after a single signal")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestWaitAndStop_SecondSignal_ForcesExit(t *testing.T) {
	sched := scheduler.New(time.Hour, func() {})
	sigCh := make(chan os.Signal, 2)
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	forceExitCalled := make(chan struct{}, 1)

	go waitAndStop(sched, sigCh, logger, func() { forceExitCalled <- struct{}{} })
	sigCh <- os.Interrupt
	sigCh <- os.Interrupt

	select {
	case <-forceExitCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("forceExit was not called after a second signal")
	}
}
