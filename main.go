package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chill-and-code/event-bus/controllers"
	"github.com/chill-and-code/event-bus/logging"
	"github.com/chill-and-code/event-bus/server"
	"github.com/chill-and-code/event-bus/services"

	"github.com/dgraph-io/badger/v2"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	err := logging.Init()
	if err != nil {
		log.Fatal("could not initialize logger: ", err)
	}

	db, err := badger.Open(badger.DefaultOptions("badger"))
	if err != nil {
		log.Fatal("could not open badger db: ", err)
	}
	bus := services.NewBus(db)
	router := controllers.NewRouter(bus)

	serverSettings := server.Settings{
		Addr:   "localhost:8080",
		Router: router,
	}

	srv := server.ListenAndServe(serverSettings)

	select {
	case <-stop:
		srv.Stop()
	}
}
