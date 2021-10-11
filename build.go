package rct

import (
	"bytes"
	"fmt"
)

// Builds RCT datagrams into an internal buffer, with escaping and CRC correction
type DatagramBuilder struct {
	buffer bytes.Buffer
	crc    *CRC
}

// Returns a new DatagramBuilder
func NewDatagramBuilder() (b *DatagramBuilder) {
	return &DatagramBuilder{
		buffer: bytes.Buffer{},
		crc:    NewCRC(),
	}
}

// Resets the internal buffer and CRC
func (rdb *DatagramBuilder) Reset() {
	rdb.buffer.Reset()
	rdb.crc.Reset()
}

// Adds a byte to the internal buffer, handling escaping and CRC calculation
func (rdb *DatagramBuilder) WriteByte(b byte) {
	if (b == 0x2b) || (b == 0x2d) {
		rdb.buffer.WriteByte(0x2d) // escape in byte stream (not in CRC stream)
	}
	rdb.buffer.WriteByte(b)
	rdb.crc.Update(b)
}

// Adds a byte to the internal buffer, without escaping or CRC calculation
func (rdb *DatagramBuilder) WriteByteUnescapedNoCRC(b byte) {
	rdb.buffer.WriteByte(b)
}

// Writes the CRC into the current datastream, handling CRC calcuation padding to an even number of bytes
func (rdb *DatagramBuilder) WriteCRC() {
	crc := rdb.crc.Get()
	rdb.buffer.WriteByte(byte(crc >> 8))
	rdb.buffer.WriteByte(byte(crc & 0xff))
}

// Builds a complete datagram into the buffer
func (rdb *DatagramBuilder) Build(dg *Datagram) {
	rdb.Reset()
	rdb.WriteByteUnescapedNoCRC(0x2b) // Start byte
	rdb.WriteByte(byte(dg.Cmd))
	rdb.WriteByte(byte(len(dg.Data) + 4))
	rdb.WriteByte(byte(dg.Id >> 24))
	rdb.WriteByte(byte((dg.Id >> 16) & 0xff))
	rdb.WriteByte(byte((dg.Id >> 8) & 0xff))
	rdb.WriteByte(byte(dg.Id & 0xff))
	for _, d := range dg.Data {
		rdb.WriteByte(d)
	}
	rdb.WriteCRC()
}

// Returns the datagram built so far as an array of bytes
func (r *DatagramBuilder) Bytes() []byte {
	return r.buffer.Bytes()
}

// Converts the datagram into a string representation for printing
func (r *DatagramBuilder) String() string {
	buf := bytes.Buffer{}
	buf.WriteByte(byte('['))
	for i, b := range r.buffer.Bytes() {
		if i != 0 {
			buf.WriteByte(byte(' '))
		}
		fmt.Fprintf(&buf, "%02X", b)
	}
	buf.WriteByte(byte(']'))
	return buf.String()
}
