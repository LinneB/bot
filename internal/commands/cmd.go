package commands

import (
	"bot/internal/database"
	"bot/internal/models"
	"fmt"
	"strings"
	"time"
)

var cmd = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) == 0 {
			return fmt.Sprintf("Missing subcommand: Usage: %s <add|remove|edit> [args].", ctx.Command), nil
		}

		subcommand := ctx.Parameters[0]
		if subcommand != "add" && subcommand != "remove" && subcommand != "edit" {
			return fmt.Sprintf("Invalid subcommand: Usage: %s <add|remove|edit> [args].", ctx.Command), nil
		}

		switch subcommand {
		case "add":
			if len(ctx.Parameters) < 2 {
				return fmt.Sprintf("Missing command name: Usage: %s add <name> <reply>.", ctx.Command), nil
			}
			if len(ctx.Parameters) < 3 {
				return fmt.Sprintf("Missing command reply: Usage: %s add <name> <reply>.", ctx.Command), nil
			}
			commandName := strings.ToLower(ctx.Parameters[1])
			reply := strings.Join(ctx.Parameters[2:], " ")
			if len(commandName) >= 100 {
				return "Command name is too long! (max 100 characters).", nil
			}
			if len(reply) >= 400 {
				return "Command reply is too long! (max 400 characters).", nil
			}
			if _, found := Handler.GetCommandByAlias(commandName); found {
				return fmt.Sprintf("%s is already a command.", commandName), nil
			}

			_, found, err := database.GetCommand(state.DB, ctx.ChannelID, commandName)
			if err != nil {
				return "", fmt.Errorf("Could not get command: %w", err)
			}
			if found {
				return fmt.Sprintf("%s is already a command.", commandName), nil
			}

			err = database.CreateCommand(state.DB, models.Command{
				ChatID: ctx.ChannelID,
				Name:   commandName,
				Reply:  reply,
			})
			if err != nil {
				return "", fmt.Errorf("Could not create command: %w", err)
			}
			return fmt.Sprintf("Added command \"%s\".", commandName), nil

		case "remove":
			if len(ctx.Parameters) < 2 {
				return fmt.Sprintf("Missing command name: Usage: %s remove <name>.", ctx.Command), nil
			}
			commandName := strings.ToLower(ctx.Parameters[1])

			command, found, err := database.GetCommand(state.DB, ctx.ChannelID, commandName)
			if err != nil {
				return "", fmt.Errorf("Could not get command: %w", err)
			}
			if !found {
				return fmt.Sprintf("%s is not a command.", commandName), nil
			}

			err = database.DeleteCommand(state.DB, command)
			if err != nil {
				return "", fmt.Errorf("Could not delete command: %w", err)
			}
			return fmt.Sprintf("Removed command \"%s\".", commandName), nil

		case "edit":
			if len(ctx.Parameters) < 2 {
				return fmt.Sprintf("Missing command name: Usage: %s edit <name> <reply>.", ctx.Command), nil
			}
			if len(ctx.Parameters) < 3 {
				return fmt.Sprintf("Missing command reply: Usage: %s edit <name> <reply>.", ctx.Command), nil
			}
			commandName := strings.ToLower(ctx.Parameters[1])
			reply := strings.Join(ctx.Parameters[2:], " ")
			if len(reply) >= 400 {
				return "Command reply is too long! (max 400 characters).", nil
			}

			command, found, err := database.GetCommand(state.DB, ctx.ChannelID, commandName)
			if err != nil {
				return "", fmt.Errorf("Could not get command: %w", err)
			}
			if !found {
				return fmt.Sprintf("%s is not a command.", commandName), nil
			}

			err = database.UpdateCommand(state.DB, command, reply)
			if err != nil {
				return "", fmt.Errorf("Could not update command: %w", err)
			}
			return fmt.Sprintf("Edited command \"%s\".", commandName), nil
		}
		return "", fmt.Errorf("This error is impossible and will never happen")
	},
	Metadata: metadata{
		Name:        "cmd",
		Description: "Add/remove/edit static commands.",
		Cooldown:    1 * time.Second,
		MinimumRole: RMod,
		Aliases:     []string{"cmd", "command"},
		Usage:       "#cmd <add|remove|edit> [args]",
		Examples: []example{
			{
				Description: "Add a command to the current chat:",
				Command:     "#cmd add test This is a test command! :D",
				Response:    "@linneb, Added command \"test\".",
			},
			{
				Description: "Edit the command:",
				Command:     "#cmd edit test This is the 2nd iteration of the test command! :o",
				Response:    "@linneb, Edited command \"test\".",
			},
			{
				Description: "Remove the command:",
				Command:     "#cmd remove test",
				Response:    "@linneb, Removed command \"test\".",
			},
		},
	},
}
