package commands

import (
	"bot/helix"
	"bot/models"
	"encoding/json"
	"fmt"
	"time"
)

var live = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) < 1 {
			return fmt.Sprintf("Missing argument: %s <channel>.", ctx.Command), nil
		}
		channel := ctx.Parameters[0]
		req, err := state.Helix.NewRequest("GET", "/streams?user_login="+channel)
		if err != nil {
			return "", fmt.Errorf("Could not create request: %w", err)
		}
		res, err := state.Helix.HttpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("Could not send request: %w", err)
		}
		if res.StatusCode == 400 {
			return fmt.Sprintf("User %s not found.", channel), nil
		}
		if res.StatusCode != 200 {
			return "", fmt.Errorf("Helix returned unhandled error code: %d", res.StatusCode)
		}

		decoder := json.NewDecoder(res.Body)
		var responseStruct struct {
			Data []helix.Stream `json:"data"`
		}
		err = decoder.Decode(&responseStruct)
		if err != nil {
			return "", fmt.Errorf("Could not parse json body: %w", err)
		}

		if len(responseStruct.Data) < 1 {
			return fmt.Sprintf("%s is offline.", channel), nil
		}
		stream := responseStruct.Data[0]
		liveDuration := time.Since(stream.StartedAt)
		hours := int(liveDuration.Hours())
		minutes := int(liveDuration.Minutes()) % 60
		message := fmt.Sprintf(
			"https://twitch.tv/%s has been live for %dh %dm playing \"%s\" with %d viewers. %s",
			stream.UserLogin,
			hours,
			minutes,
			stream.GameName,
			stream.ViewerCount,
			stream.Title,
		)
		return message, nil
	},
	Metadata: metadata{
		Name:        "live",
		Description: "Sends information about a livestream.",
		Cooldown:    3 * time.Second,
		Aliases:     []string{"live", "stream"},
		Usage:       "#live <channel>",
	},
}
