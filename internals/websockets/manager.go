package websockets

import (
	"errors"
	"go-chat/internals/api"
	"go-chat/internals/contexkeys"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

var (
	webSocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
)

type Manager struct {
	logger              *log.Logger
	MessageHandler      *api.MessageHandler
	ConversationHandler *api.ConversationHandler
	clientsList         ClientList
	sync.RWMutex
	handlers map[string]EventHandler
}

func (m *Manager) GetStats() map[string]interface{} {
	m.RLock()
	defer m.RUnlock()

	return map[string]interface{}{
		"total_clients": len(m.clientsList),
		"timestamp":     time.Now(),
	}
}
func (m *Manager) Shutdown() {
	m.Lock()
	defer m.Unlock()

	m.logger.Println("shutting down websocket manager...")

	for client := range m.clientsList {
		// Send close message to client
		client.Connection.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"),
		)
		client.Connection.Close()
		close(client.egress)
	}

	m.clientsList = make(ClientList)
	m.logger.Println("websocket manager shutdown complete")
}

func NewManager(Logger *log.Logger, messageHandler *api.MessageHandler, conversationHandler *api.ConversationHandler) *Manager {
	m := &Manager{
		logger:              Logger,
		MessageHandler:      messageHandler,
		ConversationHandler: conversationHandler,
		clientsList:         make(ClientList),
		handlers:            make(map[string]EventHandler),
	}
	m.SetUpEventHandlers()
	return m
}
func (m *Manager) SetUpEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler
	m.handlers[EventCreateConversation] = CreateConversationHandler
	m.handlers[EventGetConversations] = GetConversationHandler
	m.handlers[EventGetMessages] = GetMessagesHandler
	m.handlers[EventMarkAsRead] = MarkAsReadHandler
	m.handlers[EventStartTyping] = TypingIndicatorHandler
	m.handlers[EventStopTyping] = TypingIndicatorHandler

}

func (m *Manager) routeEvent(e Event, c *Client) error {
	if handler, ok := m.handlers[e.Type]; ok {
		if err := handler(e, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contexkeys.UserID).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := webSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Println("upgrade error:", err)
		return
	}

	m.logger.Println("client connected:", userID)

	client := NewClient(conn, m, m.logger, userID)
	m.AddClient(client)

	go client.ReadMessages()
	go client.WriteMessages()
}

func (m *Manager) AddClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	m.clientsList[client] = true

}

func (m *Manager) RemoveClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clientsList[client]; ok {
		m.logger.Printf("removing client: %s", client.UserID)

		// Close websocket connection
		client.Connection.Close()

		// Close egress channel (will cause WriteMessages to exit)
		close(client.egress)

		// Remove from list
		delete(m.clientsList, client)

		m.logger.Printf("client disconnected: %s (total clients: %d)", client.UserID, len(m.clientsList))
	}
}

//BroadcastToUsers sends an event to specific users

func (m *Manager) BroadcastToUsers(event Event, userIDs []uuid.UUID) {
	userIdMap := make(map[uuid.UUID]bool)
	for _, id := range userIDs {
		userIdMap[id] = true
	}

	m.RLock()
	defer m.RUnlock()

	m.logger.Printf("broadcasting event %s to %d users", event.Type, len(userIDs))

	successCount := 0
	for client := range m.clientsList {
		if userIdMap[client.UserID] {
			select {
			case client.egress <- event:
				successCount++
			default:
				m.logger.Printf("failed to send to client %s (channel full, dropping message)", client.UserID)
			}
		}
	}

	m.logger.Printf("successfully sent to %d/%d target users", successCount, len(userIDs))
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:8080":
		return true
	case "http://localhost:5500":
		return true
	default:
		return true
	}
}
