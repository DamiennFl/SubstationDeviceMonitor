package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// General SSE handler, which sends events back to frontend.
// Subscribes to the shared event broker in the server's Store.
func (s *Server) getEventsSSEHandler(w http.ResponseWriter, r *http.Request) {
	// Set header to event stream
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Get Flusher, confirm streaming is supported
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Initial status as OK, which signals to the client the stream is open
	w.WriteHeader(http.StatusOK)

	// Send initial status, otherwise it will hang.
	flusher.Flush()

	// Sub before looping, defer unsub until looping is done.
	// When r.Context().Done() is true, loop will exit and unsub will trigger
	subscriber := s.eventBroker.Subscribe()
	defer s.eventBroker.Unsubscribe(subscriber.SubscriberID)

	// Infinite loop. SSE stream closes once client context is closed.
	// .Done sends a channel that is closed when client closes sse connection.
	for {
		select {
		// New event
		case event, open := <-subscriber.Events:
			// Channel is not open, return
			if !open {
				return
			}

			// Marshal the full event envelope. JSON output has no raw newlines,
			// so it fits on a single SSE data line.
			payload, err := json.Marshal(event)
			if err != nil {
				log.Printf("Failed to create event JSON payload for event %s: %v", event.EventID, err)
				continue
			}

			// SSE frame: id and event name for browser dispatch, envelope as data.
			// The trailing blank line terminates the frame.
			if _, err := fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", event.EventID, event.EventType, payload); err != nil {
				log.Printf("failed to write SSE event: %v", err)
				return
			}

			// Flush written data, immediately sending to client
			flusher.Flush()

		// If Done returns a closed channel, exit loop as client no longer wants to receive SSEs.
		case <-r.Context().Done():
			return
		}
	}
}
