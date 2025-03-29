package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	var serverPort string
	var debug bool

	flag.BoolVar(&debug, "debug", true, "debug option")
	flag.StringVar(&serverPort, "port", ":5000", "default server port")
	flag.Parse()

	var logLevel slog.Level
	if debug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	muxRouter := http.NewServeMux()

	hubManager := &Hub{clients: make(map[*Client]bool)}
	go hubManager.Run()

	muxRouter.HandleFunc("/ws", wrap(hubManager, handleWS))

	srv := &http.Server{
		Addr:    serverPort,
		Handler: muxRouter,
	}

	go func() {
		slog.Info("server", "started on port", serverPort)
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("error starting the server on port: %s err: %v", serverPort, err)
			}
		}
	}()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

	sig := <-sigch
	slog.Info("signal", "got", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
