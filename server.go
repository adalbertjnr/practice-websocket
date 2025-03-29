package main

import (
	"log/slog"
	"net/http"
)

func handleWS(hubManager *Hub, w http.ResponseWriter, r *http.Request) {
	slog.Info("websocket", "status", "new connection")

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket", "status err", err)
		return
	}

	sendChannel := make(chan []byte)
	c := NewClient(conn, hubManager, sendChannel)

	hubManager.register(c)

	go c.read()
	go c.write()
}

type wrapperFunc func(hub *Hub, w http.ResponseWriter, r *http.Request)

func wrap(hub *Hub, wrapperFn wrapperFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wrapperFn(hub, w, r)
	}
}
