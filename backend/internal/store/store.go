package store

import (
	"sort"
	"sync"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

// Store holds device data and protects it from concurrent access.
type Store struct {
	mu      sync.RWMutex               // Mutex for locking Store
	devices map[uuid.UUID]model.Device // Device map
}

// Store constructor, which initializes the map and a pointer to the Store
func newStore() *Store {
	return &Store{ // Return pointer
		devices: make(map[uuid.UUID]model.Device), // Make new map
	}
}

// List devices present in this Store's map, alphabetically.
func (s *Store) listDevices() []model.Device {
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

	return deviceList // Return sorted list
}
