package server

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/go-web-dev/event-bus/logging"
)

type router interface {
	Switch(io.Writer, io.Reader) (bool, error)
}

// Settings represents the Event Bus server settings
type Settings struct {
	Addr   string
	Router router
	DB     io.Closer
}

// Server represents the Event Bus TCP server
type Server struct {
	listener    net.Listener
	quit        chan struct{}
	exited      chan struct{}
	connections map[int]net.Conn
	router      router
	db          io.Closer
}

// ListenAndServe spins up the Event Bus TCP server
func ListenAndServe(settings Settings) *Server {
	li, err := net.Listen("tcp", settings.Addr)
	if err != nil {
		log.Fatal("could not create server listener:", err)
	}
	srv := &Server{
		listener:    li,
		quit:        make(chan struct{}),
		exited:      make(chan struct{}),
		connections: map[int]net.Conn{},
		router:      settings.Router,
		db:          settings.DB,
	}
	go srv.serve()
	return srv
}

// Stop is responsible for cleanup process before application server shutdown
func (srv *Server) Stop() error {
	logger := logging.Logger
	close(srv.quit)
	logger.Info("stopping the database")
	err := srv.db.Close()
	if err != nil {
		return err
	}
	<-srv.exited
	return nil
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
				srv.closeConn(conn, connID)
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
		srv.closeConn(conn, id)
	}
}

func (srv *Server) closeConn(conn net.Conn, connID int) {
	logger := logging.Logger
	err := conn.Close()
	if err != nil {
		logger.Debug(
			"could not close connection",
			zap.Int("client_id", connID),
		)
	}
}
