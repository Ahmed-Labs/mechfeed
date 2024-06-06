package redditportal

type RedditResponse struct {
	Data struct {
		Children []struct {
			Data RawRedditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type RawRedditPost struct {
	ID            string  `json:"id"`
	Author        string  `json:"author"`
	URL           string  `json:"url"`
	Created       float64 `json:"created"`
	Title         string  `json:"title"`
	LinkFlairText string  `json:"link_flair_text"`
	HTMLText      string  `json:"selftext_html"`
	Content       string  `json:"selftext"`
}

type imgurAlbumResponse struct {
	Media []struct {
		URL string `json:"url"`
	} `json:"media"`
}