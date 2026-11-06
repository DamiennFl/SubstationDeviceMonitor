package model

import (
	"time"

	"github.com/google/uuid"
)

// Uniform event type
type Event struct {
	EventID        uuid.UUID `json:"eventId"`
	DeviceID       uuid.UUID `json:"deviceId"`
	EventType      EventType `json:"eventType"`
	EventTimestamp time.Time `json:"eventTimeStamp"`
	Data           string    `json:"data"`
}

// Event types
type EventType string

const (
	ReadingEvent      EventType = "reading"
	AlarmRaisedEvent  EventType = "alarmRaised"
	AlarmClearedEvent EventType = "alarmCleared"
	AlarmAckedEvent   EventType = "alarmAcked"
)

func (e EventType) IsValid() bool {
	switch e {
	case ReadingEvent:
		return true
	case AlarmRaisedEvent:
		return true
	case AlarmClearedEvent:
		return true
	case AlarmAckedEvent:
		return true
	default:
		return false
	}
}

type Subscriber struct {
	SubscriberID uuid.UUID
	Events       chan Event
}
