package discordclient_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/Sut103/discord-evenotify/internal/discordclient"
)

func TestDryRunClient_SendDM_DoesNotCallUnderlyingSendDM(t *testing.T) {
	called := false
	fake := &discordclient.Fake{
		SendDMFunc: func(userID, message string) error {
			called = true
			return nil
		},
	}
	client := discordclient.NewDryRun(fake, log.New(&bytes.Buffer{}, "", 0))

	if err := client.SendDM("u1", "hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("underlying SendDM was called in dry-run mode")
	}
}

func TestDryRunClient_SendDM_LogsUserAndMessage(t *testing.T) {
	fake := &discordclient.Fake{}
	var buf bytes.Buffer
	client := discordclient.NewDryRun(fake, log.New(&buf, "", 0))

	if err := client.SendDM("u1", "hello world"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "u1") || !strings.Contains(out, "hello world") {
		t.Fatalf("log output %q does not contain user id and message", out)
	}
}

func TestDryRunClient_DelegatesReadMethods(t *testing.T) {
	guilds := []discordclient.Guild{{ID: "g1"}}
	events := []discordclient.Event{{ID: "e1"}}
	users := []discordclient.User{{ID: "u1"}}
	fake := &discordclient.Fake{
		GuildsFunc: func() []discordclient.Guild { return guilds },
		ScheduledEventsFunc: func(guildID string) ([]discordclient.Event, error) {
			return events, nil
		},
		EventUsersFunc: func(guildID, eventID, after string, limit int) ([]discordclient.User, error) {
			return users, nil
		},
	}
	client := discordclient.NewDryRun(fake, log.New(&bytes.Buffer{}, "", 0))

	if got := client.Guilds(); len(got) != 1 || got[0].ID != "g1" {
		t.Fatalf("Guilds() = %v, want delegated result %v", got, guilds)
	}
	gotEvents, err := client.ScheduledEvents("g1")
	if err != nil || len(gotEvents) != 1 || gotEvents[0].ID != "e1" {
		t.Fatalf("ScheduledEvents() = %v, %v, want delegated result %v", gotEvents, err, events)
	}
	gotUsers, err := client.EventUsers("g1", "e1", "", 100)
	if err != nil || len(gotUsers) != 1 || gotUsers[0].ID != "u1" {
		t.Fatalf("EventUsers() = %v, %v, want delegated result %v", gotUsers, err, users)
	}
}
