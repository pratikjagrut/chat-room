package client

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/pratikjagrut/chat-room/room"
)

// Forwarder defines the interface for forwarding messages to the room.
type Forwarder interface {
	Forward(msg room.WsJsonResponse)
}

// Client represents a single chat user's WebSocket connection.
type Client struct {
	conn     *websocket.Conn
	receive  chan room.WsJsonResponse
	room     Forwarder
	username string
}

// NewClient is the factory function that main will use.
func NewClient(conn *websocket.Conn, r Forwarder, username string) *Client {
	return &Client{
		conn:     conn,
		receive:  make(chan room.WsJsonResponse, 256),
		room:     r,
		username: username,
	}
}

// Read pumps messages from the websocket connection to the room.
func (c *Client) Read() {
	defer c.conn.Close()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Client %s read error: %v", c.username, err)
			return
		}

		response := room.WsJsonResponse{
			Username:    c.username,
			Message:     string(message),
			MessageType: "chat",
		}
		c.room.Forward(response)
	}
}

// Write pumps messages from the room to the websocket connection.
func (c *Client) Write() {
	defer c.conn.Close()
	for message := range c.receive {
		err := c.conn.WriteJSON(message)
		if err != nil {
			log.Printf("Client %s write error: %v", c.username, err)
			return
		}
	}
}

// Username returns the client's username.
func (c *Client) Username() string {
	return c.username
}

// Send sends a message to the client's receive channel.
func (c *Client) Send(msg room.WsJsonResponse) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in Send:", r)
		}
	}()
	c.receive <- msg
}

// Close closes the client's receive channel.
func (c *Client) Close() {
	close(c.receive)
}
