package api

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

// Make event
func testEvent() model.Event {
	return model.Event{
		EventID:        uuid.New(),
		DeviceID:       uuid.New(),
		EventType:      model.ReadingEvent,
		EventTimestamp: time.Date(2026, 11, 3, 20, 0, 0, 0, time.UTC),
		Data:           `{"value":10}`,
	}
}

func TestEventBrokerPublish(t *testing.T) {
	broker := NewEventBroker()
	first := broker.Subscribe()
	second := broker.Subscribe()
	event := testEvent()

	if err := broker.Publish(event); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	for name, subscriber := range map[string]model.Subscriber{
		"first":  first,
		"second": second,
	} {
		t.Run(name, func(t *testing.T) {
			select {
			case got := <-subscriber.Events:
				if got != event {
					t.Fatalf("event = %#v, want %#v", got, event)
				}
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for published event")
			}
		})
	}
}

func TestEventBrokerUnsubscribe(t *testing.T) {
	broker := NewEventBroker()
	subscriber := broker.Subscribe()

	if err := broker.Unsubscribe(subscriber.SubscriberID); err != nil {
		t.Fatalf("Unsubscribe() error = %v", err)
	}

	select {
	case _, open := <-subscriber.Events:
		if open {
			t.Fatal("subscriber channel is still open")
		}
	default:
		t.Fatal("subscriber channel was not closed")
	}

	if err := broker.Unsubscribe(subscriber.SubscriberID); !errors.Is(err, ErrSubscriberNotFound) {
		t.Fatalf("second Unsubscribe() error = %v, want %v", err, ErrSubscriberNotFound)
	}
}

