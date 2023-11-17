package main

import (
	"livestream-service/internal/router"
	"livestream-service/internal/ws"
)

func main() {
	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub)

	go hub.Run()

	router.InitRouter(wsHandler)
	router.Start("127.0.0.1:8080")
}
