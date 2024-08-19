package models

import (
	"database/sql"
	"log"
	"time"

	irc "github.com/gempir/go-twitch-irc/v4"
)

type State struct {
	Config    Config
	DB        *sql.DB
	IRC       *irc.Client
	Logger    *log.Logger
	StartedAt *time.Time
}

type Config struct {
	DatabasePath string `toml:"database_path"`
	Channel      string `toml:"channel"`
	Prefix       string `toml:"prefix"`
	Identity     struct {
		BotUsername string `toml:"bot_username"`
		HelixToken  string `toml:"helix_token"`
		ClientID    string `toml:"client_id"`
		// TODO: Client secret (for webhooks)
	}
}
