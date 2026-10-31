package store

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrDeviceNotFound  = errors.New("device not found")
	ErrDuplicateDevice = errors.New("device name or IP already exists")
	ErrDuringUpdate    = errors.New("error occurred during device update")
	ErrAlreadyDeleted  = errors.New("device has already been deleted")
	ErrNilUUID         = errors.New("uuid for a device was nil")
)

// Store holds device data and protects it from concurrent access.
type Store struct {
	mu       sync.RWMutex                  // Mutex for locking Store
	devices  map[uuid.UUID]model.Device    // Device map
	rules    map[uuid.UUID][]model.Rule    // Rules: Device ID maps to list of Rules
	readings map[uuid.UUID][]model.Reading // Readings: Device ID maps to list of Readings
	now      func() time.Time              // Current time, replaceable in tests
}

// Store constructor, which initializes the map and a pointer to the Store
func NewStore(ctx context.Context) *Store {
	return &Store{ // Return pointer
		devices:  make(map[uuid.UUID]model.Device),    // Make new devices map
		rules:    make(map[uuid.UUID][]model.Rule),    // Make new rules map
		readings: make(map[uuid.UUID][]model.Reading), // Make new readings map
		now: func() time.Time { // Use the real UTC clock by default
			return time.Now().UTC()
		},
	}
}

// List devices present in this Store's map, alphabetically.
func (s *Store) ListDevices(ctx context.Context) ([]model.Device, error) {
	s.mu.RLock()         // Lock Store
	defer s.mu.RUnlock() // Defer the Unlock until the function is done executing

	deviceList := make([]model.Device, 0, len(s.devices)) // Make device list with capacity equal to map length
	for _, device := range s.devices {                    // Iterate through map
		if device.DeletedAt != nil { // Check that device is not soft-deleted, skip if it is
			continue
		}
		deviceList = append(deviceList, device) // Add device to list
	}

	sort.Slice(deviceList, func(i, j int) bool { // Sort alphabetically.
		return deviceList[i].Name < deviceList[j].Name // Compare
	})

	return deviceList, nil // Return sorted list
}

// Get a specific device based on its UUID
func (s *Store) GetDevice(ctx context.Context, deviceID uuid.UUID) (model.Device, error) {
	s.mu.RLock()         // Lock Store
	defer s.mu.RUnlock() // Defer the Unlock until the function is done executing

	device, ok := s.devices[deviceID] // Get device, if it exists
	if !ok || device.DeletedAt != nil {
		return model.Device{}, ErrDeviceNotFound // If it does not exist, return zero Device and error.
	}

	return device, nil // Return device
}

// CreateDevice adds passed in device to this Store's map
func (s *Store) CreateDevice(ctx context.Context, device model.Device) (model.Device, error) {
	s.mu.Lock()         // Lock Store
	defer s.mu.Unlock() // Defer the Unlock until function is done executing

	if s.hasDuplicateDevice(device, nil) {
		return model.Device{}, ErrDuplicateDevice // Return nil struct and error
	}

	// Do not create if UUID is nil
	if device.ID == uuid.Nil {
		return model.Device{}, ErrNilUUID
	}
	// Update times
	now := s.now()
	device.CreatedAt = now
	device.UpdatedAt = now

	s.devices[device.ID] = device // Set device

	return device, nil // Return it
}

// UpdateDevice updates a device's values
func (s *Store) UpdateDevice(ctx context.Context, deviceID uuid.UUID, edit func(device model.Device) (model.Device, error)) (model.Device, error) {
	// Lock and deferred Unlock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get device and validate
	device, ok := s.devices[deviceID]
	if !ok || device.DeletedAt != nil {
		return model.Device{}, ErrDeviceNotFound
	}

	if edit == nil {
		return model.Device{}, ErrDuringUpdate
	}

	// Run edit on device, get output
	newDevice, err := edit(device)

	// If error during edit, return empty device and edit's error
	if err != nil {
		return model.Device{}, err
	}

	// Duplicate check
	if s.hasDuplicateDevice(newDevice, &deviceID) {
		return model.Device{}, ErrDuplicateDevice // Return nil struct and error
	}

	// Verify ID is the same
	if newDevice.ID != device.ID {
		return model.Device{}, ErrDuringUpdate
	}

	// Update updated time to now
	newDevice.UpdatedAt = s.now()

	s.devices[deviceID] = newDevice

	return s.devices[deviceID], nil
}

// hasDuplicateDevice checks active devices for a matching name or IP.
func (s *Store) hasDuplicateDevice(device model.Device, excludedID *uuid.UUID) bool {
	for _, existing := range s.devices { // Iterate devices
		if existing.DeletedAt != nil { // Device has been deleted, skip
			continue
		}
		if excludedID != nil && existing.ID == *excludedID { // Optionally exclude the current device ID, for update
			continue
		}
		if existing.Name == device.Name || existing.IP == device.IP { // Duplicate Name or IP
			return true
		}
	}

	return false
}

func (s *Store) DeleteDevice(ctx context.Context, deviceID uuid.UUID) error {
	// Lock and Unlock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get device and status
	device, ok := s.devices[deviceID]
	if !ok {
		return ErrDeviceNotFound
	}

	if device.DeletedAt != nil {
		return ErrAlreadyDeleted
	}

	// Get time
	currentTime := s.now()

	// Set deleted and updated to same time
	device.DeletedAt = &currentTime
	device.UpdatedAt = currentTime

	s.devices[deviceID] = device

	// Return
	return nil
}
