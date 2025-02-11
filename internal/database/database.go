package database

import (
	"bot/internal/models"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTables(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(), `
CREATE TABLE IF NOT EXISTS chats (
    chatid INTEGER PRIMARY KEY NOT NULL,
    chatname VARCHAR(50) NOT NULL
);
CREATE TABLE IF NOT EXISTS subscriptions (
    chatid INTEGER NOT NULL,
    subscription_username VARCHAR(50) NOT NULL,
    subscription_userid INTEGER NOT NULL,
    subscription_id SERIAL NOT NULL,
    PRIMARY KEY (subscription_id),
    CONSTRAINT fk_chats FOREIGN KEY (chatid) REFERENCES chats (chatid) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS subscribers (
    chatid INTEGER NOT NULL,
    subscription_id INTEGER NOT NULL,
    subscriber_username VARCHAR(50) NOT NULL,
    CONSTRAINT fk_subscriptions FOREIGN KEY (subscription_id) REFERENCES subscriptions (subscription_id) ON DELETE CASCADE,
    CONSTRAINT fk_chats FOREIGN KEY (chatid) REFERENCES chats (chatid) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS commands (
    chatid INTEGER NOT NULL,
    name VARCHAR(100) UNIQUE NOT NULL,
    reply VARCHAR(400) NOT NULL,
    CONSTRAINT fk_chats FOREIGN KEY (chatid) REFERENCES chats (chatid) ON DELETE CASCADE
);
    `)
	if err != nil {
		return models.NewDatabaseError(err)
	}
	return nil
}
