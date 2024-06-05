package main

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
	Heartbeat_Interval int `json:"heartbeat_interval,omitempty"`
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

// ---------- Replace Session payload --------------

type GatewayReplaceSessionPayload struct {
	GatewayEvent
	Data []GatewayReplaceSessionData `json:"d"`
}

type GatewayReplaceSessionData struct {
	SessionID string `json:"session_id"`
	// More info about client but seems broken
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

// rsume
// {
// 	"op": 6,
// 	"d": {
// 	  "token": "my_token",
// 	  "session_id": "session_id_i_stored",
// 	  "seq": 1337
// 	}
//   }
