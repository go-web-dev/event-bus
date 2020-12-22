package server

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"

	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/testutils"
)

var (
	errTest = errors.New("some test error")
)

type serverSuite struct {
	testutils.Suite
	db          *dbMock
	router      *routerMock
	settings    Settings
	server      *Server
	loggerEntry zapcore.Entry
}

func (s *serverSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), &s.loggerEntry)
	s.db = new(dbMock)
	s.router = new(routerMock)
	s.settings = Settings{
		Addr:   "localhost:9000",
		DB:     s.db,
		Router: s.router,
	}
}

func (s *serverSuite) SetupTest() {
	srv, err := ListenAndServe(s.settings)
	s.Require().NoError(err)
	s.server = srv
	s.Require().NotNil(s.server)
	s.Equal(s.db, s.server.db)
	s.Equal(s.router, s.server.router)
}

func (s *serverSuite) TearDownTest() {
	s.db.AssertExpectations(s.T())
	s.router.AssertExpectations(s.T())
	s.Require().NoError(s.server.listener.Close())
}

func (s *serverSuite) Test_Server_OpenConnections() {
	s.router.
		On("Switch", mock.AnythingOfType("*net.TCPConn"), mock.AnythingOfType("*bytes.Reader")).
		Return(false, nil).
		Twice()

	conn1 := s.newConn(s.settings.Addr)
	s.connWrite(conn1, `{"operation": "conn1"}`)
	conn2 := s.newConn(s.settings.Addr)
	s.connWrite(conn2, `{"operation": "conn2"}`)

	s.waitForConnections(s.server, 2)
	s.Equal(2, len(s.server.connections))
}

func (s *serverSuite) Test_Server_ClosedConnections() {
	s.router.
		On("Switch", mock.AnythingOfType("*net.TCPConn"), mock.AnythingOfType("*bytes.Reader")).
		Return(true, nil).
		Twice()

	conn1 := s.newConn(s.settings.Addr)
	s.connWrite(conn1, `{"operation": "conn1"}`)
	conn2 := s.newConn(s.settings.Addr)
	s.connWrite(conn2, `{"operation": "conn2"}`)

	s.waitForConnections(s.server, 0)
	s.Equal(0, len(s.server.connections))
}

func (s *serverSuite) Test_Server_Stop_Success() {
	s.db.
		On("Close").
		Return(nil).
		Once()

	srv, err := ListenAndServe(Settings{
		Addr:   "localhost:8080",
		DB:     s.db,
		Router: s.router,
	})

	s.Require().NoError(err)
	s.Require().NoError(srv.Stop())
}

func (s *serverSuite) Test_Server_Stop_Error() {
	s.db.
		On("Close").
		Return(errTest).
		Once()

	srv, err := ListenAndServe(Settings{
		Addr:   "localhost:9090",
		DB:     s.db,
		Router: s.router,
	})

	s.Require().NoError(err)
	s.Equal(errTest, srv.Stop())
}

func (s *serverSuite) Test_Server_Switch_Error() {
	s.router.
		On("Switch", mock.AnythingOfType("*net.TCPConn"), mock.AnythingOfType("*bytes.Reader")).
		Return(false, errTest).
		Once()

	conn1 := s.newConn(s.settings.Addr)
	s.connWrite(conn1, `{"operation": "conn1"}`)

	s.waitForConnections(s.server, 1)
	s.Equal("switch error", s.loggerEntry.Message)
}

func (s *serverSuite) Test_Server_Switch_CloseConnectionsError() {
	s.db.
		On("Close").
		Return(nil).
		Once()

	srv, err := ListenAndServe(Settings{
		Addr:   "localhost:6000",
		DB:     s.db,
		Router: s.router,
	})
	s.Require().NoError(err)
	tcpListener := srv.listener.(*net.TCPListener)
	s.Require().NoError(tcpListener.SetDeadline(time.Now().Add(1 * time.Minute)))

	_ = s.newConn("localhost:6000")
	s.NoError(srv.Stop())
	s.waitForConnections(srv, 0)
	s.Equal("could not close connection", s.loggerEntry.Message)
}

func (s *serverSuite) Test_Server_ListenError() {
	srv, err := ListenAndServe(Settings{
		Addr:   "9000",
		Router: s.router,
		DB:     s.db,
	})

	s.EqualError(err, "listen tcp: address 9000: missing port in address")
	s.Nil(srv)
}

func (s *serverSuite) newConn(addr string) net.Conn {
	conn, err := net.Dial("tcp", addr)
	s.Require().NoError(err)
	return conn
}

func (s *serverSuite) connWrite(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message + "\n"))
	s.Require().NoError(err)
}

func (s *serverSuite) waitForConnections(srv *Server, connCount int) {
	timeout := time.After(1 * time.Second)
	tick := time.Tick(10 * time.Millisecond)
	for {
		select {
		case <-timeout:
			s.Fail(fmt.Sprintf("waiting for %d connections timed out", connCount))
			return
		case <-tick:
			if len(srv.connections) == connCount {
				return
			}
		}
	}
}

func Test_ServerSuite(t *testing.T) {
	suite.Run(t, new(serverSuite))
}
