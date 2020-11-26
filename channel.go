package pool

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

var (
	InvalidCapSetting = errors.New("invalid capacity settings")
)

// channelPool implements the Pool interface based on buffered channels.
type channelPool struct {
	// storage our net.Conn connections
	mu      sync.RWMutex
	conns   chan net.Conn
	factory Factory
}

// Factory create new connections
type Factory func() (net.Conn, error)

func NewChannelPool(InitCap int, MaxCap int, factory Factory) (Pool, error) {
	if InitCap < 0 {
		return nil, InvalidCapSetting
	}
	c := &channelPool{
		conns:   make(chan net.Conn, MaxCap),
		factory: factory,
	}
	for i := 0; i < InitCap; i++ {
		conn, err := factory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.conns <- conn
	}
	return c, nil
}

func (c *channelPool) getConnectionsAndFactory() (chan net.Conn, Factory) {
	c.mu.RLock()
	conns := c.conns
	factory := c.factory
	c.mu.RUnlock()
	return conns, factory
}

// Get return a usable connection, if there is no new conn available in the pool,
// a new connection will be created via the Factory() method.
func (c *channelPool) Get() (net.Conn, error) {
	conns, factory := c.getConnectionsAndFactory()
	if conns == nil {
		return nil, ErrClosed
	}

	// remove and
	select {
	case conn := <-conns:
		// coon is closed
		if conn == nil {
			return nil, ErrClosed
		}
		return c.wrapConn(conn), nil
	default:
		conn, err := factory()
		if err != nil {
			return nil, err
		}
		return c.wrapConn(conn), nil
	}
}

// put puts the connection back to the pool
// If the pool is full or closed, coon is simply closed.
// A nil conn will be rejected.
func (c *channelPool) put(coon net.Conn) error {
	if coon == nil {
		return errors.New("connection is nil")
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conns == nil {
		return coon.Close()
	}

	select {
	case c.conns <- coon:
		return nil
	default:
		// pool is full, close the connection
		return coon.Close()
	}
}

// Close closes the pool and destroy the connections
func (c *channelPool) Close() {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.mu.Unlock()

	if conns == nil {
		return
	}
	close(conns)
	for conn := range conns {
		conn.Close()
	}
}

func (c *channelPool) Len() int {
	conns, _ := c.getConnectionsAndFactory()
	return len(conns)
}
