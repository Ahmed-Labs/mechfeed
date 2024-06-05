package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
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

	DEBUG = os.Getenv("DEBUG") == "true"

	// Establish connection with gateway
	c, _, err := websocket.DefaultDialer.Dial(GATEWAY_URL, nil)
	if err != nil {
		return GatewayConnection{}, errors.New("failed to dial discord gateway")
	}

	return GatewayConnection{
		token:              discordToken,
		is_connected:       false,
		heartbeat_interval: 0,
		conn:               c,
	}, nil
}

func main() {
	gateway, err := initApp()
	if err != nil {
		log.Fatal(err)
	}
	defer gateway.conn.Close()
	gateway.conn.SetCloseHandler(func(code int, text string) error {
		if code != 4007 {
			gateway.resume_connection()
		}
		return nil
	})

	go gateway.on_message()
	for {
		continue
	}
}
func (g GatewayConnection) on_message() {
	for {
		// Read raw message (byte array)
		_, json_msg, err := g.conn.ReadMessage()

		if err != nil {
			log.Println(string(json_msg))
			log.Fatal("read:", err)
		}

		// Unmarshal Event name and OP code first
		var event GatewayEvent
		unmarshalJSON(json_msg, &event)

		if event.Sequence != nil {
			g.sequence = *(event.Sequence)
		}

		if DEBUG {
			// Print raw json string in DEBUG mode
			if event.Name != "READY" {
				log.Println(string(json_msg))
			}
		}
		log.Printf("Event %+v", event)
		
		if event.OP == 7 {
			g.resume_connection()
		}
		if event.OP == 10 { // Hello event
			var payload GatewayHelloPayload
			unmarshalJSON(json_msg, &payload)
			log.Printf("Unmarshalled %+v", payload)
			g.is_connected = true
			g.heartbeat_interval = payload.Data.Heartbeat_Interval
			go g.send_heartbeat()
		}
		if !g.is_identified && event.OP == 11 {
			err := g.send_identify()
			if err != nil {
				log.Println(err.Error())
			} else {
				g.is_identified = true
			}
		}
		// if event.Name == "SESSIONS_REPLACE" {
		// 	var payload GatewayReplaceSessionPayload
		// 	unmarshalJSON(json_msg, &payload)
		// 	g.session_id = payload.Data[0].SessionID
		// 	err := g.resume_connection()
		// 	if err != nil {
		// 		log.Fatal(err.Error())
		// 	}
		// }
		if event.Name == "READY" {
			var payload GatewayReadyPayload
			unmarshalJSON(json_msg, &payload)
			g.resume_gateway_url = payload.Data.ResumeURL
			g.session_id = payload.Data.SessionID
			log.Printf("Gateway Connection %+v", g)
		}
		if event.Name == "MESSAGE_CREATE" {
			// log.Println(res.Data.Content)
			log.Println("Message create event")
		}
	}
}

func (g *GatewayConnection) send_heartbeat() {
	for g.is_connected {
		heartbeat_payload := GatewayHeartbeat{
			GatewayEvent: GatewayEvent{OP: 1},
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
		GatewayEvent: GatewayEvent{OP: 2},
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
	c, _, err := websocket.DefaultDialer.Dial(g.resume_gateway_url, nil)
	if err != nil {
		return errors.New("failed to resume connection")
	}

	resume_payload := GatewayResumePayload{
		GatewayEvent: GatewayEvent{OP: 6},
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
	return nil
}

func unmarshalJSON(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	if err != nil {
		log.Fatal("read:", err)
	}
}
