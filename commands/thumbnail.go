package commands

import (
	"bot/helix"
	"bot/models"
	"encoding/json"
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

		var responseStruct struct {
			Data []helix.Stream `json:"data"`
		}
		err = json.NewDecoder(res.Body).Decode(&responseStruct)
		if err != nil {
			return "", fmt.Errorf("Could not parse json body: %w", err)
		}

		if len(responseStruct.Data) < 1 {
			return fmt.Sprintf("%s is offline.", channel), nil
		}
		link := strings.Replace(responseStruct.Data[0].ThumbnailURL, "{width}x{height}", "1920x1080", 1)
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
