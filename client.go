package main

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WRITE_WAIT       = 30 * time.Second
	PONG_WAIT        = 60 * time.Second
	PING_PERIOD      = (PONG_WAIT * 9) / 10
	MAX_MESSAGE_SIZE = 512
)

type Client struct {
	connection *websocket.Conn
	hub        *Hub
	sendch     chan []byte
}

func NewClient(conn *websocket.Conn, hub *Hub, sendch chan []byte) *Client {
	return &Client{
		connection: conn,
		hub:        hub,
		sendch:     sendch,
	}
}

func (c *Client) read() {
	defer func() {
		c.hub.close(c)
		c.hub.deregister(c)
	}()

	c.connection.SetReadLimit(MAX_MESSAGE_SIZE)
	c.connection.SetReadDeadline(time.Now().Add(PONG_WAIT))
	c.connection.SetPongHandler(func(appData string) error {
		slog.Debug("pong handler", "read dead line limit", "reseted", "to", c.connection.RemoteAddr())
		err := c.connection.SetReadDeadline(time.Now().Add(PONG_WAIT))
		if err != nil {
			slog.Error("failed to set readDeadLine", "error", err)
		}
		return err
	})

	for {
		_, message, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("read message", "read message error", err)
			}
			break
		}

		slog.Info("read message", "message", string(message), "from", c.connection.RemoteAddr())

		for wsClient := range c.hub.clients {
			wsClient.sendch <- message
		}
	}
}

func (c *Client) write() {
	defer func() {
		c.hub.close(c)
		c.hub.deregister(c)
	}()

	ticker := time.NewTicker(PING_PERIOD)

	for {
		select {
		case message, ok := <-c.sendch:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					slog.Error("write message", "connection status", "closed")
				}
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				slog.Error("write message", "failed to send message", err)
				return
			}

			slog.Error("write message", "message sent", string(message), "to", c.connection.RemoteAddr())
		case <-ticker.C:
			c.connection.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err := c.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			slog.Debug("write ping", "status", "sent", "to", c.connection.RemoteAddr())
		}
	}
}
