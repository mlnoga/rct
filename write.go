package rct

import (
	"encoding/binary"
	"fmt"
	"math"
)

/*
 * See https://github.com/do-gooder/rctpower_writesupport?tab=readme-ov-file#usage
 */

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

// SetSocTarget sets the SOC target (power_mng.soc_target_set) with the given value
func (c *Connection) SetSocTarget(target float32) error {
	if target < 0.00 || target > 1.00 {
		return fmt.Errorf("invalid SOC target value: %.2f, valid range is 0.00 to 1.00", target)
	}

	// Round to 2 decimal places
	target = float32(math.Round(float64(target)*100) / 100)

	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, math.Float32bits(target))

	builder := NewDatagramBuilder()
	builder.Build(&Datagram{
		Cmd:  Write,
		Id:   PowerMngSocTargetSet,
		Data: data,
	})

	_, err := c.Send(builder)
	if err != nil {
		return fmt.Errorf("failed to set SOC target: %w", err)
	}

	return nil
}

// SetBatteryPowerExtern sets the external battery power (power_mng.battery_power_extern) with the given float32 value in W
func (c *Connection) SetBatteryPowerExtern(power float32) error {
	if power < -6000 || power > 6000 {
		return fmt.Errorf("invalid battery power value: %.2f, valid range is -6000 to 6000", power)
	}

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

// SetSocMin sets the minimum SOC target (power_mng.soc_min) with the given value
func (c *Connection) SetSocMin(min float32) error {
	if min < 0.00 || min > 1.00 {
		return fmt.Errorf("invalid SOC min value: %.2f, valid range is 0.00 to 1.00", min)
	}

	// Round to 2 decimal places
	min = float32(math.Round(float64(min)*100) / 100)

	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, math.Float32bits(min))

	builder := NewDatagramBuilder()
	builder.Build(&Datagram{
		Cmd:  Write,
		Id:   PowerMngSocMin,
		Data: data,
	})

	_, err := c.Send(builder)
	if err != nil {
		return fmt.Errorf("failed to set SOC min: %w", err)
	}

	return nil
}

// SetSocMax sets the maximum SOC target (power_mng.soc_max) with the given value
func (c *Connection) SetSocMax(max float32) error {
	if max < 0.00 || max > 1.00 {
		return fmt.Errorf("invalid SOC max value: %.2f, valid range is 0.00 to 1.00", max)
	}

	// Round to 2 decimal places
	max = float32(math.Round(float64(max)*100) / 100)

	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, math.Float32bits(max))

	builder := NewDatagramBuilder()
	builder.Build(&Datagram{
		Cmd:  Write,
		Id:   PowerMngSocMax,
		Data: data,
	})

	_, err := c.Send(builder)
	if err != nil {
		return fmt.Errorf("failed to set SOC max: %w", err)
	}

	return nil
}

// SetSocChargePower sets the charging power to reach SOC target (power_mng.soc_charge_power)
func (c *Connection) SetSocChargePower(power uint16) error {
	// Valid range is not defined, assume it’s a valid unsigned integer
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, power)

	builder := NewDatagramBuilder()
	builder.Build(&Datagram{
		Cmd:  Write,
		Id:   PowerMngSocChargePowerW,
		Data: data,
	})

	_, err := c.Send(builder)
	if err != nil {
		return fmt.Errorf("failed to set SOC charge power: %w", err)
	}

	return nil
}

// SetSocCharge sets the trigger for charging to SOC_min (power_mng.soc_charge)
func (c *Connection) SetSocCharge(charge float32) error {
	if charge < 0.00 || charge > 1.00 {
		return fmt.Errorf("invalid SOC charge value: %.2f, valid range is 0.00 to 1.00", charge)
	}

	// Round to 2 decimal places
	charge = float32(math.Round(float64(charge)*100) / 100)

	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, math.Float32bits(charge))

	builder := NewDatagramBuilder()
	builder.Build(&Datagram{
		Cmd:  Write,
		Id:   PowerMngSocCharge,
		Data: data,
	})

	_, err := c.Send(builder)
	if err != nil {
		return fmt.Errorf("failed to set SOC charge: %w", err)
	}

	return nil
}

// SetGridPowerLimit sets the maximum battery-to-grid power (p_rec_lim[1])
func (c *Connection) SetGridPowerLimit(power uint16) error {
	if power > 6000 {
		return fmt.Errorf("invalid grid power limit value: %d, valid range is 0 to 6000", power)
	}

	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, power)

	builder := NewDatagramBuilder()
	builder.Build(&Datagram{
		Cmd:  Write,
		Id:   PowerMngGridPowerLimitW,
		Data: data,
	})

	_, err := c.Send(builder)
	if err != nil {
		return fmt.Errorf("failed to set grid power limit: %w", err)
	}

	return nil
}

// SetUseGridPower sets the enable/disable flag for grid power usage (power_mng.use_grid_power_enable)
func (c *Connection) SetUseGridPower(enable bool) error {
	var data byte
	if enable {
		data = 1
	} else {
		data = 0
	}

	builder := NewDatagramBuilder()
	builder.Build(&Datagram{
		Cmd:  Write,
		Id:   PowerMngUseGridPowerEnable,
		Data: []byte{data},
	})

	_, err := c.Send(builder)
	if err != nil {
		return fmt.Errorf("failed to set grid power usage: %w", err)
	}

	return nil
}
