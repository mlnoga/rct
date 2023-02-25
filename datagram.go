package rct

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Command type for the RCT device
type Command uint8

// Command values for the RCT device
const (
	Read Command = iota + 1
	Write
	LongWrite
	Reserved1
	Response
	LongResponse
	Reserved2
	ReadPeriodically
	Extension = iota + 0x3c - 0x09
)

// Helper to convert command values to a human-readable representation
var rctCommandToString = []string{
	"#INVALID",
	"Read",
	"Write",
	"LongWrite",
	"Reserved1",
	"Response",
	"LongResponse",
	"Reserved2",
	"ReadPeriodically",
}

// Converts a RCT command to human-readable represenation
func (c Command) String() string {
	if c <= ReadPeriodically {
		return rctCommandToString[c]
	}
	if c == Extension {
		return "Extension"
	}
	return rctCommandToString[0]
}

// Identifier type for variables on the RCT device
type Identifier uint32

// Identifier values for variables on the RCT device, see https://rctclient.readthedocs.io/en/latest/inverter_registry.html
const (
	// power
	//
	SolarGenAPowerW  Identifier = 0xB5317B78 // float32
	SolarGenBPowerW  Identifier = 0xAA9AA253 // float32
	BatteryPowerW    Identifier = 0x400f015b // float32, positive = discharge, negative = charge
	InverterACPowerW Identifier = 0xDB2D69AE // float32
	RealPowerW       Identifier = 0x4E49AEC5 // float32
	TotalGridPowerW  Identifier = 0x91617C58 // float32, positive = taken from grid, negative = feed into grid
	BatterySoC       Identifier = 0x959930BF // float32, range 0 ... 1
	S0ExternalPowerW Identifier = 0xE96F1844 // float32

	// voltage
	//
	SolarGenAVoltage Identifier = 0xB298395D // float32
	SolarGenBVoltage Identifier = 0x5BB8075A // float32
	BatteryVoltage   Identifier = 0xA7FA5C5D // float32

	// energy
	//
	TotalEnergyWh           Identifier = 0xB1EF67CE // float32
	TotalEnergySolarGenAWh  Identifier = 0xFC724A9E // float32
	TotalEnergySolarGenBWh  Identifier = 0x68EEFD3D // float32
	TotalEnergyBattInWh     Identifier = 0x5570401B // float32
	TotalEnergyBattOutWh    Identifier = 0xA9033880 // float32
	TotalEnergyHouseholdWh  Identifier = 0xEFF4B537 // float32
	TotalEnergyGridWh       Identifier = 0xA59C8428 // float32
	TotalEnergyGridFeedInWh Identifier = 0x44D4C533 // float32
	TotalEnergyGridLoadWh   Identifier = 0x62FBE7DC // float32

	// other
	//
	InverterState             Identifier = 0x5F33284E // uint8
	BatteryCapacityAh         Identifier = 0xB57B59BD // float32
	BatteryTemperatureC       Identifier = 0x902AFAFB // float32
	BatterySoCTarget          Identifier = 0x8B9FF008 // float32 0 ... 1
	BatterySoCTargetHigh      Identifier = 0xB84A38AB // float32 0 ... 1
	BatterySoCTargetMin       Identifier = 0xCE266F0F // float32 0 ... 1
	BatterySoCTargetMinIsland Identifier = 0x8EBF9574 // float32 0 ... 1
)

