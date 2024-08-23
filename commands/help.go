package commands

import (
	"bot/models"
	"bot/utils"
	"fmt"
	"strings"
	"time"
)

var help = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		commandName := "help"
		if len(ctx.Parameters) > 0 {
			commandName = ctx.Parameters[0]
		}
		command, found := Handler.GetCommandByName(commandName)
		if !found {
			command, found = Handler.GetCommandByAlias(commandName)
			if !found {
				return "Command name/alias not found.", nil
			}
		}
		return fmt.Sprintf("%s: %s Aliases: [%s]. Usage: \"%s\".",
			utils.CapitalizeFirstCharacter(command.Metadata.Name),
			command.Metadata.Description,
			strings.Join(command.Metadata.Aliases, ", "),
			command.Metadata.Usage,
		), nil
	},
	Metadata: metadata{
		Name:        "help",
		Description: "Shows a help message for a specific command.",
		Cooldown:    1 * time.Second,
		Aliases:     []string{"help", "usage"},
		Usage:       "#help [command]",
		Examples: []example{
			{
				Description: "Get some information about a command:",
				Command:     "#help live",
				Response:    "@linneb, Live: Sends information about a livestream. Aliases: [live, stream]. Usage: \"#live <channel>\".",
			},
			{
				Description: "You can also use an alias for a command:",
				Command:     "#help stream",
				Response:    "@linneb, Live: Sends information about a livestream. Aliases: [live, stream]. Usage: \"#live <channel>\".",
			},
		},
	},
}
