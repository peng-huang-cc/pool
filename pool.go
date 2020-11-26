package pool

import (
	"errors"
	"net"
)

var (
	ErrClosed = errors.New("pool is closed")
)

type Pool interface {
	// Get return a usable collection form the pool. Closing connections will back to the pool.
	// Closing it when the pool is destroyed or full will be counted as an error
	Get() (net.Conn, error)

	// Close closes the pool and all its connections. After Close() the pool is no longer usable
	Close()

	// Len returns the current number of connections of the pool.
	Len() int
}
