package notifications

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)



func SendWebhook(webhookURL string, message DiscordNoti) error {

	jsonPayload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func CreateNotification(postData ProcessedRedditPost) DiscordNoti {
	return DiscordNoti{
		Content: nil,
		Embeds: []Embed{
			{
				Title:     postData.Title,
				URL:       postData.URL,
				Color:     16734296,
				Fields: []Field{
					{Name: "Posted by", Value: "u/"+postData.Author+" [[PM]](https://www.reddit.com/message/compose/?to=" + postData.Author + ")", Inline: true},
					{Name: "Category", Value: postData.Category, Inline: true},
					{Name: "Imgur Link", Value: postData.Imgur},
				},
				Footer:    Footer{Text: "mechfeed"},
				Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
				Image:     Image{URL: postData.Thumbnail},
			},
		},
		Username:   "mechfeed",
	}
}
