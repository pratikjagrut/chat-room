package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pratikjagrut/chat-room/client"
	"github.com/pratikjagrut/chat-room/room"
)

func main() {
	log.Println("Starting server...")
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/", fs)

	// The factory function that knows how to create a client.
	clientFactory := func(conn *websocket.Conn, r *room.Room, username string) room.Chatter {
		return client.NewClient(conn, r, username)
	}

	// Create the room and give it the factory.
	chatRoom := room.NewRoom(clientFactory)
	go chatRoom.Run()

	http.Handle("/room", chatRoom)

	log.Println("🚀 Server starting on http://localhost:8080")
	log.Println("📁 Serving static files from ./static/")
	log.Println("🔌 WebSocket endpoint: ws://localhost:8080/room")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
