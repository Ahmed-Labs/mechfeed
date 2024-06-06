package main

import (
	"encoding/json"
	"log"
	"mechfeed/channels"
	"mechfeed/discord-portal"
	"mechfeed/reddit-portal"
)

var DiscordConfig = make(map[string]Server)

func init_config(){
	for _, server := range ServerList {
		for _, channel := range server.Channels {
			DiscordConfig[channel.ID] = server
		}
	}
}
func main() {
	init_config()
	go discordportal.Listen()
	go redditportal.Monitor()

	for {
		select {
		case discord_msg := <-channels.DiscordChannel:
			// log.Printf("Discord: %+v", discord_msg)
			// log.Printf("[ DISCORD ]: ")
			discord_handler(discord_msg)

		case reddit_msg := <-channels.RedditChannel:
			// log.Printf("Reddit: %+v", reddit_msg)
			// log.Printf("[ REDDIT ]: ")
			PrettyPrint(reddit_msg)
		}

	}
}

func discord_handler(msg channels.DiscordMessage){
	_, exists := DiscordConfig[msg.ChannelID]
	if !exists {
		return
	}
	PrettyPrint(msg)
}

func PrettyPrint(data interface{}) {
	var p []byte
	//    var err := error
	p, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("%s \n", p)
}
