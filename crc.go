package rct

// Checksum for a RCT datagram
type CRC struct {
	crc   uint16
	isOdd bool
}

// Returns a new CRC
func NewCRC() (c *CRC) {
	return &CRC{
		crc:   0xffff,
		isOdd: false,
	}
}

// Resets the CRC
func (c *CRC) Reset() {
	c.crc = 0xffff
	c.isOdd = false
}

// Updates CRC with the given byte
func (c *CRC) Update(b byte) {
	crc := c.crc
	for i := 0; i < 8; i++ {
		bit := (b >> (7 - i) & 1) == 1
		c15 := ((crc >> 15) & 1) == 1
		crc <<= 1
		if c15 != bit {
			crc ^= 0x1021
		}
	}
	c.crc = crc
	c.isOdd = !c.isOdd
}

// Finalizes CRC, padding if required, and returns value
func (c *CRC) Get() uint16 {
	if c.isOdd {
		c.Update(0) // pad CRC stream (not byte stream) to even length
	}
	return c.crc
}
