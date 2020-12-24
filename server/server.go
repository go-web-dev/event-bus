package server

import (
	"bufio"
	"bytes"
	"io"
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
	Addr     string
	Router   router
	DB       io.Closer
	Deadline time.Time
}

// Server represents the Event Bus TCP server
type Server struct {
	listener    net.Listener
	quit        chan struct{}
	exited      chan struct{}
	connections connections
	router      router
	db          io.Closer
	deadline    time.Time
}

// ListenAndServe spins up the Event Bus TCP server
func ListenAndServe(settings Settings) (*Server, error) {
	li, err := net.Listen("tcp", settings.Addr)
	if err != nil {
		return nil, err
	}
	srv := &Server{
		listener:    li,
		quit:        make(chan struct{}),
		exited:      make(chan struct{}),
		connections: connections{connMap: map[int]*connection{}},
		router:      settings.Router,
		db:          settings.DB,
		deadline:    settings.Deadline,
	}
	go srv.serve()
	return srv, nil
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
	logger.Info(
		"event bus server is up and running on address",
		zap.String("addr", srv.listener.Addr().String()),
	)
	for {
		select {
		case <-srv.quit:
			logger.Info("shutting down the event bus server")
			// avoid accepting new connections
			err := srv.listener.Close()
			if err != nil {
				logger.Error("could not close listener", zap.Error(err))
			}
			srv.connections.closeAll()
			close(srv.exited)
			return
		default:
			tcpListener := srv.listener.(*net.TCPListener)
			err := tcpListener.SetDeadline(srv.deadline)
			if srv.closedConnection(err) {
				return
			}
			if err != nil {
				logger.Error("failed to set listener deadline", zap.Error(err))
			}

			conn, err := tcpListener.Accept()
			if oppErr, ok := err.(*net.OpError); ok && oppErr.Timeout() {
				continue
			}
			if srv.closedConnection(err) {
				return
			}
			if err != nil {
				logger.Error("failed to accept connection", zap.Error(err))
				return
			}

			c := srv.connections.add(conn)
			go func() {
				srv.handle(conn)
				srv.connections.close(c.id)
			}()
		}
	}
}

func (srv *Server) closedConnection(err error) bool {
	if oppErr, ok := err.(*net.OpError); ok && oppErr.Unwrap().Error() == "use of closed network connection" {
		return true
	}
	return false
}

func (srv *Server) handle(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	logger := logging.Logger
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			logger.Error("empty request line")
			continue
		}

		exited, err := srv.router.Switch(conn, bytes.NewReader(scanner.Bytes()))
		if err != nil {
			logger.Error("switch error", zap.Error(err))
		}
		if exited {
			break
		}
	}
}
