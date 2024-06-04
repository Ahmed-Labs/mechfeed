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

const GATEWAY_URL = "wss://gateway.discord.gg/?v=9&encoding=json"
var DEBUG bool

type GatewayPayload struct {
	Event string `json:"t,omitempty"`
	OP int `json:"op"`
	Data GatewayPayloadData `json:"d,omitempty"`
}

type GatewayPayloadData struct {
	Heartbeat_Interval int `json:"heartbeat_interval,omitempty"`
	Token string `json:"token,omitempty"`
	Properties DeviceProperties `json:"properties,omitempty"`
	ResumeURL string `json:"resume_gateway_url,omitempty"`
}

type DeviceProperties struct {
	OS string `json:"$os,omitempty"`
	Browser string `json:"$broswer,omitempty"`
	Device string `json:"$device,omitempty"`
}

type GatewayConnection struct {
	token string
	is_connected bool
	ready bool
	heartbeat_interval int
	conn *websocket.Conn
}

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

	return GatewayConnection {
		token : discordToken,
		is_connected: false,
		ready: false,
		heartbeat_interval: 0,
		conn: c,
	}, nil
}

func main() {
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
func (g GatewayConnection) on_message() {
	for {
		var res GatewayPayload
		var err error

		if DEBUG {
			var json_msg []byte
			_, json_msg, err = g.conn.ReadMessage()
			if err != nil {
				log.Fatal("read:", err)
			}
			err = json.Unmarshal(json_msg, &res)
			if err != nil {
				log.Fatal("read:", err)
			}

			if res.Event == "READY" {
				os.WriteFile("ready.json", json_msg, os.ModePerm)
			} else {
				log.Println(string(json_msg))
			}
		} else {
			err = g.conn.ReadJSON(&res)
			log.Printf("%#v", res)
		}

		if err != nil {
			log.Println("read:", err)
			return
		}

		if res.OP == 10 { // hello event
			g.is_connected = true
			g.heartbeat_interval = res.Data.Heartbeat_Interval
			go g.send_heartbeat()
		}
		if !g.ready && res.OP == 11 {
			g.ready = true
			g.send_identify()
		}
	}
}

func (g *GatewayConnection) send_heartbeat() {
	for g.is_connected {
		heartbeat_payload := GatewayPayload{OP: 1}
		err := g.conn.WriteJSON(heartbeat_payload)
		if err != nil {
			log.Println("Failed to send heartbeat")
			return
		}
		log.Println("Sent heartbeat")
		time.Sleep(time.Duration(g.heartbeat_interval)*time.Millisecond)
	}
}

func (g GatewayConnection) send_identify() error {
	identify_payload := GatewayPayload{
		OP: 2,
		Data: GatewayPayloadData{
			Token: g.token,
			Properties: DeviceProperties{
				OS: "windows",
				Browser: "Discord",
				Device: "desktop",
			},
		},
	}
	err := g.conn.WriteJSON(identify_payload)
	if err != nil {
		return errors.New("failed to send gateway identify")
	}
	return nil
}