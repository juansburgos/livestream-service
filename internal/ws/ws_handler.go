package ws

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
)

type Handler struct {
	hub *Hub
}

func NewHandler(h *Hub) *Handler {
	return &Handler{
		hub: h,
	}
}

type CreateRoomReq struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OwnerName string `json:"ownerName"`
	OwnerID   string `json:"ownerId"`
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var newRoomID = strconv.Itoa(len(h.hub.Rooms))
	req.ID = newRoomID

	owner := &Client{
		Conn:   nil,
		Stream: make(chan *VideoMessage),
		ID:     req.OwnerID,
		RoomID: newRoomID,
	}

	h.hub.Rooms[newRoomID] = &Room{
		ID:              newRoomID,
		Name:            req.Name,
		Owner:           owner,
		Clients:         make(map[string]*Client),
		StreamBroadcast: make(chan *VideoMessage),
	}

	go h.hub.Rooms[newRoomID].Run()

	c.JSON(http.StatusOK, req)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		//origin := r.Header.Get("Origin")
		//return origin == url_del_snap_middle_end
		return true
	},
}

func (h *Handler) JoinRoom(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	roomID := c.Param("roomId")
	clientID := c.Query("Id")

	var cl *Client

	if clientID == h.hub.Rooms[roomID].Owner.ID {
		cl = h.hub.Rooms[roomID].Owner
		cl.Conn = conn
		go h.hub.Rooms[roomID].Owner.readStream(h.hub)
	} else {
		cl = &Client{
			Conn:   conn,
			Stream: make(chan *VideoMessage),
			ID:     clientID,
			RoomID: roomID,
		}
		h.hub.Register <- cl
		go cl.writeStream()
		go cl.readStream(h.hub)
	}
}

type RoomResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OwnerName string `json:"ownerName"`
}

func (h *Handler) GetRooms(c *gin.Context) {
	rooms := make([]RoomResponse, 0)

	for _, r := range h.hub.Rooms {
		rooms = append(rooms, RoomResponse{
			ID:   r.ID,
			Name: r.Name,
		})
	}
	c.JSON(http.StatusOK, rooms)
}

type ClientResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (h *Handler) GetClients(c *gin.Context) {
	var clients []ClientResponse
	roomId := c.Param("roomId")

	if _, ok := h.hub.Rooms[roomId]; !ok {
		clients = make([]ClientResponse, 0)
		c.JSON(http.StatusOK, clients)
	}

	for _, c := range h.hub.Rooms[roomId].Clients {
		clients = append(clients, ClientResponse{
			ID: c.ID,
		})
	}
	c.JSON(http.StatusOK, clients)
}
