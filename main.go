package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mechfeed/errors"
	"mechfeed/notifications"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type RedditResponse struct {
	Data struct {
		Children []struct {
			Data RawRedditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type RawRedditPost struct {
	ID   string `json:"id"`
	Author          string `json:"author"`
	URL          string `json:"url"`
	Created         float64  `json:"created"`
	Title           string `json:"title"`
	LinkFlairText   string `json:"link_flair_text"`
}

func main() {
	godotenv.Load()
	webhook := os.Getenv("DISCORD_WEBHOOK")
	monitor(webhook)
}

func monitor(webhook string){
	started := false
	var currID string

	for {
		time.Sleep(2*time.Second)
		fmt.Printf("[%s] Monitoring...\n", time.Now().Format("2006-01-02 15:04:05"))
		var res RedditResponse;
	
		if err:= getLatest(&res); err != nil {
			continue
		}
		latestID := res.Data.Children[0].Data.ID
		if !started {
			started = true
			currID = latestID
			continue
		}
		if currID == latestID {
			continue;
		}
		postPivot := false
		for i:= len(res.Data.Children)-1; i >= 0; i-- {
			post := res.Data.Children[i].Data
			if postPivot{
				fmt.Printf("Author: %s\n", post.Author)
				fmt.Printf("Created: %f\n", post.Created)
				fmt.Printf("Title: %s\n", post.Title)
				fmt.Printf("Link Flair: %s\n", post.LinkFlairText)
				fmt.Println("------------")
				processedPost := notifications.CreateWebhook(notifications.ProcessedRedditPost{
					Title:     post.Title,
					URL:       post.URL,
					Author:    post.Author,
					Category:  post.LinkFlairText,
					Imgur:     "https://imgur.com/oVhfdG7",
					Thumbnail: "https://i.imgur.com/a3Cgynr.jpeg",
				})
				notifications.SendWebhook(webhook, processedPost)
			} else if post.ID == currID {
				postPivot = true
				continue
			}
		}
		currID = latestID
	}
}

func getLatest(result *RedditResponse) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.reddit.com/r/mechmarket/new.json", nil)
	if err != nil {
		log.Panic(err)
		return err
	}
	
	req.Header.Set("User-Agent", "myApp")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200{
		return errors.FetchError{
			Code: resp.StatusCode,
			Message: resp.Status,
		}
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	extractPrettifiedJson(bodyText)
	if err := json.Unmarshal(bodyText, result); err != nil {
		return err
	}
	return nil
}

func extractPrettifiedJson(bodyText []byte){
	var jsonData interface{}

    // Unmarshal JSON into the struct
    if err := json.Unmarshal(bodyText, &jsonData); err != nil {
        log.Fatal(err)
    }

	file, _ := json.MarshalIndent(jsonData, "", " ")
 
	_ = os.WriteFile("test.json", file, 0644)
}
