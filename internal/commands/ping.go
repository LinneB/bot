package commands

import (
	"bot/internal/models"
	"bot/internal/utils"
	"context"
	"fmt"
	"runtime"
	"time"
)

var ping = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		uptime := utils.PrettyDuration(time.Since(state.StartedAt))
		dbStart := time.Now()
		err = state.DB.Ping(context.Background())
		if err != nil {
			return reply, fmt.Errorf("Could not ping database: %w", err)
		}
		dbPing := utils.PrettyDuration(time.Since(dbStart))

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		kb := float32(memStats.HeapAlloc) / 1000.0
		memory := ""
		if kb/1000.0 > 1 {
			memory = fmt.Sprintf("%.1f MB", kb/1000.0)
		} else {
			memory = fmt.Sprintf("%.0f KB", kb)
		}

		return fmt.Sprintf("Pong! Bot has been up for %s. Database ping is %s. Heap usage: %s.", uptime, dbPing, memory), nil
	},
	Metadata: metadata{
		Name:        "ping",
		Description: "Returns uptime and other information.",
		Cooldown:    1 * time.Second,
		MinimumRole: RGeneric,
		Aliases:     []string{"ping", "uptime"},
		Usage:       "#ping",
		Examples: []example{
			{
				Description: "Check that the bot is alive:",
				Command:     "#ping",
				Response:    "@linneb, Pong! Bot has been up for 9 seconds. Database ping is 12 μs. Heap usage: 1.5 MB.",
			},
			{
				Description: "If the bot is offline, it wont respond!",
				Command:     "#ping",
				Response:    "",
			},
		},
	},
}
