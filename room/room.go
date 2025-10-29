package room

import (
	"log"
	"net/http"
	"sort"

	"github.com/gorilla/websocket"
)

// WsJsonResponse defines the structure for JSON responses sent over the WebSocket.
type WsJsonResponse struct {
	Username       string   `json:"username"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

// Chatter defines the interface for a chat client.
// This allows the room to manage clients without depending on the concrete client implementation.
type Chatter interface {
	Username() string
	Send(msg WsJsonResponse)
	Close()
	Read()
	Write()
}

// Room represents a single chat room.
type Room struct {
	clients   map[Chatter]bool
	forward   chan WsJsonResponse
	join      chan Chatter
	leave     chan Chatter
	newClient func(conn *websocket.Conn, room *Room, username string) Chatter
}

// NewRoom creates a new chat room.
func NewRoom(newClientFunc func(conn *websocket.Conn, room *Room, username string) Chatter) *Room {
	return &Room{
		forward:   make(chan WsJsonResponse, 256),
		join:      make(chan Chatter),
		leave:     make(chan Chatter),
		clients:   make(map[Chatter]bool),
		newClient: newClientFunc,
	}
}

// Run is the single-threaded gatekeeper for all room state changes.
func (r *Room) Run() {
	for {
		select {
		case chatter := <-r.join:
			r.clients[chatter] = true
			r.broadcastUserList()

		case chatter := <-r.leave:
			if _, ok := r.clients[chatter]; ok {
				delete(r.clients, chatter)
				chatter.Close() // Close the chatter's resources.
				r.broadcastUserList()
			}

		case msg := <-r.forward:
			for chatter := range r.clients {
				chatter.Send(msg)
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

// ServeHTTP handles WebSocket requests from the peer.
func (r *Room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("ServeHTTP:", err)
		return
	}

	username := req.URL.Query().Get("username")
	if username == "" {
		username = "Anonymous"
	}

	chatter := r.newClient(socket, r, username)
	log.Println("New client joined: ", username)
	r.join <- chatter

	// Defer the leave action until the connection is closed.
	defer func() { r.leave <- chatter }()

	// Start the write pump in a separate goroutine.
	go chatter.Write()

	// Block on the read pump. This will keep the connection alive.
	// When Read() returns, the client has disconnected, and the defer will be executed.
	chatter.Read()
}

// broadcastUserList sends the updated user list to all clients.
func (r *Room) broadcastUserList() {
	users := r.getUserList()
	response := WsJsonResponse{
		ConnectedUsers: users,
		MessageType:    "users_list",
	}
	for chatter := range r.clients {
		chatter.Send(response)
	}
}

// getUserList returns a sorted list of unique usernames in the room.
func (r *Room) getUserList() []string {
	var users []string
	for chatter := range r.clients {
		users = append(users, chatter.Username())
	}
	sort.Strings(users)
	return users
}

// Forward sends a message to the room's forward channel.
func (r *Room) Forward(msg WsJsonResponse) {
	r.forward <- msg
}
