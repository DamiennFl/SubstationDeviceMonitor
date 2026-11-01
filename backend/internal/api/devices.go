package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
	"github.com/google/uuid"
)

// Limit request body to 1 MB
const maxBodySize = 1024 * 1024

// getDeviceHandler handles getting device
// Handles "GET /api/device/{device_id}" via store.GetDevice
func (s *Server) getDeviceHandler(w http.ResponseWriter, r *http.Request) {

	// Error out if not GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get deviceID from path string
	deviceIDString := r.PathValue("device_id")
	// Parse the UUID, return if it is invalid
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	// GetDevice through the deviceService
	device, err := s.store.GetDevice(r.Context(), deviceID)

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

// createDeviceHandler handles creating device.
// Handles "POST /api/device" via store.CreateDevice
func (s *Server) createDeviceHandler(w http.ResponseWriter, r *http.Request) {
	// Close body since it is data stream once function is finished, to free system resources and prevent data leak.
	defer r.Body.Close()

	// Validate method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	/* type CreateDeviceInput struct {
	Name       string     `json:"name"`
	IP         string     `json:"ip"`
	DeviceType DeviceType `json:"type"`
	}
	*/
	var deviceInput model.CreateDeviceInput

	// Make new decoder with a 1MB limit
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxBodySize))
	decoder.DisallowUnknownFields()

	// Decode into deviceInput, handle error
	if err := decoder.Decode(&deviceInput); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Using &struct{}{} checks that only one JSON value was decoded.
	// If more than one JSON value is present, Decode will return nil, instead of
	// io.EOF.
	if decoder.Decode(&struct{}{}) != io.EOF {
		http.Error(w, "Request body must contain one JSON object", http.StatusBadRequest)
		return
	}

	// Validate input fields
	if deviceInput.Name == "" {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if net.ParseIP(deviceInput.IP) == nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if !deviceInput.DeviceType.IsValid() {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Define new Device with deviceInput fields
	device := model.Device{
		ID:         uuid.New(),
		Name:       deviceInput.Name,
		IP:         string(net.ParseIP(deviceInput.IP)),
		DeviceType: deviceInput.DeviceType,
	}

	// Create Device in Store
	createdDevice, err := s.store.CreateDevice(r.Context(), device)

	// If error during Device creation, print corresponding error
	if err != nil {
		switch {
		case errors.Is(err, store.ErrDuplicateDevice):
			http.Error(w, "Device with that name or IP already exists", http.StatusConflict)
		case errors.Is(err, store.ErrNilDeviceUUID):
			http.Error(w, "Could not create device ID", http.StatusInternalServerError)
		default:
			http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return newly created device encoded into JSON; will be used by frontend to display
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdDevice); err != nil {
		log.Printf("failed to encode created device response: %v", err)
	}
}

// updateDeviceHandler handles updating device
// Handles "PATCH /api/device/{device_id}" via Store.UpdateDevice
func (s *Server) updateDeviceHandler(w http.ResponseWriter, r *http.Request) {
	// Close body since it is data stream once function is finished, to free system resources and prevent data leak.
	defer r.Body.Close()

	// Validate method
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get deviceID from path string
	deviceIDString := r.PathValue("device_id")
	// Parse the UUID, return if it is invalid
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	/* type UpdateDeviceInput struct {
	Name *string `json:"name"`
	IP   *string `json:"ip"`
	}
	*/
	var deviceInput model.UpdateDeviceInput
	// Make new decoder with a 1MB limit
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxBodySize))
	decoder.DisallowUnknownFields()

	// Decode into deviceInput, handle error
	if err := decoder.Decode(&deviceInput); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Using &struct{}{} checks that only one JSON value was decoded.
	// If more than one JSON value is present, Decode will return nil, instead of
	// io.EOF.
	if decoder.Decode(&struct{}{}) != io.EOF {
		http.Error(w, "Request body must contain one JSON object", http.StatusBadRequest)
		return
	}
	// Validate input fields
	if deviceInput.Name != nil && *deviceInput.Name == "" {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if deviceInput.IP != nil && net.ParseIP(*deviceInput.IP) == nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	updatedDevice, err := s.store.UpdateDevice(
		r.Context(),
		deviceID,
		func(device model.Device) (model.Device, error) {
			if deviceInput.Name != nil {
				device.Name = *deviceInput.Name
			}
			if deviceInput.IP != nil {
				device.IP = string(net.ParseIP(*deviceInput.IP))
			}
			return device, nil
		})

	if err != nil {
		switch {
		case errors.Is(err, store.ErrDeviceNotFound):
			http.Error(w, "Device not found", http.StatusNotFound)
		case errors.Is(err, store.ErrDuplicateDevice):
			http.Error(w, "Device with that name or IP already exists", http.StatusConflict)
		default:
			http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return the updated device encoded into JSON.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updatedDevice); err != nil {
		log.Printf("failed to encode updated device response: %v", err)
	}
}

// deleteDeviceHandler handles deleting device.
// Handles "DELETE /api/device/{device_id}"
func (s *Server) deleteDeviceHandler(w http.ResponseWriter, r *http.Request) {

	// Validate method
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get deviceID from path string
	deviceIDString := r.PathValue("device_id")
	// Parse the UUID, return if it is invalid
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	err = s.store.DeleteDevice(r.Context(), deviceID)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrDeviceNotFound):
			http.Error(w, "Device not found", http.StatusNotFound)
		case errors.Is(err, store.ErrDeviceAlreadyDeleted):
			http.Error(w, "Device is already deleted", http.StatusConflict)
		case errors.Is(err, store.ErrUnacknowledgedAlarm):
			http.Error(w, "Unacknowledged alarm for device", http.StatusConflict)
		default:
			http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return 204
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listDevicesHandler(w http.ResponseWriter, r *http.Request) {
	// Error out if not GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// GetDevice through the deviceService
	deviceList, err := s.store.ListDevices(r.Context())

	// Does nothing for now: Will be useful when Postgres is implemented
	if err != nil {
		http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
	}

	// Set return header to JSON content-type
	w.Header().Set("Content-Type", "application/json")

	// Try and encode device return to writer. Error out if error occurs, otherwise the device JSON will be written out.
	if err := json.NewEncoder(w).Encode(deviceList); err != nil {
		log.Printf("failed to encode device response: %v", err)
		return
	}
}
