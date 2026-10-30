package model

import (
	"time"
	"uuid"
)

type Event struct {
	Reading           *Reading         `json:"reading,omitempty"`           // Reading Event
	AlarmRaised       *Alarm           `json:"alarmRaised,omitempty"`       // Alarm Raised Event
	AlarmCleared      *AlarmCleared    `json:"alarmCleared,omitempty"`      // Alarm Cleared Event
	AlarmAcknowledged *Acknowledgement `json:"alarmAcknowledged,omitempty"` // Alarm Acknowledged Event
}

type AlarmCleared struct {
	Id        uuid.UUID `json:"id"`        // Alarm Cleared ID
	RuleID    uuid.UUID `json:"ruleId"`    // Matching Rule ID
	DeviceID  uuid.UUID `json:"deviceId"`  // Matching Device ID
	ClearedAt time.Time `json:"clearedAt"` // When Alarm was cleared
}