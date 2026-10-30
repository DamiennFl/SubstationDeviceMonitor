package model

import (
	"time"
	"uuid"
)

type Acknowledgement struct {
	AlarmID        uuid.UUID // Alarm ID
	AcknowledgedBy string	 // Who Acknowledged
	AcknowledgedAt time.Time // When Alarm was Acknowledged
}
