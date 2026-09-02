package batch_test

import (
	"testing"
	"time"

	"github.com/Sut103/discord-evenotify/internal/batch"
	"github.com/Sut103/discord-evenotify/internal/discordclient"
)

func TestFilterTargetEvents_OnlyTomorrowUTC(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)

	yesterday := mkEvent("yesterday", time.Date(2026, 8, 17, 5, 0, 0, 0, time.UTC), discordclient.EventStatusScheduled)
	today := mkEvent("today", time.Date(2026, 8, 18, 23, 0, 0, 0, time.UTC), discordclient.EventStatusScheduled)
	tomorrow := mkEvent("tomorrow", time.Date(2026, 8, 19, 0, 30, 0, 0, time.UTC), discordclient.EventStatusScheduled)
	dayAfterTomorrow := mkEvent("day-after-tomorrow", time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC), discordclient.EventStatusScheduled)

	got := batch.FilterTargetEvents([]discordclient.Event{yesterday, today, tomorrow, dayAfterTomorrow}, now)

	assertEventIDs(t, got, []string{"tomorrow"})
}

func TestFilterTargetEvents_ExcludesCanceledAndCompleted(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	tomorrowStart := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)

	canceled := mkEvent("canceled", tomorrowStart, discordclient.EventStatusCanceled)
	completed := mkEvent("completed", tomorrowStart, discordclient.EventStatusCompleted)

	got := batch.FilterTargetEvents([]discordclient.Event{canceled, completed}, now)

	assertEventIDs(t, got, nil)
}

func TestFilterTargetEvents_IncludesScheduledAndActive(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	tomorrowStart := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)

	scheduled := mkEvent("scheduled", tomorrowStart, discordclient.EventStatusScheduled)
	active := mkEvent("active", tomorrowStart, discordclient.EventStatusActive)

	got := batch.FilterTargetEvents([]discordclient.Event{scheduled, active}, now)

	assertEventIDs(t, got, []string{"scheduled", "active"})
}

func TestFilterTargetEvents_EmptyInput_ReturnsEmpty(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)

	got := batch.FilterTargetEvents([]discordclient.Event{}, now)

	assertEventIDs(t, got, nil)
}

func mkEvent(id string, start time.Time, status string) discordclient.Event {
	return discordclient.Event{
		ID:                 id,
		GuildID:            "guild-1",
		Name:               "event-" + id,
		ScheduledStartTime: start,
		Status:             status,
	}
}

func assertEventIDs(t *testing.T, got []discordclient.Event, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d events (%+v), want %d (%v)", len(got), got, len(want), want)
	}
	for i, e := range got {
		if e.ID != want[i] {
			t.Fatalf("got[%d].ID = %q, want %q", i, e.ID, want[i])
		}
	}
}
