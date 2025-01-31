package commands

import (
	"bot/internal/database"
	"bot/internal/models"
	"fmt"
	"strings"
	"time"
)

var subscribe = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) < 1 {
			return fmt.Sprintf("Missing channel. Usage: %s <channel>", ctx.Command), nil
		}
		channel := strings.ToLower(ctx.Parameters[0])
		sub, found, err := database.GetSubscriptionByName(state.DB, ctx.ChannelID, channel)
		if err != nil {
			return "", fmt.Errorf("Could not get subscription: %w", err)
		}
		if !found {
			return fmt.Sprintf("This chat is not subscribed to %s. Moderators can use %snotify to add/remove channels.", channel, state.Config.Prefix), nil
		}

		isSubbed, err := database.IsUserSubscribed(state.DB, ctx.SenderUsername, sub.SubscriptionID)
		if err != nil {
			return "", fmt.Errorf("Could not get subscriber: %w", err)
		}

		if isSubbed {
			err = database.DeleteSubscriber(state.DB, ctx.SenderUsername, sub.SubscriptionID)
			if err != nil {
				return "", fmt.Errorf("Could not delete subscriber: %w", err)
			}
			return fmt.Sprintf("Unsubscribed from %s. You will no longer be notified when they go live.", channel), nil
		} else {
			err = database.AddSubscriber(state.DB, ctx.SenderUsername, sub.SubscriptionID, ctx.ChannelID)
			if err != nil {
				return "", fmt.Errorf("Could not add subscriber: %w", err)
			}
			return fmt.Sprintf("Subscribed to %s. You will be notified when they go live.", channel), nil
		}
	},
	Metadata: metadata{
		Name:        "subscribe",
		Description: "Subscribe/unsubscribe from live notifications.",
		Cooldown:    1 * time.Second,
		MinimumRole: RGeneric,
		Aliases:     []string{"subscribe"},
		Usage:       "#subscribe <channel>",
		Examples: []example{
			{
				Description: "Subscribe to a channel (this requires that a mod has added them to live notifications):",
				Command:     "#subscribe forsen",
				Response:    "@linneb, Subscribed to forsen. You will be notified when they go live.",
			},
			{
				Description: "This command is a toggle, run it again to unsubscribe:",
				Command:     "#subscribe forsen",
				Response:    "@linneb, Unsubscribed from forsen. You will no longer be notified when they go live.",
			},
		},
	},
}
