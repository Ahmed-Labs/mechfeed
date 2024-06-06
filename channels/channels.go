package channels

var (
	DiscordChannel = make(chan DiscordMessage)
	RedditChannel  = make(chan RedditMessage)
)

type DiscordMessage struct {
	ID        string                     `json:"id"`
	Content   string                     `json:"content"`
	GuildID   string                     `json:"guild_id"`
	ChannelID string                     `json:"channel_id"`
	Timestamp string                     `json:"timestamp"`
	Author    DiscordMessageAuthor `json:"author"`
}

type DiscordMessageAuthor struct {
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Discriminator string `json:"discriminator"`
	ID            string `json:"id"`
}

type RedditMessage struct {
	ID        string
	Title     string
	URL       string
	Author    string
	Category  string
	Imgur     string
	Thumbnail string
	Content   string
}