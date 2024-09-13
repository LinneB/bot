package commands

import (
	"bot/models"
	"encoding/json"
	"fmt"
	"time"
)

var randomLog = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) < 2 {
			return fmt.Sprintf("Missing user/channel. Usage: %s <user> <channel>", ctx.Command), nil
		}
		user := ctx.Parameters[0]
		channel := ctx.Parameters[1]
		res, err := state.Rustlog.NewRequest("GET", fmt.Sprintf("/channel/%s/user/%s/random?jsonBasic=true", channel, user))
		if err != nil {
			return "", fmt.Errorf("Could not send request: %w", err)
		}
		if res.StatusCode == 404 {
			return "User/channel not found, make sure they are being logged at https://logs.ivr.fi", nil
		}
		if res.StatusCode != 200 {
			return "", fmt.Errorf("Rustlog returned unhandled status code: %d", res.StatusCode)
		}
		var jsonBody struct {
			Messages []struct {
				Text        string
				DisplayName string
				Timestamp   time.Time
			}
		}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return "", fmt.Errorf("Could not decode JSON: %w", err)
		}
		if len(jsonBody.Messages) == 0 {
			return "", fmt.Errorf("Rustlog returned 0 messages")
		}
		msg := jsonBody.Messages[0]
		return fmt.Sprintf("[%s] %s: %s", msg.Timestamp.Format(time.DateTime), msg.DisplayName, msg.Text), nil
	},
	Metadata: metadata{
		Name:                "randomLog",
		Description:         "Sends a random message from a user in a chat.",
		ExtendedDescription: "This command requires that the channel is being logged on https://logs.ivr.fi and that the user has not opted out.",
		Cooldown:            3 * time.Second,
		MinimumRole:         RGeneric,
		Aliases:             []string{"randomlog", "rl"},
		Usage:               "#rl <user> <channel>",
		Examples: []example{{
			Description: "Get a random log from LinneB in forsen's chat:",
			Command:     "#rl linneb forsen",
			Response:    "@linneb, [2023-09-10 16:25:13] LinneB: elisDancing LETS elisDancing GO elisDancing FORSEN elisDancing",
		}},
	},
}
