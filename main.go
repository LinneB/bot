package main

import (
	"bot/commands"
	"bot/database"
	"bot/helix"
	"bot/models"
	"bot/utils"
	"bot/web"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/LinneB/twitchwh"
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
		IsMod:             msg.Tags["mod"] == "1",
		IsBroadcaster:     SenderUserID == ChannelID,
		IsAdmin:           slices.Contains(state.Config.Admins, msg.User.Name),
	}, nil
}

func onMessage(state *models.State) func(irc.PrivateMessage) {
	return func(msg irc.PrivateMessage) {
		if !strings.HasPrefix(msg.Message, state.Config.Prefix) {
			return
		}
		context, err := newContext(state, msg)
		if err != nil {
			log.Printf("Could not create command context: %s", err)
		}

		// Interactive command
		command, found := commands.Handler.GetCommandByAlias(context.Invocation)
		if found {
			if commands.Handler.IsOnCooldown(context.SenderUserID, command.Metadata.Name, command.Metadata.Cooldown) {
				return
			}
			commands.Handler.SetCooldown(context.SenderUserID, command.Metadata.Name)
			reply, err := command.Run(state, context)
			if err != nil {
				log.Printf("Command execution failed: %s", err)
			}
			state.IRC.Say(msg.Channel, fmt.Sprintf("@%s, %s", msg.User.Name, reply))
			return
		}

		// Static command
		var reply string
		err = state.DB.QueryRow("SELECT reply FROM commands WHERE chatid = $1 AND name = $2", context.ChannelID, context.Invocation).Scan(&reply)
		if err != nil {
			if err == sql.ErrNoRows {
				return
			}
			log.Printf("Could not query database: %s", err)
			return
		}
		if commands.Handler.IsOnCooldown(context.SenderUserID, context.Invocation, 1*time.Second) {
			return
		}
		commands.Handler.SetCooldown(context.SenderUserID, context.Invocation)
		state.IRC.Say(msg.Channel, fmt.Sprintf("@%s, %s", msg.User.Name, reply))
	}
}

func onLive(state *models.State) func(twitchwh.StreamOnline) {
	return func(event twitchwh.StreamOnline) {
		streamUserID, err := strconv.Atoi(event.BroadcasterUserID)
		if err != nil {
			log.Printf("UserID \"%s\" is not convertable to int: %s", event.BroadcasterUserID, err)
			return
		}
		// Map of chats and their subscribers
		log.Printf("%s went live!", event.BroadcasterUserName)
		subscribers := make(map[string][]string)

		// Get subscribed chats
		rows, err := state.DB.Query(`
SELECT
  c.chatname,
  c.chatid
FROM
  subscriptions su
  JOIN chats c ON c.chatid = su.chatid
WHERE
  su.subscription_userid = $1;`, streamUserID)
		if err != nil {
			log.Printf("Could not query database: %s", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var chat database.Chat
			err := rows.Scan(&chat.ChatName, &chat.ChatID)
			if err != nil {
				log.Printf("Could not scan row: %s", err)
				return
			}
			subscribers[chat.ChatName] = []string{}
		}

		// Get subscribed users
		rows, err = state.DB.Query(`
SELECT
  c.chatname,
  s.subscriber_username
FROM
  subscribers s
  JOIN chats c ON c.chatid = s.chatid
  JOIN subscriptions su ON su.subscription_id = s.subscription_id
WHERE
  su.subscription_userid = $1;`, streamUserID)
		if err != nil {
			log.Printf("Could not query database: %s", err)
			return
		}
		for rows.Next() {
			var (
				chatName string
				username string
			)
			err := rows.Scan(&chatName, &username)
			if err != nil {
				log.Printf("Could not scan row: %s", err)
				return
			}
			subscribers[chatName] = append(subscribers[chatName], username)
		}

		liveMessage := fmt.Sprintf("https://twitch.tv/%s just went live!", event.BroadcasterUserLogin)
		for chat, users := range subscribers {
			for _, message := range utils.SplitStreamOnlineMessage(liveMessage, users, 450) {
				state.IRC.Say(chat, message)
			}
		}
	}
}

func loadSubscriptions(s *models.State) error {
	var databaseIDs []int
	rows, err := s.DB.Query("SELECT subscription_userid FROM subscriptions GROUP BY subscription_userid")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return err
		}
		databaseIDs = append(databaseIDs, id)
	}

	var activeIDs []int
	subscriptions, err := s.TwitchWH.GetSubscriptionsByStatus("enabled")
	if err != nil {
		return err
	}
	for _, sub := range subscriptions {
		if sub.Type != "stream.online" {
			continue
		}
		id, err := strconv.Atoi(sub.Condition.BroadcasterUserID)
		if err != nil {
			return err
		}
		activeIDs = append(activeIDs, id)
	}

	for _, id := range databaseIDs {
		if !slices.Contains(activeIDs, id) {
			go func() {
				log.Printf("Creating subscription for %d", id)
				err := s.TwitchWH.AddSubscription("stream.online", "1", twitchwh.Condition{
					BroadcasterUserID: fmt.Sprint(id),
				})
				if err != nil {
					log.Printf("Could not create subscription: %s", err)
				}
			}()
		}
	}
	return nil
}

