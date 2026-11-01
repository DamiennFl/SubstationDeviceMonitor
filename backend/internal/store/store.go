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

	ErrAlarmNotFound            = errors.New("alarm not found")
	ErrNilAlarmUUID             = errors.New("uuid for an alarm was nil")
	ErrExistingRaisedAlarm      = errors.New("existing raised alarm for rule was found")
	ErrNoAssociatedRule         = errors.New("no associated rule was set for alarm")
	ErrAlarmAlreadyAcknowledged = errors.New("alarm is already acknowledged")
	ErrUnacknowledgedAlarm      = errors.New("active alarm must be acknowledged before deletion")
)

// Store holds device data and protects it from concurrent access.
type Store struct {
	mu               sync.RWMutex                        // Mutex for locking Store
	devices          map[uuid.UUID]model.Device          // Device map
	rules            map[uuid.UUID][]model.Rule          // Rules: Device ID maps to list of Rules
	alarms           map[uuid.UUID][]model.Alarm         // Alarms: Device ID maps to list of Alarms
	acknowledgements map[uuid.UUID]model.Acknowledgement // Acknowledgements: Alarm ID maps to its Acknowledgement
	readings         map[uuid.UUID][]model.Reading       // Readings: Device ID maps to list of Readings
	now              func() time.Time                    // Current time, replaceable in tests
}

