package commands

import (
	"bot/internal/helix"
	"bot/internal/http"
	"bot/internal/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var title = command{
	Run: func(state *models.State, ctx Context) (reply string, err error) {
		id := ctx.ChannelID
		if len(ctx.Parameters) > 0 {
			login := strings.ToLower(ctx.Parameters[0])
			userid, found, err := helix.LoginToID(state.Http, login)
			if err != nil {
				return "", fmt.Errorf("Could not get user ID: %w", err)
			}
			if !found {
				return fmt.Sprintf("User %s not found.", login), nil
			}
			id = userid
		}
		req := http.Request{
			Method: "GET",
			URL:    helix.HelixURL + fmt.Sprintf("/channels?broadcaster_id=%d", id),
		}
		res, err := state.Http.GenericRequest(req)
		if err != nil {
			return "", &models.APIError{
				URL: req.Url(),
				Err: err,
			}
		}
		if res.StatusCode != 200 {
			return "", &models.APIError{
				Status: res.StatusCode,
				URL:    req.Url(),
			}
		}

		var jsonBody struct {
			Data []struct {
				Broadcaster_name string `json:"broadcaster_name"`
				Title            string `json:"title"`
			} `json:"data"`
		}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return "", fmt.Errorf("Could not decode JSON: %w", err)
		}
		if len(jsonBody.Data) == 0 {
			return "", &models.APIError{
				Status: res.StatusCode,
				URL:    req.Url(),
			}
		}
		channel := jsonBody.Data[0]
		return fmt.Sprintf("Title of %s is: %s", channel.Broadcaster_name, channel.Title), nil
	},
	Metadata: metadata{
		Name:        "title",
		Description: "Gets the title of a channel. Defaults to current chat.",
		Cooldown:    3 * time.Second,
		MinimumRole: RGeneric,
		Aliases:     []string{"title"},
		Usage:       "#title [channel]",
		Examples: []example{
			{
				Description: "Get the current chats title:",
				Command:     "#title",
				Response:    "@linneb, Title of LinneB is: gaycatwithsweetbabysrayhoneymustart",
			},
			{
				Description: "Get the title of another channel:",
				Command:     "#title sodapoppin",
				Response:    "@linneb, Title of sodapoppin is: Draft just... couldn't be simple could it | RAID PREP 9AM CST PULLS 1PM SATURDAY | !starforge !gamersupps",
			},
		},
	},
}
