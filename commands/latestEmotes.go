package commands

import (
	"bot/models"
	"bot/utils"
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"time"
)

var latestEmotes = command{
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
		type emote struct {
			Name      string `json:"name"`
			Timestamp int64  `json:"timestamp"`
		}
		var jsonBody struct {
			EmoteSet struct {
				Emotes []emote `json:"emotes"`
			} `json:"emote_set"`
		}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return "", fmt.Errorf("Could not decode JSON body: %w", err)
		}

		slices.SortFunc(jsonBody.EmoteSet.Emotes, func(a, b emote) int {
			return cmp.Compare(b.Timestamp, a.Timestamp)
		})
		emotes := jsonBody.EmoteSet.Emotes[0:min(len(jsonBody.EmoteSet.Emotes), 5)]
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
		Name:        "latestEmotes",
		Description: "Posts the 5 most recent 7TV emotes added to the current chat.",
		Cooldown:    3 * time.Second,
		MinimumRole: RGeneric,
		Aliases:     []string{"latestemotes", "le"},
		Usage:       "#le [channel]",
		Examples: []example{
			{
				Description: "Get the 5 most recent emotes.",
				Command:     "#le",
				Response:    "@linneb, buh (1 day ago), buh (3 days ago), buh (5 days ago), buh (6 days ago), buh (9 days ago)",
			},
			{
				Description: "Get the 5 most recent emotes added to a different channel.",
				Command:     "#le psp1g",
				Response:    "@linneb, buh (21 hours ago) buh (1 day ago) buh (1 day ago) buh (1 day ago) buh (1 day ago)",
			},
		},
	},
}
