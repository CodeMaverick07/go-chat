package websockets

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	Connection *websocket.Conn
	Manager    *Manager
	Logger     *log.Logger
	egress     chan Event
}

func NewClient(connection *websocket.Conn, manager *Manager, logger *log.Logger) *Client {
	return &Client{
		Connection: connection,
		Manager:    manager,
		Logger:     logger,
		egress:     make(chan Event),
	}
}

func (c *Client) ReadMessages() {
	c.Logger.Println("client read loop started")

	defer func() {
		c.Manager.RemoveClient(c)
	}()

	for {
		_, payload, err := c.Connection.ReadMessage()
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
		var req Event
		if err := json.Unmarshal(payload, &req); err != nil {
			c.Logger.Printf("error marshalling message: %v", err)
			break
		}

		if err := c.Manager.routeEvent(req, c); err != nil {
			c.Logger.Println("Error handeling Message: ", err)

		}

	}
}

func (c *Client) WriteMessages() {
	defer func() {
		c.Manager.RemoveClient(c)
	}()
	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.Connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					c.Logger.Printf("connection closed %v: ", err)
				}
				return
			}
			data, err := json.Marshal(message)
			if err != nil {
				c.Logger.Println(err)
				return // closes the connection, should we really
			}

			if err := c.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
				c.Logger.Println(err)
			}
			c.Logger.Println("message sent")
		}
	}
}
