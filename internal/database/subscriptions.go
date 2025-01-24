package database

import (
	"bot/internal/models"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Get all chats that have a subscription to streamUserID.
func GetSubscribedChats(db *pgxpool.Pool, streamUserID int) ([]models.Chat, error) {
	var chats []models.Chat
	rows, err := db.Query(context.Background(), `
SELECT c.chatname, c.chatid
FROM subscriptions su
JOIN chats c ON c.chatid = su.chatid
WHERE su.subscription_userid = $1`, streamUserID)
	if err != nil {
		return nil, err
	}
	chats, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Chat])
	if err != nil {
		return nil, err
	}
	return chats, nil
}

// Get all chats and subscribers that should be notified when streamUserID goes live.
// Returns a map of chatname to a slice of users to notify.
func GetSubscribers(db *pgxpool.Pool, streamUserID int) (map[string][]string, error) {
	subscribers := make(map[string][]string)
	rows, err := db.Query(context.Background(), `
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
		return nil, err
	}
	for rows.Next() {
		var chatname, subscriber string
		err := rows.Scan(&chatname, &subscriber)
		if err != nil {
			return nil, err
		}
		subscribers[chatname] = append(subscribers[chatname], subscriber)
	}
	return subscribers, nil
}
