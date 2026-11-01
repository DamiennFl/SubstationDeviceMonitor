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
func (server *Server) Routes() http.Handler {
	// New mutex
	mux := http.NewServeMux()

	// Handle GET device
	mux.HandleFunc("GET /api/devices/{deviceID}", server.getDeviceHandler)
	return mux
}
