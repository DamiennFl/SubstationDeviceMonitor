package model

import (
	"time"
	"uuid"
)

type Alarm struct {
	Id              uuid.UUID       // Alarm ID
	RuleID          uuid.UUID       // Matching Rule ID
	DeviceID        uuid.UUID       // Matching Device ID
	RaisedAt        time.Time       // Raise Time and Date
	ClearedAt       time.Time       // Cleared Time and Date
	ClearReason     string          // Reason for Alarm Clear
	TriggerValue    any             // Number or String depending on Rule type
	RuleSnapshot    Rule            // Snapshot of Rule which triggered Alarm
	Acknowledgement Acknowledgement // Acknowledgement
}
