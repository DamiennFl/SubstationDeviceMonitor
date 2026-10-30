package model

import (
	"time"

	"github.com/google/uuid"
)

// Acknowledgement records that an alarm was acknowledged.
type Acknowledgement struct {
	AlarmID        uuid.UUID `json:"alarmId"`        // Alarm ID
	AcknowledgedBy string    `json:"acknowledgedBy"` // Who Acknowledged
	AcknowledgedAt time.Time `json:"acknowledgedAt"` // When Alarm was Acknowledged
}
