package batch_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Sut103/discord-evenotify/internal/batch"
	"github.com/Sut103/discord-evenotify/internal/discordclient"
)

func TestFetchAllInterestedUsers_NoUsers_ReturnsEmpty(t *testing.T) {
	client := &discordclient.Fake{
		EventUsersFunc: func(guildID, eventID, after string, limit int) ([]discordclient.User, error) {
			return nil, nil
		},
	}

	got, err := batch.FetchAllInterestedUsers(client, "guild-1", "event-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d users, want 0", len(got))
	}
}

func TestFetchAllInterestedUsers_SinglePage_OneCall(t *testing.T) {
	calls := 0
	users := makeUsers(1, 50)
	client := &discordclient.Fake{
		EventUsersFunc: func(guildID, eventID, after string, limit int) ([]discordclient.User, error) {
			calls++
			if limit != 100 {
				t.Fatalf("limit = %d, want 100", limit)
			}
			if after != "" {
				t.Fatalf("after = %q, want empty on first call", after)
			}
			return users, nil
		},
	}

	got, err := batch.FetchAllInterestedUsers(client, "guild-1", "event-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if len(got) != 50 {
		t.Fatalf("got %d users, want 50", len(got))
	}
}

func TestFetchAllInterestedUsers_MultiplePages_AllReturned(t *testing.T) {
	pages := [][]discordclient.User{
		makeUsers(1, 100),
		makeUsers(101, 100),
		makeUsers(201, 30),
	}
	var seenAfter []string
	callIndex := 0
	client := &discordclient.Fake{
		EventUsersFunc: func(guildID, eventID, after string, limit int) ([]discordclient.User, error) {
			seenAfter = append(seenAfter, after)
			page := pages[callIndex]
			callIndex++
			return page, nil
		},
	}

	got, err := batch.FetchAllInterestedUsers(client, "guild-1", "event-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 230 {
		t.Fatalf("got %d users, want 230", len(got))
	}
	seen := map[string]bool{}
	for _, u := range got {
		if seen[u.ID] {
			t.Fatalf("duplicate user id %q", u.ID)
		}
		seen[u.ID] = true
	}
	if seenAfter[0] != "" {
		t.Fatalf("first call after = %q, want empty", seenAfter[0])
	}
	if seenAfter[1] != pages[0][len(pages[0])-1].ID {
		t.Fatalf("second call after = %q, want last id of page 1", seenAfter[1])
	}
	if seenAfter[2] != pages[1][len(pages[1])-1].ID {
		t.Fatalf("third call after = %q, want last id of page 2", seenAfter[2])
	}
}

func TestFetchAllInterestedUsers_ErrorPropagates(t *testing.T) {
	wantErr := errors.New("api error")
	client := &discordclient.Fake{
		EventUsersFunc: func(guildID, eventID, after string, limit int) ([]discordclient.User, error) {
			return nil, wantErr
		},
	}

	_, err := batch.FetchAllInterestedUsers(client, "guild-1", "event-1")

	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func makeUsers(startID, count int) []discordclient.User {
	users := make([]discordclient.User, count)
	for i := 0; i < count; i++ {
		users[i] = discordclient.User{ID: idString(startID + i)}
	}
	return users
}

func idString(n int) string {
	return fmt.Sprintf("user-%d", n)
}
