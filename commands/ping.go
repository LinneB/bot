package commands

import (
	"bot/models"
	"bot/utils"
	"fmt"
	"time"
)

var ping = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		uptime := utils.PrettyDuration(time.Since(*state.StartedAt))
		dbStart := time.Now()
		err = state.DB.Ping()
		if err != nil {
			return reply, fmt.Errorf("Could not ping database: %w", err)
		}
		dbPing := utils.PrettyDuration(time.Since(dbStart))
		return fmt.Sprintf("Pong! Bot has been up for %s. Database ping is %s.", uptime, dbPing), nil
	},
	Metadata: metadata{
		Name:        "ping",
		Description: "Returns uptime and other information.",
		Cooldown:    1 * time.Second,
		Aliases:     []string{"ping", "uptime"},
	},
}
