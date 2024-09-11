package commands

import (
	"bot/models"
	"bot/utils"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

var randomEmotes = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		userid := ctx.ChannelID
		if len(ctx.Parameters) > 0 {
			login := ctx.Parameters[0]
			id, err := state.Helix.LoginToID(login)
			if err != nil {
				return "", fmt.Errorf("Could not get ID: %w", err)
			}
			if id == nil {
				return fmt.Sprintf("User %s not found.", login), nil
			}
			userid = *id
		}
		res, err := state.SevenTV.NewRequest("GET", "/users/twitch/"+fmt.Sprint(userid))
		if err != nil {
			return "", fmt.Errorf("Could not send request: %w", err)
		}
		if res.StatusCode == 404 {
			return "User does not have a 7TV profile.", nil
		}
		if res.StatusCode != 200 {
			return "", fmt.Errorf("SevenTV returned unhandled status code: %d", res.StatusCode)
		}
		var jsonBody struct {
			EmoteSet struct {
				Emotes []struct {
					Name      string `json:"name"`
					Timestamp int64  `json:"timestamp"`
				} `json:"emotes"`
			} `json:"emote_set"`
		}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return "", fmt.Errorf("Could not decode JSON body: %w", err)
		}

		emotes := jsonBody.EmoteSet.Emotes
		rand.Shuffle(len(emotes), func(i, j int) {
			emotes[i], emotes[j] = emotes[j], emotes[i]
		})
		emotes = emotes[0:min(len(emotes), 5)]
		if len(emotes) == 0 {
			return "This channel does not have any 7tv emotes.", nil
		}
		for _, e := range emotes {
			timestamp := time.Unix(e.Timestamp/1000, 0)
			reply += fmt.Sprintf("%s (%s ago) ", e.Name, utils.PrettyDuration(time.Since(timestamp)))
		}
		return reply, nil
	},
	Metadata: metadata{
		Name:        "randomEmotes",
		Description: "Posts 5 random 7TV emotes.",
		Cooldown:    3 * time.Second,
		Aliases:     []string{"randomemotes", "re"},
		Usage:       "#re [channel]",
		Examples: []example{
			{
				Description: "Post 5 random emotes:",
				Command:     "#re",
				Response:    "@linneb, buh (2 seconds ago), buh (3 weeks ago), buh (5 days ago), buh (3 months ago), buh (1 week ago)",
			},
		},
	},
}
