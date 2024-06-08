package main

import (
	"encoding/json"
	"errors"
	"log"
	"mechfeed/channels"
	"mechfeed/discord-portal"
	"mechfeed/filter"
	"mechfeed/reddit-portal"
	"os"
	"time"

	"github.com/joho/godotenv"
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
	go portal_supervisor(discordportal.Listen, "discordportal", time.Millisecond*300)
	go portal_supervisor(redditportal.Monitor, "redditportal", time.Millisecond*300)

	for {
		select {
		case discord_msg := <-channels.DiscordChannel:
			go discord_handler(discord_msg)

		case reddit_msg := <-channels.RedditChannel:
			go reddit_handler(reddit_msg)
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

func discord_handler(msg channels.DiscordMessage) error {
	msg_channel, ok := DISCORD_CHANNELS[msg.ChannelID]
	if !ok {
		return nil // Channel not being monitored
	}
	msg_server := DISCORD_SERVERS[msg.ChannelID]

	for _, keyword := range Keywords {
		if filter.FilterKeywords(msg.Content, keyword) {
			log.Printf("Server: %s, Channel: %s ", msg_server.Name, msg_channel.Name)
			PrettyPrint(msg)
		}
	}
	return nil
}

func reddit_handler(msg channels.RedditMessage) {
	for _, keyword := range Keywords {
		if filter.FilterKeywords(msg.Content, keyword) {
			PrettyPrint(msg)
		}
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
