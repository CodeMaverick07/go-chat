package websockets

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, client *Client) error

const (
	EventSeedMessage = "new_message"
	EventChangeRoom  = "change_room"
	EventSendMessage = "send_message"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}
type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

func SendMessageHandler(event Event, c *Client) error {
	var chatevent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return err
	}

	broadMessage := NewMessageEvent{
		SendMessageEvent: SendMessageEvent{
			Message: chatevent.Message,
			From:    c.UserID,
		},
		Sent: time.Now(),
	}

	data, _ := json.Marshal(broadMessage)
	outgoingEvent := Event{
		Type:    EventSeedMessage,
		Payload: data,
	}

	var toRemove []*Client

	c.Manager.RLock()
	for client := range c.Manager.clientsList {
		if client.chatroom == c.chatroom {
			select {
			case client.egress <- outgoingEvent:
			default:
				toRemove = append(toRemove, client)
			}
		}
	}
	c.Manager.RUnlock()

	for _, client := range toRemove {
		c.Manager.RemoveClient(client)
	}
	return nil
}

type ChangeRoomEvent struct {
	Name string `json:"name"`
}

func ChatRoomHandler(event Event, c *Client) error {
	var changeRoomEvent ChangeRoomEvent
	if err := json.Unmarshal(event.Payload, &changeRoomEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}
	c.chatroom = changeRoomEvent.Name

	return nil

}
