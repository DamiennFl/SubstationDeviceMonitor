package api

import (
	"errors"
	"sync"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

// Event Broker code

var (
	ErrSubscriberNotFound = errors.New("subscriber not found")

	ErrInvalidEventID        = errors.New("invalid event ID")
	ErrMissingEventTimestamp = errors.New("event is missing timestamp")
	ErrInvalidEventType      = errors.New("event type is missing or invalid")
	ErrEmptyEventPayload     = errors.New("event payload is empty")
)

// Event Envelope:
// Common structure defining each event. Event types will also need to be defined.
// Why choose this structure: Arbitrary event handling by SSE event handler. By using a common structure,
// we allow the SSE event handler to take just one type.
// See model/event_types.go

// Broker state:
// Maintains a map of currently connected subscribers. Allows subbing and unsubbing at will.
// A mutex protecting the map of subs.
// Broker will stay up as long as the app is up, so no need for a closed/stopped state for now.

type EventBroker struct {
	subscribers map[uuid.UUID]model.Subscriber // Map of UUID -> Subscriber
	mu          sync.Mutex                     // Mutex for sublist
}

// Broker constructor:
// Constructor should initialize map and mutex.
func NewEventBroker() *EventBroker {
	return &EventBroker{
		subscribers: make(map[uuid.UUID]model.Subscriber),
	}
}

// Broker subscribe:
// The broker should allow clients (SSE handlers) to subscribe to it.
func (e *EventBroker) Subscribe() model.Subscriber {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Make new Subscriber
	subscriber := model.Subscriber{
		SubscriberID: uuid.New(),
		Events:       make(chan model.Event, 60), // Buffered channel with size 60
		// If buffer fills, we can reasonably assume frontend
		// is not receiving events for some reason or another.
	}

	// Add to list of Subscribers
	e.subscribers[subscriber.SubscriberID] = subscriber

	return subscriber
}

// Broker unsubscribe:
// The broker should allow clients to unsubscribe to it.
func (e *EventBroker) Unsubscribe(subID uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	err := e.unsubscribeLockAware(subID)

	if err != nil {
		return err
	}

	return nil
}

// Broker publish:
// The broker should allow publishing events to its list of clients via iteration. This will be
// called by mutation handlers, e.g. create and update handlers once a successful creation or update has occurred.
func (e *EventBroker) Publish(event model.Event) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Validate event
	if event.EventID == uuid.Nil {
		return ErrInvalidEventID
	}

	// No check for Device ID, maybe in the future will want to publish non-device specific events.

	if event.EventTimestamp.IsZero() {
		return ErrMissingEventTimestamp
	}

	if !event.EventType.IsValid() {
		return ErrInvalidEventType
	}

	if len(event.Payload) == 0 {
		return ErrEmptyEventPayload
	}

	// Publish event to all subscribers
	for _, sub := range e.subscribers {
		select {
		// Valid path
		case sub.Events <- event:
		// Full buffer. Drop subscriber
		default:
			e.unsubscribeLockAware(sub.SubscriberID) // Drop slow sub. Use lock-aware helper since Publish holds lock
		}
	}

	return nil
}

func (e *EventBroker) unsubscribeLockAware(subID uuid.UUID) error {
	// Check for Subscriber. Error if not found
	subscriber, ok := e.subscribers[subID]
	if !ok {
		return ErrSubscriberNotFound
	}

	// Delete sub from map
	delete(e.subscribers, subID)

	// Close Events channel for subscriber
	close(subscriber.Events)

	return nil
}

// How this fits into the rest of the code:
// The server will have an event broker, it is independent of the store.
// Broker sits and waits for events. When events come through, it broadcasts events to all subscribers.
// The broker knows nothing about the business layer. The event handlers for different endpoints are
// the bridge. Example flow:

// Reading comes in from Simulator. Server multiplexer catches it, and sends the request to createReadingHandler.
// createReadingHandler does all of its validation and creating the Reading. createReading returns the reading,
// and (soon) any alarms raised or cleared. Events are created for each(likely using a helper), and then
// broker.Publish is called with said events (IMPORTANT: Events MUST be valid here, broker should not and
// has no way of verifying the event content for validity.) The broker then broadcasts events to all subscribers.

// Design notes:
// Handlers should publish in deterministic order. First reading, then alarm raised, then alarm cleared.
// The reading should always be first in the list of live events.
