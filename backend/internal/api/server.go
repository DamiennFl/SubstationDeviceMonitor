package api

import (
	"net/http"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
)

type Server struct {
	store         store.Store
	deviceService DeviceService
}

func NewServer(deviceService DeviceService) *Server {
	return &Server{
		store:         *store.NewStore(),
		deviceService: deviceService,
	}
}

func (server *Server) ListenAndServe() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/devices/{deviceID}", server.getDeviceHandler)
	return mux
}
