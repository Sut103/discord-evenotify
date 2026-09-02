package batch

import (
	"fmt"
	"time"

	"github.com/Sut103/discord-evenotify/internal/discordclient"
	"github.com/Sut103/discord-evenotify/internal/reminder"
)

// BatchResult summarizes the outcome of one RunDailyBatch execution.
type BatchResult struct {
	TargetEvents int
	SentSuccess  int
	SentFailure  int
	Errors       []error
}

// RunDailyBatch walks every guild the bot is in, finds each guild's events
// starting tomorrow (UTC), and sends a reminder DM to every interested
// user. A failure fetching one guild's events, or sending to one user,
// is recorded in the result and does not stop processing of the rest.
func RunDailyBatch(client discordclient.Client, now time.Time) BatchResult {
	var result BatchResult

	for _, guild := range client.Guilds() {
		events, err := client.ScheduledEvents(guild.ID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("fetch scheduled events for guild %s: %w", guild.ID, err))
			continue
		}

		targetEvents := FilterTargetEvents(events, now)
		result.TargetEvents += len(targetEvents)

		for _, event := range targetEvents {
			users, err := FetchAllInterestedUsers(client, event.GuildID, event.ID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("fetch interested users for event %s: %w", event.ID, err))
				continue
			}

			message := reminder.FormatReminder(event)
			for _, user := range users {
				if err := SendReminderDM(client, user.ID, message); err != nil {
					result.SentFailure++
					result.Errors = append(result.Errors, err)
					continue
				}
				result.SentSuccess++
			}
		}
	}

	return result
}
