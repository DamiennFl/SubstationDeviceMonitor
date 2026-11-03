package api

// Event Broker code

// Event Envelope:
// Common structure defining each event. Event types will also need to be defined.
// Why choose this structure: Arbitrary event handling by SSE event handler. By using a common structure,
// we allow the SSE event handler to take just one type.

// Broker state:
// Maintains a list of currently connected subscribers. Allows subbing and unsubbing at will.
// A mutex protecting the list of subs.
// Broker will stay up as long as the app is up, so no need for a closed/stopped state for now.

// Broker constructor:
// Constructor should initialize list and mutex.

// Broker subscribe:
// The broker should allow clients (SSE handler) to subscribe to it.

// Broker unsubscribe:
// The broker should allow clients to unsubscribe to it.

// Broker publish:
// The broker should allow publishing events to its list of clients via iteration. This will be
// called by mutation handlers, e.g. create and update handlers once a successful creation or update has occurred.

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
