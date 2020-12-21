//+build integration

package integration

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/go-web-dev/event-bus/config"
	"github.com/go-web-dev/event-bus/controllers"
	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/server"
	"github.com/go-web-dev/event-bus/services"
	"github.com/go-web-dev/event-bus/testutils"
)

const (
	addr        = "localhost:8000"
	testCfgFile = "./config_integration_test.yaml"
)

type appSuite struct {
	testutils.Suite
	server *server.Server
	auth   string
	cfg    *config.Manager
}

func (s *appSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), nil)
	cfg, err := config.NewManager(testCfgFile)
	s.Require().NoError(err)
	s.cfg = cfg
	db := testutils.NewBadger(s.T())
	bus := services.NewBus(db)
	s.Require().NoError(bus.Init())
	router := controllers.NewRouter(bus, cfg)
	settings := server.Settings{
		Addr:   addr,
		Router: router,
		DB:     db,
	}
	srv, err := server.ListenAndServe(settings)
	s.Require().NoError(err)
	s.server = srv
	s.waitForServer()
}

func (s *appSuite) SetupTest() {
	auth := s.cfg.GetAuth()
	integrationClient, ok := auth["integration"]
	s.Require().True(ok)
	s.auth = fmt.Sprintf(
		`{"client_id": "%s", "client_secret": "%s"}`,
		integrationClient.ClientID,
		integrationClient.ClientSecret,
	)
}

func (s *appSuite) TearDownSuite() {
	s.Require().NoError(s.server.Stop())
}

func (s *appSuite) waitForServer() {
	timeout := time.After(3 * time.Second)
	tick := time.Tick(50 * time.Millisecond)
	for {
		select {
		case <-timeout:
			s.Fail("could not connect to server listener")
		case <-tick:
			conn, err := net.Dial("tcp", addr)
			if err == nil {
				s.Require().NoError(conn.Close())
				return
			}
		}
	}
}

func (s *appSuite) newConn() net.Conn {
	conn, err := net.Dial("tcp", addr)
	s.Require().NoError(err)
	return conn
}

func (s *appSuite) write(conn net.Conn, operation, body string) {
	req := fmt.Sprintf(`{"operation": "%s"`, operation)
	if s.auth != "" {
		req += fmt.Sprintf(`, "auth": %s`, s.auth)
	}
	if body != "" {
		req += fmt.Sprintf(`, "body": %s`, s.auth)
	}
	req += "}\n"

	_, err := conn.Write([]byte(req))
	s.Require().NoError(err)
}

func Test_App(t *testing.T) {
	suite.Run(t, new(appSuite))
}
