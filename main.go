package main

import (
	"log"
	"net/http"
	"sort"

	"github.com/gorilla/websocket"
)

func main() {
	log.Println("Starting server...")
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/", fs)

	log.Println("🚀 Server starting on http://localhost:8080")
	log.Println("📁 Serving static files from ./static/")
	log.Println("🔌 WebSocket endpoint: ws://localhost:8080/room")

	room := newRoom()
	go room.run()
	http.Handle("/room", room)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type WsJsonResponse struct {
	Username       string   `json:"username"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

// Client implementation
type Client struct {
	Conn     *websocket.Conn
	receive  chan WsJsonResponse
	room     *room
	username string
}

// read pumps messages from the websocket connection to the room.
func (c *Client) read() {
	defer c.Conn.Close()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			// A read error is treated as a disconnect.
			// The defer in ServeHTTP will trigger the leave event.
			return
		}

		// FIX #2: Create a new response object for each message to prevent data races.
		response := WsJsonResponse{
			Username:    c.username,
			Message:     string(message),
			MessageType: "chat",
		}
		c.room.forward <- response
	}
}

// write pumps messages from the room to the websocket connection.
func (c *Client) write() {
	defer c.Conn.Close()
	for message := range c.receive {
		err := c.Conn.WriteJSON(message)
		if err != nil {
			// A write error is also a disconnect.
			return
		}
	}
}

type room struct {
	clients map[*Client]string
	forward chan WsJsonResponse
	join    chan *Client
	leave   chan *Client
}

func newRoom() *room {
	return &room{
		forward: make(chan WsJsonResponse, 256),
		join:    make(chan *Client),
		leave:   make(chan *Client),
		clients: make(map[*Client]string),
	}
}

// run is the single-threaded gatekeeper for all room state changes.
func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// A client has connected.
			r.clients[client] = client.username

			// FIX #1: Broadcast the new user list DIRECTLY, not via the forward channel.
			users := r.getUserList()
			response := WsJsonResponse{
				ConnectedUsers: users,
				MessageType:    "users_list",
			}
			for c := range r.clients {
				c.receive <- response
			}

		case client := <-r.leave:
			// A client has disconnected.
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.receive)

				// FIX #1: Broadcast the new user list DIRECTLY.
				users := r.getUserList()
				response := WsJsonResponse{
					ConnectedUsers: users,
					MessageType:    "users_list",
				}
				for c := range r.clients {
					c.receive <- response
				}
			}

		case msg := <-r.forward:
			// A chat message has been received. Broadcast it to all clients.
			for client := range r.clients {
				client.receive <- msg
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	username := req.URL.Query().Get("username")
	if username == "" {
		username = "Anonymous"
	}

	client := &Client{
		Conn:     socket,
		receive:  make(chan WsJsonResponse, 256),
		room:     r,
		username: username,
	}
	log.Println("New client joined: ", username)
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

func (r *room) getUserList() []string {
	var users []string
	for _, username := range r.clients {
		users = append(users, username)
	}
	sort.Strings(users)
	return users
}
