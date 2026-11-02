package api

import (
	"net/http"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
)

// Limit request body to 1 MB
const maxBodySize = 1024 * 1024

// Define Server with services
type Server struct {
	store *store.Store
}

// Server constructor
func NewServer(store *store.Store) *Server {
	return &Server{
		store: store,
	}
}

// Define Routes for server
// Runs once on startup, to register Routes. Then everything is handled by mux it returns.
// The mux matches the method and path, then calls the handler.
func (server *Server) Routes() http.Handler {
	// New multiplexer
	mux := http.NewServeMux()

	// Handle GET device
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

	return mux
}
