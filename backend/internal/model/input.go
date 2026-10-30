package model

import (
	"time"
)

// Use Structs for Input. When the client sends JSON, it needs something to decode it into, which is these creation structs.

// CreateDeviceInput contains fields accepted when creating a device.
type CreateDeviceInput struct {
	Name       string     `json:"name"`
	IP         string     `json:"ip"`
	DeviceType DeviceType `json:"type"`
}

// UpdateDeviceInput contains optional fields accepted when updating a device.
type UpdateDeviceInput struct {
	Name *string `json:"name"`
	IP   *string `json:"ip"`
}

// CreateRuleInput contains fields accepted when creating a rule.
type CreateRuleInput struct {
	Name        string      `json:"name"`
	Enabled     *bool       `json:"enabled"`
	Type        RuleType    `json:"type"`
	Measurement Measurement `json:"measurement"`
	Comparison  *Comparison `json:"comparison"`
	Limit       *float64    `json:"limit"`
	LowerLimit  *float64    `json:"lowerLimit"`
	UpperLimit  *float64    `json:"upperLimit"`
	Deadband    *float64    `json:"deadband"`
	TripState   *State      `json:"tripState"`
}

// UpdateRuleInput contains optional fields accepted when updating a rule.
type UpdateRuleInput struct {
	Name        *string      `json:"name"`
	Enabled     *bool        `json:"enabled"`
	Measurement *Measurement `json:"measurement"`
	Comparison  *Comparison  `json:"comparison"`
	Limit       *float64     `json:"limit"`
	LowerLimit  *float64     `json:"lowerLimit"`
	UpperLimit  *float64     `json:"upperLimit"`
	Deadband    *float64     `json:"deadband"`
	TripState   *State       `json:"tripState"`
}

// ReadingInput contains a reading submitted for the device in the URL path.
type ReadingInput struct {
	Timestamp     time.Time `json:"timestamp"`
	CurrentA      *float64  `json:"currentA"`
	CurrentB      *float64  `json:"currentB"`
	CurrentC      *float64  `json:"currentC"`
	VoltageA      *float64  `json:"voltageA"`
	VoltageB      *float64  `json:"voltageB"`
	VoltageC      *float64  `json:"voltageC"`
	Frequency     *float64  `json:"frequency"`
	RealPower     *float64  `json:"realPower"`
	ReactivePower *float64  `json:"reactivePower"`
	BreakerStatus *State    `json:"breakerStatus"`
}
