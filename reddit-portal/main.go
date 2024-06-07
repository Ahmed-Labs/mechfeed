package redditportal

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mechfeed/channels"
	"mechfeed/fetch-errors"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	REDDIT_SECRET string
	DEBUG bool
)

const (
	REDDIT_POST_ENDPOINT = "https://www.reddit.com/r/mechmarket/new.json"
	IMGUR_ALBUM_ENDPOINT = "https://api.imgur.com/post/v1/albums/"
)

func initApp() error {
	godotenv.Load()
	DEBUG = os.Getenv("DEBUG_REDDIT_PORTAL") == "true"
	REDDIT_SECRET = os.Getenv("REDDIT_SECRET")

	if REDDIT_SECRET == "" {
		return errors.New("no reddit secret found")
	}

	return nil
}

func Monitor() {
	if err := initApp(); err != nil {
		log.Fatal(err.Error())
	}

	var currID string

	for {
		time.Sleep(2 * time.Second)
		log.Println("Monitoring...")
		var res RedditResponse

		if err := getLatest(&res); err != nil {
			log.Println(err)
			continue
		}
		latestID := res.Data.Children[0].Data.ID

		if currID == "" {
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
				process_reddit_post(post)
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
	req, err := http.NewRequest("GET", REDDIT_POST_ENDPOINT, nil)
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
	if DEBUG {
		extract_pretified_json(bodyText)
	}
	if err := json.Unmarshal(bodyText, result); err != nil {
		return err
	}
	return nil
}

func process_reddit_post(post RawRedditPost) {
	imgurLinks := extract_imgur_links(post.HTMLText)
	var imgurAlbumLink string = "No Imgur link found"
	var thumbnailLink string

	if len(imgurLinks) > 0 {
		imgurAlbumLink = imgurLinks[0]
		splitLinkDash := strings.Split(imgurLinks[0], "/")

		if strings.Contains(splitLinkDash[len(splitLinkDash)-1], ".") {
			thumbnailLink = imgurLinks[0]
		} else {
			imgurAlbumID := splitLinkDash[len(splitLinkDash)-1]
			albumImages := get_imgur_thumbnail(imgurAlbumID)
			if len(albumImages) > 0 {
				thumbnailLink = albumImages[0]
			}
		}
	}

	channels.RedditChannel <- channels.RedditMessage{
		ID:        post.ID,
		Title:     post.Title,
		URL:       post.URL,
		Author:    post.Author,
		Category:  post.LinkFlairText,
		Imgur:     imgurAlbumLink,
		Thumbnail: thumbnailLink,
		Content:   post.Content,
	}
}

func extract_imgur_links(postBody string) []string {
	regexPattern := `href="([^"]*imgur[^"]*)"`
	regex, _ := regexp.Compile(regexPattern)
	matches := regex.FindAllStringSubmatch(postBody, -1)
	var imgurLinks []string

	for _, match := range matches {
		imgurLinks = append(imgurLinks, match[1])
	}

	return imgurLinks
}

func get_imgur_thumbnail(imgurAlbumID string) []string {
	client := &http.Client{}
	reqURL := IMGUR_ALBUM_ENDPOINT + imgurAlbumID + "?client_id=546c25a59c58ad7&include=media%2Cadconfig%2Caccount"
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

	if err := json.Unmarshal(bodyText, &albumImages); err != nil {
		log.Fatal(err)
	}

	var albumImageURLs []string
	for i := 0; i < len(albumImages.Media); i++ {
		albumImageURLs = append(albumImageURLs, albumImages.Media[i].URL)
	}
	return albumImageURLs
}

func extract_pretified_json(bodyText []byte) {
	var jsonData interface{}

	if err := json.Unmarshal(bodyText, &jsonData); err != nil {
		log.Fatal(err)
	}

	file, _ := json.MarshalIndent(jsonData, "", " ")

	_ = os.WriteFile("test.json", file, 0644)
}
