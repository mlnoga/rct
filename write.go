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

	if err := c.Write(PowerMngSocStrategy, []byte{strategy}); err != nil {
		return fmt.Errorf("failed to set SOC strategy: %w", err)
	}

	return nil
}

// SetSocTarget sets the SOC target (power_mng.soc_target_set) with the given value
func (c *Connection) SetSocTarget(target float32) error {
	if target < 0.00 || target > 1.00 {
		return fmt.Errorf("invalid SOC target value: %.2f, valid range is 0.00 to 1.00", target)
	}

	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, math.Float32bits(target))

	if err := c.Write(PowerMngSocTargetSet, data); err != nil {
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

	if err := c.Write(PowerMngBatteryPowerExternW, data); err != nil {
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

	if err := c.Write(BatterySoCTargetMin, data); err != nil {
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

	if err := c.Write(PowerMngSocMax, data); err != nil {
		return fmt.Errorf("failed to set SOC max: %w", err)
	}

	return nil
}

// SetSocChargePower sets the charging power to reach SOC target (power_mng.soc_charge_power)
func (c *Connection) SetSocChargePower(power uint16) error {
	// Valid range is not defined, assume it’s a valid unsigned integer
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, power)

	if err := c.Write(PowerMngSocChargePowerW, data); err != nil {
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

	if err := c.Write(PowerMngSocCharge, data); err != nil {
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

	if err := c.Write(PowerMngGridPowerLimitW, data); err != nil {
		return fmt.Errorf("failed to set grid power limit: %w", err)
	}

	return nil
}

// SetUseGridPower sets the enable/disable flag for grid power usage (power_mng.use_grid_power_enable)
func (c *Connection) SetUseGridPower(enable bool) error {
	var data byte
	if enable {
		data = 1
	}

	if err := c.Write(PowerMngUseGridPowerEnable, []byte{data}); err != nil {
		return fmt.Errorf("failed to set grid power usage: %w", err)
	}

	return nil
}
