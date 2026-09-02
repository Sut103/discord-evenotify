// Package reminder formats the DM text sent to users interested in a
// scheduled event.
package reminder

import (
	"fmt"

	"github.com/Sut103/discord-evenotify/internal/discordclient"
)

// FormatReminder builds the DM body for an upcoming event: its name, start
// time (UTC, human-readable), and a link to the event on Discord.
func FormatReminder(event discordclient.Event) string {
	return fmt.Sprintf(
		"「%s」が明日開始します！\n開始日時: %s\nイベントページ: %s",
		event.Name,
		event.ScheduledStartTime.UTC().Format("2006-01-02 15:04 UTC"),
		eventURL(event),
	)
}

func eventURL(event discordclient.Event) string {
	return fmt.Sprintf("https://discord.com/events/%s/%s", event.GuildID, event.ID)
}
