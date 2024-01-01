package internal

import (
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

func ConnectToWebSocket(symbols []string) {
	var streams string
	for _, symbol := range symbols {
		// Symbols need to be lowercase
		streams += fmt.Sprintf("%s@markPrice/", strings.ToLower(symbol))
	}
	streams = strings.TrimRight(streams, "/") // Remove the trailing slash

	wsURL := fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s", streams)
	log.Println(wsURL)
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)
	}
}