func main() {
	// TODO: Log to both stdout and a log file
	log.SetPrefix("Bot: ")
	log.SetFlags(log.Ltime | log.Lshortfile)
	startedAt := time.Now()

	log.Println("Loading config file")
	var config models.Config
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		log.Fatalf("Could not read and parse config file: %s", err)
	}

	log.Println("Opening sqlite database")
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_fk=true", config.DatabasePath))
	if err != nil {
		log.Fatalf("Could not open sqlite database: %s", err)
	}
	err = database.CreateTables(db)
	if err != nil {
		log.Fatalf("Could not create required SQL tables: %s", err)
	}

	log.Println("Creating Helix client")
	helix := helix.Client{
		ClientID:    config.Identity.ClientID,
		HelixURL:    "https://api.twitch.tv/helix",
		HttpClient:  &http.Client{},
		Token:       config.Identity.HelixToken,
		UserIDCache: make(map[string]int),
	}
	valid, err := helix.ValidateToken()
	if err != nil {
		log.Fatalf("Could not validate Helix token: %s", err)
	}
	if !valid {
		log.Fatalf("Helix token invalid")
	}

	ircClient := irc.NewClient(
		config.Identity.BotUsername,
		fmt.Sprintf("oauth:%s", config.Identity.HelixToken),
	)

	// Get chats from database
	rows, err := db.Query("SELECT chatname FROM chats GROUP BY chatid")
	if err != nil {
		log.Fatalf("Could not get chats from database: %s", err)
	}
	var chats []string
	for rows.Next() {
		var chat string
		err := rows.Scan(&chat)
		if err != nil {
			log.Fatalf("Could not scan row: %s", err)
		}
		chats = append(chats, chat)
	}
	if len(chats) > 0 {
		log.Printf("Found %d channels in database. Joining...", len(chats))
		ircClient.Join(chats...)
	} else {
		// Init chat from config
		log.Println("No chats found in database, checking config file")
		if config.InitialChannel == "" {
			log.Fatal("No channels found in database or config")
		}
		id, err := helix.LoginToID(config.InitialChannel)
		if err != nil {
			log.Fatalf("Could not get ID of user: %s", err)
		}
		_, err = db.Exec("INSERT INTO chats (chatname, chatid) VALUES ($1, $2)", config.InitialChannel, id)
		if err != nil {
			log.Fatalf("Could not insert to database: %s", err)
		}
		ircClient.Join(config.InitialChannel)
	}

	log.Println("Creating twitchwh client")
	whClient, err := twitchwh.New(twitchwh.ClientConfig{
		ClientID:      config.Identity.ClientID,
		ClientSecret:  config.Identity.ClientSecret,
		WebhookSecret: config.Eventsub.WebhookSecret,
		WebhookURL:    config.Eventsub.WebhookURL,
		Debug:         true,
	})
	if err != nil {
		log.Fatalf("Could not create twitchwh client: %s", err)
	}

	log.Println("Starting web server")
	router, err := web.New(config.BindAddr)
	if err != nil {
		log.Fatalf("Could not create web server: %s", err)
	}
	router.HandleFunc("POST /eventsub", whClient.Handler)
	server := &http.Server{
		Addr:    config.BindAddr,
		Handler: web.Logging(router),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Could not start http server: %s", err)
		}
	}()

	state := models.State{
		Config:    config,
		DB:        db,
		Helix:     &helix,
		IRC:       ircClient,
		StartedAt: &startedAt,
		TwitchWH:  whClient,
	}

	ircClient.OnPrivateMessage(onMessage(&state))
	ircClient.OnConnect(func() { log.Println("Connected to chat") })

	whClient.OnStreamOnline = onLive(&state)
	err = loadSubscriptions(&state)
	if err != nil {
		log.Fatalf("Could not load eventsub subscriptions: %s", err)
	}

	if err := ircClient.Connect(); err != nil {
		log.Fatalf("Twitch chat connection failed: %s", err)
	}
}
