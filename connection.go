package rct

import (
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	// DialTimeout is the default cache for connecting to a RCT device
	DialTimeout = time.Second * 5

	// Map of active connections
	connectionCache = make(map[string]*Connection)
)

// Connection to a RCT device
type Connection struct {
	mu     sync.Mutex
	host   string
	conn   net.Conn
	parser *DatagramParser
	cache  *Cache
}

// Creates a new connection to a RCT device at the given address.
// Must not be called concurrently.
func NewConnection(host string, cache time.Duration) (*Connection, error) {
	if conn, ok := connectionCache[host]; ok {
		if conn.conn != nil { // there might be dead connection in the cache, e.g. when connection was disconnected
			return conn, nil
		}
	}

	conn := &Connection{
		host:   host,
		parser: NewDatagramParser(),
		cache:  NewCache(cache),
	}

	if err := conn.connect(); err != nil {
		return nil, err
	}

	connectionCache[host] = conn
	return conn, nil
}

// Connects an uninitialized RCT connection to the device at the given address
func (c *Connection) connect() (err error) {
	address := net.JoinHostPort(c.host, "8899") // default port for RCT
	c.conn, err = net.DialTimeout("tcp", address, DialTimeout)
	return err
}

// Closes the RCT device connection
func (c *Connection) Close() {
	c.conn.Close()
	c.conn = nil
	delete(connectionCache, c.host) // connection is dead, no need to cache any more
}

// Sends the given RCT datagram via the connection
func (c *Connection) Send(rdb *DatagramBuilder) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.send(rdb)
}

func (c *Connection) send(rdb *DatagramBuilder) (int, error) {
	// ensure active connection
	if c.conn == nil {
		if err := c.connect(); err != nil {
			return 0, err
		}
	}

	// fmt.Printf("Sending %v\n", c.Builder.String())
	n, err := c.conn.Write(rdb.Bytes())
	// single retry on error when sending
	if err != nil {
		// fmt.Printf("Read %d bytes error %v\n", n, err)
		c.conn.Close()
		// fmt.Printf("Error reconnecting: %v\n", err)
		if err := c.connect(); err != nil {
			return 0, err
		}
		n, err = c.conn.Write(rdb.Bytes())
		// fmt.Printf("Read %d bytes error %v\n", n, err)
	}
	return n, err
}

// Receives an RCT response via the connection
func (c *Connection) Receive() (*Datagram, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.receive()
}

// Receives an RCT response via the connection
func (c *Connection) receive() (dg *Datagram, err error) {
	// ensure active connection
	if c.conn == nil {
		if err := c.connect(); err != nil {
			return nil, err
		}
	}

	c.parser.Reset()
	c.parser.length, err = c.conn.Read(c.parser.buffer)
	if err != nil {
		return dg, err
	}
	// fmt.Printf("Received %d bytes: %v\n", c.Parser.Len, c.Parser.Buffer[:c.Parser.Len])

	return c.parser.Parse()

	// dg, err=c.Parser.Parse()
	// fmt.Printf("Received datagram %s error %v\n", dg.String(), err)
	// return dg, err
}

// Queries the given identifier on the RCT device, returning its value as a datagram
func (c *Connection) Query(id Identifier) (*Datagram, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if dg, ok := c.cache.Get(id); ok {
		return dg, nil
	}

	builder := NewDatagramBuilder()
	builder.Build(&Datagram{Read, id, nil})
	if _, err := c.send(builder); err != nil {
		return nil, err
	}

	dg, err := c.receive()
	if err != nil {
		return nil, err
	}
	if dg.Cmd != Response || dg.Id != id {
		return nil, RecoverableError{fmt.Sprintf("invalid response to read of %08X: %v", id, dg)}
	}
	c.cache.Put(dg)

	return dg, nil
}

// Queries the given identifier on the RCT device, returning its value as a float32
func (c *Connection) QueryFloat32(id Identifier) (val float32, err error) {
	dg, err := c.Query(id)
	if err != nil {
		return 0, err
	}
	return dg.Float32()
}

// Queries the given identifier on the RCT device, returning its value as a uint16
func (c *Connection) QueryUint16(id Identifier) (val uint16, err error) {
	dg, err := c.Query(id)
	if err != nil {
		return 0, err
	}
	return dg.Uint16()
}

// Queries the given identifier on the RCT device, returning its value as a uint8
func (c *Connection) QueryUint8(id Identifier) (val uint8, err error) {
	dg, err := c.Query(id)
	if err != nil {
		return 0, err
	}
	return dg.Uint8()
}

// Writes the given identifier with the given value on the RCT device
func (c *Connection) Write(id Identifier, data []byte) error {
	b := NewDatagramBuilder()
	b.Build(&Datagram{Write, id, data})
	_, err := c.Send(b)
	return err
}
