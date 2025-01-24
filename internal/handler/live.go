package handler

import (
	"bot/internal/helix"
	"bot/internal/models"
	"bot/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/LinneB/twitchwh"
)

func OnLive(state *models.State) func(twitchwh.StreamOnline) {
	return func(event twitchwh.StreamOnline) {
		streamUserID, err := strconv.Atoi(event.BroadcasterUserID)
		if err != nil {
			log.Printf("UserID \"%s\" is not convertable to int: %s", event.BroadcasterUserID, err)
			return
		}
		// Map of chats and their subscribers
		subscribers := make(map[string][]string)

		// Get subscribed chats
		query := "SELECT c.chatname, c.chatid FROM subscriptions su JOIN chats c ON c.chatid = su.chatid WHERE su.subscription_userid = $1"
		rows, err := state.DB.Query(context.Background(), query, streamUserID)
		if err != nil {
			log.Printf("Could not query database: %s", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var chat models.Chat
			err := rows.Scan(&chat.ChatName, &chat.ChatID)
			if err != nil {
				log.Printf("Could not scan row: %s", err)
				return
			}
			subscribers[chat.ChatName] = []string{}
		}

		// Get subscribed users
		rows, err = state.DB.Query(context.Background(), `
SELECT
  c.chatname,
  s.subscriber_username
FROM
  subscribers s
  JOIN chats c ON c.chatid = s.chatid
  JOIN subscriptions su ON su.subscription_id = s.subscription_id
WHERE
  su.subscription_userid = $1;`, streamUserID)
		if err != nil {
			log.Printf("Could not query database: %s", err)
			return
		}
		for rows.Next() {
			var (
				chatName string
				username string
			)
			err := rows.Scan(&chatName, &username)
			if err != nil {
				log.Printf("Could not scan row: %s", err)
				return
			}
			subscribers[chatName] = append(subscribers[chatName], username)
		}

		// Get stream information
		req, err := state.Helix.NewRequest("GET", "/streams?user_login="+event.BroadcasterUserLogin)
		if err != nil {
			log.Printf("Could not create request: %s", err)
			return
		}
		res, err := state.Helix.HttpClient.Do(req)
		if err != nil {
			log.Printf("Could not send request: %s", err)
			return
		}
		if res.StatusCode != 200 {
			log.Printf("Helix returned unhandled error code: %d", res.StatusCode)
			return
		}

		decoder := json.NewDecoder(res.Body)
		var responseStruct struct {
			Data []helix.Stream `json:"data"`
		}
		err = decoder.Decode(&responseStruct)
		if err != nil {
			log.Printf("Could not parse json body: %s", err)
			return
		}

		var liveMessage string
		if len(responseStruct.Data) > 0 {
			stream := responseStruct.Data[0]
			liveMessage = fmt.Sprintf("https://twitch.tv/%s just went live playing %s! \"%s\"", stream.UserLogin, stream.GameName, stream.Title)
		} else {
			liveMessage = fmt.Sprintf("https://twitch.tv/%s just went live!", event.BroadcasterUserLogin)
		}
		for chat, users := range subscribers {
			for _, message := range utils.SplitStreamOnlineMessage(liveMessage, users, 450) {
				state.IRC.Say(chat, message)
			}
		}
	}
}
