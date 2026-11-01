package api

import (
	"net/http"
)

// Define Server with services
type Server struct {
	deviceService DeviceService
}

// Server constructor
func NewServer(deviceService DeviceService) *Server {
	return &Server{
		deviceService: deviceService,
	}
}

// Define Routes for server
// Runs once on startup, to register Routes. Then everything is handled by mux it returns.
// The mux matches the method and path, then calls the handler.
func (server *Server) Routes() http.Handler {
	// New multiplexer
	mux := http.NewServeMux()

	// Handle GET device
	mux.HandleFunc("GET /api/devices/{deviceID}", server.getDeviceHandler)
	return mux
}
