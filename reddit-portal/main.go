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
const (
	REDDIT_AUTH_ENDPOINT = "https://www.reddit.com/api/v1/access_token"
	REDDIT_POST_ENDPOINT = "https://oauth.reddit.com/r/mechmarket/new.json"
	IMGUR_ALBUM_ENDPOINT = "https://api.imgur.com/post/v1/albums/"
)

var (
	DEBUG                bool
	REDDIT_CLIENT_ID     string
	REDDIT_CLIENT_SECRET string
	REDDIT_AUTH          RedditAuth
)

type RedditAuth struct {
	access_token string
	expires_at   time.Time
}

func init_app() error {
	godotenv.Load()
	DEBUG = os.Getenv("DEBUG_REDDIT_PORTAL") == "true"
	REDDIT_CLIENT_ID = os.Getenv("REDDIT_CLIENT_ID")
	REDDIT_CLIENT_SECRET = os.Getenv("REDDIT_CLIENT_SECRET")

	if REDDIT_CLIENT_ID == "" {
		return errors.New("no reddit client id found")
	}
	if REDDIT_CLIENT_SECRET == "" {
		return errors.New("no reddit client secret found")
	}
	var err error
	REDDIT_AUTH, err = get_reddit_auth()

	if err != nil {
		return err
	}
	return nil
}

func Monitor() {
	var err error
	if err = init_app(); err != nil {
		log.Fatal(err)
	}
	defer panic("exited redditportal")

	var currID string
	check_expiry := 300

	for {
		if check_expiry <= 0 && time.Now().After(REDDIT_AUTH.expires_at) {
			get_reddit_auth()
			log.Println("Refreshed reddit access token")
			check_expiry = 300
		}
		check_expiry--
		time.Sleep(2 * time.Second)
		var res RedditResponse

		if err := getLatest(&res); err != nil {
			log.Print(err.Error())
			continue
		}
		log.Println("Monitoring...")
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

func get_reddit_auth() (RedditAuth, error) {
	client := &http.Client{}
	auth_payload := strings.NewReader("grant_type=client_credentials")
	req, err := http.NewRequest("POST", REDDIT_AUTH_ENDPOINT, auth_payload)

	if err != nil {
		return RedditAuth{}, err
	}

	req.SetBasicAuth(REDDIT_CLIENT_ID, REDDIT_CLIENT_SECRET)
	req.Header.Set("User-Agent", "mechfeed/0.1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return RedditAuth{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return RedditAuth{}, err
	}

	var auth_info struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}

	json.Unmarshal(data, &auth_info)

	if auth_info.Error != "" {
		return RedditAuth{}, errors.New(auth_info.Error)
	}

	expiration_time := time.Now().Add(time.Duration(int(float64(auth_info.ExpiresIn)*0.9)) * time.Second)

	return RedditAuth{access_token: auth_info.AccessToken, expires_at: expiration_time}, nil
}

func getLatest(result *RedditResponse) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", REDDIT_POST_ENDPOINT, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "mechfeed/0.1")
	req.Header.Set("Authorization", "Bearer " + REDDIT_AUTH.access_token)

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
		log.Println(err)
		return []string{}
	}
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return []string{}
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	var albumImages imgurAlbumResponse

	if err := json.Unmarshal(bodyText, &albumImages); err != nil {
		log.Println(err)
		return []string{}
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
