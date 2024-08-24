package commands

import (
	"bot/models"
	"database/sql"
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
		var subID int
		err = state.DB.QueryRow("SELECT subscription_id FROM subscriptions WHERE chatid = $1 AND subscription_username = $2", ctx.ChannelID, channel).Scan(&subID)
		if err == sql.ErrNoRows {
			return fmt.Sprintf("This chat is not subscribed to %s.", channel), nil
		}
		if err != nil {
			return "", fmt.Errorf("Could not query database: %w", err)
		}

		var count int
		err = state.DB.QueryRow("SELECT COUNT(*) FROM subscribers WHERE chatid = $1 AND subscriber_username = $2", ctx.ChannelID, ctx.SenderUsername).Scan(&subID)
		if err != nil {
			return "", fmt.Errorf("Could not query database: %w", err)
		}

		if count > 0 {
			_, err = state.DB.Exec("DELETE FROM subscribers WHERE chatid = $1 AND subscriber_username = $2 AND subscription_id = $3", ctx.Command, ctx.SenderUsername, subID)
			if err != nil {
				return "", fmt.Errorf("Could not delete from database: %w", err)
			}
			return fmt.Sprintf("Unsubscribed from %s. You will no longer be notified when they go live.", channel), nil
		} else {
			_, err = state.DB.Exec("INSERT INTO subscribers (chatid, subscriber_username, subscription_id) VALUES ($1, $2, $3)", ctx.ChannelID, ctx.SenderUsername, subID)
			if err != nil {
				return "", fmt.Errorf("Could not insert into database: %w", err)
			}
			return fmt.Sprintf("Subscribed to %s. You will be notified when they go live.", channel), nil
		}
	},
	Metadata: metadata{
		Name:        "subscribe",
		Description: "Subscribe/unsubscribe from live notifications.",
		Cooldown:    1 * time.Second,
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
