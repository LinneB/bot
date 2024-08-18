package models

import (
	"log"
	"time"

	irc "github.com/gempir/go-twitch-irc/v4"
)

type State struct {
	Config    Config
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
