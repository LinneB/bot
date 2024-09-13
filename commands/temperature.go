package commands

import (
	"bot/models"
	"fmt"
	"strconv"
	"time"
)

var temperature = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		if len(ctx.Parameters) < 1 {
			return fmt.Sprintf("Missing temperature. Usage: %s <temperature>", ctx.Command), nil
		}

		t, err := strconv.Atoi(ctx.Parameters[0])
		if err != nil {
			if nerr, ok := err.(*strconv.NumError); ok && nerr.Err == strconv.ErrRange {
				return "Temperature is too large!", nil
			}
			return "Temperature is not a number.", nil
		}
		if ctx.Invocation == "ctof" || ctx.Invocation == "ctf" {
			f := (t*9)/5 + 32
			return fmt.Sprintf("%dC is %dF.", t, f), nil
		}
		if ctx.Invocation == "ftoc" || ctx.Invocation == "ftc" {
			c := (t - 32) * 5 / 9
			return fmt.Sprintf("%dF is %dC.", t, c), nil
		}
		return "", fmt.Errorf("This error is impossible and will never happen")
	},
	Metadata: metadata{
		Name:        "temperature",
		Description: "Convert temperature to various other units.",
		Cooldown:    1 * time.Second,
		MinimumRole: RGeneric,
		Aliases:     []string{"ctof", "ctf", "ftoc", "ftc"},
		Usage:       "#<ctof|ftoc> <temperature>",
		Examples: []example{
			{
				Description: "Convert 20 celsius to fahrenheit:",
				Command:     "#ctof 20",
				Response:    "@linneb, 20C is 68F.",
			},
			{
				Description: "Convert 80 fahrenheit to real units:",
				Command:     "#ftoc 80",
				Response:    "@linneb, 80F is 26C.",
			},
		},
	},
}
