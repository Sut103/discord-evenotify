package reminder_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Sut103/discord-evenotify/internal/discordclient"
	"github.com/Sut103/discord-evenotify/internal/reminder"
)

func TestFormatReminder_IncludesNameStartTimeAndURL(t *testing.T) {
	event := discordclient.Event{
		ID:                 "event-1",
		GuildID:            "guild-1",
		Name:               "Fleet Op",
		ScheduledStartTime: time.Date(2026, 8, 19, 20, 0, 0, 0, time.UTC),
		Status:             discordclient.EventStatusScheduled,
	}

	got := reminder.FormatReminder(event)

	if !strings.Contains(got, "Fleet Op") {
		t.Fatalf("message %q does not contain event name", got)
	}
	if !strings.Contains(got, "2026-08-19 20:00 UTC") {
		t.Fatalf("message %q does not contain human-readable start time", got)
	}
	wantURL := "https://discord.com/events/guild-1/event-1"
	if !strings.Contains(got, wantURL) {
		t.Fatalf("message %q does not contain event URL %q", got, wantURL)
	}
}
