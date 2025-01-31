package commands

import (
	"bot/internal/helix"
	"bot/internal/models"
	"fmt"
	"time"
)

var id = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) < 1 {
			return fmt.Sprintf("Your ID is %d", ctx.SenderUserID), nil
		}
		id, found, err := helix.LoginToID(state.Http, ctx.Parameters[0])
		if err != nil {
			return "", fmt.Errorf("Could not get ID: %w", err)
		}
		if found {
			return fmt.Sprintf("ID of %s is %d", ctx.Parameters[0], id), nil
		} else {
			return fmt.Sprintf("User %s not found.", ctx.Parameters[0]), nil
		}
	},
	Metadata: metadata{
		Name:        "id",
		Description: "Gets the Twitch user ID for you or another user.",
		Cooldown:    1 * time.Second,
		MinimumRole: RGeneric,
		Aliases:     []string{"id", "userid"},
		Usage:       "#id [user]",
		Examples: []example{
			{
				Description: "Get the senders user ID:",
				Command:     "#id",
				Response:    "@linneb, Your ID is 215185844",
			},
			{
				Description: "Get the user ID of a different user:",
				Command:     "#id forsen",
				Response:    "@linneb, ID of forsen is 22484632",
			},
		},
	},
}
