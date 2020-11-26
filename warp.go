package pool

import "net"

func (c *channelPool) wrapConn(conn net.Conn) net.Conn  {
	p := &PoolConn{
		c: c,
	}
	p.Conn = conn
	return p
}