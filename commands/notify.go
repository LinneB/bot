package commands

import (
	"bot/models"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/LinneB/twitchwh"
)

var notify = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) < 1 {
			return fmt.Sprintf("No subcommand provided. Usage: %s <add|remove> <channel>.", ctx.Command), nil
		}
		if len(ctx.Parameters) < 2 {
			return fmt.Sprintf("No channel provided. Usage: %s <add|remove> <channel>.", ctx.Command), nil
		}
		subcommand := ctx.Parameters[0]
		channel := strings.ToLower(ctx.Parameters[1])
		if !slices.Contains([]string{"add", "remove"}, subcommand) {
			return fmt.Sprintf("Invalid subcommand. Usage: %s <add|remove> <channel>.", ctx.Command), nil
		}

		id, err := state.Helix.LoginToID(channel)
		if err != nil {
			return "", fmt.Errorf("Could not get user ID: %w", err)
		}
		if id == nil {
			return fmt.Sprintf("User %s not found.", channel), nil
		}

		switch subcommand {
		case "add":
			var count int
			err := state.DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE chatid = $1 AND subscription_userid = $2", ctx.ChannelID, *id).Scan(&count)
			if err != nil {
				return "", fmt.Errorf("Could not query database: %w", err)
			}
			if count != 0 {
				return "Channel is already added to live notifications. Use #subscribe to be pinged when they go live.", nil
			}

			_, err = state.DB.Exec("INSERT INTO subscriptions (chatid, subscription_username, subscription_userid) VALUES ($1, $2, $3)", ctx.ChannelID, channel, *id)
			if err != nil {
				return "", fmt.Errorf("Could not insert into database: %w", err)
			}
			err = state.TwitchWH.AddSubscription("stream.online", "1", twitchwh.Condition{
				BroadcasterUserID: fmt.Sprint(*id),
			})
			var duplicate *twitchwh.DuplicateSubscriptionError
			if err == nil || errors.As(err, &duplicate) {
				return fmt.Sprintf("Added %s to notifications! Use #subscribe to be pinged when they go live.", channel), nil
			}
			return "", fmt.Errorf("Could not add subscription: %w", err)

		case "remove":
			var count int
			err := state.DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE chatid = $1 AND subscription_userid = $2", ctx.ChannelID, *id).Scan(&count)
			if err != nil {
				return "", fmt.Errorf("Could not query database: %w", err)
			}
			if count == 0 {
				return "Channel is not added to live notifications.", nil
			}

			_, err = state.DB.Exec("DELETE FROM subscriptions WHERE chatid = $1 AND subscription_userid = $2", ctx.ChannelID, *id)
			if err != nil {
				return "", fmt.Errorf("Could not delete from database: %w", err)
			}

			// Remove eventsub subscription if neccesary
			err = state.DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE subscription_userid = $1", *id).Scan(&count)
			if err != nil {
				return "", fmt.Errorf("Could not query database: %w", err)
			}
			if count == 0 {
				err := state.TwitchWH.RemoveSubscriptionByType("stream.online", twitchwh.Condition{
					BroadcasterUserID: fmt.Sprint(*id),
				})
				if err != nil {
					return "", fmt.Errorf("Could not remove subscription: %w", err)
				}
			}
			return fmt.Sprintf("Removed %s from live notifications.", channel), nil
		}
		return "Invalid subcommand.", nil
	},
	Metadata: metadata{
		Name:        "notify",
		Description: "Add/remove channels from live notifications.",
		Cooldown:    3 * time.Second,
		MinimumRole: RMod,
		Aliases:     []string{"notify", "notif", "livenotif"},
		Usage:       "#notify <add|remove> channel",
		Examples: []example{
			{
				Description: "Add a channel to the chats live notifications:",
				Command:     "#notify add forsen",
				Response:    "@linneb, Added forsen to notifications! Use #subscribe to be pinged when they go live.",
			},
			{
				Description: "The bot will send a message when that user goes live:",
				Response:    "https://twitch.tv/forsen just went live!",
			},
			{
				Description: "Users who have subscribed using #subscribe will also be @'d:",
				Response:    "https://twitch.tv/forsen just went live! @linneb",
			},
			{
				Description: "Removing a channel will permanently remove all subscribers, so be careful:",
				Command:     "#notify remove forsen",
				Response:    "@linneb, Removed forsen from live notifications.",
			},
		},
	},
}
