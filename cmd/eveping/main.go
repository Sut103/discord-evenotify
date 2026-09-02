// Command eveping is a stateless Discord bot that sends a DM reminder to
// every user interested in a guild's scheduled event, one day before it
// starts. It runs as a resident process: once per day it scans every guild
// the bot has joined, so there is no database or cache to keep in sync.
package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/Sut103/discord-evenotify/internal/batch"
	"github.com/Sut103/discord-evenotify/internal/discordclient"
	"github.com/Sut103/discord-evenotify/internal/scheduler"
)

const (
	tokenEnvVar   = "EVEPING_DISCORD_TOKEN"
	dryRunEnvVar  = "EVEPING_DRY_RUN"
	batchInterval = 24 * time.Hour
)

var errMissingToken = errors.New(tokenEnvVar + " environment variable is not set")

func loadToken(getenv func(string) string) (string, error) {
	token := getenv(tokenEnvVar)
	if token == "" {
		return "", errMissingToken
	}
	return token, nil
}

// dryRunEnabled reports whether EVEPING_DRY_RUN requests dry-run mode, so a
// manual smoke-test startup can confirm connectivity and target
// events/users without sending duplicate reminder DMs to real users.
func dryRunEnabled(getenv func(string) string) bool {
	value := getenv(dryRunEnvVar)
	return value == "1" || strings.EqualFold(value, "true")
}

func logBatchResult(logger *log.Logger, start time.Time, result batch.BatchResult) {
	logger.Printf(
		"daily batch finished: duration=%s target_events=%d sent_success=%d sent_failure=%d error_count=%d",
		time.Since(start).Round(time.Millisecond),
		result.TargetEvents,
		result.SentSuccess,
		result.SentFailure,
		len(result.Errors),
	)
	for _, err := range result.Errors {
		logger.Printf("daily batch error: %v", err)
	}
}

func runDailyBatch(client discordclient.Client, logger *log.Logger) {
	logger.Println("daily batch starting")
	start := time.Now()
	result := batch.RunDailyBatch(client, time.Now())
	logBatchResult(logger, start, result)
}

func main() {
	logger := log.Default()

	token, err := loadToken(os.Getenv)
	if err != nil {
		logger.Fatal(err)
	}

	// discordgo.New sets Identify.Intents to IntentsAllWithoutPrivileged by
	// default, which already includes IntentGuildScheduledEvents — no
	// explicit intent configuration is needed here.
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Fatalf("create discord session: %v", err)
	}

	if err := session.Open(); err != nil {
		logger.Fatalf("open discord session: %v", err)
	}
	defer session.Close()

	client := discordclient.New(session)
	if dryRunEnabled(os.Getenv) {
		logger.Printf("%s is set: dry-run mode enabled, reminder DMs will be logged but not sent", dryRunEnvVar)
		client = discordclient.NewDryRun(client, logger)
	}

	sched := scheduler.New(batchInterval, func() {
		runDailyBatch(client, logger)
	})

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go waitAndStop(sched, sigCh, logger, func() { os.Exit(1) })

	sched.Start()
}
