package handler

import (
	"bot/internal/database"
	"bot/internal/helix"
	"bot/internal/models"
	"bot/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

func OnLive(state *models.State) func(rawEvent json.RawMessage) {
	return func(rawEvent json.RawMessage) {
		var event struct {
			BroadcasterUserID    string `json:"broadcaster_user_id"`
			BroadcasterUserLogin string `json:"broadcaster_user_login"`
		}
		if err := json.Unmarshal(rawEvent, &event); err != nil {
			log.Printf("Could not unmarshal event: %s", err)
		}
		streamUserID, err := strconv.Atoi(event.BroadcasterUserID)
		if err != nil {
			log.Printf("UserID \"%s\" is not convertable to int: %s", event.BroadcasterUserID, err)
			return
		}

		subscribedChats, err := database.GetSubscribedChats(state.DB, streamUserID)
		if err != nil {
			log.Printf("Could not get subscribed chats: %s", err)
			return
		}
		if len(subscribedChats) == 0 {
			log.Printf("No subscribed chats found for stream %d (%s)", streamUserID, event.BroadcasterUserLogin)
			return
		}
		subscribers, err := database.GetSubscribers(state.DB, streamUserID)
		if err != nil {
			log.Printf("Could not get subscribers: %s", err)
			return
		}

		// Add chats that have a subscription but no subscribers to the subscribers map
		for _, chat := range subscribedChats {
			if _, ok := subscribers[chat.ChatName]; !ok {
				subscribers[chat.ChatName] = []string{}
			}
		}

		// Get stream information
		stream, found, err := helix.GetStream(state.Http, event.BroadcasterUserLogin)
		if err != nil {
			log.Printf("Could not get stream information: %s", err)
			return
		}

		var liveMessage string
		if found {
			if stream.GameName != "" {
				liveMessage = fmt.Sprintf("https://twitch.tv/%s just went live playing %s! \"%s\"", stream.UserLogin, stream.GameName, stream.Title)
			} else {
				liveMessage = fmt.Sprintf("https://twitch.tv/%s just went live! \"%s\"", stream.UserLogin, stream.Title)
			}
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
