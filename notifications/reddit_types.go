package notifications

type ProcessedRedditPost struct {
	ID        string
	Title     string
	URL       string
	Author    string
	Category  string
	Imgur     string
	Thumbnail string
	Content   string
}
