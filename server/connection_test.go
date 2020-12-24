package server

import (
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"

	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/testutils"
)

type connectionSuite struct {
	testutils.Suite
	conns       connections
	li          net.Listener
	loggerEntry zapcore.Entry
}

func (s *connectionSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), &s.loggerEntry)
	li, err := net.Listen("tcp", "localhost:8080")
	s.Require().NoError(err)
	s.li = li
	go func() {
		_, err = li.Accept()
		s.Require().NoError(err)
	}()
}

func (s *connectionSuite) TearDownSuite() {
	s.Require().NoError(s.li.Close())
}

func (s *connectionSuite) SetupTest() {
	s.conns = connections{connMap: map[int]*connection{}}
}

func (s *connectionSuite) Test_add() {
	conn := s.newConn()
	expected := &connection{
		conn: conn,
		id:   0,
	}

	c := s.conns.add(conn)

	s.Equal(expected, c)
	s.Equal(expected, s.conns.connMap[0])
	s.Equal(1, s.conns.index)
}

func (s *connectionSuite) Test_close_Success() {
	conn := s.newConn()
	s.conns.connMap[0] = &connection{
		conn: conn,
		id:   0,
	}

	s.conns.close(0)

	s.Len(s.conns.connMap, 0)
}

func (s *connectionSuite) Test_close_ConnNotFound() {
	conn := s.newConn()
	s.conns.connMap[0] = &connection{
		conn: conn,
		id:   0,
	}

	s.conns.close(1)

	s.Len(s.conns.connMap, 1)
}

func (s *connectionSuite) Test_close_ConnCloseError() {
	conn := s.newConn()
	s.conns.connMap[0] = &connection{
		conn: conn,
		id:   0,
	}
	s.Require().NoError(conn.Close())

	s.conns.close(0)

	s.Len(s.conns.connMap, 0)
	s.Equal("could not close connection", s.loggerEntry.Message)
}

func (s *connectionSuite) Test_closeAll_Success() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	s.conns.connMap[0] = &connection{
		conn: conn1,
		id:   0,
	}
	s.conns.connMap[1] = &connection{
		conn: conn2,
		id:   1,
	}

	s.conns.closeAll()

	s.Len(s.conns.connMap, 0)
}

func (s *connectionSuite) Test_closeAll_ConnCloseError() {
	conn1 := s.newConn()
	conn2 := s.newConn()
	s.conns.connMap[0] = &connection{
		conn: conn1,
		id:   0,
	}
	s.conns.connMap[1] = &connection{
		conn: conn2,
		id:   1,
	}
	s.Require().NoError(conn1.Close())
	s.Require().NoError(conn2.Close())

	s.conns.closeAll()

	s.Len(s.conns.connMap, 0)
	s.Equal("could not close connection", s.loggerEntry.Message)
}

func (s *connectionSuite) newConn() net.Conn {
	conn, err := net.Dial("tcp", "localhost:8080")
	s.Require().NoError(err)
	return conn
}

func Test_ConnectionSuite(t *testing.T) {
	suite.Run(t, new(connectionSuite))
}
