package model

import (
	"time"
	"uuid"
)

type Acknowledgement struct {
	AlarmID        uuid.UUID `json:"alarmId"`        // Alarm ID
	AcknowledgedBy string    `json:"acknowledgedBy"` // Who Acknowledged
	AcknowledgedAt time.Time `json:"acknowledgedAt"` // When Alarm was Acknowledged
}
