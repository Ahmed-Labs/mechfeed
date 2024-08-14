package main

import (
	"encoding/json"
	"errors"
	"log"
	"mechfeed/channels"
	"mechfeed/discord-portal"
	"mechfeed/filter"
	"mechfeed/notifications"
	"mechfeed/reddit-portal"
	"mechfeed/users"
	"mechfeed/bot"
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
)

func load_config() error {
	godotenv.Load()
	DISCORD_WEBHOOK_URL = os.Getenv("DISCORD_WEBHOOK")
	if DISCORD_WEBHOOK_URL == "" {
		return errors.New("no discord weebhook found")
	}

	for _, server := range ServerList {
		for _, channel := range server.Channels {
			DISCORD_CHANNELS[channel.ID] = channel
			DISCORD_SERVERS[channel.ID] = server
		}
	}
	return nil
}


func main() {
	if err := load_config(); err != nil {
		log.Fatal(err)
	}

	repo, err := users.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Db.Close()

	go bot.MechfeedBot()
	go portal_supervisor(discordportal.Listen, "discordportal", time.Millisecond*300)
	go portal_supervisor(redditportal.Monitor, "redditportal", time.Millisecond*300)

	for {
		select {
		case discord_msg := <-channels.DiscordChannel:
			go discord_handler(repo, discord_msg)

		case reddit_msg := <-channels.RedditChannel:
			go reddit_handler(repo, reddit_msg)
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

func discord_handler(r *users.Repository, msg channels.DiscordMessage) {
	msg_channel, ok := DISCORD_CHANNELS[msg.ChannelID]
	if !ok {
		return // Channel not being monitored
	}
	msg_server := DISCORD_SERVERS[msg.ChannelID]

	alerts, err := get_grouped_alerts(r)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("grouped alerts ", alerts)

	for keyword, user_ids := range alerts {
		if filter.FilterKeywords(msg.Content, keyword) {
			for user_id := range user_ids {
				go func(user_id, curr_keyword string) {
					user, err := r.Queries.GetUser(r.Ctx, user_id)
					if err != nil {
						log.Println("failed to fetch user: ", user_id, " , error: ", err)
						return
					}
					
					log.Println("Sending notification via DM to user :", user.Username)
					bot.SendEmbedDM(
						user_id, 
						notifications.CreateDiscordNotificationMessageEmbed(msg_server.Name, msg_channel.Name, curr_keyword, msg),
					)
					if user.WebhookUrl.Valid {
						log.Println("Notifying user through webhook: ", user.WebhookUrl)
						notifications.SendWebhook(
							user.WebhookUrl.String, 
							notifications.CreateNotificationDiscord(
								msg_server.Name, msg_channel.Name, curr_keyword, msg,
							),
						)
					} else {
						log.Println("user did not set Webhook URL.")
					}
				}(user_id, keyword)
			}
			log.Printf("Server: %s, Channel: %s ", msg_server.Name, msg_channel.Name)
			PrettyPrint(msg)
		}
	}
}

func reddit_handler(r *users.Repository, msg channels.RedditMessage) {
	alerts, err := get_grouped_alerts(r)
	if err != nil {
		log.Println(err)
		return
	}
	for keyword, user_ids := range alerts {
		if filter.FilterKeywords(msg.Content, keyword) {
			for user_id := range user_ids {
				go func(user_id string) {
					user, err := r.Queries.GetUser(r.Ctx, user_id)
					if err != nil {
						log.Println("failed to fetch user: ", user_id, " , error: ", err)
						return
					}
					log.Println("Sending notification via DM to user:", user.Username)
					bot.SendEmbedDM(
						user_id, 
						notifications.CreateRedditNotificationMessageEmbed(msg),
					)
					if user.WebhookUrl.Valid {
						log.Println("Notifying user through webhook: ", user.WebhookUrl)
						notifications.SendWebhook(user.WebhookUrl.String, notifications.CreateNotificationReddit(msg))
					} else {
						log.Println("Webhook URL invalid: ", user.WebhookUrl)
					}
				}(user_id)
			}
			PrettyPrint(msg)
		}
	}
}

func get_grouped_alerts(r *users.Repository) (map[string]map[string]bool, error) {
	alerts, err := r.Queries.GetAlerts(r.Ctx)

	if err != nil {
		return nil, err
	}
	var group_alerts = make(map[string]map[string]bool)

	for _, alert := range alerts {
		split_keywords := strings.Split(alert.Keyword, ",")
		sort.Strings(split_keywords)
		joined_keywords := strings.ReplaceAll(strings.ToLower(strings.Join(split_keywords, ",")), " ", "")

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
