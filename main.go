package main

import (
	"bot/commands"
	"bot/helix"
	"bot/models"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	irc "github.com/gempir/go-twitch-irc/v4"
	_ "github.com/mattn/go-sqlite3"
)

func newContext(state *models.State, msg irc.PrivateMessage) (context commands.Context, err error) {
	SenderUserID, err := strconv.Atoi(msg.User.ID)
	if err != nil {
		return context, err
	}
	ChannelID, err := strconv.Atoi(msg.RoomID)
	if err != nil {
		return context, err
	}
	Arguments := strings.Fields(msg.Message)
	Invocation := strings.TrimPrefix(Arguments[0], state.Config.Prefix)

	return commands.Context{
		SenderUserID:      SenderUserID,
		SenderUsername:    msg.User.Name,
		SenderDisplayname: msg.User.DisplayName,
		ChannelID:         ChannelID,
		ChannelName:       msg.Channel,
		Message:           msg.Message,
		Arguments:         Arguments,
		Parameters:        Arguments[1:],
		Command:           Arguments[0],
		Invocation:        Invocation,
	}, nil
}

func onMessage(state *models.State) func(irc.PrivateMessage) {
	return func(msg irc.PrivateMessage) {
		context, err := newContext(state, msg)
		if err != nil {
			state.Logger.Printf("Could not create command context: %s", err)
		}
		command, found := commands.Handler.GetCommandByAlias(state.Config.Prefix, context.Command)
		if !found {
			state.Logger.Printf("No command found for %s", msg.Message)
			return
		}
		if commands.Handler.IsOnCooldown(context.SenderUserID, &command) {
			return
		}
		commands.Handler.SetCooldown(context.SenderUserID, &command)
		reply, err := command.Run(state, context)
		if err != nil {
			state.Logger.Printf("Command execution failed: %s", err)
		}
		state.IRC.Say(msg.Channel, fmt.Sprintf("@%s, %s", msg.User.Name, reply))
	}
}

func main() {
	// TODO: Log to both stdout and a log file
	logger := log.New(os.Stdout, "Bot: ", log.Ltime|log.Lshortfile)
	startedAt := time.Now()

	logger.Println("Loading config file")
	var config models.Config
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		logger.Fatalf("Could not read and parse config file: %s", err)
	}

	logger.Println("Opening sqlite database")
	db, err := sql.Open("sqlite3", config.DatabasePath)
	if err != nil {
		logger.Fatalf("Could not open sqlite database: %s", err)
	}

	logger.Println("Creating Helix client")
	helix := helix.Client{
		ClientID:   config.Identity.ClientID,
		HelixURL:   "https://api.twitch.tv/helix",
		HttpClient: &http.Client{},
		Token:      config.Identity.HelixToken,
	}
	valid, err := helix.ValidateToken()
	if err != nil {
		logger.Fatalf("Could not validate Helix token: %s", err)
	}
	if !valid {
		logger.Fatalf("Helix token invalid")
	}

	ircClient := irc.NewClient(
		config.Identity.BotUsername,
		fmt.Sprintf("oauth:%s", config.Identity.HelixToken),
	)

	state := models.State{
		Config:    config,
		DB:        db,
		IRC:       ircClient,
		Logger:    logger,
		StartedAt: &startedAt,
	}

	ircClient.OnPrivateMessage(onMessage(&state))
	ircClient.OnConnect(func() { logger.Println("Connected to chat") })
	ircClient.Join(config.Channel)

	if err := ircClient.Connect(); err != nil {
		logger.Fatalf("Twitch chat connection failed: %s", err)
	}
}
