package main

import (
	"context"
	"log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"strconv"
)

func main() {
	ctx := context.Background()

	var count int
	for count < 10 {
		conn, _, err := websocket.Dial(ctx, "ws://localhost:8080/ws", nil)
		if err != nil {
			log.Fatal("Failed to connect:", err)
		}

		err = wsjson.Write(ctx, conn, strconv.Itoa(count)+" Hello World!")
		if err != nil {
			log.Fatal("Failed to send message", err)
		}
		var msg string
		err = wsjson.Read(ctx, conn, &msg)
		if err != nil {
			log.Fatal("Failed to read message:", err)
		}
		log.Println("Received message:", msg)
		_ = conn.Close(websocket.StatusInternalError, "Internal Server Error")
		count++
	}
	log.Println("Client Exit")
}
