package main

import (
	"errors"
	"log"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var GATEWAY_URL string = "wss://gateway.discord.gg/?v=9&encoding=json"

type GatewayPayload struct {
	Event string `json:"t"`
	OP int `json:"op"`
}

type GatewayConnection struct {
	token string
	is_connected bool
	heartbeat_interval int
	dialer *websocket.Dialer
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
	return GatewayConnection {
		token : discordToken,
		is_connected: false,
		heartbeat_interval: 0,
		dialer: websocket.DefaultDialer,
	}, nil
}

func main() {
	gateway, err := initApp()
	if err != nil {
		log.Fatal(err)
	}

	// Establish connection with gateway
	c, _, err := gateway.dialer.Dial(GATEWAY_URL, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			var res GatewayPayload
			err = c.ReadJSON(&res)

			if err != nil {
				log.Println("read:", err)
				return
			}

			log.Printf("%+v", res)

			if res.OP == 10 {
				go gateway.send_heartbeat()
			}
		}
	}()

	for {
		continue
	}
}

func (g GatewayConnection) send_heartbeat(){}
