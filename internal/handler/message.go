package handler

import (
	"bot/internal/commands"
	"bot/internal/database"
	"bot/internal/models"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	irc "github.com/gempir/go-twitch-irc/v4"
)

func OnMessage(state *models.State) func(irc.PrivateMessage) {
	return func(msg irc.PrivateMessage) {
		if !strings.HasPrefix(msg.Message, state.Config.Prefix) {
			return
		}
		ctx, err := commands.NewContext(state, msg)
		if err != nil {
			log.Printf("Could not create command context: %s", err)
		}

		// Interactive command
		command, found := commands.Handler.GetCommandByAlias(ctx.Invocation)
		if found {
			if ctx.Role < command.Metadata.MinimumRole {
				return
			}
			if commands.Handler.IsOnCooldown(ctx.SenderUserID, command.Metadata.Name, command.Metadata.Cooldown) {
				return
			}
			commands.Handler.SetCooldown(ctx.SenderUserID, command.Metadata.Name)
			now := time.Now()
			reply, err := command.Run(state, ctx)
			if err != nil {
				var ae *models.APIError
				if errors.As(err, &ae) {
					log.Printf("%s", err)
					state.IRC.Say(msg.Channel, fmt.Sprintf("@%s, :( 3rd party API failure.", msg.User.Name))
				} else {
					log.Printf("Command execution failed: %s", err)
				}
				return
			}
			log.Printf("Executed %s in %s", command.Metadata.Name, time.Since(now))
			if reply != "" {
				state.IRC.Say(msg.Channel, fmt.Sprintf("@%s, %s", msg.User.Name, reply))
			} else {
				log.Printf("Command returned empty reply")
			}
			return
		}

		// Static command
		cmd, found, err := database.GetCommand(state.DB, ctx.ChannelID, ctx.Invocation)
		if err != nil {
			log.Printf("Could not query database: %s", err)
			return
		}
		if found {
			if commands.Handler.IsOnCooldown(ctx.SenderUserID, ctx.Invocation, 1*time.Second) {
				return
			}
			commands.Handler.SetCooldown(ctx.SenderUserID, ctx.Invocation)
			state.IRC.Say(msg.Channel, fmt.Sprintf("@%s, %s", msg.User.Name, cmd.Reply))
		}
	}
}