// Store constructor, which initializes the map and a pointer to the Store
func NewStore() *Store {
	return &Store{ // Return pointer
		devices:          make(map[uuid.UUID]model.Device),          // Make new devices map
		rules:            make(map[uuid.UUID][]model.Rule),          // Make new rules map
		alarms:           make(map[uuid.UUID][]model.Alarm),         // Make new alarms map
		acknowledgements: make(map[uuid.UUID]model.Acknowledgement), // Make new acknowledgements map
		readings:         make(map[uuid.UUID][]model.Reading),       // Make new readings map
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

	// Return error if device has unacknowledged active alarm.
	if s.hasUnacknowledgedActiveAlarm(deviceID, nil) {
		return ErrUnacknowledgedAlarm
	}

	// Clear all Alarms
	for index, alarm := range s.alarms[deviceID] {
		if alarm.ClearedAt == nil {
			alarm.ClearedAt = &currentTime
			reason := model.ClearReasonDeviceDeleted
			alarm.ClearReason = &reason
			s.alarms[deviceID][index] = alarm
		}
	}

	// Set deleted and updated to same time
	device.DeletedAt = &currentTime
	device.UpdatedAt = currentTime

	s.devices[deviceID] = device

	// Return
	return nil
}

// hasDuplicateDevice checks active devices for a matching name or IP.
func (s *Store) hasDuplicateDevice(device model.Device, excludedID *uuid.UUID) bool {

	// TODO: Once Alarm logic is implemented, add check to ensure Devices with Rules with active Alarms cannot be deleted.

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
		returnRuleList = append(returnRuleList, rule) // Add rule to list
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
	if hasDuplicateRule(rule, rulesList, nil) {
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
	if hasDuplicateRule(newRule, rulesList, &ruleID) {
		return model.Rule{}, ErrDuplicateRule
	}

	currentTime := s.now()

	// Rule has been disabled, clear alarm
	if !newRule.Enabled {
		s.clearActiveAlarm(deviceID, ruleID, currentTime, model.ClearReasonRuleDisabled)
	}

	// Update updated time to now
	newRule.UpdatedAt = currentTime

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
		return ErrDeviceNotFound
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

	// Return error if rule has unacknowledged alarm (s)
	if s.hasUnacknowledgedActiveAlarm(deviceID, &ruleID) {
		return ErrUnacknowledgedAlarm
	}

	currentTime := s.now()

	// Clear all Alarms
	for index, alarm := range s.alarms[deviceID] {
		if alarm.RuleID != nil && *alarm.RuleID == ruleID && alarm.ClearedAt == nil {
			alarm.ClearedAt = &currentTime
			reason := model.ClearReasonRuleDeleted
			alarm.ClearReason = &reason
			s.alarms[deviceID][index] = alarm
		}
	}

	// Delete element in list
	s.rules[deviceID] = slices.Delete(rulesList, ruleIndex, ruleIndex+1)

	// Return
	return nil
}

// hasDuplicateRule checks if a rules list contains a duplicate rule, optionally skips an item for update function
func hasDuplicateRule(rule model.Rule, rules []model.Rule, excludedID *uuid.UUID) bool {

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

//----------------------------------------------------------------------------------
// Alarm Functions
//----------------------------------------------------------------------------------

func (s *Store) ListAlarms(ctx context.Context, deviceID uuid.UUID) ([]model.AlarmWithAcknowledgement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check that device actually exists first, before getting alarms
	device, ok := s.devices[deviceID]
	if !ok || device.DeletedAt != nil {
		return []model.AlarmWithAcknowledgement{}, ErrDeviceNotFound
	}

	// Get Alarm list, can be empty
	alarmList, ok := s.alarms[deviceID]

	if !ok { // No list of Alarms, but that is valid. Return empty list with no error
		return []model.AlarmWithAcknowledgement{}, nil
	}

	// Make new list copy
	returnAlarmList := make([]model.AlarmWithAcknowledgement, 0, len(alarmList))

	// Copy alarms into the result slice
	for _, alarm := range alarmList {
		// Get alarm acknowledgement state
		acknowledgement, acknowledged := s.acknowledgements[alarm.ID]
		var acknowledgementPtr *model.Acknowledgement
		// If acknowledged, set acknowledgement pointer. Will be nil if alarm is not acknowledged
		if acknowledged {
			acknowledgementPtr = &acknowledgement
		}
		// Add alarm, ack pair to return list
		returnAlarmList = append(returnAlarmList, model.AlarmWithAcknowledgement{
			Alarm:           alarm,
			Acknowledgement: acknowledgementPtr,
		})
	}

	return returnAlarmList, nil // Return alarms newest first

}

func (s *Store) RaiseAlarm(ctx context.Context, alarm model.Alarm) (model.Alarm, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for existing device
	device, ok := s.devices[alarm.DeviceID]
	if !ok || device.DeletedAt != nil {
		return model.Alarm{}, ErrDeviceNotFound
	}

	// Check for existing alarmList
	alarmList, ok := s.alarms[alarm.DeviceID]

	if !ok { // No alarm list for the device yet, initialize it
		alarmList = make([]model.Alarm, 0)
	}

	// Do not create if UUID is nil
	if alarm.ID == uuid.Nil {
		return model.Alarm{}, ErrNilAlarmUUID
	}

	// RuleID nil pointer dereference check
	if alarm.RuleID == nil {
		return model.Alarm{}, ErrNoAssociatedRule
	}

	// Verify that the associated rule belongs to the device.
	rulesList, ok := s.rules[alarm.DeviceID]
	if !ok {
		return model.Alarm{}, ErrRuleNotFound
	}

	// Verify that Rule for Alarm exists
	ruleExists := false
	for _, rule := range rulesList {
		if rule.ID == *alarm.RuleID {
			ruleExists = true
			break
		}
	}

	// If Rule does not exist, return error
	if !ruleExists {
		return model.Alarm{}, ErrRuleNotFound
	}

	// Check if active alarm exists. If so, return with error, we don't want to infinitely add new alarms at each check.
	_, exists := s.findActiveAlarm(alarm.DeviceID, *alarm.RuleID)
	if exists {
		return model.Alarm{}, ErrExistingRaisedAlarm
	}

	// Add alarm to alarm list at the beginning.
	s.alarms[alarm.DeviceID] = slices.Insert(alarmList, 0, alarm)
	return alarm, nil
}

func (s *Store) ClearAlarm(ctx context.Context, alarm model.Alarm, clearedTime time.Time, reason model.ClearReason) (model.Alarm, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Do not clear if UUID is nil
	if alarm.ID == uuid.Nil {
		return model.Alarm{}, ErrNilAlarmUUID
	}

	// RuleID nil pointer dereference check
	if alarm.RuleID == nil {
		return model.Alarm{}, ErrNoAssociatedRule
	}

	// Clear Alarm
	clearedAlarm, err := s.clearActiveAlarm(alarm.DeviceID, *alarm.RuleID, clearedTime, reason)

	if err != nil {
		return model.Alarm{}, err
	}

	return clearedAlarm, nil
}

// AcknowledgeAlarm acknowledges an alarm and ties an Acknowledgement object to it.
func (s *Store) AcknowledgeAlarm(ctx context.Context, alarm model.Alarm, ack model.Acknowledgement) (model.Alarm, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if alarm.ID == uuid.Nil {
		return model.Alarm{}, ErrNilAlarmUUID
	}

	alarmList, ok := s.alarms[alarm.DeviceID]
	if !ok {
		return model.Alarm{}, ErrAlarmNotFound
	}

	// Find the exact stored alarm by ID. Cleared alarms may still be acknowledged.
	alarmIndex := -1
	for index, storedAlarm := range alarmList {
		if storedAlarm.ID == alarm.ID {
			alarmIndex = index
			break
		}
	}
	if alarmIndex == -1 {
		return model.Alarm{}, ErrAlarmNotFound
	}

	// The acknowledgement map is the single source of truth.
	if _, acknowledged := s.acknowledgements[alarm.ID]; acknowledged {
		return model.Alarm{}, ErrAlarmAlreadyAcknowledged
	}

	ack.AlarmID = alarm.ID
	ack.AcknowledgedAt = s.now()
	s.acknowledgements[alarm.ID] = ack

	newAlarm := alarmList[alarmIndex]
	return newAlarm, nil
}

// findActiveAlarm is a helper to find an active Alarm based on a Device and Rule. It must only be used within functions that have locked the Store.
func (s *Store) findActiveAlarm(deviceID uuid.UUID, ruleID uuid.UUID) (int, bool) {
	for index, alarm := range s.alarms[deviceID] {
		if alarm.RuleID != nil && *alarm.RuleID == ruleID && alarm.ClearedAt == nil {
			return index, true
		}
	}

	return -1, false
}

// hasUnacknowledgedActiveAlarm reports whether a device or specific rule has an
// active alarm without a corresponding acknowledgement.
func (s *Store) hasUnacknowledgedActiveAlarm(deviceID uuid.UUID, ruleID *uuid.UUID) bool {
	// Iterate alarms for device
	for _, alarm := range s.alarms[deviceID] {
		// Skip cleared alarms
		if alarm.ClearedAt != nil {
			continue
		}
		// Skip non-matching rules for ruleID (optional)
		if ruleID != nil && (alarm.RuleID == nil || *alarm.RuleID != *ruleID) {
			continue
		}
		// Found alarm, get acknowledged state. If not acknowledged, return true.
		if _, acknowledged := s.acknowledgements[alarm.ID]; !acknowledged {
			return true
		}
	}

	return false
}

// clearActiveAlarm is a helper to clear an active Alarm based on a Device and Rule. It must only be used within functions that have locked the Store.
func (s *Store) clearActiveAlarm(deviceID uuid.UUID, ruleID uuid.UUID, clearedTime time.Time, reason model.ClearReason) (model.Alarm, error) {
	// Check for existing device
	device, ok := s.devices[deviceID]
	if !ok || device.DeletedAt != nil {
		return model.Alarm{}, ErrDeviceNotFound
	}

	// Check for existing alarmList
	alarmList, ok := s.alarms[deviceID]

	if !ok || len(alarmList) == 0 { // No alarm list or alarms for the device yet, error out
		return model.Alarm{}, ErrAlarmNotFound
	}

	// Verify that the associated rule belongs to the device.
	rulesList, ok := s.rules[deviceID]
	if !ok {
		return model.Alarm{}, ErrRuleNotFound
	}

	// Verify that Rule for Alarm exists
	ruleExists := false
	for _, rule := range rulesList {
		if rule.ID == ruleID {
			ruleExists = true
			break
		}
	}

	// If Rule does not exist, return error
	if !ruleExists {
		return model.Alarm{}, ErrRuleNotFound
	}

	index, exists := s.findActiveAlarm(deviceID, ruleID)

	// If Alarm does not exist, nothing to clear. Return error
	if !exists {
		return model.Alarm{}, ErrAlarmNotFound
	}

	alarm := alarmList[index]

	alarm.ClearedAt = &clearedTime
	alarm.ClearReason = &reason

	alarmList[index] = alarm

	return alarm, nil
}
