package batch_test

import (
	"errors"
	"testing"

	"github.com/Sut103/discord-evenotify/internal/batch"
	"github.com/Sut103/discord-evenotify/internal/discordclient"
)

func TestSendReminderDM_Success_ReturnsNil(t *testing.T) {
	client := &discordclient.Fake{
		SendDMFunc: func(userID, message string) error { return nil },
	}

	err := batch.SendReminderDM(client, "user-1", "hello")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendReminderDM_Failure_ReturnsError(t *testing.T) {
	wantErr := errors.New("50007: Cannot send messages to this user")
	client := &discordclient.Fake{
		SendDMFunc: func(userID, message string) error { return wantErr },
	}

	err := batch.SendReminderDM(client, "user-1", "hello")

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, does not wrap %v", err, wantErr)
	}
}

func TestSendReminderDM_Failure_DoesNotPanic(t *testing.T) {
	client := &discordclient.Fake{
		SendDMFunc: func(userID, message string) error { return errors.New("boom") },
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SendReminderDM panicked: %v", r)
		}
	}()

	_ = batch.SendReminderDM(client, "user-1", "hello")
}
