package ws

import (
	"github.com/gorilla/websocket"
	"livestream-service/internal/logger"
)

type Client struct {
	Conn *websocket.Conn
	//Message  chan *Message
	Stream chan *VideoMessage
	ID     string `json:"Id"`
	RoomID string `json:"roomId"`
	//Username string `json:"username"`
}

type Message struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

type VideoMessage struct {
	Content []byte
}

//func (c *Client) writeMessage() {
//	defer func() {
//		c.Conn.Close()
//	}()
//
//	for {
//		message, ok := <-c.Message
//		if !ok {
//			return
//		}
//		c.Conn.WriteJSON(message)
//		logger.Get().Printf("WRITE MSG client:  %s, %s al roomid %s", c.ID, c.Username, c.RoomID)
//		//logger := log.New(os.Stdout, "", 0)
//		//logger.Printf("WRITE MSG client:  %s, %s al roomid %s", c.ID, c.Username, c.RoomID)
//	}
//}

//func (c *Client) readMessage(h *Hub) {
//	defer func() {
//		h.Unregister <- c
//		c.Conn.Close()
//	}()
//
//	for {
//		_, m, err := c.Conn.ReadMessage()
//		if err != nil {
//			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
//				log.Printf("error %s", err.Error())
//			}
//			break
//		}
//		msg := &Message{
//			Content:  string(m),
//			Username: c.Username,
//		}
//		logger := log.New(os.Stdout, "", 0)
//		logger.Printf("READMSG client:  %s, %s al roomid %s", c.ID, c.Username, c.RoomID)
//		h.Rooms[c.RoomID].ChatBroadcast <- msg
//	}
//}

// Send data to a specific client
func (c *Client) writeStream() {
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			logger.Get().Print("Error while trying to close outgoing conn for client %s, err:", c.ID, err.Error())
			return
		}
	}()

	for {
		data, ok := <-c.Stream

		logger.Get().Print("Writing stream for client %s", c.ID)
		if !ok {
			logger.Get().Print("Error on stream for client %s", c.ID)
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
			logger.Get().Print("Error while trying to close ingoing conn for streamer %s, err:", c.ID, err.Error())
			return
		}
	}()

	for {
		_, m, err := c.Conn.ReadMessage()

		logger.Get().Print("Read data from streamer %s", c.ID)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Get().Print("Unexpected close for streamer %s, err:", c.ID, err.Error())
			}
			break
		}
		msg := &VideoMessage{
			Content: m,
		}
		h.Rooms[c.RoomID].StreamBroadcast <- msg
	}
}
