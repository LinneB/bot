package models

import (
	"bot/helix"
	"bot/http"
	"database/sql"
	"time"

	"github.com/LinneB/twitchwh"
	irc "github.com/gempir/go-twitch-irc/v4"
)

type State struct {
	Config    Config
	DB        *sql.DB
	Helix     helix.Client
	IRC       *irc.Client
	IVR       http.Client
	Rustlog   http.Client
	SevenTV   http.Client
	StartedAt time.Time
	TwitchWH  *twitchwh.Client
}

type Config struct {
	Admins         []string `toml:"admins"`
	BindAddr       string   `toml:"bind_addr"`
	DatabasePath   string   `toml:"database_path"`
	InitialChannel string   `toml:"initial_channel"`
	Prefix         string   `toml:"prefix"`
	Identity       struct {
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
