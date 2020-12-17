package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-web-dev/event-bus/config"
	"github.com/go-web-dev/event-bus/controllers"
	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/server"
	"github.com/go-web-dev/event-bus/services"

	"github.com/dgraph-io/badger/v2"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "config file path")
	flag.Parse()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	err := logging.Init()
	if err != nil {
		log.Fatal("could not initialize logger: ", err)
	}

	cfg, err := config.NewManager(*configPath)
	if err != nil {
		log.Fatal("could not create config manager: ", err)
	}

	dbOptions := badger.DefaultOptions("badger")
	dbOptions.Logger = nil
	db, err := badger.Open(dbOptions)
	if err != nil {
		log.Fatal("could not open badger db: ", err)
	}

	bus := services.NewBus(db)
	err = bus.Init()
	if err != nil {
		log.Fatal("could not initialize event bus: ", err)
	}

	router := controllers.NewRouter(bus, cfg)

	serverSettings := server.Settings{
		Addr:   "localhost:8080",
		Router: router,
		DB:     db,
	}

	srv := server.ListenAndServe(serverSettings)

	select {
	case <-stop:
		err := srv.Stop()
		if err != nil {
			log.Fatal("could not stop server: ", err)
		}
	}
}
