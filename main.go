package main

import (
	"encoding/json"
	"errors"
	"log"
	"mechfeed/channels"
	"mechfeed/discord-portal"
	"mechfeed/reddit-portal"
	"os"

	"github.com/joho/godotenv"
)

var DISCORD_CHANNELS = make(map[string]Channel) // Discord channels indexed by channel ID
var DISCORD_SERVERS = make(map[string]Server) // Discord servers indexed by channel ID
var DISCORD_WEBHOOK_URL string

func load_config() error{

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

	go discordportal.Listen()
	go redditportal.Monitor()

	for {
		select {
		case discord_msg := <-channels.DiscordChannel:
			go discord_handler(discord_msg)

		case reddit_msg := <-channels.RedditChannel:
			go reddit_handler(reddit_msg)
		}

	}
}

func discord_handler(msg channels.DiscordMessage) error {
	msg_channel, ok := DISCORD_CHANNELS[msg.ChannelID]
	if !ok {
		return nil // Channel not being monitored
	}
	msg_server := DISCORD_SERVERS[msg.ChannelID]

	log.Printf("Server: %s, Channel: %s ", msg_server.Name, msg_channel.Name)
	PrettyPrint(msg)
	return nil
}

func reddit_handler(msg channels.RedditMessage) {
	PrettyPrint(msg)
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
