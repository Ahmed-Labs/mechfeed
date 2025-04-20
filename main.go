package main

import (
	"log"
	"mechfeed/bot"
	"mechfeed/channels"
	discordportal "mechfeed/discord-portal"
	"mechfeed/filter"
	"mechfeed/notifications"
	redditportal "mechfeed/reddit-portal"
	"mechfeed/users"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	discordChannels            = make(map[string]Channel) // Discord channels indexed by channel ID
	discordServers             = make(map[string]Server)  // Discord servers indexed by channel ID
	discordWebhookURL          string
	publicMechmarketWebhookURL string
	sentryDSN                  string
)

func loadConfig() error {
	godotenv.Load()
	discordWebhookURL = os.Getenv("DISCORD_WEBHOOK")
	publicMechmarketWebhookURL = os.Getenv("PUBLIC_MECHMARKET_WEBHOOK")
	sentryDSN = os.Getenv("SENTRY_DSN")

	for _, server := range ServerList {
		for _, channel := range server.Channels {
			discordChannels[channel.ID] = channel
			discordServers[channel.ID] = server
		}
	}
	return nil
}

func main() {
	// Load env
	if err := loadConfig(); err != nil {
		log.Fatal(err)
	}

	// Get user DB connection
	conn, err := users.DatabaseConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Db.Close()

	// Intiialize sentry
	err = sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSN,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)

	// Mechfeed client discord bot
	go bot.Run()

	// Wrapped goroutines for Discord & Reddit monitors
	go runPortal(discordportal.Run, "discordportal", time.Millisecond*300)
	go runPortal(redditportal.Run, "redditportal", time.Millisecond*300)

	for {
		select {
		case discord_msg := <-channels.DiscordChannel:
			go discordHandler(conn, discord_msg)

		case reddit_msg := <-channels.RedditChannel:
			go redditHandler(conn, reddit_msg)
		}
	}
}

func runPortal(portal func(), name string, restartDelay time.Duration) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[ %s ] Crashed with error: %v. Restarting...\n", name, r)

					switch r_asserted := r.(type) {
					case string:
						sentry.CaptureMessage(r_asserted)
					case error:
						sentry.CaptureException(r_asserted)
					}
					time.Sleep(restartDelay)
				}
			}()
			portal()
		}()
	}
}

func discordHandler(r *users.Connection, msg channels.DiscordMessage) {
	_, ok := discordChannels[msg.ChannelID]
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
			go notifyDiscordMessage(r, msg, alert)
		}

	}
}

func redditHandler(r *users.Connection, msg channels.RedditMessage) {
	// Notify public mechmarket channel
	if publicMechmarketWebhookURL != "" {
		notifications.SendWebhook(publicMechmarketWebhookURL, notifications.CreateNotificationReddit(msg))
	}

	// User alerts
	alerts, err := r.Queries.GetAlerts(r.Ctx)
	if err != nil {
		log.Println(err)
		return
	}

	for _, alert := range alerts {
		// Notify user if alert matches
		if filter.FilterKeywords(msg.Content, alert.Keyword) {
			go notifyRedditMessage(r, msg, alert)
		}
	}
}

func notifyDiscordMessage(r *users.Connection, msg channels.DiscordMessage, alert users.UserAlert) {
	msg_server := discordServers[msg.ChannelID]
	msg_channel := discordChannels[msg.ChannelID]

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

func notifyRedditMessage(r *users.Connection, msg channels.RedditMessage, alert users.UserAlert) {
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
