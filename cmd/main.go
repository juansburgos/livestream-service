package main

import (
	"livestream-service/internal/logger"
	"livestream-service/internal/router"
	"livestream-service/internal/ws"
)

func main() {
	var address = "127.0.0.1:8080"
	logger.Get().Print("Starting app on port %s", address)

	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub)
	go hub.Run()

	router.InitRouter(wsHandler)
	err := router.Start(address)

	if err != nil {
		logger.Get().Printf("Error on router start: %s", err)
		return
	}
}
