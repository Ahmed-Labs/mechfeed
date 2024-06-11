package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"mechfeed/channels"
	"mechfeed/discord-portal"
	"mechfeed/filter"
	"mechfeed/notifications"
	"mechfeed/reddit-portal"
	"mechfeed/users"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	DISCORD_CHANNELS    = make(map[string]Channel) // Discord channels indexed by channel ID
	DISCORD_SERVERS     = make(map[string]Server)  // Discord servers indexed by channel ID
	DISCORD_WEBHOOK_URL string
	POSTGRES_CONNECTION string
)

type Repository struct {
	db          *sql.DB
	ctx         context.Context
	queries     *users.Queries
	alerts_pool []users.UserAlert
	users_pool  []users.User
}

func load_config() error {
	godotenv.Load()
	DISCORD_WEBHOOK_URL = os.Getenv("DISCORD_WEBHOOK")
	if DISCORD_WEBHOOK_URL == "" {
		return errors.New("no discord weebhook found")
	}

	POSTGRES_CONNECTION = os.Getenv("POSTGRES_CONNECTION")
	if POSTGRES_CONNECTION == "" {
		return errors.New("no postgres connection string found")
	}

	for _, server := range ServerList {
		for _, channel := range server.Channels {
			DISCORD_CHANNELS[channel.ID] = channel
			DISCORD_SERVERS[channel.ID] = server
		}
	}
	return nil
}

func init_db() (*Repository, error) {
	db, err := sql.Open("postgres", POSTGRES_CONNECTION)

	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	log.Println("Successfully connected to database")

	return &Repository{
		db:      db,
		ctx:     context.Background(),
		queries: users.New(db),
	}, nil
}

func main() {
	if err := load_config(); err != nil {
		log.Fatal(err)
	}
    repo, err := init_db()
    if err != nil {
        log.Fatal(err)
    }
    defer repo.db.Close() 

	go portal_supervisor(discordportal.Listen, "discordportal", time.Millisecond*300)
	go portal_supervisor(redditportal.Monitor, "redditportal", time.Millisecond*300)

	for {
		select {
		case discord_msg := <-channels.DiscordChannel:
			go repo.discord_handler(discord_msg)

		case reddit_msg := <-channels.RedditChannel:
			go repo.reddit_handler(reddit_msg)
		}

	}
}

func portal_supervisor(portal func(), name string, restartDelay time.Duration) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[ %s ] Crashed with error: %v. Restarting...\n", name, r)
					time.Sleep(restartDelay)
				}
			}()
			portal()
		}()
	}
}

func (r *Repository) discord_handler(msg channels.DiscordMessage) {
	msg_channel, ok := DISCORD_CHANNELS[msg.ChannelID]
	if !ok {
		return // Channel not being monitored
	}
	msg_server := DISCORD_SERVERS[msg.ChannelID]

	alerts, err := r.get_grouped_alerts()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("grouped alerts ", alerts)

	for keyword, user_ids := range alerts {
		if filter.FilterKeywords(msg.Content, keyword) {
			for user_id := range user_ids {
				user, err := r.queries.GetUser(r.ctx, user_id)
				if err != nil {
					log.Println("failed to fetch user: ", user_id, " , error: ", err)
					continue
				}
				if user.WebhookUrl.Valid {
					log.Println("Notifying user through webhook: ", user.WebhookUrl)
					notifications.SendWebhook(user.WebhookUrl.String, notifications.CreateNotificationDiscord(msg_server.Name, msg_channel.Name, keyword, msg))
				} else {
					log.Println("Webhook URL invalid: ", user.WebhookUrl)
				}
			}
			log.Printf("Server: %s, Channel: %s ", msg_server.Name, msg_channel.Name)
			PrettyPrint(msg)
		}
	}
}

func (r *Repository) reddit_handler(msg channels.RedditMessage) {
	alerts, err := r.get_grouped_alerts()
	if err != nil {
		log.Println(err)
		return
	}

	for keyword, user_ids := range alerts {
		if filter.FilterKeywords(msg.Content, keyword) {
			for user_id := range user_ids {
				user, err := r.queries.GetUser(r.ctx, user_id)
				if err != nil {
					log.Println("failed to fetch user: ", user_id, " , error: ", err)
					continue
				}
				if user.WebhookUrl.Valid {
					log.Println("Notifying user through webhook: ", user.WebhookUrl)
					notifications.SendWebhook(user.WebhookUrl.String, notifications.CreateNotificationReddit(msg))
				} else {
					log.Println("Webhook URL invalid: ", user.WebhookUrl)
				}
			}
			PrettyPrint(msg)
		}
	}
}

func (r *Repository) get_grouped_alerts() (map[string]map[string]bool, error) {
	alerts, err := r.queries.GetAlerts(r.ctx)

	if err != nil {
		return nil, err
	}
	var group_alerts = make(map[string]map[string]bool)

	for _, alert := range alerts {
		split_keywords := strings.Split(alert.Keyword, ",")
		sort.Strings(split_keywords)
		joined_keywords := strings.Join(split_keywords, ",")

		if _, ok := group_alerts[joined_keywords]; !ok {
			group_alerts[joined_keywords] = make(map[string]bool)
		}
		group_alerts[joined_keywords][alert.ID] = true
	}
	return group_alerts, nil
}

func PrettyPrint(data interface{}) {
	var p []byte

	p, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("%s \n", p)
}
