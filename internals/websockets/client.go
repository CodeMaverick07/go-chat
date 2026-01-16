package websockets

import (
	"log"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	Connection *websocket.Conn
	Manager    *Manager
	Logger     *log.Logger
}

func NewClient(connection *websocket.Conn, manager *Manager, logger *log.Logger) *Client {
	return &Client{
		Connection: connection,
		Manager:    manager,
		Logger:     logger,
	}
}

func (c *Client) ReadMessages() {
	c.Logger.Println("client read loop started")

	defer func() {
		c.Manager.RemoveClient(c)
	}()

	for {
		messageType, payload, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				c.Logger.Printf("unexpected close: %v", err)
			}
			break
		}

		c.Logger.Println("MessageType:", messageType)
		c.Logger.Println("Payload:", string(payload))
	}
}

func (c *Client) WriteMessages() {
	defer func() {
		c.Manager.RemoveClient(c)
	}()
}
