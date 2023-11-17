package router

import (
	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	"livestream-service/internal/ws"
	"net/http"
)

var r *gin.Engine

func InitRouter(wsHandler *ws.Handler) {
	r = gin.Default()

	r.Use(cors.Middleware(cors.Config{
		Origins:        "*",
		Methods:        "GET, POST",
		RequestHeaders: "*",
	}))

	//De esto se deberia ocupar el front
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "Hello")
	})

	r.GET("/rooms", wsHandler.GetRooms)
	r.GET("/rooms/:roomId/clients", wsHandler.GetClients)

	r.POST("/rooms", wsHandler.CreateRoom)
	r.GET("/rooms/:roomId", wsHandler.JoinRoom)
}

func Start(addr string) error {
	return r.Run(addr)
}
