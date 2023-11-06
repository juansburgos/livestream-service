package ws

import (
	"fmt"
	"log"
)

type Room struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Owner           *Client
	Clients         map[string]*Client
	ChatBroadcast   chan *Message
	StreamBroadcast chan *VideoMessage
}

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case cl := <-h.Register:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				log.Printf("Primer check")
				r := h.Rooms[cl.RoomID]
				if _, ok := r.Clients[cl.ID]; !ok {
					r.Clients[cl.ID] = cl
				}
			}
		case cl := <-h.Unregister:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				if _, ok := h.Rooms[cl.RoomID].Clients[cl.ID]; ok {
					h.Rooms[cl.RoomID].ChatBroadcast <- &Message{
						Content:  fmt.Sprintf("user %s has left the chat", cl.Username),
						Username: cl.Username,
					}
					delete(h.Rooms[cl.RoomID].Clients, cl.ID)
					close(cl.Message)
				}
			}
		}
	}
}

func (r *Room) Run() {
	for {
		select {
		case m := <-r.ChatBroadcast:
			for _, cl := range r.Clients {
				cl.Message <- m
			}
			//r.Owner.Message <- m
		case b := <-r.StreamBroadcast:
			for _, cl := range r.Clients {
				cl.Stream <- b
			}
		}
	}
}
