package commands

import (
	"bot/models"
	"encoding/json"
	"fmt"
	"time"
)

var title = command{
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
		req, err := state.Helix.NewRequest("GET", fmt.Sprintf("/channels?broadcaster_id=%d", userid))
		if err != nil {
			return "", fmt.Errorf("Could not create request: %w", err)
		}
		res, err := state.Helix.HttpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("Could not send request: %w", err)
		}
		if res.StatusCode != 200 {
			return "", fmt.Errorf("Helix returned unexpected status code: %d", res.StatusCode)
		}

		var jsonBody struct {
			Data []struct {
				Broadcaster_name string `json:"broadcaster_name"`
				Title            string `json:"title"`
			} `json:"data"`
		}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return "", fmt.Errorf("Could not decode JSON: %w", err)
		}
		if len(jsonBody.Data) == 0 {
			return "", fmt.Errorf("Could not get channel information for id: %d", userid)
		}
		channel := jsonBody.Data[0]
		return fmt.Sprintf("Title of %s is: %s", channel.Broadcaster_name, channel.Title), nil
	},
	Metadata: metadata{
		Name:        "title",
		Description: "Gets the title of a channel. Defaults to current chat.",
		Cooldown:    3 * time.Second,
		MinimumRole: RGeneric,
		Aliases:     []string{"title"},
		Usage:       "#title [channel]",
		Examples: []example{
			{
				Description: "Get the current chats title:",
				Command:     "#title",
				Response:    "@linneb, Title of LinneB is: gaycatwithsweetbabysrayhoneymustart",
			},
			{
				Description: "Get the title of another channel:",
				Command:     "#title psp1g",
				Response:    "@linneb, Title of PSP1G is: PLAYING UNRAILED [WITH @deme & @zoil] !drama !facereveal !lawsuit",
			},
		},
	},
}
