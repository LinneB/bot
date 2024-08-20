package models

import (
	"bot/helix"
	"database/sql"
	"log"
	"time"

	"github.com/LinneB/twitchwh"
	irc "github.com/gempir/go-twitch-irc/v4"
)

type State struct {
	Config    Config
	DB        *sql.DB
	IRC       *irc.Client
	Helix     *helix.Client
	Logger    *log.Logger
	StartedAt *time.Time
	TwitchWH  *twitchwh.Client
}

type Config struct {
	DatabasePath string `toml:"database_path"`
	Channel      string `toml:"channel"`
	Prefix       string `toml:"prefix"`
	Identity     struct {
		BotUsername  string `toml:"bot_username"`
		HelixToken   string `toml:"helix_token"`
		ClientID     string `toml:"client_id"`
		ClientSecret string `toml:"client_secret"`
	}
	Eventsub struct {
		WebhookURL    string `toml:"webhook_url"`
		WebhookSecret string `toml:"webhook_secret"`
	}
}