// Table to convert identifier values to human-readable strings
var identifiersToString = map[Identifier]string{
	// power
	//
	SolarGenAPowerW:  "Solar generator A power [W]",
	SolarGenBPowerW:  "Solar generator B power [W]",
	BatteryPowerW:    "Battery power [W]",
	InverterACPowerW: "Inverter AC power [W]",
	RealPowerW:       "Real power [W]",
	TotalGridPowerW:  "Total grid power [W]",
	BatterySoC:       "Battery state of charge",

	// voltage
	//
	SolarGenAVoltage: "Solar generator A voltage [V]",
	SolarGenBVoltage: "Solar generator B voltage [V]",
	BatteryVoltage:   "Battery voltage [V]",

	// energy
	//
	TotalEnergyWh:           "Total energy [Wh]",
	TotalEnergySolarGenAWh:  "Total energy solarGenA [Wh]",
	TotalEnergySolarGenBWh:  "Total energy solarGenB [Wh]",
	TotalEnergyBattInWh:     "Total energy batt in [Wh]",
	TotalEnergyBattOutWh:    "Total energy batt out [Wh]",
	TotalEnergyHouseholdWh:  "Total energy household [Wh]",
	TotalEnergyGridWh:       "Total energy grid [Wh]",
	TotalEnergyGridFeedInWh: "Total energy grid feed in [Wh]",
	TotalEnergyGridLoadWh:   "Total energy grid load [Wh]",

	// other
	//
	InverterState:             "Inverter state",
	BatteryCapacityAh:         "Battery capacity [Ah]",
	BatteryTemperatureC:       "Battery temperature [°C]",
	BatterySoCTarget:          "Battery SoC target",
	BatterySoCTargetHigh:      "Battery SoC target high",
	BatterySoCTargetMin:       "Battery SoC target min",
	BatterySoCTargetMinIsland: "Battery SoC target min island",
}

// Converts an identifier to a human-readable representation
func (i Identifier) String() string {
	s, ok := identifiersToString[i]
	if !ok {
		return "#INVALID"
	}
	return s
}

// Inverter state type for InverterState responses from the RCT
type InverterStates uint8

// Inverter state values for InverterState responses from the RCT
const (
	StateStandby InverterStates = iota
	StateInitialization
	StateStandby2
	StateEfficiency
	StateInsulationCheck
	StateIslandCheck
	StatePowerCheck
	StateSymmetry
	StateRelayTest
	StateGridPassive
	StatePrepareBattPassive
	StateBattPassive
	StateHWCheck
	StateFeedIn
)

// Table to convert an inverter state value to a human-readable string
var inverterStateToString []string = []string{
	"Standby",
	"Initialization",
	"Standby2",
	"Efficiency",
	"Insulation check",
	"Island check",
	"Power check",
	"Symmetry",
	"Relay test",
	"Grid passive",
	"Prepare battery passive",
	"Battery passive",
	"Hardware check",
	"Feed in",
}

// Converts an inverter state value to a human-readable string
func (i InverterStates) String() string {
	if i > StateFeedIn {
		return "#INVALID"
	}
	return inverterStateToString[i]
}

// A RCT datagram
type Datagram struct {
	Cmd  Command
	Id   Identifier
	Data []byte
}

// Prints a RCT datagram in a human-readable representation
func (d *Datagram) String() string {
	return fmt.Sprintf("Cmd %s (%02X) Id %s (%08X) Data %v", d.Cmd.String(), uint8(d.Cmd), d.Id.String(), uint32(d.Id), d.Data)
}

// Returns datagram body value as a float32
func (d *Datagram) Float32() (val float32, err error) {
	if len(d.Data) != 4 {
		return 0, fmt.Errorf("invalid data length %d", len(d.Data))
	}

	return math.Float32frombits(binary.BigEndian.Uint32(d.Data)), nil
}

// Returns datagram body value as a uint16
func (d *Datagram) Uint16() (val uint16, err error) {
	if len(d.Data) != 2 {
		return 0, fmt.Errorf("invalid data length %d", len(d.Data))
	}

	return binary.BigEndian.Uint16(d.Data), nil
}

// Returns datagram body value as a uint8
func (d *Datagram) Uint8() (val uint8, err error) {
	if len(d.Data) != 1 {
		return 0, fmt.Errorf("invalid data length %d", len(d.Data))
	}

	return uint8(d.Data[0]), nil
}
