package channels

var (
	DiscordChannel = make(chan DiscordMessage)
	RedditChannel  = make(chan RedditMessage)
)
