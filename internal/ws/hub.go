package ws

import "livestream-service/internal/logger"

type Room struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Owner   *Client
	Clients map[string]*Client
	//ChatBroadcast   chan *Message
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
				r := h.Rooms[cl.RoomID]
				if _, ok := r.Clients[cl.ID]; !ok {
					r.Clients[cl.ID] = cl

					// notify streamer of new viewer
					r.StreamBroadcast <- &VideoMessage{
						Content: make([]byte, 20),
					}
				}
			}
		case cl := <-h.Unregister:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				if _, ok := h.Rooms[cl.RoomID].Clients[cl.ID]; ok {
					logger.Get().Print("Client %s has deregistered", cl.ID)
					//h.Rooms[cl.RoomID].ChatBroadcast <- &Message{
					//	Content:  fmt.Sprintf("user %s has left the chat", cl.Username),
					//	Username: cl.Username,
					//}
					delete(h.Rooms[cl.RoomID].Clients, cl.ID)
					//close(cl.Message)
				}
			}
		}
	}
}

func (r *Room) Run() {
	for {
		select {
		//case m := <-r.ChatBroadcast:
		//	for _, cl := range r.Clients {
		//		cl.Message <- m
		//	}
		//	//r.Owner.Message <- m
		case b := <-r.StreamBroadcast:

			// send video chunk to each client in hub
			for _, cl := range r.Clients {
				cl.Stream <- b
			}
		}
	}
}
