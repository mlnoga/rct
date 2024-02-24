package rct

import (
	"net"
	"time"
)

// DialTimeout is the default cache for connecting to a RCT device
var DialTimeout = time.Second * 5

// Connection to a RCT device
type Connection struct {
	host    string
	conn    net.Conn
	builder *DatagramBuilder
	parser  *DatagramParser
	cache   *Cache
}

// Map of active connections
var connectionCache map[string]*Connection = make(map[string]*Connection)

// Creates a new connection to a RCT device at the given address
func NewConnection(host string, cache time.Duration) (*Connection, error) {
	if conn, ok := connectionCache[host]; ok {
		return conn, nil
	}

	conn := &Connection{
		host:    host,
		builder: NewDatagramBuilder(),
		parser:  NewDatagramParser(),
		cache:   NewCache(cache),
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
}

// Sends the given RCT datagram via the connection
func (c *Connection) Send(rdb *DatagramBuilder) (n int, err error) {
	// ensure active connection
	if c.conn == nil {
		if err := c.connect(); err != nil {
			return 0, err
		}
	}

	// fmt.Printf("Sending %v\n", c.Builder.String())
	n, err = c.conn.Write(rdb.Bytes())

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
func (c *Connection) Receive() (dg *Datagram, err error) {
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
func (c *Connection) Query(id Identifier) (dg *Datagram, err error) {
	if dg, ok := c.cache.Get(id); ok {
		return dg, nil
	}

	const maxRetries = 10                     // Define maximum number of retry attempts.
	const startDelay = 100 * time.Millisecond // Initial delay before retries.
	const backoffMultiplier = 2               // Multiplier to increase delay between retries.
	var delay = startDelay                    // Current delay, starting with startDelay.

	for attempt := 0; attempt < maxRetries; attempt++ {
		c.builder.Build(&Datagram{Read, id, nil})
		_, err = c.Send(c.builder)
		if err != nil {
			return nil, err
		}

		dg, err = c.Receive()
		if err == nil && dg.Cmd == Response && dg.Id == id {
			c.cache.Put(dg) // Cache and return on success, unchanged from original.
			return dg, nil
		} else {
			// Loop on recoverable error or if response doesn't match request.
			time.Sleep(delay)                         // Wait before retrying.
			delay *= time.Duration(backoffMultiplier) // Increase delay for next attempt.
			continue                                  // Continue to next attempt.
		}
	}

	// Return the last occurred error after all attempts, if not successful.
	return nil, err
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
