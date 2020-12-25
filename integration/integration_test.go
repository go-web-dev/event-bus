//+build integration

package integration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/suite"

	"github.com/go-web-dev/event-bus/config"
	"github.com/go-web-dev/event-bus/controllers"
	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/models"
	"github.com/go-web-dev/event-bus/server"
	"github.com/go-web-dev/event-bus/services"
	"github.com/go-web-dev/event-bus/testutils"
)

const (
	addr        = "localhost:8000"
	testCfgFile = "./config_integration_test.yaml"
	testTimeStr = "2020-12-15T05:28:31.490416Z"
)

var (
	testTime, _ = time.Parse(time.RFC3339, testTimeStr)
	s1          = models.Stream{
		ID:        "s1-id",
		Name:      "s1-name",
		CreatedAt: testTime,
	}
	s2 = models.Stream{
		ID:        "s2-id",
		Name:      "s2-name",
		CreatedAt: testTime,
	}
	s3 = models.Stream{
		ID:        "s3-id",
		Name:      "s3-name",
		CreatedAt: testTime,
	}
	s1evt1 = models.Event{
		ID:        "evt1-id",
		StreamID:  s1.ID,
		Status:    0,
		CreatedAt: testTime,
		Body:      []byte(`{"f1":"v1"}`),
	}
	s1evt2 = models.Event{
		ID:        "evt2-id",
		StreamID:  s1.ID,
		Status:    0,
		CreatedAt: testTime,
		Body:      []byte(`{"f2":"v2"}`),
	}
	s1evt3 = models.Event{
		ID:        "evt3-id",
		StreamID:  s1.ID,
		Status:    1,
		CreatedAt: testTime,
		Body:      []byte(`{"f3":"v3"}`),
	}
	s1evt4 = models.Event{
		ID:        "evt4-id",
		StreamID:  s1.ID,
		Status:    2,
		CreatedAt: testTime,
		Body:      []byte(`{"f4":"v4"}`),
	}
	s2evt1 = models.Event{
		ID:        "evt1-id",
		StreamID:  s2.ID,
		Status:    0,
		CreatedAt: testTime,
		Body:      []byte(`{"f1":"v1"}`),
	}
	operations = []string{
		"create_stream",
		"delete_stream",
		"get_stream_info",
		"get_stream_events",
		"write_event",
		"mark_event",
		"process_events",
		"retry_events",
	}
	opNotFoundReason = "operation must be one of: 'health', 'create_stream', 'delete_stream', 'get_stream_info', 'get_stream_events', 'write_event', 'process_events', 'retry_events', 'retry_events', 'mark_event', 'exit'"
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
	wg     sync.WaitGroup
}

