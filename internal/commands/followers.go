package commands

import (
	"bot/internal/helix"
	"bot/internal/http"
	"bot/internal/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var followers = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		login := ctx.SenderUsername
		id := ctx.SenderUserID
		if len(ctx.Parameters) > 0 {
			login = strings.ToLower(ctx.Parameters[0])
			userid, found, err := helix.LoginToID(state.Http, login)
			if err != nil {
				return "", fmt.Errorf("Could not get user ID: %w", err)
			}
			if !found {
				return fmt.Sprintf("User %s not found.", login), nil
			}
			id = userid
		}

		// TODO: This could be moved to a function like helix.GetFollowers, but it seems pretty specific
		req := http.Request{
			Method: "GET",
			URL:    helix.HelixURL + fmt.Sprintf("/channels/followers?broadcaster_id=%d", id),
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
			}
		}
		var jsonBody struct {
			Total int `json:"total"`
		}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return "", fmt.Errorf("Could not decode JSON body: %w", models.NewSystemError(err))
		}
		return fmt.Sprintf("%s has %d followers.", login, jsonBody.Total), nil
	},
	Metadata: metadata{
		Name:        "followers",
		Description: "Show the number of followers for a user. Defaults to the sender.",
		Cooldown:    3 * time.Second,
		MinimumRole: RGeneric,
		Aliases:     []string{"followers", "followcount"},
		Usage:       "#followers [user]",
		Examples: []example{{
			Description: "Get the follow count of forsen:",
			Command:     "#followers forsen",
			Response:    "@linneb, forsen has 1755483 followers.",
		}},
	},
}
