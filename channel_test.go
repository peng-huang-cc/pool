package pool

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

var (
	InitCap = 5
	MaxCap = 30
	network = "tcp"
	address = "127.0.0.1:8888"
	factory = func() (net.Conn, error) { return  net.Dial(network, address) }
)

func init()  {
	go simpleTcpServer()
	time.Sleep(time.Second)
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestNewChannelPool(t *testing.T) {
	_, err := NewChannelPool(InitCap, MaxCap, factory)
	if err != nil {
		t.Errorf("New error: %s", err)
	}

	_, err = NewChannelPool(0, 10, factory)
	if err != nil {
		t.Errorf("New error: %s", err)
	}
}

func TestChannelPool_Get_Impl(t *testing.T) {
	p, _ := NewChannelPool(InitCap, MaxCap, factory)
	defer p.Close()

	conn, err := p.Get()
	if err != nil {
		t.Errorf("Get connection err:  %s", err)
	}
	_, ok := conn.(*PoolConn)
	if !ok {
		t.Errorf("Conn is not of type PoolConn")
	}
}

func TestChannelPool_Get(t *testing.T) {
	p, _ := NewChannelPool(InitCap, MaxCap, factory)
	defer p.Close()

	_, err := p.Get()
	if err != nil {
		t.Errorf("Get connection err:  %s", err)
	}

	expected := InitCap - 1
	actual := p.Len()
	if p.Len() != actual {
		t.Errorf("Get error. Expecting %d, got %d", expected, actual)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < InitCap - 1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := p.Get()
			l := p.Len()
			t.Logf("Get len %d", l)
			if err != nil {
				t.Errorf("Get error: %s", err)
			}
		}()
	}
	wg.Wait()

	if p.Len() != 0 {
		t.Errorf("Get error. Expecting 0, got %d", p.Len())
	}

	_, err = p.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}

}

func TestChannelPool_Put(t *testing.T) {
	p, _ := NewChannelPool(0, MaxCap, factory)
	defer p.Close()

	conns := make([]net.Conn, MaxCap)
	for i := 0; i < MaxCap; i++  {
		conn, _ := p.Get()
		conns[i] = conn
	}

	for _, conn := range conns {
		conn.Close()
	}

	if p.Len() != MaxCap {
		t.Errorf("Put error length. Expecting %d, got %d", 10, p.Len())
	}

	conn, _ := p.Get()
	p.Close()

	conn.Close()

	if p.Len() != 0 {
		t.Errorf("Put error. Closed pool shouldn't allow to put connections")
	}
}

func simpleTcpServer()  {
	l, err := net.Listen(network, address)
	if err != nil{
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			buffer := make([]byte, 256)
			conn.Read(buffer)
		}()
	}
}