package main

import (
	"bot/database"
	"bot/handler"
	"bot/helix"
	httpclient "bot/http"
	"bot/models"
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
	db, err := loadDB(config.DatabasePath)
	if err != nil {
		log.Fatalf("Could not load DB: %s", err)
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
	chats, err := getChatsFromDatabase(db)
	if err != nil {
		log.Fatalf("Could not get chats from database: %s", err)
	}
	if len(chats) > 0 {
		log.Printf("Found %d chat(s) in database. Joining...", len(chats))
		ircClient.Join(chats...)
	} else {
		log.Println("No chats found in database, checking config file")
		if config.InitialChannel == "" {
			log.Fatal("No channels found in database or config")
		}
		chat := strings.ToLower(config.InitialChannel)
		id, err := helix.LoginToID(chat)
		if err != nil {
			log.Fatalf("Could not get ID of user: %s", err)
		}
		_, err = db.Exec("INSERT INTO chats (chatname, chatid) VALUES ($1, $2)", chat, id)
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

	ivr := httpclient.Client{
		Client:         &http.Client{},
		BaseURL:        "https://api.ivr.fi/v2",
		DefaultHeaders: make(map[string]string),
	}
	rustlog := httpclient.Client{
		Client:         &http.Client{},
		BaseURL:        "https://logs.ivr.fi",
		DefaultHeaders: make(map[string]string),
	}
	seventv := httpclient.Client{
		Client:         &http.Client{},
		BaseURL:        "https://7tv.io/v3",
		DefaultHeaders: make(map[string]string),
	}

	state := models.State{
		Config:    config,
		DB:        db,
		Helix:     helix,
		IRC:       ircClient,
		IVR:       ivr,
		Rustlog:   rustlog,
		SevenTV:   seventv,
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

func loadDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_fk=true", path))
	if err != nil {
		return nil, fmt.Errorf("Could not open sqlite database: %w", err)
	}
	err = database.CreateTables(db)
	if err != nil {
		return nil, fmt.Errorf("Could not create required tables: %w", err)
	}
	return db, nil
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

func getChatsFromDatabase(db *sql.DB) ([]string, error) {
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
	return chats, nil
}
