package model

import (
	"time"

	"github.com/google/uuid"
)

type Alarm struct {
	Id	            uuid.UUID        `json:"id"`                        // Alarm ID
	DeviceID        uuid.UUID        `json:"deviceId"`                  // Matching Device ID
	RaisedAt        time.Time        `json:"raisedAt"`                  // Raise Time and Date
	ClearedAt       *time.Time       `json:"clearedAt,omitempty"`       // Cleared Time and Date
	ClearReason     *string          `json:"clearReason,omitempty"`     // Reason for Alarm Clear
	TriggerValue    any              `json:"triggerValue"`              // Number or String depending on Rule type
	RuleSnapshot    Rule             `json:"ruleSnapshot"`              // Snapshot of Rule which triggered Alarm
	Acknowledgement *Acknowledgement `json:"acknowledgement,omitempty"` // Acknowledgement
}

