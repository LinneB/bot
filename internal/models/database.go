package models

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

// TODO: ChatID is redundant in this struct, and should be removed
type Subscriber struct {
	ChatID             int    `db:"chatid"`
	SubscriptionID     int    `db:"subscription_id"`
	SubscriberUsername string `db:"subscriber_username"`
}

type Command struct {
	ChatID int    `db:"chatid"`
	Name   string `db:"subscription_id"`
	Reply  string `db:"subscriber_username"`
}
