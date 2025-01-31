package commands

import (
	"bot/internal/database"
	"bot/internal/helix"
	"bot/internal/models"
	"errors"
	"fmt"
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
		if subcommand != "add" && subcommand != "remove" {
			return fmt.Sprintf("Invalid subcommand. Usage: %s <add|remove> <channel>.", ctx.Command), nil
		}

		id, found, err := helix.LoginToID(state.Http, channel)
		if err != nil {
			return "", fmt.Errorf("Could not get user ID: %w", err)
		}
		if !found {
			return fmt.Sprintf("User %s not found.", channel), nil
		}

		if subcommand == "add" {
			_, found, err := database.GetSubscription(state.DB, ctx.ChannelID, id)
			if err != nil {
				return "", fmt.Errorf("Could not get subscription: %w", err)
			}
			if found {
				return fmt.Sprintf("Channel is already added to live notifications. Use %ssubscribe to be pinged when they go live.", state.Config.Prefix), nil
			}

			err = database.CreateSubscription(state.DB, models.Subscription{
				ChatID:               ctx.ChannelID,
				SubscriptionUsername: channel,
				SubscriptionUserID:   id,
			})
			if err != nil {
				return "", fmt.Errorf("Could not create subscription: %w", err)
			}
			err = state.TwitchWH.AddSubscription("stream.online", "1", twitchwh.Condition{
				BroadcasterUserID: fmt.Sprint(id),
			})
			var duplicate *twitchwh.DuplicateSubscriptionError
			if err == nil || errors.As(err, &duplicate) {
				return fmt.Sprintf("Added %s to notifications! Use %ssubscribe to be pinged when they go live.", channel, state.Config.Prefix), nil
			}
			return "", fmt.Errorf("Could not add eventsub subscription: %w", err)
		}
		if subcommand == "remove" {
			subscription, found, err := database.GetSubscription(state.DB, ctx.ChannelID, id)
			if err != nil {
				return "", fmt.Errorf("Could not get subscription: %w", err)
			}
			if !found {
				return "Channel is not added to live notifications.", nil
			}

			err = database.DeleteSubscription(state.DB, subscription)
			if err != nil {
				return "", fmt.Errorf("Could not delete subscription: %w", err)
			}

			// Remove eventsub subscription if neccesary
			subbed, err := database.IsChannelSubscribed(state.DB, subscription.SubscriptionUserID)
			if err != nil {
				return "", fmt.Errorf("Could not check if channel is subscribed: %w", err)
			}
			if !subbed {
				err = state.TwitchWH.RemoveSubscriptionByType("stream.online", twitchwh.Condition{
					BroadcasterUserID: fmt.Sprint(id),
				})
				if err != nil {
					return "", fmt.Errorf("Could not remove subscription: %w", err)
				}
			}
			return fmt.Sprintf("Removed %s from live notifications.", channel), nil
		}
		return "", fmt.Errorf("This error is impossible and will never happen")
	},
	Metadata: metadata{
		Name:                "notify",
		Description:         "Add/remove channels from live notifications.",
		ExtendedDescription: "The bot can send notifications to a chat when a channel goes live. This command is used to add/remove channels. If you want to be pinged for an existing notification, you can use the \"subscribe\" command.",
		Cooldown:            3 * time.Second,
		MinimumRole:         RMod,
		Aliases:             []string{"notify", "notif", "livenotif"},
		Usage:               "#notify <add|remove> <channel>",
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
