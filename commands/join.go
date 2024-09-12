package commands

import (
	"bot/models"
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
			err = state.DB.QueryRow("SELECT COUNT(*) FROM chats WHERE chatname = ?", channel).Scan(&chatCount)
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
			_, err = state.DB.Exec("INSERT INTO chats(chatid, chatname) VALUES (?, ?)", channelID, channel)
			if err != nil {
				return "", fmt.Errorf("Could not insert chat: %w", err)
			}
			state.IRC.Join(channel)
			return fmt.Sprintf("Joining channel %s.", channel), nil
		}

		if ctx.Invocation == "part" {
			if ctx.IsBroadcaster {
				if len(ctx.Parameters) < 1 || ctx.Parameters[0] != "DELETEME" {
					return fmt.Sprintf("This command will part this chat and DELETE all commands and live notifications PERMANENTLY. Use %s DELETEME to confirm.", ctx.Command), nil
				}
				_, err := state.DB.Exec("DELETE FROM chats WHERE chatid = ?", ctx.ChannelID)
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
				meta, err := state.DB.Exec("DELETE FROM chats WHERE chatname = ?", channel)
				if err != nil {
					return "", fmt.Errorf("Could not delete from database: %w", err)
				}
				affected, _ := meta.RowsAffected() // err is always nil in go-sqlite3
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
		Name:        "join",
		Description: "Join/part channels. Broadcaster required to part, admin required to add.",
		Cooldown:    1 * time.Second,
		Aliases:     []string{"join", "part"},
		Usage:       "#ping",
		Examples:    []example{{}},
	},
}