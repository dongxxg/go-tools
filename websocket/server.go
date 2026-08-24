package main

import (
	"context"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)

func handleWebsocket(conn *websocket.Conn) {
	defer conn.Close(websocket.StatusInternalError, "Internal Server Error")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var msg string
	err := wsjson.Read(ctx, conn, &msg)
	if err != nil {
		log.Println("Failed to read message:", err)
		return
	}

	log.Println("Received message:", msg)
	err = wsjson.Write(ctx, conn, "Server received: "+msg)
	if err != nil {
		log.Println("Failed to send message:", err)
		return
	}
}

func main() {
	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		conn, err := websocket.Accept(writer, request, nil)
		if err != nil {
			log.Println("Failed to accept connection:", err)
			return
		}
		defer conn.Close(websocket.StatusInternalError, "Internal Server Error")
		handleWebsocket(conn)
	})

	log.Println("websocket server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