func TestEventBrokerDropsFullSubscriber(t *testing.T) {
	broker := NewEventBroker()
	subscriber := broker.Subscribe()

	for index := 0; index < cap(subscriber.Events); index++ {
		if err := broker.Publish(testEvent()); err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	if err := broker.Publish(testEvent()); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	for index := 0; index < cap(subscriber.Events); index++ {
		if _, open := <-subscriber.Events; !open {
			t.Fatalf("subscriber channel closed after %d buffered events, want %d", index, cap(subscriber.Events))
		}
	}

	if _, open := <-subscriber.Events; open {
		t.Fatal("full subscriber channel is still open after draining buffered events")
	}
}

func TestEventBrokerRejectsInvalidEvents(t *testing.T) {
	tests := []struct {
		name string
		edit func(*model.Event)
		want error
	}{
		{
			name: "missing ID",
			edit: func(event *model.Event) { event.EventID = uuid.Nil },
			want: ErrInvalidEventID,
		},
		{
			name: "missing timestamp",
			edit: func(event *model.Event) { event.EventTimestamp = time.Time{} },
			want: ErrMissingEventTimestamp,
		},
		{
			name: "invalid type",
			edit: func(event *model.Event) { event.EventType = "invalid" },
			want: ErrInvalidEventType,
		},
		{
			name: "empty payload",
			edit: func(event *model.Event) { event.Data = "" },
			want: ErrEmptyEventPayload,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := testEvent()
			test.edit(&event)

			if err := NewEventBroker().Publish(event); !errors.Is(err, test.want) {
				t.Fatalf("Publish() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestPublishReadingResultEvents(t *testing.T) {
	broker := NewEventBroker()
	server := &Server{eventBroker: broker}
	subscriber := broker.Subscribe()

	deviceID := uuid.New()
	ruleID := uuid.New()
	raisedAt := time.Date(2026, 11, 3, 20, 0, 0, 0, time.UTC)
	clearedAt := raisedAt.Add(time.Minute)
	reading := model.Reading{
		Timestamp:     raisedAt,
		BreakerStatus: model.StateClosed,
	}
	raisedAlarm := model.Alarm{
		ID:           uuid.New(),
		RuleID:       &ruleID,
		DeviceID:     deviceID,
		RaisedAt:     raisedAt,
		TriggerValue: 401,
	}
	clearedAlarm := raisedAlarm
	clearedAlarm.ClearedAt = &clearedAt
	clearedAlarm.ClearReason = func() *model.ClearReason {
		reason := model.ClearReasonNormal
		return &reason
	}()

	if err := server.publishReadingResultEvents(deviceID, model.ReadingResult{
		Reading: reading,
		RuleEvaluationResult: model.RuleEvaluationResult{
			RaisedAlarms:  []model.Alarm{raisedAlarm},
			ClearedAlarms: []model.Alarm{clearedAlarm},
		},
	}); err != nil {
		t.Fatalf("publishReadingResultEvents() error = %v", err)
	}

	wantTypes := []model.EventType{
		model.ReadingEvent,
		model.AlarmRaisedEvent,
		model.AlarmClearedEvent,
	}
	for _, wantType := range wantTypes {
		select {
		case event := <-subscriber.Events:
			if event.EventType != wantType {
				t.Fatalf("event type = %q, want %q", event.EventType, wantType)
			}
			if event.DeviceID != deviceID {
				t.Fatalf("event device ID = %v, want %v", event.DeviceID, deviceID)
			}
			if event.Data == "" {
				t.Fatal("event payload is empty")
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for %q event", wantType)
		}
	}
}

func TestCreateEventPayloads(t *testing.T) {
	reading := model.Reading{Timestamp: time.Now().UTC()}
	alarm := model.Alarm{ID: uuid.New(), DeviceID: uuid.New(), RaisedAt: reading.Timestamp}

	tests := []struct {
		name     string
		create   func() (string, error)
		contains string
	}{
		{
			name:     "reading",
			create:   func() (string, error) { return createReadingPayload(reading) },
			contains: `"timestamp"`,
		},
		{
			name:     "raised alarm",
			create:   func() (string, error) { return createRaisedAlarmPayload(alarm) },
			contains: `"id"`,
		},
		{
			name:     "cleared alarm",
			create:   func() (string, error) { return createClearedAlarmPayload(alarm) },
			contains: `"id"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload, err := test.create()
			if err != nil {
				t.Fatalf("payload helper error = %v", err)
			}
			if payload == "" {
				t.Fatal("payload is empty")
			}
			if !strings.Contains(payload, test.contains) {
				t.Fatalf("payload = %s, missing %s", payload, test.contains)
			}
		})
	}
}

func TestAcknowledgeAlarmEvents(t *testing.T) {
	broker := NewEventBroker()
	server := &Server{eventBroker: broker}
	subscriber := broker.Subscribe()

	deviceID := uuid.New()
	ruleID := uuid.New()
	raisedAt := time.Date(2026, 11, 3, 20, 0, 0, 0, time.UTC)

	raisedAlarm := model.Alarm{
		ID:           uuid.New(),
		RuleID:       &ruleID,
		DeviceID:     deviceID,
		RaisedAt:     raisedAt,
		TriggerValue: 401,
	}

	acknowledgement := model.Acknowledgement{
		AlarmID:        raisedAlarm.ID,
		AcknowledgedBy: "admin",
		AcknowledgedAt: raisedAt,
	}

	alarmWithAck := model.AlarmWithAcknowledgement{
		Alarm:           raisedAlarm,
		Acknowledgement: &acknowledgement,
	}

	server.publishAcknowledgeAlarmEvent(alarmWithAck)

	select {
	case event := <-subscriber.Events:
		if event.DeviceID != deviceID {
			t.Fatalf("device ID = %v, want %v", event.DeviceID, deviceID)
		}
		if !event.EventTimestamp.Equal(raisedAt) {
			t.Fatalf("event timestamp = %v, want %v, which is alarm raised timestamp", event.EventTimestamp, raisedAt)
		}
		if event.EventType != model.AlarmAckedEvent {
			t.Fatalf("event type = %v, want %v", event.EventType, model.AlarmAckedEvent)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %q event", model.AlarmAckedEvent)
	}
}
