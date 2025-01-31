package database

import (
	"bot/internal/models"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Check if a user is subscribed to a subscription.
func IsUserSubscribed(db *pgxpool.Pool, user string, subID int) (bool, error) {
	var exists bool
	err := db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM subscribers WHERE subscriber_username = $1 AND subscription_id = $2)", user, subID).Scan(&exists)
	if err != nil {
		return false, models.NewDatabaseError(err)
	}
	return exists, nil
}

func DeleteSubscriber(db *pgxpool.Pool, user string, subID int) error {
	_, err := db.Exec(context.Background(), "DELETE FROM subscribers WHERE subscriber_username = $1 AND subscription_id = $2", user, subID)
	if err != nil {
		return models.NewDatabaseError(err)
	}
	return nil
}

func AddSubscriber(db *pgxpool.Pool, user string, subID int, chatid int) error {
	_, err := db.Exec(context.Background(), "INSERT INTO subscribers (chatid, subscriber_username, subscription_id) VALUES ($1, $2, $3)", chatid, user, subID)
	if err != nil {
		return models.NewDatabaseError(err)
	}
	return nil
}
