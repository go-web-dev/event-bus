//+build integration

package integration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
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

type JSON = map[string]interface{}

type response struct {
	Operation string          `json:"operation"`
	Status    bool            `json:"status"`
	Reason    string          `json:"reason,omitempty"`
	Body      json.RawMessage `json:"body,omitempty"`
	Context   interface{}     `json:"context,omitempty"`
}

type appSuite struct {
	testutils.Suite
	server *server.Server
	auth   string
	cfg    *config.Manager
	db     *badger.DB
	bus    *services.Bus
}

func (s *appSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), nil)
	cfg, err := config.NewManager(testCfgFile)
	s.Require().NoError(err)
	s.cfg = cfg
	s.db = testutils.NewBadger(s.T())
	s.bus = services.NewBus(s.db)
	s.Require().NoError(s.bus.Init())
	router := controllers.NewRouter(s.bus, cfg)
	settings := server.Settings{
		Addr:   addr,
		Router: router,
		DB:     s.db,
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

func (s *appSuite) TearDownTest() {
	s.Require().NoError(s.db.DropAll())
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
		req += fmt.Sprintf(`, "body": %s`, body)
	}
	req += "}\n"

	_, err := conn.Write([]byte(req))
	s.Require().NoError(err)
}

func (s *appSuite) read(conn net.Conn, res interface{}) {
	bs, err := bufio.NewReader(conn).ReadBytes('\n')
	s.Require().NoError(err)
	s.JSONUnmarshal(bs, res)
}

func (s *appSuite) dbGet(key []byte) string {
	var v string
	s.Require().NoError(s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			v = string(val)
			return nil
		})
	}))
	return v
}

func (s *appSuite) dbSet(key []byte, value []byte) {
	s.Require().NoError(s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	}))
}

func Test_App(t *testing.T) {
	suite.Run(t, new(appSuite))
}