func (s *appSuite) SetupSuite() {
	logging.Logger = testutils.Logger(s.T(), nil)
	cfg, err := config.NewManager(testCfgFile)
	s.Require().NoError(err)
	s.cfg = cfg
	s.db = testutils.NewBadger(s.T())
	s.bus = services.NewBus(s.db)
	router := controllers.NewRouter(s.bus, cfg)
	settings := server.Settings{
		Addr:     addr,
		Router:   router,
		DB:       s.db,
		Deadline: 500 * time.Millisecond,
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
	s.Require().NoError(s.db.DropAll())
	s.dbInit()
	s.Require().NoError(s.bus.Init())
}

func (s *appSuite) TearDownSuite() {
	s.Require().NoError(s.server.Stop())
}

func (s *appSuite) Test_Health() {
	conn := s.newConn()

	s.write(conn, "health", "")

	var res response
	s.read(conn, &res)
	s.Equal("health", res.Operation)
	s.True(res.Status)
	s.Empty(res.Body)
}

func (s *appSuite) Test_Exit() {
	conn := s.newConn()

	s.write(conn, "exit", "")

	var res response
	s.read(conn, &res)
	s.Equal("exit", res.Operation)
	s.True(res.Status)
	s.Empty(res.Body)
}

func (s *appSuite) Test_Operations_DecodeRequestError() {
	conn := s.newConn()
	ops := append(operations, "health", "exit")
	for _, op := range ops {
		s.Run(op, func() {
			_, err := conn.Write([]byte("{}\n"))
			s.Require().NoError(err)

			var res response
			s.read(conn, &res)
			s.False(res.Status)
			s.Equal("decode_request", res.Operation)
			s.Equal(opNotFoundReason, res.Reason)
			s.Empty(res.Body)
			s.Empty(res.Context)
		})
	}
}

func (s *appSuite) Test_OperationNotFound() {
	conn := s.newConn()

	s.write(conn, "operation_does_not_exist", "")

	var res response
	s.read(conn, &res)
	s.False(res.Status)
	s.Equal("decode_request", res.Operation)
	s.Equal(opNotFoundReason, res.Reason)
	s.Empty(res.Body)
}

func (s *appSuite) Test_Operations_UnAuthorizedError() {
	s.auth = ""
	conn := s.newConn()
	for _, op := range operations {
		s.Run(op, func() {
			s.write(conn, op, "")

			var res response
			s.read(conn, &res)
			s.False(res.Status)
			s.Equal(op, res.Operation)
			s.Equal("unauthorized to make request", res.Reason)
			s.Empty(res.Body)
			s.Empty(res.Context)
		})
	}
}

func (s *appSuite) Test_Operations_ParseRequestBody_MissingFieldsError() {
	requests := []struct {
		operation string
		context   map[string]interface{}
	}{
		{
			operation: "create_stream",
			context: JSON{
				"body": []interface{}{
					JSON{
						"name":     "stream_name",
						"type":     "string",
						"required": true,
					},
				},
			},
		},
		{
			operation: "delete_stream",
			context: JSON{
				"body": []interface{}{
					JSON{
						"name":     "stream_name",
						"type":     "string",
						"required": true,
					},
				},
			},
		},
		{
			operation: "get_stream_info",
			context: JSON{
				"body": []interface{}{
					JSON{
						"name":     "stream_name",
						"type":     "string",
						"required": true,
					},
				},
			},
		},
		{
			operation: "get_stream_events",
			context: JSON{
				"body": []interface{}{
					JSON{
						"name":     "stream_name",
						"type":     "string",
						"required": true,
					},
				},
			},
		},
		{
			operation: "write_event",
			context: JSON{
				"body": []interface{}{
					JSON{
						"name":     "stream_name",
						"type":     "string",
						"required": true,
					},
					JSON{
						"name":     "event",
						"type":     "[]byte",
						"required": true,
					},
				},
			},
		},
		{
			operation: "mark_event",
			context: JSON{
				"body": []interface{}{
					JSON{
						"name":     "event_id",
						"type":     "string",
						"required": true,
					},
					JSON{
						"name":     "status",
						"type":     "uint8",
						"required": true,
					},
				},
			},
		},
		{
			operation: "process_events",
			context: JSON{
				"body": []interface{}{
					JSON{
						"name":     "stream_name",
						"type":     "string",
						"required": true,
					},
				},
			},
		},
		{
			operation: "retry_events",
			context: JSON{
				"body": []interface{}{
					JSON{
						"name":     "stream_name",
						"type":     "string",
						"required": true,
					},
				},
			},
		},
	}
	conn := s.newConn()
	for _, req := range requests {
		s.Run(req.operation, func() {
			s.write(conn, req.operation, "")

			var res response
			s.read(conn, &res)
			s.False(res.Status)
			s.Equal(req.operation, res.Operation)
			s.Equal("missing required fields", res.Reason)
			s.Equal(req.context, res.Context)
			s.Empty(res.Body)
		})
	}
}

func (s *appSuite) Test_Operations_ParseRequestBody_DecodeJSONError() {
	s.auth = ""
	conn := s.newConn()
	for _, op := range operations {
		s.Run(op, func() {
			s.write(conn, op, "{")

			var res response
			s.read(conn, &res)
			s.False(res.Status)
			s.Equal("decode_request", res.Operation)
			s.Equal("invalid json provided", res.Reason)
			s.Empty(res.Context)
			s.Empty(res.Body)
		})
	}
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

func (s *appSuite) dbGet(key []byte) (string, error) {
	var v string
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			v = string(val)
			return nil
		})
	})
	if err != nil {
		return "", err
	}
	return v, nil
}

func (s *appSuite) dbSet(key []byte, value []byte) {
	s.Require().NoError(s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	}))
}

func (s *appSuite) dbInit() {
	s.dbSet(s1.Key(), s1.Value())
	s.dbSet(s2.Key(), s2.Value())
	s.dbSet(s3.Key(), s3.Value())
	s.dbSet(s1evt1.Key(s1evt1.Status), s1evt1.Value())
	s.dbSet(s1evt2.Key(s1evt2.Status), s1evt2.Value())
	s.dbSet(s1evt3.Key(s1evt3.Status), s1evt3.Value())
	s.dbSet(s1evt4.Key(s1evt4.Status), s1evt4.Value())
	s.dbSet(s2evt1.Key(s2evt1.Status), s2evt1.Value())
}

func Test_App(t *testing.T) {
	suite.Run(t, new(appSuite))
}
