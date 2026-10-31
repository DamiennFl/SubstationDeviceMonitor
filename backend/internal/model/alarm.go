package model

import (
	"time"

	"github.com/google/uuid"
)

// Alarm records one occurrence of a rule being triggered.
type Alarm struct {
	ID           uuid.UUID    `json:"id"`                    // Alarm ID
	RuleID       *uuid.UUID   `json:"ruleId,omitempty"`      // Matching Rule ID
	DeviceID     uuid.UUID    `json:"deviceId"`              // Matching Device ID
	RaisedAt     time.Time    `json:"raisedAt"`              // Raise Time and Date
	ClearedAt    *time.Time   `json:"clearedAt,omitempty"`   // Cleared Time and Date
	ClearReason  *ClearReason `json:"clearReason,omitempty"` // Reason for Alarm Clear
	TriggerValue any          `json:"triggerValue"`          // Number or String depending on Rule type
	RuleSnapshot RuleSnapshot `json:"ruleSnapshot"`          // Snapshot of Rule which triggered Alarm
}

// RuleSnapshot preserves the rule fields that explain why an alarm fired.
type RuleSnapshot struct {
	Name        string      `json:"name"`
	Type        RuleType    `json:"type"`
	Measurement Measurement `json:"measurement"`
	Comparison  *Comparison `json:"comparison,omitempty"`
	Limit       *float64    `json:"limit,omitempty"`
	LowerLimit  *float64    `json:"lowerLimit,omitempty"`
	UpperLimit  *float64    `json:"upperLimit,omitempty"`
	Deadband    *float64    `json:"deadband,omitempty"`
	TripState   *State      `json:"tripState,omitempty"`
}
