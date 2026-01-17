package websockets

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

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
	logger      *log.Logger
	clientsList ClientList
	sync.RWMutex
	handlers map[string]EventHandler
}

func NewManager(Logger *log.Logger) *Manager {
	m := &Manager{
		logger:      Logger,
		clientsList: make(ClientList),
		handlers:    make(map[string]EventHandler),
	}
	m.SetUpEventHandlers()
	return m
}
func (m *Manager) SetUpEventHandlers() {
	m.handlers[EventSeedMessage] = func(e Event, c *Client) error {
		fmt.Println(e)
		return nil
	}
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
	conn, err := webSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Println("upgrade error:", err)
		return
	}

	m.logger.Println("client connected")

	client := NewClient(conn, m, m.logger)
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
		client.Connection.Close()
		delete(m.clientsList, client)
		m.logger.Println("client disconnected")
	}
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
