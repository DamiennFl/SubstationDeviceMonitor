package model

import (
	"time"
	"uuid"
)

type Reading struct {
	DeviceID      uuid.UUID `json:"deviceId"`      // Device ID
	Timestamp     time.Time `json:"timestamp"`     // Timestamp of Reading
	CurrentA      float64   `json:"currentA"`      // Current A (Amps)
	CurrentB      float64   `json:"currentB"`      // Current B (Amps)
	CurrentC      float64   `json:"currentC"`      // Current C (Amps)
	VoltageA      float64   `json:"voltageA"`      // Voltage A (Volts)
	VoltageB      float64   `json:"voltageB"`      // Voltage B (Volts)
	VoltageC      float64   `json:"voltageC"`      // Voltage C (Volts)
	Frequency     float64   `json:"frequency"`     // Frequency (Hz)
	RealPower     float64   `json:"realPower"`     // Real Power (Kilowatts)
	ReactivePower float64   `json:"reactivePower"` // Reactive Power (Kilovolt-Ampere Reactive)
	BreakerStatus string    `json:"breakerStatus"` // Breaker Status (Open or Closed)
}
