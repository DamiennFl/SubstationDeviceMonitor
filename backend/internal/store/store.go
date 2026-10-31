package store

import (
	"errors"
	"sort"
	"sync"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrDeviceNotFound  = errors.New("device not found")
	ErrDuplicateDevice = errors.New("device name or IP already exists")
)

// Store holds device data and protects it from concurrent access.
type Store struct {
	mu      sync.RWMutex               // Mutex for locking Store
	devices map[uuid.UUID]model.Device // Device map
}

// Store constructor, which initializes the map and a pointer to the Store
func NewStore() *Store {
	return &Store{ // Return pointer
		devices: make(map[uuid.UUID]model.Device), // Make new map
	}
}

// List devices present in this Store's map, alphabetically.
func (s *Store) ListDevices() ([]model.Device, error) {
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
func (s *Store) GetDevice(deviceID uuid.UUID) (model.Device, error) {
	s.mu.RLock()         // Lock Store
	defer s.mu.RUnlock() // Defer the Unlock until the function is done executing

	device, ok := s.devices[deviceID] // Get device, if it exists
	if !ok || device.DeletedAt != nil {
		return model.Device{}, ErrDeviceNotFound // If it does not exist, return zero Device and error.
	}

	return device, nil // Return device
}

// Adds passed in device to this Store's map
func (s *Store) CreateDevice(device model.Device) (model.Device, error) {
	s.mu.Lock()         // Lock Store
	defer s.mu.Unlock() // Defer the Unlock until function is done executing

	for _, existing := range s.devices { // Iterate devices
		if existing.DeletedAt != nil { // Skip soft-deleted
			continue
		}
		if existing.Name == device.Name || existing.IP == device.IP { // If a field is matching, it is duplicate, don't add
			return model.Device{}, ErrDuplicateDevice // Return nil struct and error
		}
	}

	s.devices[device.ID] = device // Set device

	return device, nil // Return it
}
