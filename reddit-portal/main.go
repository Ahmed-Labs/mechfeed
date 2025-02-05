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
)

const (
	redditAuthEndpoint   = "https://www.reddit.com/api/v1/accessToken"
	redditPostEndpoint = "https://oauth.reddit.com/r/mechmarket/new.json"
	imgurAlbumEndpoint = "https://api.imgur.com/post/v1/albums/"
)

var (
	debug              bool
	redditClientID     string
	redditClientSecret string
	globalRedditAuth   RedditAuth
)

type RedditAuth struct {
	accessToken string
	expiresAt   time.Time
}

func initApp() error {
	debug = os.Getenv("DEBUG_REDDIT_PORTAL") == "true"
	redditClientID = os.Getenv("REDDIT_CLIENT_ID")
	redditClientSecret = os.Getenv("REDDIT_CLIENT_SECRET")

	if redditClientID == "" {
		return errors.New("no reddit client id found")
	}
	if redditClientSecret == "" {
		return errors.New("no reddit client secret found")
	}
	var err error
	globalRedditAuth, err = redditAuth()

	if err != nil {
		return err
	}
	return nil
}

func Monitor() {
	var err error
	if err = initApp(); err != nil {
		log.Fatal(err)
	}
	defer panic("exited redditportal")

	var currID string
	checkExpiry := 300

	for {
		if checkExpiry <= 0 && time.Now().After(globalRedditAuth.expiresAt) {
			globalRedditAuth, err = redditAuth()
			checkExpiry = 300
			if err != nil {
				log.Println("failed to refresh reddit access token")
			} else {
				log.Println("refreshed reddit access token")
			}
		}
		checkExpiry--
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

func redditAuth() (RedditAuth, error) {
	client := &http.Client{}
	authPayload := strings.NewReader("grant_type=client_credentials")
	req, err := http.NewRequest("POST", redditAuthEndpoint, authPayload)

	if err != nil {
		return RedditAuth{}, err
	}

	req.SetBasicAuth(redditClientID, redditClientSecret)
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

	// reddit's auth json response
	var authInfoResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}

	json.Unmarshal(data, &authInfoResponse)

	if authInfoResponse.Error != "" {
		return RedditAuth{}, errors.New(authInfoResponse.Error)
	}

	expirationTime := time.Now().Add(time.Duration(int(float64(authInfoResponse.ExpiresIn)*0.9)) * time.Second)

	return RedditAuth{accessToken: authInfoResponse.AccessToken, expiresAt: expirationTime}, nil
}

func getLatest(result *RedditResponse) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", redditPostEndpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "mechfeed/0.1")
	req.Header.Set("Authorization", "Bearer "+globalRedditAuth.accessToken)

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
	if debug {
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
	category := "No Category"
	if post.LinkFlairText != "" {
		category = post.LinkFlairText
	}

	channels.RedditChannel <- channels.RedditMessage{
		ID:        post.ID,
		Title:     post.Title,
		URL:       post.URL,
		Author:    post.Author,
		Category:  category,
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
	reqURL := imgurAlbumEndpoint + imgurAlbumID + "?client_id=546c25a59c58ad7&include=media%2Cadconfig%2Caccount"
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
