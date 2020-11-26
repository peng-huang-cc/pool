package pool

import (
	"net"
	"sync"
)

type PoolConn struct {
	net.Conn
	mu sync.RWMutex // read more than write
	c *channelPool
	unusable bool
}

// Close put the conn back to the pool instead of closing it
// 1. lock the conn
// 2. if usable close, else put it back to pool
// 3. unlock it
func (p *PoolConn) Close() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.unusable {
		if p.Conn != nil {
			p.Conn.Close()
		}
	}
	return p.c.put(p.Conn)
}

func (p *PoolConn) MarkUnusable()  {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}
