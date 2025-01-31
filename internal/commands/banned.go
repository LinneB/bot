package commands

import (
	"bot/internal/http"
	"bot/internal/models"
	"encoding/json"
	"fmt"
	"time"
)

var banned = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) < 1 {
			return fmt.Sprintf("Missing username. Usage: %s <user>", ctx.Command), nil
		}
		req := http.Request{
			Method: "GET",
			URL:    "https://api.ivr.fi/v2/twitch/user?login=" + ctx.Parameters[0],
		}
		res, err := state.Http.GenericRequest(req)
		if err != nil {
			return "", &models.APIError{
				URL: req.Url(),
				Err: err,
			}
		}
		if res.StatusCode != 200 {
			return "", &models.APIError{
				Status: res.StatusCode,
				URL:    req.Url(),
				Err:    err,
			}
		}

		var body []struct {
			Banned      bool
			BanReason   string
			DisplayName string
		}
		err = json.NewDecoder(res.Body).Decode(&body)
		if err != nil {
			return "", fmt.Errorf("Could not decode json: %w", models.NewSystemError(err))
		}
		if len(body) == 0 {
			return fmt.Sprintf("User %s not found.", ctx.Parameters[0]), nil
		}
		if body[0].Banned {
			return fmt.Sprintf("%s is BANNED: %s BOP", body[0].DisplayName, body[0].BanReason), nil
		}
		return fmt.Sprintf("%s is not banned.", body[0].DisplayName), nil
	},
	Metadata: metadata{
		Name:        "banned",
		Description: "Checks if a user is banned on Twitch.",
		Cooldown:    3 * time.Second,
		MinimumRole: RGeneric,
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
