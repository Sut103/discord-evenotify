package scheduler_test

import (
	"testing"
	"time"

	"github.com/Sut103/discord-evenotify/internal/scheduler"
)

func TestScheduler_Start_CallsRunImmediately(t *testing.T) {
	calls := make(chan struct{}, 10)
	// A long interval that would never fire during the test, so the only
	// way a call can arrive is the immediate on-start invocation.
	s := scheduler.New(time.Hour, func() { calls <- struct{}{} })

	go s.Start()
	defer s.Stop()

	select {
	case <-calls:
	case <-time.After(2 * time.Second):
		t.Fatal("Start() did not call run() immediately")
	}
}

func TestScheduler_CallsRunPeriodically(t *testing.T) {
	calls := make(chan struct{}, 10)
	s := scheduler.New(5*time.Millisecond, func() { calls <- struct{}{} })

	go s.Start()
	defer s.Stop()

	for i := 0; i < 3; i++ {
		select {
		case <-calls:
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for call %d", i+1)
		}
	}
}

func TestScheduler_Stop_StopsFurtherCalls(t *testing.T) {
	calls := make(chan struct{}, 10)
	s := scheduler.New(5*time.Millisecond, func() { calls <- struct{}{} })

	go s.Start()

	select {
	case <-calls:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first call")
	}

	s.Stop()

	// Drain any call already in flight when Stop was invoked, then confirm
	// no further calls arrive.
	select {
	case <-calls:
	default:
	}

	select {
	case <-calls:
		t.Fatal("received a call after Stop")
	case <-time.After(50 * time.Millisecond):
	}
}
