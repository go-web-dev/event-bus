package server

import (
	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/testutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net"
	"testing"
	"time"
)

type serverSuite struct {
	testutils.Suite
	db     *dbMock
	router *routerMock
	server *Server
}

func (s *serverSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), nil)
	s.db = new(dbMock)
	s.router = new(routerMock)
}

func (s *serverSuite) TearDownTest() {
	s.db.AssertExpectations(s.T())
	s.router.AssertExpectations(s.T())
}

func (s *serverSuite) Test_Server_ListenAndServe_Success() {
	settings := Settings{
		Addr:   "localhost:9000",
		DB:     s.db,
		Router: s.router,
	}
	body := []byte(`{"some": "operation"}` + "\n")
	s.router.
		On("Switch", mock.AnythingOfType("*net.TCPConn"), mock.AnythingOfType("*bytes.Reader")).
		Return(true, nil).
		Once()

	srv := ListenAndServe(settings)

	s.Require().NotNil(srv)
	s.Equal(s.db, srv.db)
	s.Equal(s.router, srv.router)
	s.Equal(s.router, srv.router)
	s.Len(srv.connections, 0)

	conn := s.newConn(settings.Addr)
	conn.Write(body)
	// make the server block and handle it inside main
	time.Sleep(time.Millisecond * 100) // solve this
	close(s.server.exited)
}

func (s *serverSuite) newConn(addr string) net.Conn {
	conn, err := net.Dial("tcp", addr)
	s.Require().NoError(err)
	return conn
}

func Test_ServerSuite(t *testing.T) {
	suite.Run(t, new(serverSuite))
}
