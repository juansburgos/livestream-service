package ws

import (
	"github.com/gorilla/websocket"
	"log"
)

type Client struct {
	Conn     *websocket.Conn
	Message  chan *Message
	Stream   chan *VideoMessage
	ID       string `json:"Id"`
	RoomID   string `json:"roomId"`
	Username string `json:"username"`
}

type Message struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

type VideoMessage struct {
	Content []byte
}

func (c *Client) writeMessage() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Message
		if !ok {
			return
		}
		c.Conn.WriteJSON(message)
	}
}

func (c *Client) readMessage(h *Hub) {
	defer func() {
		h.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error %s", err.Error())
			}
			break
		}
		msg := &Message{
			Content:  string(m),
			Username: c.Username,
		}

		h.Rooms[c.RoomID].ChatBroadcast <- msg
	}
}

func (c *Client) writeStream() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		data, ok := <-c.Stream
		if !ok {
			return
		}
		c.Conn.WriteMessage(websocket.BinaryMessage, data.Content)
	}
}

func (c *Client) readStream(h *Hub) {
	defer func() {
		c.Conn.Close()
	}()

	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error %s", err.Error())
			}
			break
		}
		vmsg := &VideoMessage{
			Content: m,
		}

		h.Rooms[c.RoomID].StreamBroadcast <- vmsg
	}
}
