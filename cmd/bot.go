package main

import (
	"bot/internal/database"
	"bot/internal/handler"
	"bot/internal/helix"
	httpclient "bot/internal/http"
	"bot/internal/models"
	"bot/web"
	"context"
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
	"github.com/jackc/pgx/v5/pgxpool"
)

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

	log.Println("Creating PostgreSQL pool")
	db, err := loadDB(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Could not load DB: %s", err)
	}

	log.Println("Creating HTTP client")
	httpClient := httpclient.Client{
		Client: &http.Client{},
		DefaultHeaders: map[string]string{
			"User-Agent": "LinneB/bot (https://github.com/LinneB/bot)",
		},
		URLHeaders: map[string]map[string]string{
			"api.twitch.tv": {
				"Client-ID":     config.Identity.ClientID,
				"Authorization": fmt.Sprintf("Bearer %s", config.Identity.HelixToken),
			},
			"id.twitch.tv": {
				"Authorization": fmt.Sprintf("Bearer %s", config.Identity.HelixToken),
			},
		},
	}

	log.Println("Validating Helix token")
	valid, err := helix.ValidateToken(httpClient)
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
	chats, err := database.GetChats(db)
	if err != nil {
		log.Fatalf("Could not get chats from database: %s", err)
	}
	if len(chats) > 0 {
		log.Printf("Found %d chat(s) in database. Joining...", len(chats))
		// Go please give me slices.Map ;)
		chatNames := make([]string, len(chats))
		for _, chat := range chats {
			chatNames = append(chatNames, chat.ChatName)
		}
		ircClient.Join(chatNames...)
	} else {
		// Load channel from config file
		log.Println("No chats found in database, checking config file")
		if config.InitialChannel == "" {
			log.Fatal("No channels found in database or config")
		}
		chat := strings.ToLower(config.InitialChannel)
		id, found, err := helix.LoginToID(httpClient, chat)
		if err != nil {
			log.Fatalf("Could not get ID of user: %s", err)
		}
		if !found {
			log.Fatalf("Could not find user %s", chat)
		}
		err = database.InsertChat(db, models.Chat{
			ChatID:   id,
			ChatName: chat,
		})
		if err != nil {
			log.Fatalf("Could not insert to database: %s", err)
		}
		ircClient.Join(chat)
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

	state := models.State{
		Config:    config,
		DB:        db,
		Http:      httpClient,
		IRC:       ircClient,
		StartedAt: startedAt,
		TwitchWH:  whClient,
	}

	ircClient.OnPrivateMessage(handler.OnMessage(&state))
	ircClient.OnConnect(func() { log.Println("Connected to chat") })

	whClient.OnStreamOnline = handler.OnLive(&state)
	err = loadSubscriptions(&state)
	if err != nil {
		log.Fatalf("Could not load eventsub subscriptions: %s", err)
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

	if err := ircClient.Connect(); err != nil {
		log.Fatalf("Twitch chat connection failed: %s", err)
	}
}

func loadDB(connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to database: %w", err)
	}
	err = database.CreateTables(pool)
	if err != nil {
		return nil, fmt.Errorf("Could not create required tables: %w", err)
	}
	return pool, nil
}

func loadSubscriptions(s *models.State) error {
	var databaseIDs []int
	rows, err := s.DB.Query(context.Background(), "SELECT subscription_userid FROM subscriptions GROUP BY subscription_userid")
	if err != nil {
		return fmt.Errorf("Could not query database: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return fmt.Errorf("Could not scan row: %w", err)
		}
		databaseIDs = append(databaseIDs, id)
	}

	var activeIDs []int
	subscriptions, err := s.TwitchWH.GetSubscriptionsByStatus("enabled")
	if err != nil {
		return fmt.Errorf("Could not get subscriptions: %w", err)
	}
	for _, sub := range subscriptions {
		if sub.Type != "stream.online" {
			continue
		}
		id, err := strconv.Atoi(sub.Condition.BroadcasterUserID)
		if err != nil {
			return fmt.Errorf("Condition ID \"%s\" is not a number: %w", sub.Condition.BroadcasterUserID, err)
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

func getChatsFromDatabase(db *pgxpool.Pool) ([]string, error) {
	rows, err := db.Query(context.Background(), "SELECT chatname FROM chats GROUP BY chatid")
	if err != nil {
		return nil, fmt.Errorf("Could not query database: %w", err)
	}
	var chats []string
	for rows.Next() {
		var chat string
		err := rows.Scan(&chat)
		if err != nil {
			return nil, fmt.Errorf("Could not scan row: %w", err)
		}
		chats = append(chats, chat)
	}
	return chats, nil
}
