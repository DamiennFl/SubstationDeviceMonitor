package model

import (
	"time"

	"github.com/google/uuid"
)

// AlarmCleared describes an alarm-cleared live event.
type AlarmCleared struct {
	ID          uuid.UUID   `json:"id"`               // Alarm Cleared ID
	RuleID      *uuid.UUID  `json:"ruleId,omitempty"` // Matching Rule ID
	DeviceID    uuid.UUID   `json:"deviceId"`         // Matching Device ID
	ClearedAt   time.Time   `json:"clearedAt"`        // When Alarm was cleared
	ClearReason ClearReason `json:"clearReason"`      // Reason Alarm was Cleared
}
