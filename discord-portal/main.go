package discordportal

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"mechfeed/channels"
)

var DEBUG bool

func initApp() (GatewayConnection, error) {
	err := godotenv.Load()

	if err != nil {
		return GatewayConnection{}, errors.New("error loading .env")
	}

	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		return GatewayConnection{}, errors.New("no discord token found")
	}

	DEBUG = os.Getenv("DEBUG_DISCORD_PORTAL") == "true"

	// Establish connection with gateway
	c, _, err := websocket.DefaultDialer.Dial(GATEWAY_URL, nil)
	if err != nil {
		return GatewayConnection{}, errors.New("failed to dial discord gateway")
	}

	return GatewayConnection{
		token:        discordToken,
		is_connected: false,
		conn:         c,
	}, nil
}

// Connect to discord gateway websocket server and pipe messages through channel
func Listen() {
	gateway, err := initApp()
	if err != nil {
		log.Fatal(err)
	}
	defer gateway.conn.Close()
	go gateway.on_message()

	for {
		continue
	}
}

func (g *GatewayConnection) on_message() {
	for {
		// Read raw message (byte array)
		_, json_msg, err := g.conn.ReadMessage()

		if err != nil {
			log.Println(string(json_msg), err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
				g.resume_connection()
			}
			continue
		}

		// Unmarshal Event name and OP code first
		var event GatewayEvent
		unmarshalJSON(json_msg, &event)

		if event.Sequence != nil {
			g.sequence = *(event.Sequence)
		}

		if DEBUG {
			if event.Name != "READY" {
				log.Println(string(json_msg))
			} else {
				log.Printf("Event %+v", event)
			}
		}

		if event.OP == OP_RECONNECT {
			g.resume_connection()
		}
		if event.OP == OP_HELLO {
			var payload GatewayHelloPayload
			unmarshalJSON(json_msg, &payload)
			g.is_connected = true
			g.heartbeat_interval = payload.Data.HeartbeatInterval
			go g.send_heartbeat()
		}
		if !g.is_identified && event.OP == OP_HEARTBEAT_ACK {
			err := g.send_identify()
			if err != nil {
				log.Println(err.Error())
			} else {
				g.is_identified = true
			}
		}
		if event.Name == READY {
			var payload GatewayReadyPayload
			unmarshalJSON(json_msg, &payload)
			g.resume_gateway_url = payload.Data.ResumeURL
			g.session_id = payload.Data.SessionID
			log.Printf("Gateway Connection %+v", g)
		}
		if event.Name == MESSAGE_CREATE {
			var payload GatewayMessageCreatePayload
			unmarshalJSON(json_msg, &payload)
			// log.Printf("Message create event %+v", payload)
			channels.DiscordChannel <- channels.DiscordMessage{
				ID:        payload.Data.ID,
				Content:   payload.Data.Content,
				GuildID:   payload.Data.GuildID,
				ChannelID: payload.Data.ChannelID,
				Timestamp: payload.Data.Timestamp,
				Author: channels.DiscordMessageAuthor{
					Username:      payload.Data.Author.Username,
					GlobalName:    payload.Data.Author.GlobalName,
					Discriminator: payload.Data.Author.Discriminator,
					ID:            payload.Data.Author.ID,
				},
			}
		}
	}
}

func (g *GatewayConnection) send_heartbeat() {
	for g.is_connected {
		heartbeat_payload := GatewayHeartbeat{
			GatewayEvent: GatewayEvent{OP: OP_HEARTBEAT},
			Sequence:     g.sequence,
		}
		err := g.conn.WriteJSON(heartbeat_payload)
		if err != nil {
			log.Println("Failed to send heartbeat")
			return
		}
		log.Println("Sent heartbeat")
		time.Sleep(time.Duration(g.heartbeat_interval) * time.Millisecond)
	}
}

func (g GatewayConnection) send_identify() error {
	identify_payload := GatewayIdentifyPayload{
		GatewayEvent: GatewayEvent{OP: OP_IDENTIFY},
		Data: GatewayIdentifyData{
			Token:   g.token,
			Intents: GUILD_MESSAGE_INTENT,
			Properties: DeviceProperties{
				OS:      "windows",
				Browser: "Discord",
				Device:  "desktop",
			},
		},
	}
	err := g.conn.WriteJSON(identify_payload)
	if err != nil {
		return errors.New("failed to send gateway identify")
	}
	return nil
}

func (g *GatewayConnection) resume_connection() error {
	g.is_connected = false
	g.conn.Close()
	c, _, err := websocket.DefaultDialer.Dial(g.resume_gateway_url, nil)
	if err != nil {
		return errors.New("failed to resume connection")
	}

	resume_payload := GatewayResumePayload{
		GatewayEvent: GatewayEvent{OP: OP_RESUME},
		Data: GatewayResumeData{
			Token:     g.token,
			SessionID: g.session_id,
			Sequence:  g.sequence,
		},
	}
	err = g.conn.WriteJSON(resume_payload)
	if err != nil {
		return errors.New("failed to send gateway resume connection")
	}
	g.conn = c
	g.is_connected = true
	g.is_identified = true
	log.Println("resumed gateway reconnection")
	return nil
}

func unmarshalJSON(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	if err != nil {
		log.Fatal("read:", err)
	}
}
