package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
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

	// Get deviceID from path string
	deviceIDString := r.PathValue("deviceID")
	// Parse the UUID, return if it is invalid
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	// GetDevice through the deviceService
	device, err := s.deviceService.GetDevice(r.Context(), deviceID)

	// Return generic error if device is not found, do not want to expose internal error.
	if errors.Is(err, store.ErrDeviceNotFound) {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
		// If any other error, return unexpected error
	} else if err != nil {
		http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		return
	}

	// Set return header to JSON content-type
	w.Header().Set("Content-Type", "application/json")

	// Try and encode device return to writer. Error out if error occurs, otherwise the device JSON will be written out.
	if err := json.NewEncoder(w).Encode(device); err != nil {
		log.Printf("failed to encode device response: %v", err)
		return
	}
}
