package server

import (
	"github.com/go-web-dev/event-bus/logging"
	"go.uber.org/zap"
	"net"
	"sync"
)

type connection struct {
	id   int
	conn net.Conn
}

type connections struct {
	connMap map[int]*connection
	mu      sync.RWMutex
	index   int
}

func (c *connections) add(conn net.Conn) *connection {
	logger := logging.Logger
	c.mu.Lock()
	defer c.mu.Unlock()
	res := &connection{
		id:   c.index,
		conn: conn,
	}
	c.connMap[c.index] = res
	logger.Info("client joined", zap.Int("client_id", c.index))
	c.index++
	return res
}

func (c *connections) close(connID int) {
	logger := logging.Logger
	c.mu.Lock()
	defer c.mu.Unlock()
	connection, ok := c.connMap[connID]
	if !ok {
		return
	}
	conn := connection.conn

	err := conn.Close()
	if err != nil {
		logger.Error(
			"could not close connection",
			zap.Int("client_id", connID),
			zap.Error(err),
		)
	}
	delete(c.connMap, connID)
	logger.Info("client left", zap.Int("client_id", connID))
}

func (c *connections) closeAll() {
	for id := range c.connMap {
		c.close(id)
	}
}
