package commands

import (
	"bot/internal/helix"
	"bot/internal/models"
	"errors"
	"fmt"
	"strings"
	"time"
)

var live = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) < 1 {
			return fmt.Sprintf("Missing argument: %s <channel>.", ctx.Command), nil
		}

		channel := strings.ToLower(ctx.Parameters[0])
		stream, found, err := helix.GetStream(state.Http, channel)
		if err != nil {
			var ae *models.APIError
			if errors.As(err, &ae) {
				if ae.Status == 400 {
					return fmt.Sprintf("User %s not found.", channel), nil
				}
			}
			return "", fmt.Errorf("Could not get stream: %w", err)
		}
		if !found {
			return fmt.Sprintf("%s is offline.", channel), nil
		}

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
		MinimumRole: RGeneric,
		Aliases:     []string{"live", "stream"},
		Usage:       "#live <channel>",
		Examples: []example{
			{
				Description: "Check if a channel is live:",
				Command:     "#live forsen",
				Response:    "@linneb, https://twitch.tv/forsen has been live for 2h 24m playing \"TEKKEN 8\" with 5757 viewers. EWC 1 million USD top 8!",
			},
		},
	},
}
