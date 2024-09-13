package commands

import (
	"bot/models"
	"fmt"
	"strings"
	"time"
)

var cmd = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) == 0 {
			return fmt.Sprintf("Missing subcommand: Usage: %s <add|remove|edit> [args].", ctx.Command), nil
		}

		switch ctx.Parameters[0] {
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
				return "Command name conflicts with an existing command.", nil
			}

			var count int
			err := state.DB.QueryRow("SELECT COUNT(*) FROM commands WHERE chatid = $1 AND name = $2", ctx.ChannelID, commandName).Scan(&count)
			if err != nil {
				return "", fmt.Errorf("Could not query database: %w", err)
			}
			if count != 0 {
				return fmt.Sprintf("%s is already a command.", commandName), nil
			}

			_, err = state.DB.Exec("INSERT INTO commands (chatid, name, reply) VALUES ($1, $2, $3)", ctx.ChannelID, commandName, reply)
			if err != nil {
				return "", fmt.Errorf("Could not insert into database: %w", err)
			}
			return fmt.Sprintf("Added command \"%s\".", commandName), nil

		case "remove":
			if len(ctx.Parameters) < 2 {
				return fmt.Sprintf("Missing command name: Usage: %s remove <name>.", ctx.Command), nil
			}
			commandName := strings.ToLower(ctx.Parameters[1])

			var count int
			err := state.DB.QueryRow("SELECT COUNT(*) FROM commands WHERE chatid = $1 AND name = $2", ctx.ChannelID, commandName).Scan(&count)
			if err != nil {
				return "", fmt.Errorf("Could not query database: %w", err)
			}
			if count == 0 {
				return fmt.Sprintf("%s is not a command.", commandName), nil
			}

			_, err = state.DB.Exec("DELETE FROM commands WHERE chatid = $1 AND name = $2", ctx.ChannelID, commandName)
			if err != nil {
				return "", fmt.Errorf("Could not delete from database: %w", err)
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

			var count int
			err := state.DB.QueryRow("SELECT COUNT(*) FROM commands WHERE chatid = $1 AND name = $2", ctx.ChannelID, commandName).Scan(&count)
			if err != nil {
				return "", fmt.Errorf("Could not query database: %w", err)
			}
			if count == 0 {
				return fmt.Sprintf("%s is not a command.", commandName), nil
			}

			_, err = state.DB.Exec("UPDATE commands SET reply = $1 WHERE chatid = $2 AND name = $3", reply, ctx.ChannelID, commandName)
			if err != nil {
				return "", fmt.Errorf("Could not update database: %w", err)
			}
			return fmt.Sprintf("Edited command \"%s\".", commandName), nil
		}
		return fmt.Sprintf("Invalid subcommand: Usage: %s <add|remove|edit> [args].", ctx.Command), nil
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
