package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mechfeed/channels"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

func SendWebhook(webhookURL string, message interface{}) error {
	json_payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(json_payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}

func CreateNotificationReddit(data channels.RedditMessage) DiscordNoti {
	return DiscordNoti{
		Content: nil,
		Embeds: []Embed{
			{
				Title: data.Title,
				URL:   data.URL,
				Color: 16734296,
				Fields: []Field{
					{Name: "Posted by", Value: "u/" + data.Author + " [[PM]](https://www.reddit.com/message/compose/?to=" + data.Author + ")", Inline: true},
					{Name: "Category", Value: data.Category, Inline: true},
					{Name: "Imgur Link", Value: data.Imgur},
				},
				Footer:    Footer{Text: "mechfeed"},
				Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
				Image:     Image{URL: data.Thumbnail},
			},
		},
		Username: "mechfeed",
	}
}

func CreateNotificationDiscord(server, channel, alert string, data channels.DiscordMessage) DiscordNoti {
	return DiscordNoti{
		Content:  nil,
		Embeds: []Embed{
			{
				Color: 5727730,
				Fields: []Field{
					{Name: "Server", Value: server, Inline: true},
					{Name: "Channel", Value: "#" + channel, Inline: true},
					{Name: "Sent by", Value: data.Author.GlobalName + " (" + data.Author.Username + ")", Inline: true},
					{Name: "Jump to message", Value: fmt.Sprintf("https://discord.com/channels/%s/%s/%s", data.GuildID, data.ChannelID, data.ID)},
					{Name: "Matched alert", Value: fmt.Sprintf("`%s`", alert), Inline: true},
					{Name: "Message", Value: data.Content},
				},
				Footer:    Footer{Text: "mechfeed"},
				Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			},
		},
		Username: "mechfeed",
	}
}


func CreateRedditNotificationMessageEmbed(data channels.RedditMessage, alert string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       data.Title,
		URL:         data.URL,
		Color:       0xe671dc, // Color in decimal format
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Posted by",
				Value:  "u/" + data.Author + " [[PM]](https://www.reddit.com/message/compose/?to=" + data.Author + ")",
				Inline: true,
			},
			{
				Name:   "Category",
				Value:  data.Category,
				Inline: true,
			},
			{
				Name:   "Imgur Link",
				Value:  data.Imgur,
			},
			{
				Name:   "Matched alert",
				Value:  fmt.Sprintf("`%s`", alert),
			},
		},
		Footer:    &discordgo.MessageEmbedFooter{Text: "mechfeed"},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Image:     &discordgo.MessageEmbedImage{URL: data.Thumbnail},
	}
}

func CreateDiscordNotificationMessageEmbed(server, channel, alert string, data channels.DiscordMessage) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Color: 0xe671dc, // Color in decimal format
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Server",
				Value:  server,
				Inline: true,
			},
			{
				Name:   "Channel",
				Value:  "#" + channel,
				Inline: true,
			},
			{
				Name:   "Sent by",
				Value:  data.Author.GlobalName + " (" + data.Author.Username + ")",
				Inline: true,
			},
			{
				Name:   "Jump to message",
				Value:  fmt.Sprintf("https://discord.com/channels/%s/%s/%s", data.GuildID, data.ChannelID, data.ID),
			},
			{
				Name:   "Matched alert",
				Value:  fmt.Sprintf("`%s`", alert),
				Inline: true,
			},
			{
				Name:   "Message",
				Value:  data.Content,
			},
		},
		Footer:    &discordgo.MessageEmbedFooter{Text: "mechfeed"},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}