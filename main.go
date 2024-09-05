package main

import (
	"encoding/json"
	"log"
	"mechfeed/channels"
	"mechfeed/discord-portal"
	"mechfeed/filter"
	"mechfeed/notifications"
	"mechfeed/reddit-portal"
	"mechfeed/users"
	"mechfeed/bot"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	DISCORD_CHANNELS    = make(map[string]Channel) // Discord channels indexed by channel ID
	DISCORD_SERVERS     = make(map[string]Server)  // Discord servers indexed by channel ID
	DISCORD_WEBHOOK_URL string
	PUBLIC_MECHMARKET_WEBHOOK_URL string
)

func load_config() error {
	godotenv.Load()
	DISCORD_WEBHOOK_URL = os.Getenv("DISCORD_WEBHOOK")
	if DISCORD_WEBHOOK_URL == "" {
		log.Println("no discord webhook found")
	}

	PUBLIC_MECHMARKET_WEBHOOK_URL = os.Getenv("PUBLIC_MECHMARKET_WEBHOOK")
	if PUBLIC_MECHMARKET_WEBHOOK_URL == "" {
		log.Println("no webhook for mechmarket channel found")
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
			log.Println(discord_msg)
			go discord_handler(repo, discord_msg)

		case reddit_msg := <-channels.RedditChannel:
			log.Println(reddit_msg)
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

	alerts, err := r.Queries.GetAlerts(r.Ctx)
	if err != nil {
		log.Println(err)
		return
	}
	
	// log.Printf("Server: %s, Channel: %s ", msg_server.Name, msg_channel.Name)
	for _, a := range alerts {
		go func(alert users.UserAlert) {
			user, err := r.Queries.GetUser(r.Ctx, alert.ID)
			if err != nil {
				log.Println("failed to fetch user: ", alert.ID, " , error: ", err)
				return
			}

			if !filter.FilterKeywords(msg.Content, alert.Keyword) {
				return
			}

			for _, u := range alert.Ignored {
				if u == msg.Author.Username {
					log.Printf("Skipping alert... '%s' is ignored by %s", u, user.Username)
					return
				}
			}

			log.Println("Sending Discord notification via DM to user:", user.Username, "Keyword:", alert.Keyword, "Message:", msg)
			bot.IsolatedSendEmbedDM(
				user.ID, 
				notifications.CreateDiscordNotificationMessageEmbed(msg_server.Name, msg_channel.Name, alert.Keyword, msg),
			)
			if user.WebhookUrl.Valid {
				log.Println("Notifying user through webhook:", user.WebhookUrl)
				notifications.SendWebhook(
					user.WebhookUrl.String, 
					notifications.CreateNotificationDiscord(
						msg_server.Name, msg_channel.Name, alert.Keyword, msg,
					),
				)
			} else {
				log.Println("user did not set Webhook URL.")
			}
				
		}(a)
	}
}


func reddit_handler(r *users.Repository, msg channels.RedditMessage) {
	// Notify public mechmarket channel
	notifications.SendWebhook(PUBLIC_MECHMARKET_WEBHOOK_URL, notifications.CreateNotificationReddit(msg))

	// User alerts
	alerts, err := r.Queries.GetAlerts(r.Ctx)
	if err != nil {
		log.Println(err)
		return
	}

	for _, a := range alerts {
		go func(alert users.UserAlert) {
			user, err := r.Queries.GetUser(r.Ctx, alert.ID)
			if err != nil {
				log.Println("failed to fetch user: ", alert.ID, " , error: ", err)
				return
			}
			if !filter.FilterKeywords(msg.Content, alert.Keyword) {
				return
			}

			for _, u := range alert.Ignored {
				if u == msg.Author {
					log.Printf("Skipping alert... '%s' is ignored by %s", u, user.Username)
					return
				}
			}
			log.Println("Sending Reddit notification via DM to user:", user.Username, "Keyword:", alert.Keyword, "Message:", msg)
			bot.IsolatedSendEmbedDM(
				user.ID, 
				notifications.CreateRedditNotificationMessageEmbed(msg, alert.Keyword),
			)
			if user.WebhookUrl.Valid {
				log.Println("Notifying user through webhook: ", user.WebhookUrl)
				notifications.SendWebhook(user.WebhookUrl.String, notifications.CreateNotificationReddit(msg))
			} else {
				log.Println("Webhook URL invalid: ", user.WebhookUrl)
			}
		}(a)
	}
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
