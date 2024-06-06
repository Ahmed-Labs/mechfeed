package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mechfeed/channels"
	"mechfeed/discord-portal"
	"mechfeed/reddit-portal"
)

func main() {
	go discordportal.Listen()
	go redditportal.Monitor()

	for {
		select {
		case discord_msg := <-channels.DiscordChannel:
			// log.Printf("Discord: %+v", discord_msg)
			log.Printf("[ DISCORD ]: ")
			PrettyPrint(discord_msg)

		case reddit_msg := <-channels.RedditChannel:
			// log.Printf("Reddit: %+v", reddit_msg)
			log.Printf("[ REDDIT ]: ")
			PrettyPrint(reddit_msg)
		}

	}
}

func PrettyPrint(data interface{}) {
	var p []byte
	//    var err := error
	p, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}
