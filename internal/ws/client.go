package ws

import (
	"github.com/gorilla/websocket"
	"log"
	"os"
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
		logger := log.New(os.Stdout, "", 0)
		logger.Printf("WRITE MSG client:  %s, %s al roomid %s", c.ID, c.Username, c.RoomID)
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
		logger := log.New(os.Stdout, "", 0)
		logger.Printf("READMSG client:  %s, %s al roomid %s", c.ID, c.Username, c.RoomID)
		h.Rooms[c.RoomID].ChatBroadcast <- msg
	}
}

func (c *Client) writeStream() {
	defer func() {
		c.Conn.Close()
	}()

	logger := log.New(os.Stdout, "", 0)
	for {
		logger.Printf("for writeStream")
		data, ok := <-c.Stream
		if !ok {
			logger.Printf("DATA del writeStream no ok")
			return
		}
		c.Conn.WriteMessage(websocket.BinaryMessage, data.Content)
		logger.Printf("USer %s escribio en el socket", c.Username)
	}
}

func (c *Client) readStream(h *Hub) {
	defer func() {
		c.Conn.Close()
	}()
	logger := log.New(os.Stdout, "", 0)
	for {
		logger.Printf("for readstream")
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
