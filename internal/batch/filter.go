// Package batch implements eveping's daily reminder batch: filtering
// scheduled events, fetching interested users, and sending reminder DMs.
package batch

import (
	"time"

	"github.com/Sut103/discord-evenotify/internal/discordclient"
)

// FilterTargetEvents returns the events starting "tomorrow" (UTC calendar
// date, relative to now) whose status is SCHEDULED or ACTIVE. It is a pure
// function so date-boundary and status logic can be tested without a
// Discord client.
func FilterTargetEvents(events []discordclient.Event, now time.Time) []discordclient.Event {
	tomorrow := now.UTC().AddDate(0, 0, 1)
	tomorrowYear, tomorrowMonth, tomorrowDay := tomorrow.Date()

	var result []discordclient.Event
	for _, e := range events {
		if !isEligibleStatus(e.Status) {
			continue
		}
		year, month, day := e.ScheduledStartTime.UTC().Date()
		if year == tomorrowYear && month == tomorrowMonth && day == tomorrowDay {
			result = append(result, e)
		}
	}
	return result
}

func isEligibleStatus(status string) bool {
	return status == discordclient.EventStatusScheduled || status == discordclient.EventStatusActive
}
