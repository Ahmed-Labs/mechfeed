package discordportal

import "github.com/gorilla/websocket"

const (
	GATEWAY_URL          = "wss://gateway.discord.gg/?v=9&encoding=json"
	GUILD_MESSAGE_INTENT = 33280 // GUILD_MESSAGES + MESSAGE_CONTENT
)

type GatewayConnection struct {
	token              string
	is_connected       bool
	is_identified      bool
	heartbeat_interval int
	conn               *websocket.Conn
	resume_gateway_url string
	session_id         string
	sequence           int
}

type GatewayEvent struct {
	Name     string `json:"t,omitempty"`
	OP       int    `json:"op,omitempty"`
	Sequence *int   `json:"s,omitempty"`
}

type GatewayPayloadData struct {
	Heartbeat_Interval int              `json:"heartbeat_interval,omitempty"`
	Token              string           `json:"token,omitempty"`
	Properties         DeviceProperties `json:"properties,omitempty"`
	ResumeURL          string           `json:"resume_gateway_url,omitempty"`
	Content            string           `json:"content,omitempty"`
	SessionID          string           `json:"session_id,omitempty"`
}

type GatewayHeartbeat struct {
	GatewayEvent
	Sequence int `json:"d"`
}

type GatewayHelloPayload struct {
	GatewayEvent
	Data GatewayHelloData `json:"d,omitempty"`
}

type GatewayHelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval,omitempty"`
}

// ---------- Identify payload --------------

type GatewayIdentifyPayload struct {
	GatewayEvent
	Data GatewayIdentifyData `json:"d,omitempty"`
}

type GatewayIdentifyData struct {
	Token      string           `json:"token"`
	Intents    int              `json:"intents"`
	Properties DeviceProperties `json:"properties"`
}
type DeviceProperties struct {
	OS      string `json:"os"`
	Browser string `json:"broswer"`
	Device  string `json:"device"`
}

// ---------- Ready payload --------------

type GatewayReadyPayload struct {
	GatewayEvent
	Data GatewayReadyData `json:"d"`
}

type GatewayReadyData struct {
	ResumeURL string `json:"resume_gateway_url"`
	SessionID string `json:"session_id"`
}

// ---------- Resume payload --------------

type GatewayResumePayload struct {
	GatewayEvent
	Data GatewayResumeData `json:"d"`
}
type GatewayResumeData struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Sequence  int    `json:"seq"`
}

// ---------- Message Create payload --------------

type GatewayMessageCreatePayload struct {
	GatewayEvent
	Data GatewayMessageCreateData `json:"d"`
}

type GatewayMessageCreateData struct {
	ID        string                     `json:"id"`
	Content   string                     `json:"content"`
	GuildID   string                     `json:"guild_id"`
	ChannelID string                     `json:"channel_id"`
	Timestamp string                     `json:"timestamp"`
	Author    GatewayMessageCreateAuthor `json:"author"`
}

type GatewayMessageCreateAuthor struct {
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Discriminator string `json:"discriminator"`
	ID            string `json:"id"`
}
