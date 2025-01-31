package commands

import (
	"bot/internal/helix"
	"bot/internal/models"
	"errors"
	"fmt"
	"strings"
	"time"
)

var thumbnail = command{
	Run: func(state *models.State, ctx Context) (string, error) {
		if len(ctx.Parameters) < 1 {
			return fmt.Sprintf("Missing channel. Usage: %s <channel>", ctx.Command), nil
		}
		channel := ctx.Parameters[0]
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

		link := strings.Replace(stream.ThumbnailURL, "{width}x{height}", "1920x1080", 1)
		return link, nil
	},
	Metadata: metadata{
		Name:        "thumbnail",
		Description: "Get the thumbnail of a stream.",
		Cooldown:    3 * time.Second,
		MinimumRole: RGeneric,
		Aliases:     []string{"thumbnail"},
		Usage:       "#thumbnail <channel>",
		Examples: []example{{
			Description: "Get the thumbnail of forsen's stream:",
			Command:     "#thumbnail forsen",
			Response:    "@linneb, https://static-cdn.jtvnw.net/previews-ttv/live_user_forsen-1920x1080.jpg",
		}},
	},
}
