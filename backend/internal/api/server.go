package api

import (
	"net/http"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
)

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
	return mux
}
