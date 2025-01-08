package rct

import (
	"encoding/binary"
	"fmt"
	"math"
)

// SetSocStrategy sets the SOC strategy (power_mng.soc_strategy) with the given ENUM value
func (c *Connection) SetSocStrategy(strategy uint8) error {
	if strategy > SOCTargetSchedule {
		return fmt.Errorf("invalid SOC strategy value: %d", strategy)
	}

	builder := NewDatagramBuilder()
	builder.Build(&Datagram{
		Cmd:  Write,
		Id:   PowerMngSocStrategy,
		Data: []byte{strategy},
	})

	_, err := c.Send(builder)
	if err != nil {
		return fmt.Errorf("failed to set SOC strategy: %w", err)
	}

	return nil
}

// SetBatteryPowerExtern sets the external battery power (power_mng.battery_power_extern) with the given float32 value in W
func (c *Connection) SetBatteryPowerExtern(power float32) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, math.Float32bits(power))

	builder := NewDatagramBuilder()
	builder.Build(&Datagram{
		Cmd:  Write,
		Id:   PowerMngBatteryPowerExternW,
		Data: data,
	})

	_, err := c.Send(builder)
	if err != nil {
		return fmt.Errorf("failed to set battery power extern: %w", err)
	}

	return nil
}
