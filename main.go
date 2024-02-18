package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mechfeed/fetch-errors"
	"mechfeed/notifications"
	"net/http"
	"os"
	"regexp"
	"time"
	"strings"
	"errors"

	"github.com/joho/godotenv"
)

var DiscordWebhook string

func main() {
	err := initApp()
	if err != nil {
		fmt.Println(err)
		return
	}
	monitor()
}

func initApp() error {
	godotenv.Load()
	DiscordWebhook = os.Getenv("DISCORD_WEBHOOK")
	if DiscordWebhook == "" {
		return errors.New("No Discord Webhook Found")
	}
	return nil
}

func monitor() {
	started := false
	var currID string

	for {
		time.Sleep(2 * time.Second)
		fmt.Printf("[%s] Monitoring...\n", time.Now().Format("2006-01-02 15:04:05"))
		var res RedditResponse

		if err := getLatest(&res); err != nil {
			continue
		}
		latestID := res.Data.Children[0].Data.ID
		if !started {
			started = true
			currID = latestID
			continue
		}
		if currID == latestID {
			continue
		}
		postPivot := false
		for i := len(res.Data.Children) - 1; i >= 0; i-- {
			post := res.Data.Children[i].Data
			if postPivot {
				fmt.Printf("Author: %s\n", post.Author)
				fmt.Printf("Created: %f\n", post.Created)
				fmt.Printf("Title: %s\n", post.Title)
				fmt.Printf("Link Flair: %s\n", post.LinkFlairText)
				fmt.Println("------------")
				if err := sendWebhookNotification(post); err != nil {
					fmt.Println("Failed to send discord notification", err)
				}
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
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fetcherrors.FetchError{
			Code:    resp.StatusCode,
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

func sendWebhookNotification(post RawRedditPost) error {
	imgurLinks := extractImgurLinks(post.HTMLText)
	var imgurAlbumLink string = "No Imgur link found"
	var thumbnailLink string
	
	if len(imgurLinks) > 0 {
		imgurAlbumLink = imgurLinks[0]
		splitLinkDash := strings.Split(imgurLinks[0], "/")

		if strings.Contains(splitLinkDash[len(splitLinkDash)-1], ".") {
			thumbnailLink = imgurLinks[0]
		} else {
			imgurAlbumID := splitLinkDash[len(splitLinkDash)-1]
			albumImages := getImgurAlbumThumnail(imgurAlbumID)
			if len(albumImages) > 0{
				thumbnailLink = albumImages[0]
			}
		}
	}
	processedPost := notifications.CreateWebhook(notifications.ProcessedRedditPost{
		Title:     post.Title,
		URL:       post.URL,
		Author:    post.Author,
		Category:  post.LinkFlairText,
		Imgur:     imgurAlbumLink,
		Thumbnail: thumbnailLink,
	})
	return notifications.SendWebhook(DiscordWebhook, processedPost)
}

func extractImgurLinks(postBody string) []string {
	// regexPattern := `(^(http|https):\/\/)?(i\.)?imgur\.com\/(?:(?:gallery\/(?<galleryid>\w+))|(?:a\/(?<albumid>\w+))|#?(?<imgid>\w+))`
	regexPattern := `href="([^"]*imgur[^"]*)"`
	regex, _ := regexp.Compile(regexPattern)
	matches := regex.FindAllStringSubmatch(postBody, -1)
	var imgurLinks []string

	for _, match := range matches {
		imgurLinks = append(imgurLinks, match[1])
	}

	return imgurLinks
}

func getImgurAlbumThumnail(imgurAlbumID string) []string{
	client := &http.Client{}
	reqURL := "https://api.imgur.com/post/v1/albums/" + imgurAlbumID + "?client_id=546c25a59c58ad7&include=media%2Cadconfig%2Caccount"
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("authority", "api.imgur.com")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var albumImages imgurAlbumResponse

	if err:= json.Unmarshal(bodyText, &albumImages); err != nil {
		log.Fatal(err)
	}

	var albumImageURLs []string
	for i:=0; i < len(albumImages.Media); i++ {
		albumImageURLs = append(albumImageURLs, albumImages.Media[i].URL)
	}
	return albumImageURLs
}

func extractPrettifiedJson(bodyText []byte) {
	var jsonData interface{}

	if err := json.Unmarshal(bodyText, &jsonData); err != nil {
		log.Fatal(err)
	}

	file, _ := json.MarshalIndent(jsonData, "", " ")

	_ = os.WriteFile("test.json", file, 0644)
}
