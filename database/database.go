package database

import (
	"database/sql"
)

type Chat struct {
	ChatID   int    `db:"chatid"`
	ChatName string `db:"chatname"`
}

type Subscription struct {
	ChatID               int    `db:"chatid"`
	SubscriptionUsername string `db:"subscription_username"`
	SubscriptionUserID   int    `db:"subscription_userid"`
	SubscriptionID       int    `db:"subscription_id"`
}

type Subscriber struct {
	ChatID             int    `db:"chatid"`
	SubscriptionID     int    `db:"subscription_id"`
	SubscriberUsername string `db:"subscriber_username"`
}

func CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS chats (
    chatid INTEGER PRIMARY KEY NOT NULL,
    chatname VARCHAR(50) NOT NULL
);
CREATE TABLE IF NOT EXISTS subscriptions (
    chatid INTEGER NOT NULL,
    subscription_username VARCHAR(50) NOT NULL,
    subscription_userid INTEGER NOT NULL,
    subscription_id INTEGER PRIMARY KEY,
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
    name VARCHAR(100) NOT NULL,
    reply VARCHAR(400) NOT NULL,
    CONSTRAINT fk_chats FOREIGN KEY (chatid) REFERENCES chats (chatid) ON DELETE CASCADE
);
    `)
	return err
}
