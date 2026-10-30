package model

import (
	"time"
	"uuid"
)

type Reading struct {
	DeviceID      uuid.UUID // Device ID
	Timestamp     time.Time // Timestamp of Reading
	CurrentA      float64   // Current A (Amps)
	CurrentB      float64   // Current B (Amps)
	CurrentC      float64   // Current C (Amps)
	VoltageA      float64   // Voltage A (Volts)
	VoltageB      float64   // Voltage B (Volts)
	VoltageC      float64   // Voltage C (Volts)
	Frequency     float64   // Frequency (Hz)
	RealPower     float64   // Real Power (Kilowatts)
	ReactivePower float64   // Reactive Power (Kilovolt-Ampere Reactive)
	BreakerStatus string    // Breaker Status (Open or Closed)
}
