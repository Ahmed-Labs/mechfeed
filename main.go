package main

import (
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
	// Load env
	if err := load_config(); err != nil {
		log.Fatal(err)
	}

	// Get user DB connection
	repo, err := users.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Db.Close()

	// Mechfeed client discord bot
	go bot.MechfeedBot()

	// Wrapped goroutines for Discord & Reddit monitors 
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
	_, ok := DISCORD_CHANNELS[msg.ChannelID]
	if !ok {
		return // Channel not being monitored
	}

	alerts, err := r.Queries.GetAlerts(r.Ctx)
	if err != nil {
		log.Println(err)
		return
	}
	
	for _, alert := range alerts {
		// Notify user if alert matches
		if filter.FilterKeywords(msg.Content, alert.Keyword) {
			go discord_notify(r, msg, alert)
		}
		
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

	for _, alert := range alerts {
		// Notify user if alert matches
		if filter.FilterKeywords(msg.Content, alert.Keyword) {
			go reddit_notify(r, msg, alert)
		}
	}
}

func discord_notify(r *users.Repository, msg channels.DiscordMessage, alert users.UserAlert) {
	msg_server := DISCORD_SERVERS[msg.ChannelID]
	msg_channel := DISCORD_CHANNELS[msg.ChannelID]

	// Get user that set alert
	user, err := r.Queries.GetUser(r.Ctx, alert.ID)
	if err != nil {
		log.Println("failed to fetch user: ", alert.ID, " , error: ", err)
		return
	}

	// Skip alert if message author is ignored
	for _, u := range alert.Ignored {
		if u == msg.Author.Username {
			log.Printf("Skipping alert... '%s' is ignored by %s", u, user.Username)
			return
		}
	}

	// Send DM notification
	log.Println("Sending Discord notification via DM to user:", user.Username, "Keyword:", alert.Keyword, "Message:", msg)
	bot.IsolatedSendEmbedDM(
		user.ID, 
		notifications.CreateDiscordNotificationMessageEmbed(msg_server.Name, msg_channel.Name, alert.Keyword, msg),
	)

	// Send webhook notification if user opted in
	if user.WebhookUrl.Valid {
		log.Println("Notifying user through webhook:", user.WebhookUrl)
		notifications.SendWebhook(
			user.WebhookUrl.String, 
			notifications.CreateNotificationDiscord(
				msg_server.Name, msg_channel.Name, alert.Keyword, msg,
			),
		)
	}
}


func reddit_notify(r *users.Repository, msg channels.RedditMessage, alert users.UserAlert) {
	// Get user that set alert
	user, err := r.Queries.GetUser(r.Ctx, alert.ID)
	if err != nil {
		log.Println("failed to fetch user: ", alert.ID, " , error: ", err)
		return
	}

	// Skip alert if message author is ignored
	for _, u := range alert.Ignored {
		if u == msg.Author {
			log.Printf("Skipping alert... '%s' is ignored by %s", u, user.Username)
			return
		}
	}

	// Send DM notification
	log.Println("Sending Reddit notification via DM to user:", user.Username, "Keyword:", alert.Keyword, "Message:", msg)
	bot.IsolatedSendEmbedDM(
		user.ID, 
		notifications.CreateRedditNotificationMessageEmbed(msg, alert.Keyword),
	)

	// Send webhook notification if user opted in
	if user.WebhookUrl.Valid {
		log.Println("Notifying user through webhook: ", user.WebhookUrl)
		notifications.SendWebhook(user.WebhookUrl.String, notifications.CreateNotificationReddit(msg))
	}
}