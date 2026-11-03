package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
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

			// Try and write SSE event metadata
			if _, err := fmt.Fprintf(w, "id: %s\nevent: %s\n", event.EventID, event.EventType); err != nil {
				log.Printf("failed to write SSE event metadata: %v", err)
				return
			}

			// Split payload into separate lines, and try to print.
			for _, line := range strings.Split(event.Payload, "\n") {
				// Error if printing any line failes
				if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
					log.Printf("failed to write SSE event data: %v", err)
					return
				}
			}

			// Write last line, required for SSE format
			if _, err := fmt.Fprint(w, "\n"); err != nil {
				log.Printf("failed to terminate SSE event: %v", err)
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
