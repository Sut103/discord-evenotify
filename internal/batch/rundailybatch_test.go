package batch_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Sut103/discord-evenotify/internal/batch"
	"github.com/Sut103/discord-evenotify/internal/discordclient"
)

func TestRunDailyBatch_AllCombinations_DMsAttempted(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)

	guilds := []discordclient.Guild{{ID: "guild-1"}, {ID: "guild-2"}}
	eventsByGuild := map[string][]discordclient.Event{
		"guild-1": {mkEvent("g1-e1", tomorrow, discordclient.EventStatusScheduled)},
		"guild-2": {mkEvent("g2-e1", tomorrow, discordclient.EventStatusScheduled)},
	}
	usersByEvent := map[string][]discordclient.User{
		"g1-e1": {{ID: "u1"}, {ID: "u2"}},
		"g2-e1": {{ID: "u3"}},
	}

	var mu sync.Mutex
	var dmsSent []string
	client := &discordclient.Fake{
		GuildsFunc: func() []discordclient.Guild { return guilds },
		ScheduledEventsFunc: func(guildID string) ([]discordclient.Event, error) {
			return eventsByGuild[guildID], nil
		},
		EventUsersFunc: func(guildID, eventID, after string, limit int) ([]discordclient.User, error) {
			if after != "" {
				return nil, nil
			}
			return usersByEvent[eventID], nil
		},
		SendDMFunc: func(userID, message string) error {
			mu.Lock()
			dmsSent = append(dmsSent, userID)
			mu.Unlock()
			return nil
		},
	}

	result := batch.RunDailyBatch(client, now)

	if result.TargetEvents != 2 {
		t.Fatalf("TargetEvents = %d, want 2", result.TargetEvents)
	}
	if result.SentSuccess != 3 {
		t.Fatalf("SentSuccess = %d, want 3", result.SentSuccess)
	}
	if result.SentFailure != 0 {
		t.Fatalf("SentFailure = %d, want 0", result.SentFailure)
	}
	if len(dmsSent) != 3 {
		t.Fatalf("dmsSent = %v, want 3 entries", dmsSent)
	}
}

func TestRunDailyBatch_OneGuildEventsError_OthersContinue(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)

	guilds := []discordclient.Guild{{ID: "guild-bad"}, {ID: "guild-good"}}
	client := &discordclient.Fake{
		GuildsFunc: func() []discordclient.Guild { return guilds },
		ScheduledEventsFunc: func(guildID string) ([]discordclient.Event, error) {
			if guildID == "guild-bad" {
				return nil, errors.New("api down")
			}
			return []discordclient.Event{mkEvent("good-e1", tomorrow, discordclient.EventStatusScheduled)}, nil
		},
		EventUsersFunc: func(guildID, eventID, after string, limit int) ([]discordclient.User, error) {
			if after != "" {
				return nil, nil
			}
			return []discordclient.User{{ID: "u1"}}, nil
		},
		SendDMFunc: func(userID, message string) error { return nil },
	}

	result := batch.RunDailyBatch(client, now)

	if result.TargetEvents != 1 {
		t.Fatalf("TargetEvents = %d, want 1", result.TargetEvents)
	}
	if result.SentSuccess != 1 {
		t.Fatalf("SentSuccess = %d, want 1", result.SentSuccess)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("Errors = %v, want 1 entry", result.Errors)
	}
}

func TestRunDailyBatch_OneUserSendFails_OthersContinue(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)

	guilds := []discordclient.Guild{{ID: "guild-1"}}
	events := []discordclient.Event{
		mkEvent("e1", tomorrow, discordclient.EventStatusScheduled),
		mkEvent("e2", tomorrow, discordclient.EventStatusScheduled),
	}
	usersByEvent := map[string][]discordclient.User{
		"e1": {{ID: "bad-user"}, {ID: "u2"}},
		"e2": {{ID: "u3"}},
	}

	var attempted []string
	client := &discordclient.Fake{
		GuildsFunc: func() []discordclient.Guild { return guilds },
		ScheduledEventsFunc: func(guildID string) ([]discordclient.Event, error) {
			return events, nil
		},
		EventUsersFunc: func(guildID, eventID, after string, limit int) ([]discordclient.User, error) {
			if after != "" {
				return nil, nil
			}
			return usersByEvent[eventID], nil
		},
		SendDMFunc: func(userID, message string) error {
			attempted = append(attempted, userID)
			if userID == "bad-user" {
				return errors.New("50007")
			}
			return nil
		},
	}

	result := batch.RunDailyBatch(client, now)

	if result.SentSuccess != 2 {
		t.Fatalf("SentSuccess = %d, want 2", result.SentSuccess)
	}
	if result.SentFailure != 1 {
		t.Fatalf("SentFailure = %d, want 1", result.SentFailure)
	}
	if len(attempted) != 3 {
		t.Fatalf("attempted = %v, want 3 entries (all users across both events)", attempted)
	}
}

func TestRunDailyBatch_ResultCounts_MatchFakeResponses(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	notTomorrow := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)

	guilds := []discordclient.Guild{{ID: "guild-1"}}
	events := []discordclient.Event{
		mkEvent("target", tomorrow, discordclient.EventStatusScheduled),
		mkEvent("not-target", notTomorrow, discordclient.EventStatusScheduled),
	}
	client := &discordclient.Fake{
		GuildsFunc: func() []discordclient.Guild { return guilds },
		ScheduledEventsFunc: func(guildID string) ([]discordclient.Event, error) {
			return events, nil
		},
		EventUsersFunc: func(guildID, eventID, after string, limit int) ([]discordclient.User, error) {
			if after != "" {
				return nil, nil
			}
			return []discordclient.User{{ID: "u1"}, {ID: "u2"}, {ID: "u3"}}, nil
		},
		SendDMFunc: func(userID, message string) error {
			if userID == "u2" {
				return errors.New("fail")
			}
			return nil
		},
	}

	result := batch.RunDailyBatch(client, now)

	if result.TargetEvents != 1 {
		t.Fatalf("TargetEvents = %d, want 1 (only tomorrow's event counted)", result.TargetEvents)
	}
	if result.SentSuccess != 2 {
		t.Fatalf("SentSuccess = %d, want 2", result.SentSuccess)
	}
	if result.SentFailure != 1 {
		t.Fatalf("SentFailure = %d, want 1", result.SentFailure)
	}
}
