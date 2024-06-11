package notifications

import (
	"bytes"
	"encoding/json"
	"mechfeed/channels"
	"net/http"
	"time"
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

func CreateNotificationDiscord(server, channl, alert string, data channels.DiscordMessage) DiscordNoti {
	return DiscordNoti{
		Content:  nil,
		Embeds: []Embed{
			{
				Color: 5727730,
				Fields: []Field{
					{Name: "Server", Value: server, Inline: true},
					{Name: "Channel", Value: channl, Inline: true},
					{Name: "Sent by", Value: "<@!" + data.Author.ID + ">", Inline: true},
					{Name: "Matched alert", Value: alert, Inline: true},
					{Name: "Message", Value: data.Content},
				},
				Footer:    Footer{Text: "mechfeed"},
				Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			},
		},
		Username: "mechfeed",
	}
}
