package websockets

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type Manager struct {
	logger      *log.Logger
	clientsList ClientList
	sync.RWMutex
	egress chan []byte
}

func NewManager(Logger *log.Logger) *Manager {
	return &Manager{
		logger:      Logger,
		clientsList: make(ClientList),
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
