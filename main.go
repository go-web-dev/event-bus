package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chill-and-code/event-bus/logging"
	"github.com/chill-and-code/event-bus/server"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	err := logging.Init()
	if err != nil {
		log.Fatal("could not initialize logger:", err)
	}

	srv := server.ListenAndServe("localhost:8080")

	select {
	case <-stop:
		srv.Stop()
	}
}
