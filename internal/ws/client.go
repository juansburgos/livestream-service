package ws

import (
	"github.com/gorilla/websocket"
	"livestream-service/internal/logger"
)

type Client struct {
	Conn   *websocket.Conn
	Stream chan *VideoMessage
	ID     string `json:"Id"`
	RoomID string `json:"roomId"`
}

type Message struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

type VideoMessage struct {
	Content []byte
}

// Send data to a specific client
func (c *Client) writeStream() {
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			logger.Get().Printf("Error while trying to close outgoing conn for client %s, err:%s", c.ID, err.Error())
			return
		}
	}()

	for {
		data, ok := <-c.Stream

		logger.Get().Printf("Writing stream for client %s", c.ID)
		if !ok {
			logger.Get().Printf("Error on stream for client %s", c.ID)
			return
		}

		err := c.Conn.WriteMessage(websocket.BinaryMessage, data.Content)
		if err != nil {
			logger.Get().Printf("Error while trying to write on stream for client %s", c.ID)
			return
		}
	}
}

// Read data from a streamer
func (c *Client) readStream(h *Hub) {
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			logger.Get().Printf("Error while trying to close ingoing conn for streamer %s, err:%s", c.ID, err.Error())
			return
		}
	}()

	for {
		_, m, err := c.Conn.ReadMessage()

		logger.Get().Printf("Read data from streamer %s", c.ID)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Get().Printf("Unexpected close for streamer %s, err: %s", c.ID, err.Error())
			}
			break
		}
		msg := &VideoMessage{
			Content: m,
		}
		h.Rooms[c.RoomID].StreamBroadcast <- msg
	}
}
