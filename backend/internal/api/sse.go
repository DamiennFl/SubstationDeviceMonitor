package api

import "net/http"

// General SSE handler, which sends events back to frontend.
// Subscribes to the shared event broker in the server's Store.
func (s *Server) getEventsSSEHandler(w http.ResponseWriter, r *http.Request) {
	// Set header to event stream
	w.Header().Set("Content-Type", "text/event-stream")
	// Initial status as OK, which signals to the client the stream is open
	w.WriteHeader(http.StatusOK)

	// Infinite loop. SSE stream closes once client context is closed.
	// .Done sends a channel that is closed when client closes sse connection.
	for {
		select {
		case <-r.Context().Done():
			return
		default:
		}
	}

	// Do SSE logic here.

}
