package store

import (
	"context"
	"errors"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrDeviceNotFound       = errors.New("device not found")
	ErrDuplicateDevice      = errors.New("device name or IP already exists")
	ErrDuringDeviceUpdate   = errors.New("error occurred during device update")
	ErrDeviceAlreadyDeleted = errors.New("device has already been deleted")
	ErrNilDeviceUUID        = errors.New("uuid for a device was nil")

	ErrRuleNotFound     = errors.New("device rule not found")
	ErrDuplicateRule    = errors.New("device rule already exists")
	ErrNilRuleUUID      = errors.New("uuid for a rule was nil")
	ErrDuringRuleUpdate = errors.New("error occurred during rule update")
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

//----------------------------------------------------------------------------------
// Device Functions
//----------------------------------------------------------------------------------

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
		return model.Device{}, ErrNilDeviceUUID
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
		return model.Device{}, ErrDuringDeviceUpdate
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
		return model.Device{}, ErrDuringDeviceUpdate
	}

	// Update updated time to now
	newDevice.UpdatedAt = s.now()

	s.devices[deviceID] = newDevice

	return s.devices[deviceID], nil
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
		return ErrDeviceAlreadyDeleted
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

//----------------------------------------------------------------------------------
// Rule Functions
//----------------------------------------------------------------------------------

// ListRules returns a list of rules based on a device.
func (s *Store) ListRules(ctx context.Context, deviceID uuid.UUID) ([]model.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check that device actually exists first, before getting rule
	device, ok := s.devices[deviceID]
	if !ok || device.DeletedAt != nil {
		return []model.Rule{}, ErrDeviceNotFound
	}

	// Get Rule list, can be empty
	ruleList, ok := s.rules[deviceID]

	if !ok { // No list of rules, but that is valid. Return empty list with no error
		return []model.Rule{}, nil
	}

	// Make new list copy
	returnRuleList := make([]model.Rule, 0, len(ruleList))

	// Add rules
	for _, rule := range ruleList { // Iterate through map
		returnRuleList = append(returnRuleList, rule) // Add device to list
	}

	// Sort rules alphabetically for display
	sort.Slice(returnRuleList, func(i, j int) bool { // Sort alphabetically.
		return returnRuleList[i].Name < returnRuleList[j].Name // Compare
	})

	return returnRuleList, nil // Return sorted list
}

// GetRule returns a singular rule based on a device ID and rule ID.
func (s *Store) GetRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID) (model.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check that device actually exists first, before getting rule
	device, ok := s.devices[deviceID]
	if !ok || device.DeletedAt != nil {
		return model.Rule{}, ErrDeviceNotFound
	}

	// Get Rule list, can be empty
	ruleList, ok := s.rules[deviceID]

	if !ok { // No list of rules. We are looking for specific one, so return error.
		return model.Rule{}, ErrRuleNotFound
	}

	// Iterate until rule is found or end of list
	for _, rule := range ruleList {
		if rule.ID == ruleID {
			return rule, nil
		}
	}

	return model.Rule{}, ErrRuleNotFound

}

// CreateRule creates and stores a new Rule.
func (s *Store) CreateRule(ctx context.Context, rule model.Rule) (model.Rule, error) {
	s.mu.Lock()         // Lock Store
	defer s.mu.Unlock() // Defer the Unlock until function is done executing

	// Check for existing device
	device, ok := s.devices[rule.DeviceID]
	if !ok || device.DeletedAt != nil {
		return model.Rule{}, ErrDeviceNotFound
	}

	// Check for existing rulesList
	rulesList, ok := s.rules[rule.DeviceID]

	if !ok { // No rules list for the device yet, initialize it
		rulesList = make([]model.Rule, 0)
	}

	// Do not create if UUID is nil
	if rule.ID == uuid.Nil {
		return model.Rule{}, ErrNilRuleUUID
	}

	// Return duplicate rule error if there is one
	if s.hasDuplicateRule(rule, rulesList, nil) {
		return model.Rule{}, ErrDuplicateRule
	}

	// Update times
	now := s.now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	s.rules[rule.DeviceID] = append(rulesList, rule)

	return rule, nil
}

// UpdateRule updates one or more values of a Rule using a function value.
func (s *Store) UpdateRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID, edit func(rule model.Rule) (model.Rule, error)) (model.Rule, error) {
	// Lock and deferred Unlock
	s.mu.Lock()
	defer s.mu.Unlock()

	// No edit function, return error
	if edit == nil {
		return model.Rule{}, ErrDuringRuleUpdate
	}

	// Check device and validate
	device, ok := s.devices[deviceID]
	if !ok || device.DeletedAt != nil {
		return model.Rule{}, ErrDeviceNotFound
	}

	// Get rule list
	rulesList, ok := s.rules[deviceID]

	// No rules list exists, error out
	if !ok {
		return model.Rule{}, ErrRuleNotFound
	}

	// Find rule in existing rules list
	ruleIndex := -1
	var rule model.Rule

	for i, existingRule := range rulesList {
		if existingRule.ID == ruleID {
			ruleIndex = i
			rule = existingRule
			break
		}
	}

	// If rule was never found, error out
	if ruleIndex == -1 {
		return model.Rule{}, ErrRuleNotFound
	}

	// Run edit on device, get output
	newRule, err := edit(rule)

	// If error during edit, return empty device and edit's error
	if err != nil {
		return model.Rule{}, err
	}

	// Verify the rule identity is unchanged
	if newRule.ID != rule.ID || newRule.DeviceID != rule.DeviceID {
		return model.Rule{}, ErrDuringRuleUpdate
	}

	// Duplicate check for new rule; make sure it does not overlap with another
	if s.hasDuplicateRule(newRule, rulesList, &ruleID) {
		return model.Rule{}, ErrDuplicateRule
	}

	// Update updated time to now
	newRule.UpdatedAt = s.now()

	// Replace the existing rule
	rulesList[ruleIndex] = newRule
	s.rules[deviceID] = rulesList
	return newRule, nil
}

// DeleteRule deletes a rule from a device's list of rules.
func (s *Store) DeleteRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID) error {
	// Lock and Unlock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get device and status
	device, ok := s.devices[deviceID]
	if !ok {
		return ErrDeviceNotFound
	}

	if device.DeletedAt != nil {
		return ErrDeviceAlreadyDeleted
	}

	// Get rule list
	rulesList, ok := s.rules[deviceID]

	// No rules list exists, error out
	if !ok {
		return ErrRuleNotFound
	}

	// Get rule index in list
	ruleIndex := -1
	for i, existingRule := range rulesList {
		if existingRule.ID == ruleID {
			ruleIndex = i
			break
		}
	}

	// If rule was never found, error out
	if ruleIndex == -1 {
		return ErrRuleNotFound
	}

	// Delete element in list
	s.rules[deviceID] = slices.Delete(rulesList, ruleIndex, ruleIndex+1)

	// Return
	return nil
}

// hasDuplicateRule checks if a rules list contains a duplicate rule, optionally skips an item for update function
func (s *Store) hasDuplicateRule(rule model.Rule, rules []model.Rule, excludedID *uuid.UUID) bool {

	// No need for device check or rule list check, we only call this after we validate in other functions

	for _, existing := range rules { // Iterate rules
		if excludedID != nil && existing.ID == *excludedID { // Optionally exclude the current rule ID, for update
			continue
		}
		// Check for dup ID
		if existing.ID == rule.ID {
			return true
		}
		// Check for dup Name
		if existing.Name == rule.Name { // Duplicate Name
			return true
		}
	}

	return false
}
