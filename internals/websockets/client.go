package websockets

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 8) / 10
)

type ClientList map[*Client]bool

type Client struct {
	Connection *websocket.Conn
	Manager    *Manager
	Logger     *log.Logger
	egress     chan Event
	chatroom   string
	UserID     string
}

func NewClient(connection *websocket.Conn, manager *Manager, logger *log.Logger, userID string) *Client {
	return &Client{
		Connection: connection,
		Manager:    manager,
		Logger:     logger,
		egress:     make(chan Event, 10),
		UserID:     userID,
	}
}

func (c *Client) ReadMessages() {
	c.Logger.Println("client read loop started")

	defer func() {
		c.Manager.RemoveClient(c)
	}()
	c.Connection.SetReadLimit(512)
	if err := c.Connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}

	c.Connection.SetPongHandler(c.pongHandler)

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
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				return
			}
			c.Connection.SetWriteDeadline(time.Now().Add(5 * time.Second))
			data, err := json.Marshal(message)
			if err != nil {
				return
			}
			if err := c.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			c.Connection.SetWriteDeadline(time.Now().Add(5 * time.Second))
			c.Logger.Println("ping")
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	c.Logger.Println("pong")
	return c.Connection.SetReadDeadline(time.Now().Add(pongWait))
}
