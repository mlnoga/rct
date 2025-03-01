package rct

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// DialTimeout is the default cache for connecting to a RCT device
var DialTimeout = time.Second * 5

// Connection to a RCT device
type Connection struct {
	mu      sync.Mutex
	conn    net.Conn
	cache   *Cache
	broker  *Broker[Datagram]
	errCB   func(error)
	timeout time.Duration
}

// WithErrorCallback sets the error callback. It is only invoked after initial connection succeeds.
func WithErrorCallback(cb func(error)) func(*Connection) {
	return func(c *Connection) {
		c.errCB = cb
	}
}

// WithTimeout sets the query timeout
func WithTimeout(timeout time.Duration) func(*Connection) {
	return func(c *Connection) {
		c.timeout = timeout
	}
}

// Creates a new connection to a RCT device at the given address.
// Must not be called concurrently.
func NewConnection(ctx context.Context, host string, opt ...func(*Connection)) (*Connection, error) {
	conn := &Connection{
		cache:  NewCache(),
		broker: NewBroker[Datagram](),
	}

	for _, o := range opt {
		o(conn)
	}

	bufC := make(chan byte, 1024)
	errC := make(chan error, 1)

	go conn.receive(ctx, net.JoinHostPort(host, "8899"), bufC, errC)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errC:
		if err != nil {
			return nil, err
		}
	}

	go func() {
		for err := range errC {
			if conn.errCB != nil {
				conn.errCB(err)
			}
		}
	}()

	go conn.broker.Start(ctx)
	go ParseAsync(ctx, bufC, conn.broker.publishCh)
	go conn.handle(ctx, conn.broker.Subscribe(), errC)

	return conn, nil
}

// receive streams received bytes from the connection
func (c *Connection) receive(ctx context.Context, addr string, bufC chan<- byte, errC chan<- error) {
	buf := make([]byte, 1024)

	for {
		n, err := backoff.Retry(ctx, func() (int, error) {
			var err error

			c.mu.Lock()
			if c.conn == nil {
				var d net.Dialer

				ctx, _ := context.WithTimeout(ctx, DialTimeout)
				c.conn, err = d.DialContext(ctx, "tcp", addr)
				if err != nil {
					return 0, err
				}
			}
			conn := c.conn
			c.mu.Unlock()

			return conn.Read(buf)
		})
		if err != nil {
			c.mu.Lock()
			c.conn = nil
			c.mu.Unlock()

			errC <- err
			continue
		}

		// ack data received
		errC <- nil

		// stream received data
		for _, b := range buf[:n] {
			bufC <- b
		}
	}
}

// run is the receiver go routine
func (c *Connection) handle(ctx context.Context, dgC <-chan Datagram, errC chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		case dg := <-dgC:
			if dg.Cmd == Response || dg.Cmd == LongResponse {
				c.cache.Put(&dg)
			}
		}
	}
}

// Sends the given RCT datagram via the connection
func (c *Connection) Send(rdb *DatagramBuilder) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// ensure active connection
	if c.conn == nil {
		return 0, errors.New("disconnected")
	}

	n, err := c.conn.Write(rdb.Bytes())
	if err != nil {
		c.conn.Close()
		c.conn = nil
	}
	return n, err
}

func (c *Connection) Subscribe() chan Datagram {
	return c.broker.Subscribe()
}

func (c *Connection) Unsubscribe(ch chan Datagram) {
	c.broker.Unsubscribe(ch)
}

func (c *Connection) Get(id Identifier) (*Datagram, time.Time) {
	return c.cache.Get(id)
}

// Queries the given identifier on the RCT device, returning its value as a datagram
func (c *Connection) Query(id Identifier) (*Datagram, error) {
	if dg, ts := c.cache.Get(id); dg != nil && time.Since(ts) < c.timeout {
		return dg, nil
	}

	resC := make(chan Datagram, 1)
	data := c.broker.Subscribe()
	go func() {
		for dg := range data {
			if dg.Id == id {
				select {
				case resC <- dg:
				default:
				}
			}
		}
	}()
	defer c.broker.Unsubscribe(data)

	var rdb DatagramBuilder
	rdb.Build(&Datagram{Read, id, nil})
	if _, err := c.Send(&rdb); err != nil {
		return nil, err
	}

	select {
	case <-time.After(c.timeout):
		return nil, errors.New("timeout")
	case dg := <-resC:
		return &dg, nil
	}
}

// Queries the given identifier on the RCT device, returning its value as a float32
func (c *Connection) QueryFloat32(id Identifier) (float32, error) {
	dg, err := c.Query(id)
	if err != nil {
		return 0, err
	}
	return dg.Float32()
}

// Queries the given identifier on the RCT device, returning its value as a uint8
func (c *Connection) QueryInt32(id Identifier) (int32, error) {
	dg, err := c.Query(id)
	if err != nil {
		return 0, err
	}
	return dg.Int32()
}

// Queries the given identifier on the RCT device, returning its value as a uint16
func (c *Connection) QueryUint16(id Identifier) (uint16, error) {
	dg, err := c.Query(id)
	if err != nil {
		return 0, err
	}
	return dg.Uint16()
}

// Queries the given identifier on the RCT device, returning its value as a uint8
func (c *Connection) QueryUint8(id Identifier) (uint8, error) {
	dg, err := c.Query(id)
	if err != nil {
		return 0, err
	}
	return dg.Uint8()
}

// Writes the given identifier with the given value on the RCT device
func (c *Connection) Write(id Identifier, data []byte) error {
	var rdb DatagramBuilder
	rdb.Build(&Datagram{Write, id, data})
	_, err := c.Send(&rdb)
	return err
}
