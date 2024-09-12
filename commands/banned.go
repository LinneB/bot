package commands

import (
	"bot/models"
	"encoding/json"
	"fmt"
	"time"
)

var banned = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) < 1 {
			return fmt.Sprintf("Missing username. Usage: %s <user>", ctx.Command), nil
		}
		res, err := state.IVR.NewRequest("GET", "/twitch/user?login="+ctx.Parameters[0])
		if err != nil {
			return "", fmt.Errorf("Could not send request: %w", err)
		}
		if res.StatusCode != 200 {
			return "", fmt.Errorf("IVR returned unhandled status code: %d", res.StatusCode)
		}
		var jsonBody []struct {
			Banned      bool
			BanReason   string
			DisplayName string
		}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return "", fmt.Errorf("Could not decode JSON body: %w", err)
		}
		if len(jsonBody) == 0 {
			return fmt.Sprintf("User %s not found.", ctx.Parameters[0]), nil
		}
		if jsonBody[0].Banned {
			return fmt.Sprintf("%s is BANNED: %s BOP", jsonBody[0].DisplayName, jsonBody[0].BanReason), nil
		}
		return fmt.Sprintf("%s is not banned.", jsonBody[0].DisplayName), nil
	},
	Metadata: metadata{
		Name:        "banned",
		Description: "Checks if a user is banned on Twitch.",
		Cooldown:    3 * time.Second,
		Aliases:     []string{"banned"},
		Usage:       "#banned <user>",
		Examples: []example{
			{
				Description: "Check if a user is banned:",
				Command:     "#banned forsen",
				Response:    "@linneb, forsen is not banned.",
			},
			{
				Description: "Check if a user is banned:",
				Command:     "#banned iceposeidon",
				Response:    "@linneb, IcePoseidon is BANNED: TOS_INDEFINITE BOP",
			},
		},
	},
}
