package discordclient_test

import (
	"errors"
	"testing"

	"github.com/Sut103/discord-evenotify/internal/discordclient"
)

func TestFake_SatisfiesClientInterface(t *testing.T) {
	var _ discordclient.Client = &discordclient.Fake{}
}

func TestFake_Guilds_ReturnsConfiguredValue(t *testing.T) {
	want := []discordclient.Guild{{ID: "guild-1"}, {ID: "guild-2"}}
	f := &discordclient.Fake{
		GuildsFunc: func() []discordclient.Guild { return want },
	}

	got := f.Guilds()

	if len(got) != len(want) || got[0].ID != want[0].ID || got[1].ID != want[1].ID {
		t.Fatalf("Guilds() = %+v, want %+v", got, want)
	}
}

func TestFake_ScheduledEvents_ReturnsConfiguredValueAndError(t *testing.T) {
	wantErr := errors.New("boom")
	f := &discordclient.Fake{
		ScheduledEventsFunc: func(guildID string) ([]discordclient.Event, error) {
			if guildID != "guild-1" {
				t.Fatalf("guildID = %q, want %q", guildID, "guild-1")
			}
			return nil, wantErr
		},
	}

	_, err := f.ScheduledEvents("guild-1")

	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestFake_EventUsers_ReturnsConfiguredValue(t *testing.T) {
	want := []discordclient.User{{ID: "user-1"}}
	f := &discordclient.Fake{
		EventUsersFunc: func(guildID, eventID, after string, limit int) ([]discordclient.User, error) {
			if guildID != "g" || eventID != "e" || after != "a" || limit != 100 {
				t.Fatalf("unexpected args: %q %q %q %d", guildID, eventID, after, limit)
			}
			return want, nil
		},
	}

	got, err := f.EventUsers("g", "e", "a", 100)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != want[0].ID {
		t.Fatalf("EventUsers() = %+v, want %+v", got, want)
	}
}

func TestFake_SendDM_ReturnsConfiguredError(t *testing.T) {
	wantErr := errors.New("dm failed")
	f := &discordclient.Fake{
		SendDMFunc: func(userID, message string) error {
			if userID != "user-1" || message != "hello" {
				t.Fatalf("unexpected args: %q %q", userID, message)
			}
			return wantErr
		},
	}

	err := f.SendDM("user-1", "hello")

	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestFake_ZeroValue_ReturnsSafeDefaults(t *testing.T) {
	f := &discordclient.Fake{}

	if got := f.Guilds(); got != nil {
		t.Fatalf("Guilds() = %+v, want nil", got)
	}
	if _, err := f.ScheduledEvents("g"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := f.EventUsers("g", "e", "", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := f.SendDM("u", "m"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
