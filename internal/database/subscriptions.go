package database

import (
	"bot/internal/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetSubscriptions(db *pgxpool.Pool) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	rows, err := db.Query(context.Background(), "SELECT * FROM subscriptions")
	if err != nil {
		return nil, models.NewDatabaseError(err)
	}
	subscriptions, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Subscription])
	if err != nil {
		return nil, models.NewDatabaseError(err)
	}
	return subscriptions, nil
}

// Get all chats that have a subscription to streamUserID.
func GetSubscribedChats(db *pgxpool.Pool, streamUserID int) ([]models.Chat, error) {
	var chats []models.Chat
	rows, err := db.Query(context.Background(), `
SELECT c.chatname, c.chatid
FROM subscriptions su
JOIN chats c ON c.chatid = su.chatid
WHERE su.subscription_userid = $1`, streamUserID)
	if err != nil {
		return nil, models.NewDatabaseError(err)
	}
	chats, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Chat])
	if err != nil {
		return nil, models.NewDatabaseError(err)
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
		return nil, models.NewDatabaseError(err)
	}
	for rows.Next() {
		var chatname, subscriber string
		err := rows.Scan(&chatname, &subscriber)
		if err != nil {
			return nil, models.NewDatabaseError(err)
		}
		subscribers[chatname] = append(subscribers[chatname], subscriber)
	}
	return subscribers, nil
}

// Get a single subscription by chat ID and channel ID.
func GetSubscription(db *pgxpool.Pool, chatid, channelid int) (models.Subscription, bool, error) {
	rows, _ := db.Query(context.Background(), "SELECT * FROM subscriptions WHERE chatid = $1 AND subscription_userid = $2", chatid, channelid)
	subscription, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Subscription])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Subscription{}, false, nil
		}
		return models.Subscription{}, false, models.NewDatabaseError(err)
	}
	return subscription, true, nil
}

// Add a subscription to the database.
func CreateSubscription(db *pgxpool.Pool, sub models.Subscription) error {
	_, err := db.Exec(
		context.Background(),
		"INSERT INTO subscriptions (chatid, subscription_username, subscription_userid) VALUES ($1, $2, $3)",
		sub.ChatID,
		sub.SubscriptionUsername,
		sub.SubscriptionUserID,
	)
	if err != nil {
		return models.NewDatabaseError(err)
	}
	return nil
}

// Remove a subscription from the database.
func DeleteSubscription(db *pgxpool.Pool, sub models.Subscription) error {
	_, err := db.Exec(
		context.Background(),
		"DELETE FROM subscriptions WHERE chatid = $1 AND subscription_userid = $2",
		sub.ChatID,
		sub.SubscriptionUserID,
	)
	if err != nil {
		return models.NewDatabaseError(err)
	}
	return nil
}

// Check if any chat is subscribed to a channel.
// This basically checks if there should be an eventsub subscription for the given channel.
func IsChannelSubscribed(db *pgxpool.Pool, channelid int) (bool, error) {
	var exists bool
	err := db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM subscriptions WHERE subscription_userid = $1)", channelid).Scan(&exists)
	if err != nil {
		return false, models.NewDatabaseError(err)
	}
	return exists, nil
}
