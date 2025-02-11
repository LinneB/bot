package models

import (
	"bot/internal/http"
	"time"

	"github.com/LinneB/twitchwh"
	irc "github.com/gempir/go-twitch-irc/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

type State struct {
	Config    Config
	DB        *pgxpool.Pool
	Http      http.Client
	IRC       *irc.Client
	StartedAt time.Time
	TwitchWH  *twitchwh.Client
}

type Config struct {
	Admins         []string `toml:"admins"`
	BindAddr       string   `toml:"bind_addr"`
	DatabaseURL    string   `toml:"database_url"`
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
