package main

import (
	"log/slog"
	"sync"
)

type Hub struct {
	clients map[*Client]bool
	sync.RWMutex
}

func (h *Hub) close(client *Client) {
	client.connection.Close()
}

func (h *Hub) register(client *Client) {
	h.Lock()
	defer h.Unlock()

	slog.Info("hub manager", "registering client", client.connection.RemoteAddr())
	h.clients[client] = true
}

func (h *Hub) deregister(client *Client) {
	h.Lock()
	defer h.Unlock()

	if _, found := h.clients[client]; found {
		slog.Info("hub manager", "deregistering client", client.connection.RemoteAddr())
		client.connection.Close()
		delete(h.clients, client)
	}
}

func (h *Hub) Run() {
	slog.Info("hub", "status", "started")
}
