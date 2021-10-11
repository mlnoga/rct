package rct

import "testing"

type builderTestCase struct {
	Dg     Datagram
	Expect string
}

var builderTestCases = []builderTestCase{
	{Datagram{Read, BatteryPowerW, nil}, "[2B 01 04 40 0F 01 5B 58 B4]"},
	{Datagram{Read, InverterACPowerW, nil}, "[2B 01 04 DB 2D 2D 69 AE 55 AB]"},
}

// Test if builder returns expected byte representation
func TestBuilder(t *testing.T) {
	builder := NewDatagramBuilder()
	for _, tc := range builderTestCases {
		builder.Build(&tc.Dg)
		res := builder.String()
		if res != tc.Expect {
			t.Errorf("error got %s, should be %s", res, tc.Expect)
		}
	}
}

// Test if roundtrip from builder to parser returns the same datagram
func TestBuilderParser(t *testing.T) {
	builder := NewDatagramBuilder()
	parser := NewDatagramParser()

	for _, tc := range builderTestCases {
		builder.Build(&tc.Dg)
		parser.Reset()
		parser.buffer = builder.Bytes()
		parser.length = len(builder.Bytes())
		dg, err := parser.Parse()
		if err != nil {
			t.Errorf(err.Error())
		}
		if dg.Cmd != tc.Dg.Cmd || dg.Id != tc.Dg.Id || len(dg.Data) != len(tc.Dg.Data) {
			t.Errorf("error mismatch got %s, expect %s", dg.String(), tc.Dg.String())
		}
		for i := 0; i < len(dg.Data); i++ {
			if dg.Data[i] != tc.Dg.Data[i] {
				t.Errorf("error mismatch got %s, expect %s", dg.String(), tc.Dg.String())
			}
		}
	}
}
