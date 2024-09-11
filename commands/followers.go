package commands

import (
	"bot/models"
	"encoding/json"
	"fmt"
	"time"
)

var followers = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		login := ctx.ChannelName
		if len(ctx.Parameters) > 0 {
			login = ctx.Parameters[0]
		}
		res, err := state.IRV.NewRequest("GET", "/twitch/user?login="+login)
		if err != nil {
			return "", fmt.Errorf("Could not send request: %w", err)
		}
		if res.StatusCode != 200 {
			return "", fmt.Errorf("IVR returned unhandled status code: %d", res.StatusCode)
		}
		var jsonBody []struct {
			DisplayName string
			Followers   int
		}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return "", fmt.Errorf("Could not decode JSON body: %w", err)
		}
		if len(jsonBody) == 0 {
			return fmt.Sprintf("User %s not found.", login), nil
		}
		return fmt.Sprintf("%s has %d followers.", jsonBody[0].DisplayName, jsonBody[0].Followers), nil
	},
	Metadata: metadata{
		Name:        "followers",
		Description: "Gets the follow count for a user.",
		Cooldown:    3 * time.Second,
		Aliases:     []string{"followers", "followcount"},
		Usage:       "#followers [user]",
		Examples: []example{{
			Description: "Get the follow count of forsen:",
			Command:     "#followers forsen",
			Response:    "@linneb, forsen has 1755483 followers.",
		}},
	},
}
