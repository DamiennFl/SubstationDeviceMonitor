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

type DeviceService interface {
	GetDevice(ctx context.Context, deviceID uuid.UUID) (model.Device, error)
	ListDevices(ctx context.Context) ([]model.Device, error)
	CreateDevice(ctx context.Context, device model.Device) (model.Device, error)
	UpdateDevice(ctx context.Context, deviceID uuid.UUID, edit func(device model.Device) (model.Device, error)) (model.Device, error)
	DeleteDevice(ctx context.Context, deviceID uuid.UUID) error
}

// GetDevice
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
	} else {
		http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(device); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}
