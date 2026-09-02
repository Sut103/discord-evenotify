package batch

import (
	"fmt"

	"github.com/Sut103/discord-evenotify/internal/discordclient"
)

// SendReminderDM sends a single reminder DM and returns any failure to the
// caller (e.g. a user with DMs disabled) instead of panicking, so a batch
// loop can record it and continue with the next user.
func SendReminderDM(client discordclient.Client, userID, message string) error {
	if err := client.SendDM(userID, message); err != nil {
		return fmt.Errorf("send reminder DM to user %s: %w", userID, err)
	}
	return nil
}
