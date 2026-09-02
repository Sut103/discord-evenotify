package main

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/Sut103/discord-evenotify/internal/batch"
)

func TestLoadToken_MissingEnv_ReturnsError(t *testing.T) {
	getenv := func(key string) string { return "" }

	_, err := loadToken(getenv)

	if err == nil {
		t.Fatal("expected an error when the token env var is unset, got nil")
	}
}

func TestLoadToken_PresentEnv_ReturnsToken(t *testing.T) {
	getenv := func(key string) string {
		if key == tokenEnvVar {
			return "secret-token"
		}
		return ""
	}

	token, err := loadToken(getenv)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "secret-token" {
		t.Fatalf("token = %q, want %q", token, "secret-token")
	}
}

func TestDryRunEnabled_Unset_ReturnsFalse(t *testing.T) {
	getenv := func(key string) string { return "" }

	if dryRunEnabled(getenv) {
		t.Fatal("expected dry-run to be disabled when the env var is unset")
	}
}

func TestDryRunEnabled_True_ReturnsTrue(t *testing.T) {
	for _, value := range []string{"true", "True", "TRUE", "1"} {
		getenv := func(key string) string {
			if key == dryRunEnvVar {
				return value
			}
			return ""
		}

		if !dryRunEnabled(getenv) {
			t.Fatalf("dryRunEnabled(%q) = false, want true", value)
		}
	}
}

func TestDryRunEnabled_False_ReturnsFalse(t *testing.T) {
	for _, value := range []string{"false", "0", "no", "yes"} {
		getenv := func(key string) string {
			if key == dryRunEnvVar {
				return value
			}
			return ""
		}

		if dryRunEnabled(getenv) {
			t.Fatalf("dryRunEnabled(%q) = true, want false", value)
		}
	}
}

func TestLogBatchResult_IncludesCounts(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	result := batch.BatchResult{TargetEvents: 3, SentSuccess: 5, SentFailure: 2}

	logBatchResult(logger, time.Now(), result)

	out := buf.String()
	for _, want := range []string{"target_events=3", "sent_success=5", "sent_failure=2"} {
		if !strings.Contains(out, want) {
			t.Fatalf("log output %q does not contain %q", out, want)
		}
	}
}

func TestLogBatchResult_IncludesErrorDetails(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	result := batch.BatchResult{
		Errors: []error{
			errors.New("fetch scheduled events for guild g1: api down"),
			errors.New("send reminder DM to user u1: 50007"),
		},
	}

	logBatchResult(logger, time.Now(), result)

	out := buf.String()
	for _, want := range []string{
		"fetch scheduled events for guild g1: api down",
		"send reminder DM to user u1: 50007",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("log output %q does not contain error detail %q", out, want)
		}
	}
}
