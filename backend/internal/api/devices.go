package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
	"github.com/google/uuid"
)

// DeviceService interface. Defines the functions DeviceService can call
type DeviceService interface {
	GetDevice(ctx context.Context, deviceID uuid.UUID) (model.Device, error)
}

// Internal deviceService, with a reference to the Store
type deviceService struct {
	store *store.Store
}

// DeviceService constructor which initializes Store pointer for modification
func NewDeviceService(store *store.Store) DeviceService {
	return &deviceService{store: store}
}

// GetDevice function which matches the function definition in internal Store
func (s *deviceService) GetDevice(ctx context.Context, deviceID uuid.UUID) (model.Device, error) {
	return s.store.GetDevice(ctx, deviceID)
}

// getDeviceHandler which actually handles getting device, errors, and returning JSON
func (s *Server) getDeviceHandler(w http.ResponseWriter, r *http.Request) {

	// Error out if not GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deviceIDString := r.PathValue("deviceID")
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	device, err := s.deviceService.GetDevice(r.Context(), deviceID)

	// Return generic error, do not want to expose internal error.
	if errors.Is(err, store.ErrDeviceNotFound) {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(device); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}
