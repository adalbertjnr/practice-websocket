package main

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/gorilla/websocket"
)

func main() {
	serverUrl := "ws://localhost:5000/ws"

	conn, _, err := websocket.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		log.Fatalf("failed to connect to the server: %s. err: %v", serverUrl, conn)
	}

	conn.SetPingHandler(func(appData string) error {
		slog.Info("client ping handler", "ping from server", "received")
		err := conn.WriteMessage(websocket.PongMessage, []byte(appData))
		if err != nil {
			slog.Error("failed to send pong message to the server", "err", err)
		}
		return err
	})

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

	done := make(chan struct{})

	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			slog.Info("closing connection")
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client shutting down"))
			conn.Close()

			close(done)
		})
	}

	defer cleanup()

	go func(done chan struct{}) {
		defer func() { cleanup() }()

		slog.Info("read message func", "status", "started")
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				slog.Error("read message", "error", err)
				return
			}

			slog.Info("[server]", "message", string(message), "from", conn.RemoteAddr())
		}
	}(done)

	read := bufio.NewReader(os.Stdin)

	go func(done chan struct{}, r *bufio.Reader) {
		defer func() { cleanup() }()

		for {
			fmt.Print("> ")
			message, err := read.ReadString('\n')
			if err != nil {
				slog.Error("bufio read message from stdin", "error", err)
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				slog.Error("write message", "error", err, "to", conn.RemoteAddr())
			}
		}
	}(done, read)

	for {
		select {
		case sig := <-sigch:
			slog.Info("sigch", "got", sig.String(), "connection status", "closing")
			cleanup()
		case <-done:
			slog.Info("done", "connection status", "closed")
			return
		}
	}

}
