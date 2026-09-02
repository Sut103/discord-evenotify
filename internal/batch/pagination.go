package batch

import "github.com/Sut103/discord-evenotify/internal/discordclient"

// eventUsersPageSize is the maximum number of users Discord returns per
// EventUsers call.
const eventUsersPageSize = 100

// FetchAllInterestedUsers retrieves every user interested in a scheduled
// event, following the after-cursor pagination of
// discordclient.Client.EventUsers until a short page signals the end.
func FetchAllInterestedUsers(client discordclient.Client, guildID, eventID string) ([]discordclient.User, error) {
	var all []discordclient.User
	after := ""
	for {
		page, err := client.EventUsers(guildID, eventID, after, eventUsersPageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if len(page) < eventUsersPageSize {
			break
		}
		after = page[len(page)-1].ID
	}
	return all, nil
}
