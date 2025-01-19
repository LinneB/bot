package commands

import (
	"bot/internal/models"
	"context"
	"fmt"
	"strings"
	"time"
)

var join = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if !ctx.IsAdmin && !ctx.IsBroadcaster {
			return
		}

		if ctx.Invocation == "join" {
			if !ctx.IsAdmin {
				return
			}
			if len(ctx.Parameters) < 1 {
				return fmt.Sprintf("Missing channel. Usage: %s <channel>", ctx.Command), nil
			}
			channel := strings.ToLower(ctx.Parameters[0])
			var chatCount int
			err = state.DB.QueryRow(context.Background(), "SELECT COUNT(*) FROM chats WHERE chatname = ?", channel).Scan(&chatCount)
			if err != nil {
				return "", fmt.Errorf("Could not query database: %w", err)
			}
			if chatCount > 0 {
				return "Chat is already joined.", nil
			}

			channelID, err := state.Helix.LoginToID(channel)
			if err != nil {
				return "", fmt.Errorf("Could not get user: %w", err)
			}
			if channelID == nil {
				return fmt.Sprintf("User %s not found.", channel), nil
			}
			_, err = state.DB.Exec(context.Background(), "INSERT INTO chats(chatid, chatname) VALUES (?, ?)", channelID, channel)
			if err != nil {
				return "", fmt.Errorf("Could not insert chat: %w", err)
			}
			state.IRC.Join(channel)
			return fmt.Sprintf("Joining chat %s.", channel), nil
		}

		if ctx.Invocation == "part" {
			if ctx.IsBroadcaster {
				if len(ctx.Parameters) < 1 || ctx.Parameters[0] != "DELETEME" {
					return fmt.Sprintf("This command will part this chat and DELETE all commands and live notifications PERMANENTLY. Use %s DELETEME to confirm.", ctx.Command), nil
				}
				_, err := state.DB.Exec(context.Background(), "DELETE FROM chats WHERE chatid = ?", ctx.ChannelID)
				if err != nil {
					return "", fmt.Errorf("Could not delete from database: %w", err)
				}
				return "Parting channel. Until we meet again. :)", nil
			}
			if ctx.IsAdmin {
				if len(ctx.Parameters) < 1 {
					return fmt.Sprintf("Missing channel. Usage: %s <channel>", ctx.Command), nil
				}
				channel := strings.ToLower(ctx.Parameters[0])
				meta, err := state.DB.Exec(context.Background(), "DELETE FROM chats WHERE chatname = ?", channel)
				if err != nil {
					return "", fmt.Errorf("Could not delete from database: %w", err)
				}
				affected := meta.RowsAffected()
				if affected == 0 {
					return "Chat not found.", nil
				}
				state.IRC.Depart(channel)
				return fmt.Sprintf("Leaving chat %s.", channel), nil
			}
		}
		return "", fmt.Errorf("This error is impossible and will never happen")
	},
	Metadata: metadata{
		Name:                "join",
		Description:         "Join/part channels. Broadcaster required to part, admin required to add.",
		ExtendedDescription: "This command manages joined chats. Broadcasters can use it to remove the bot from their chat, while admins can use it to join/part any chat.",
		Cooldown:            1 * time.Second,
		MinimumRole:         RBroadcaster,
		Aliases:             []string{"join", "part"},
		Usage:               "#<join|part> <channel>",
		Examples: []example{
			{
				Description: "(Broadcaster) Remove bot from your chat:",
				Command:     "#part",
				Response:    "This command will part this chat and DELETE all commands and live notifications PERMANENTLY. Use #part DELETEME to confirm.",
			},
			{
				Description: "(Broadcaster) Confirm and remove bot:",
				Command:     "#part DELETEME",
				Response:    "Parting channel. Until we meet again. :)",
			},
			{
				Description: "(Admin) Join a channel:",
				Command:     "#join LinneB",
				Response:    "Joining chat LinneB.",
			},
			{
				Description: "(Admin) Part a channel:",
				Command:     "#part LinneB",
				Response:    "Leaving chat LinneB.",
			},
		},
	},
}
