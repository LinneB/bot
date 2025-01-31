package commands

import (
	"bot/internal/helix"
	"bot/internal/http"
	"bot/internal/models"
	"bot/internal/utils"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var randomEmote = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		id := ctx.ChannelID
		if len(ctx.Parameters) > 0 {
			login := strings.ToLower(ctx.Parameters[0])
			userid, found, err := helix.LoginToID(state.Http, login)
			if err != nil {
				return "", fmt.Errorf("Could not get user ID: %w", err)
			}
			if !found {
				return fmt.Sprintf("User %s not found.", login), nil
			}
			id = userid
		}
		req := http.Request{
			Method: "GET",
			URL:    "https://7tv.io/v3/users/twitch/" + fmt.Sprint(id),
		}
		res, err := state.Http.GenericRequest(req)
		if err != nil {
			return "", &models.APIError{
				URL: req.Url(),
				Err: err,
			}
		}
		if res.StatusCode == 404 {
			return "User does not have a 7TV profile.", nil
		}
		if res.StatusCode != 200 {
			return "", &models.APIError{
				Status: res.StatusCode,
				URL:    req.Url(),
			}
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

		emotes := jsonBody.EmoteSet.Emotes
		if len(emotes) == 0 {
			return "This channel does not have any 7tv emotes.", nil
		}
		rand.Shuffle(len(emotes), func(i, j int) {
			emotes[i], emotes[j] = emotes[j], emotes[i]
		})
		emotes = emotes[0:min(len(emotes), 5)]
		for _, e := range emotes {
			timestamp := time.Unix(e.Timestamp/1000, 0)
			reply += fmt.Sprintf("%s (%s ago) ", e.Name, utils.PrettyDuration(time.Since(timestamp)))
		}
		return reply, nil
	},
	Metadata: metadata{
		Name:        "randomEmotes",
		Description: "Posts 5 random 7TV emotes. Defaults to the current chat.",
		Cooldown:    3 * time.Second,
		MinimumRole: RGeneric,
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
