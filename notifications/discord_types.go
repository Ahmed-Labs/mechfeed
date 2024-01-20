package notifications

type Embed struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Color     int    `json:"color"`
	Fields    []Field `json:"fields"`
	Footer    Footer  `json:"footer"`
	Timestamp string  `json:"timestamp"`
	Image     Image   `json:"image"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Footer struct {
	Text string `json:"text"`
}

type Image struct {
	URL string `json:"url"`
}

type DiscordNoti struct {
	Content    interface{} `json:"content"`
	Embeds     []Embed     `json:"embeds"`
	Username   string      `json:"username"`
	Attachments []string    `json:"attachments"`
}
