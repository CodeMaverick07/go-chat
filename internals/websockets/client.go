package websockets

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	// Increased timeouts for better stability
	pongWait     = 60 * time.Second    // Wait 60 seconds for pong response
	pingInterval = (pongWait * 9) / 10 // Send ping every 54 seconds

	// Write deadline - how long to wait when writing to client
	writeWait = 10 * time.Second
)

type ClientList map[*Client]bool

type Client struct {
	Connection *websocket.Conn
	Manager    *Manager
	Logger     *log.Logger
	egress     chan Event
	chatroom   string
	UserID     uuid.UUID
}

func NewClient(connection *websocket.Conn, manager *Manager, logger *log.Logger, userID uuid.UUID) *Client {
	return &Client{
		Connection: connection,
		Manager:    manager,
		Logger:     logger,
		egress:     make(chan Event, 100), // Increased buffer from 10 to 100
		UserID:     userID,
	}
}

func (c *Client) ReadMessages() {
	c.Logger.Printf("client read loop started for user: %s", c.UserID)

	defer func() {
		c.Logger.Printf("client read loop ended for user: %s", c.UserID)
		c.Manager.RemoveClient(c)
	}()

	// INCREASED FROM 512 bytes to 10MB
	// This prevents disconnections due to large messages
	c.Connection.SetReadLimit(10 * 1024 * 1024) // 10MB

	// Set initial read deadline
	if err := c.Connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.Logger.Printf("error setting read deadline: %v", err)
		return
	}

	// Pong handler - called when pong message received
	c.Connection.SetPongHandler(c.pongHandler)

	for {
		// Read message from client
		messageType, payload, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure, // Added normal closure
			) {
				c.Logger.Printf("unexpected close for user %s: %v", c.UserID, err)
			} else {
				c.Logger.Printf("read error for user %s: %v", c.UserID, err)
			}
			break
		}

		// Only process text messages
		if messageType != websocket.TextMessage {
			c.Logger.Printf("received non-text message type: %d", messageType)
			continue
		}

		// Log received message for debugging
		c.Logger.Printf("received message from user %s: %s", c.UserID, string(payload))

		// Parse event
		var req Event
		if err := json.Unmarshal(payload, &req); err != nil {
			c.Logger.Printf("error unmarshalling message from user %s: %v", c.UserID, err)

			// Send error back to client instead of breaking connection
			c.sendError(fmt.Sprintf("invalid JSON: %v", err))
			continue // Continue instead of break
		}

		// Route event to handler
		if err := c.Manager.routeEvent(req, c); err != nil {
			c.Logger.Printf("error handling event %s for user %s: %v", req.Type, c.UserID, err)

			// Send error back to client
			c.sendError(fmt.Sprintf("error handling %s: %v", req.Type, err))
			// Don't break - continue processing other messages
		}
	}
}

func (c *Client) WriteMessages() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Logger.Printf("client write loop ended for user: %s", c.UserID)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			// Channel closed - exit
			if !ok {
				c.Logger.Printf("egress channel closed for user: %s", c.UserID)
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Set write deadline
			if err := c.Connection.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.Logger.Printf("error setting write deadline for user %s: %v", c.UserID, err)
				return
			}

			// Marshal message to JSON
			data, err := json.Marshal(message)
			if err != nil {
				c.Logger.Printf("error marshalling message for user %s: %v", c.UserID, err)
				// Don't return - continue processing other messages
				continue
			}

			// Log outgoing message for debugging
			c.Logger.Printf("sending message to user %s: type=%s", c.UserID, message.Type)

			// Send message
			if err := c.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
				c.Logger.Printf("error writing message to user %s: %v", c.UserID, err)
				return
			}

		case <-ticker.C:
			// Send ping to keep connection alive
			if err := c.Connection.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.Logger.Printf("error setting ping deadline for user %s: %v", c.UserID, err)
				return
			}

			c.Logger.Printf("sending ping to user: %s", c.UserID)

			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Logger.Printf("error writing ping to user %s: %v", c.UserID, err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	c.Logger.Printf("received pong from user: %s", c.UserID)

	// Reset read deadline when pong received
	return c.Connection.SetReadDeadline(time.Now().Add(pongWait))
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	errorEvent := Event{
		Type:    "error",
		Payload: json.RawMessage(fmt.Sprintf(`{"error": "%s"}`, errMsg)),
	}

	select {
	case c.egress <- errorEvent:
	default:
		c.Logger.Printf("failed to send error to user %s (channel full)", c.UserID)
	}
}
