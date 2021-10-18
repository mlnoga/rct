package rct

import (
	"fmt"
)

// State machine type for the RCT datagram parser
type ParserState int

// State machine values for the RCT datagram parser
const (
	AwaitingStart ParserState = iota
	AwaitingCmd
	AwaitingLen
	AwaitingId0
	AwaitingId1
	AwaitingId2
	AwaitingId3
	AwaitingData
	AwaitingCrc0
	AwaitingCrc1
	Done
)

// A parser for RCT datagrams
type DatagramParser struct {
	buffer []byte
	length int
	pos    int
	state  ParserState
}

// Returns a new datagram parser
func NewDatagramParser() (p *DatagramParser) {
	return &DatagramParser{
		buffer: make([]byte, 1024), // default buffer size
		length: 0,
		pos:    0,
		state:  AwaitingStart,
	}
}

// Resets the state, without reallocating the buffer
func (p *DatagramParser) Reset() {
	p.length, p.pos, p.state = 0, 0, AwaitingStart
}

// Parses a given transmission into a datagram
func (p *DatagramParser) Parse() (dg *Datagram, err error) {
	length := uint8(0)
	dataLength := uint8(0)
	crc := CRC{}
	crcReceived := uint16(0)
	escaped := false
	state := AwaitingStart
	dg = &Datagram{}

	//fmt.Printf("Parser ")
	for _, b := range p.buffer[p.pos : p.length-p.pos] {
		//fmt.Printf("(%v)-%02x->", state, b)

		if !escaped {
			if b == 0x2b {
				state = AwaitingCmd
				continue
			} else if b == 0x2d {
				escaped = true
				continue
			}
		} else { // escaped start or stop char
			escaped = false
			// fall through and process normally
		}

		switch state {
		case AwaitingStart:
			if b == 0x2B {
				state = AwaitingCmd
			}

		case AwaitingCmd:
			crc.Reset()
			crc.Update(b)
			dg.Cmd = Command(b)
			if dg.Cmd <= ReadPeriodically || dg.Cmd == Extension {
				state = AwaitingLen
			} else {
				state = AwaitingStart
			}

		case AwaitingLen:
			crc.Update(b)
			length = uint8(b)
			dataLength = length - 4
			state = AwaitingId0

		case AwaitingId0:
			crc.Update(b)
			dg.Id = Identifier(uint32(b) << 24)
			state = AwaitingId1

		case AwaitingId1:
			crc.Update(b)
			dg.Id |= Identifier(uint32(b) << 16)
			state = AwaitingId2

		case AwaitingId2:
			crc.Update(b)
			dg.Id |= Identifier(uint32(b) << 8)
			state = AwaitingId3

		case AwaitingId3:
			crc.Update(b)
			dg.Id |= Identifier(uint32(b))
			if dataLength > 0 {
				dg.Data = make([]byte, 0, dataLength)
				state = AwaitingData
			} else {
				dg.Data = nil
				state = AwaitingCrc0
			}

		case AwaitingData:
			crc.Update(b)
			dg.Data = append(dg.Data, b)
			if len(dg.Data) >= int(dataLength) {
				state = AwaitingCrc0
			}

		case AwaitingCrc0:
			crcReceived = uint16(b) << 8
			state = AwaitingCrc1

		case AwaitingCrc1:
			crcReceived |= uint16(b)
			crcCalculated := crc.Get()
			if crcCalculated != crcReceived {
				// fmt.Printf("[CRC error calc %04x want %04x]", crcCalculated, crcReceived)
				state = AwaitingStart // CRCError
			} else {
				state = Done
			}

		case Done:
			// ignore extra bytes
		}
	}
	//fmt.Printf("(%v)\n", state)

	if state != Done {
		return dg, fmt.Errorf("parsing failed in state %d", state)
	}
	return dg, nil
}
