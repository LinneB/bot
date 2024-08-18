package commands

import (
	"bot/models"
	"fmt"
	"time"
)

var ping = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		uptime := time.Since(*state.StartedAt).String()
		return fmt.Sprintf("Pong! Bot has been up for %s.", uptime), nil
	},
	Metadata: metadata{
		Name:        "ping",
		Description: "Returns uptime and other information.",
		Cooldown:    1,
		Aliases:     []string{"ping", "uptime"},
	},
}
