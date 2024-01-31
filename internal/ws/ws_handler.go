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
	Name      string `json:"name"`
	OwnerName string `json:"ownerName"`
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar si el cliente ya existe
	var owner *Client
	existingOwner := h.getClientByName(req.OwnerName)
	if existingOwner != nil {
		owner = existingOwner
	} else {
		// Crear un nuevo cliente
		var ownerID = strconv.Itoa(len(h.hub.Rooms))
		owner = &Client{
			ID:     ownerID, // Generar un nuevo ID para el propietario
			Name:   req.OwnerName,
			Conn:   nil,
			Stream: make(chan *VideoMessage),
		}
		// Agregar el nuevo cliente a la lista
		h.hub.Clients[owner.ID] = owner
	}

	var newRoomID = strconv.Itoa(len(h.hub.Rooms))

	h.hub.Rooms[newRoomID] = &Room{
		ID:              newRoomID,
		Name:            req.Name,
		Owner:           owner,
		Clients:         make(map[string]*Client),
		StreamBroadcast: make(chan *VideoMessage),
	}

	go h.hub.Rooms[newRoomID].Run()

	c.JSON(http.StatusOK, gin.H{
		"roomId":    newRoomID,
		"name":      req.Name,
		"ownerId":   owner.ID,
		"ownerName": req.OwnerName,
	})
}

func (h *Handler) getClientByName(name string) *Client {
	for _, client := range h.hub.Clients {
		if client.Name == name {
			return client
		}
	}
	return nil
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
