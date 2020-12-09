package server

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"time"

	"github.com/chill-and-code/event-bus/controllers"
	"github.com/chill-and-code/event-bus/logging"
	"github.com/chill-and-code/event-bus/services"

	"go.uber.org/zap"
)

type switcher interface {
	Switch(io.Writer, io.Reader) (bool, error)
}

type Server struct {
	listener    net.Listener
	quit        chan struct{}
	exited      chan struct{}
	connections map[int]net.Conn
	router      switcher
}

func ListenAndServe(addr string) *Server {
	li, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("could not create server listener:", err)
	}
	checkpoint := services.NewCheckpoint()
	bus := services.NewBus(checkpoint)
	router := controllers.NewRouter(bus)
	srv := &Server{
		listener:    li,
		quit:        make(chan struct{}),
		exited:      make(chan struct{}),
		connections: map[int]net.Conn{},
		router:      router,
	}
	go srv.serve()
	return srv
}

func (srv *Server) Stop() {
	logger := logging.Logger
	logger.Info("stopping the event bus server")
}

func (srv *Server) serve() {
	logger := logging.Logger
	var id int
	logger.Info(
		"event bus server is up and running on address",
		zap.String("addr", srv.listener.Addr().String()),
	)
	for {
		select {
		case <-srv.quit:
			logger.Info("shutting down the event bus server")
			err := srv.listener.Close()
			if err != nil {
				logger.Error("could not close listener", zap.Error(err))
			}
			srv.closeConnections()
			close(srv.exited)
			return
		default:
			tcpListener := srv.listener.(*net.TCPListener)
			err := tcpListener.SetDeadline(time.Now().Add(2 * time.Second))
			if err != nil {
				logger.Error("failed to set listener deadline", zap.Error(err))
			}

			conn, err := tcpListener.Accept()
			if oppErr, ok := err.(*net.OpError); ok && oppErr.Timeout() {
				continue
			}
			if err != nil {
				logger.Error("failed to accept connection", zap.Error(err))
				return
			}

			srv.connections[id] = conn
			go func(connID int) {
				logger.Info("client joined", zap.Int("client_id", connID))
				srv.handle(conn)
				delete(srv.connections, connID)
				closeConn(conn, connID)
				logger.Info("client left", zap.Int("client_id", connID))
			}(id)
			id++
		}
	}
}

func (srv *Server) handle(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	logger := logging.Logger
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue
		}

		exited, err := srv.router.Switch(conn, bytes.NewReader(scanner.Bytes()))
		if err != nil {
			logger.Debug("switch error", zap.Error(err))
		}
		if exited {
			break
		}
	}
}

func (srv *Server) closeConnections() {
	logging.Logger.Info("closing all connections")
	for id, conn := range srv.connections {
		closeConn(conn, id)
	}
}

func closeConn(conn net.Conn, connID int) {
	logger := logging.Logger
	err := conn.Close()
	if err != nil {
		logger.Debug(
			"could not close connection",
			zap.Int("client_id", connID),
		)
	}
}
