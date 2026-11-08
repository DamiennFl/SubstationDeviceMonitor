package api

import (
	"net/http"
)

// Limit request body to 1 MB
const maxBodySize = 1024 * 1024

// Define Server with services
type Server struct {
	store       DataStore
	eventBroker *EventBroker
}

// Server constructor
func NewServer(storage DataStore) *Server {
	return &Server{
		store:       storage,
		eventBroker: NewEventBroker(),
	}
}

func (s *Server) live(w http.ResponseWriter, r *http.Request) {
	// Just say OK
	w.WriteHeader(http.StatusOK)
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	// Just say OK
	w.WriteHeader(http.StatusOK)
}

// Define Routes for server
// Runs once on startup, to register Routes. Then everything is handled by mux it returns.
// The mux matches the method and path, then calls the handler.
func (server *Server) Routes() http.Handler {
	// New multiplexer
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/devices", server.listDevicesHandler)
	mux.HandleFunc("POST /api/devices", server.createDeviceHandler)
	mux.HandleFunc("GET /api/devices/{device_id}", server.getDeviceHandler)
	mux.HandleFunc("PATCH /api/devices/{device_id}", server.updateDeviceHandler)
	mux.HandleFunc("DELETE /api/devices/{device_id}", server.deleteDeviceHandler)

	mux.HandleFunc("GET /api/devices/{device_id}/rules", server.getRulesListHandler)
	mux.HandleFunc("POST /api/devices/{device_id}/rules", server.createRuleHandler)
	mux.HandleFunc("GET /api/devices/{device_id}/rules/{rule_id}", server.getRuleHandler)
	mux.HandleFunc("PATCH /api/devices/{device_id}/rules/{rule_id}", server.updateRuleHandler)
	mux.HandleFunc("DELETE /api/devices/{device_id}/rules/{rule_id}", server.deleteRuleHandler)

	mux.HandleFunc("POST /api/readings", server.createReadingHandler) // Only for Simulator
	mux.HandleFunc("GET /api/devices/{device_id}/readings", server.getReadingsListHandler)

	mux.HandleFunc("GET /api/devices/{device_id}/alarms", server.getDeviceAlarmListHandler)
	mux.HandleFunc("POST /api/devices/{device_id}/alarms/{alarm_id}/acknowledge", server.acknowledgeAlarmHandler)

	mux.HandleFunc("GET /live", server.live)
	mux.HandleFunc("GET /ready", server.ready)

	mux.HandleFunc("GET /api/events", server.getEventsSSEHandler)
	return mux
}
