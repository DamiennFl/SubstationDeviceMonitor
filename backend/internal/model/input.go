package model

import (
	"net"
	"strings"
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

//
// Helper Functions
//

func (input UpdateDeviceInput) UpdateFields(device *Device) {
	if input.Name != nil {
		device.Name = strings.TrimSpace(*input.Name)
	}
	if input.IP != nil {
		device.IP = net.ParseIP(*input.IP).String()
	}
}

func (input UpdateRuleInput) UpdateFields(rule *Rule) {
	if input.Name != nil {
		rule.Name = strings.TrimSpace(*input.Name)
	}
	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}
	if input.Measurement != nil {
		rule.Measurement = *input.Measurement
	}
	if input.Comparison != nil {
		rule.Comparison = input.Comparison
	}
	if input.Limit != nil {
		rule.Limit = input.Limit
	}
	if input.LowerLimit != nil {
		rule.LowerLimit = input.LowerLimit
	}
	if input.UpperLimit != nil {
		rule.UpperLimit = input.UpperLimit
	}
	if input.Deadband != nil {
		rule.Deadband = input.Deadband
	}
	if input.TripState != nil {
		rule.TripState = input.TripState
	}
}
