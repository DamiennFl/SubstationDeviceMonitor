package model

import (
	"time"

	"github.com/google/uuid"
)

// Uniform event type
type Event struct {
	EventID        uuid.UUID
	DeviceID       uuid.UUID
	EventType      EventType
	EventTimestamp time.Time
	Data           string
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
