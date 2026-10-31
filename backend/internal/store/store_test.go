package store

import (
	"errors"
	"testing"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

// testDevice minimal model
func testDevice(name, ip string) model.Device {
	return model.Device{
		ID:   uuid.New(),
		Name: name,
		IP:   ip,
	}
}

// testRule minimal model
func testRule(deviceID uuid.UUID, name string) model.Rule {
	return model.Rule{
		ID:       uuid.New(),
		DeviceID: deviceID,
		Name:     name,
	}
}

// Regular case: NewStore returns an initialized store with no devices.
func TestNewStore(t *testing.T) {
	s := NewStore(t.Context())
	if s == nil {
		t.Fatal("NewStore returned nil")
	}
	if devices, err := s.ListDevices(t.Context()); devices == nil || len(devices) != 0 || err != nil {
		t.Fatalf("ListDevices() = %#v, want a non-nil empty slice", devices)
	}
}

// Edge case: NewStore calls create independent stores with independent maps.
func TestNewStoreEdgeCase(t *testing.T) {
	first := NewStore(t.Context())
	second := NewStore(t.Context())
	device := testDevice("Relay A", "192.0.2.1")

	if _, err := first.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	if devices, err := second.ListDevices(t.Context()); len(devices) != 0 || err != nil {
		t.Fatalf("second store contains devices = %#v, want empty", devices)
	}
}

// Regular case: ListDevices returns active devices in name order.
func TestListDevices(t *testing.T) {
	s := NewStore(t.Context())
	for _, device := range []model.Device{
		testDevice("Relay C", "192.0.2.3"),
		testDevice("Relay A", "192.0.2.1"),
		testDevice("Relay B", "192.0.2.2"),
	} {
		if _, err := s.CreateDevice(t.Context(), device); err != nil {
			t.Fatalf("CreateDevice() error = %v", err)
		}
	}

	devices, _ := s.ListDevices(t.Context())
	if len(devices) != 3 {
		t.Fatalf("ListDevices() returned %d devices, want 3", len(devices))
	}
	for i, want := range []string{"Relay A", "Relay B", "Relay C"} {
		if devices[i].Name != want {
			t.Errorf("ListDevices()[%d].Name = %q, want %q", i, devices[i].Name, want)
		}
	}
}

// Edge case: ListDevices returns a non-nil empty slice for an empty
func TestListDevicesEdgeCase(t *testing.T) {
	devices, err := NewStore(t.Context()).ListDevices(t.Context())
	if devices == nil || len(devices) != 0 || err != nil {
		t.Fatalf("ListDevices() = %#v, want a non-nil empty slice", devices)
	}
}

// Invalid case: ListDevices excludes soft-deleted devices.
func TestListDevicesInvalidCase(t *testing.T) {
	s := NewStore(t.Context())
	device := testDevice("Deleted Relay", "192.0.2.10")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	if err := s.DeleteDevice(t.Context(), device.ID); err != nil {
		t.Fatalf("DeleteDevice() error = %v", err)
	}
	if devices, _ := s.ListDevices(t.Context()); len(devices) != 0 {
		t.Fatalf("ListDevices() = %#v, want no active devices", devices)
	}
}

// Regular case: GetDevice returns an existing active device.
func TestGetDevice(t *testing.T) {
	s := NewStore(t.Context())
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	got, err := s.GetDevice(t.Context(), device.ID)
	if err != nil {
		t.Fatalf("GetDevice() error = %v", err)
	}
	if got.ID != device.ID || got.Name != device.Name {
		t.Fatalf("GetDevice() = %#v, want %#v", got, device)
	}
}

// Edge case: GetDevice cannot retrieve a device whose ID is the zero UUID.
func TestGetDeviceEdgeCase(t *testing.T) {
	s := NewStore(t.Context())
	device := testDevice("Zero ID Relay", "192.0.2.2")
	device.ID = uuid.Nil
	if _, err := s.CreateDevice(t.Context(), device); !errors.Is(err, ErrNilDeviceUUID) {
		t.Fatalf("CreateDevice() error = %v, want ErrNilUUID", err)
	}

	_, err := s.GetDevice(t.Context(), uuid.Nil)
	if err == nil {
		t.Fatalf("GetDevice() returned device when error was expected.")
	}
}

// Invalid case: GetDevice returns ErrDeviceNotFound for an unknown ID.
func TestGetDeviceInvalidCase(t *testing.T) {
	s := NewStore(t.Context())
	_, err := s.GetDevice(t.Context(), uuid.New())
	if !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("GetDevice() error = %v, want ErrDeviceNotFound", err)
	}
}

// Regular case: CreateDevice adds a device and returns the created value.
func TestCreateDevice(t *testing.T) {
	s := NewStore(t.Context())
	device := testDevice("Relay A", "192.0.2.1")

	got, err := s.CreateDevice(t.Context(), device)
	if err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	if got.ID != device.ID || got.Name != device.Name || got.IP != device.IP {
		t.Fatalf("CreateDevice() = %#v, want device fields from %#v", got, device)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatalf("CreateDevice() returned zero timestamps: %#v", got)
	}
	if stored, err := s.GetDevice(t.Context(), device.ID); err != nil || stored.ID != device.ID {
		t.Fatalf("GetDevice() = %#v, %v; want device ID %v, nil", stored, err, device.ID)
	}
}

// Edge case: CreateDevice allows a new device to reuse values from a soft-deleted device.
func TestCreateDeviceEdgeCase(t *testing.T) {
	s := NewStore(t.Context())
	deleted := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), deleted); err != nil {
		t.Fatalf("CreateDevice(deleted) error = %v", err)
	}
	if err := s.DeleteDevice(t.Context(), deleted.ID); err != nil {
		t.Fatalf("DeleteDevice() error = %v", err)
	}

	active := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), active); err != nil {
		t.Fatalf("CreateDevice(active) error = %v", err)
	}
}

// Invalid case: CreateDevice rejects duplicate names and duplicate IP addresses.
func TestCreateDeviceInvalidCase(t *testing.T) {
	tests := []struct {
		name  string
		build func(model.Device) model.Device
	}{
		{
			name: "duplicate name",
			build: func(device model.Device) model.Device {
				device.IP = "192.0.2.2"
				return device
			},
		},
		{
			name: "duplicate IP",
			build: func(device model.Device) model.Device {
				device.Name = "Relay B"
				return device
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewStore(t.Context())
			existing := testDevice("Relay A", "192.0.2.1")
			if _, err := s.CreateDevice(t.Context(), existing); err != nil {
				t.Fatalf("CreateDevice(existing) error = %v", err)
			}

			_, err := s.CreateDevice(t.Context(), test.build(testDevice("Relay A", "192.0.2.1")))
			if !errors.Is(err, ErrDuplicateDevice) {
				t.Fatalf("CreateDevice() error = %v, want ErrDuplicateDevice", err)
			}
			if devices, _ := s.ListDevices(t.Context()); len(devices) != 1 {
				t.Fatalf("ListDevices() returned %d devices, want 1", len(devices))
			}
		})
	}
}

// Regular case: UpdateDevice uses the Store clock for UpdatedAt.
func TestUpdateDeviceUsesInjectedClock(t *testing.T) {
	const fixedTime = "2026-10-30T20:00:00Z"
	now, err := time.Parse(time.RFC3339, fixedTime) // Get now
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}

	store := NewStore(t.Context()) // Create new store
	store.now = func() time.Time { // Set store's clock to injected clock
		return now
	}

	device := model.Device{ // Create new device
		ID:   uuid.New(),
		Name: "Relay A",
		IP:   "192.0.2.1",
	}
	if _, err := store.CreateDevice(t.Context(), device); err != nil { // Store device
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Update device
	updated, err := store.UpdateDevice(t.Context(), device.ID, func(device model.Device) (model.Device, error) {
		device.Name = "Relay B"
		return device, nil
	})
	if err != nil {
		t.Fatalf("UpdateDevice() error = %v", err)
	}
	// Check that updatedAt equals now, which would be set to the time at which device was updated in store.
	if !updated.UpdatedAt.Equal(now) {
		t.Fatalf("UpdatedAt = %v, want %v", updated.UpdatedAt, now)
	}
}

// Regular case: DeleteDevice marks a device deleted and updates its timestamps.
func TestDeleteDevice(t *testing.T) {
	const fixedTime = "2026-10-30T21:00:00Z"
	// Get fixed time
	now, err := time.Parse(time.RFC3339, fixedTime)
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}

	// New store, set s.now to fixed time
	s := NewStore(t.Context())
	s.now = func() time.Time { return now }
	// Make new device
	device := testDevice("Relay A", "192.0.2.1")

	// Check device creation
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// Delete it
	if err := s.DeleteDevice(t.Context(), device.ID); err != nil {
		t.Fatalf("DeleteDevice() error = %v", err)
	}

	// Check trying to get device gives ErrDeviceNotFound
	if _, err := s.GetDevice(t.Context(), device.ID); !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("GetDevice() error = %v, want ErrDeviceNotFound", err)
	}
	// Get soft-deleted device
	deleted := s.devices[device.ID]
	// Check that DeletedAt is set, and that it is equal to fixed time
	if deleted.DeletedAt == nil || !deleted.DeletedAt.Equal(now) {
		t.Fatalf("DeletedAt = %v, want %v", deleted.DeletedAt, now)
	}
	//Check deleted time is same as fixed time
	if !deleted.UpdatedAt.Equal(now) {
		t.Fatalf("UpdatedAt = %v, want %v", deleted.UpdatedAt, now)
	}
}

// Edge case: DeleteDevice rejects deleting an already-deleted device.
func TestDeleteDeviceEdgeCase(t *testing.T) {
	s := NewStore(t.Context())
	// Make device and ensure it is created in Store
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// Delete it
	if err := s.DeleteDevice(t.Context(), device.ID); err != nil {
		t.Fatalf("first DeleteDevice() error = %v", err)
	}
	// Try and delete again; should error
	if err := s.DeleteDevice(t.Context(), device.ID); !errors.Is(err, ErrDeviceAlreadyDeleted) {
		t.Fatalf("second DeleteDevice() error = %v, want ErrAlreadyDeleted", err)
	}
}

// Invalid case: DeleteDevice returns ErrDeviceNotFound for an unknown device.
func TestDeleteDeviceInvalidCase(t *testing.T) {
	if err := NewStore(t.Context()).DeleteDevice(t.Context(), uuid.New()); !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("DeleteDevice() error = %v, want ErrDeviceNotFound", err)
	}
}

// Regular case: A deleted device's name and IP can be reused.
func TestCreateDeviceAfterDelete(t *testing.T) {
	s := NewStore(t.Context())
	// Make device and ensure it is created in Store
	deleted := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), deleted); err != nil {
		t.Fatalf("CreateDevice(deleted) error = %v", err)
	}
	// Delete it
	if err := s.DeleteDevice(t.Context(), deleted.ID); err != nil {
		t.Fatalf("DeleteDevice() error = %v", err)
	}
	// Test that adding new device with same values is allowed
	replacement := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), replacement); err != nil {
		t.Fatalf("CreateDevice(replacement) error = %v", err)
	}
}

// Regular case: UpdateDevice applies an edit and preserves the device ID.
func TestUpdateDevice(t *testing.T) {
	s := NewStore(t.Context())
	// Make device and ensure it is created in Store
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Update the name, via function value
	updated, err := s.UpdateDevice(t.Context(), device.ID, func(device model.Device) (model.Device, error) {
		device.Name = "Relay B"
		return device, nil
	})
	// If error, return error and fail test
	if err != nil {
		t.Fatalf("UpdateDevice() error = %v", err)
	}
	// If deviceID changed, or Name did not change, return error and fail test
	if updated.ID != device.ID || updated.Name != "Relay B" {
		t.Fatalf("UpdateDevice() = %#v, want ID %v and name Relay B", updated, device.ID)
	}
}

// Edge case: UpdateDevice rejects unknown and deleted device IDs.
func TestUpdateDeviceNotFound(t *testing.T) {
	s := NewStore(t.Context())
	// Update device but with unknown UUID.
	if _, err := s.UpdateDevice(t.Context(), uuid.New(), func(device model.Device) (model.Device, error) {
		return device, nil
	}); !errors.Is(err, ErrDeviceNotFound) { // Confirm device is not found and errors
		t.Fatalf("UpdateDevice(unknown) error = %v, want ErrDeviceNotFound", err)
	}

	// Make device and ensure it is created in Store
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// Delete it
	if err := s.DeleteDevice(t.Context(), device.ID); err != nil {
		t.Fatalf("DeleteDevice() error = %v", err)
	}
	// Try and update deleted device
	if _, err := s.UpdateDevice(t.Context(), device.ID, func(device model.Device) (model.Device, error) {
		return device, nil
	}); !errors.Is(err, ErrDeviceNotFound) { // Confirm device is not found and errors
		t.Fatalf("UpdateDevice(deleted) error = %v, want ErrDeviceNotFound", err)
	}
}

// Invalid case: UpdateDevice returns the edit error without saving changes.
func TestUpdateDeviceEditError(t *testing.T) {
	s := NewStore(t.Context())
	device := testDevice("Relay A", "192.0.2.1")
	// Make device and ensure it is created in Store
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	editErr := errors.New("edit failed")

	// Update device, but return error after edit. Function value will inherit error and return it
	// 	WITHOUT changing device Name value.
	_, err := s.UpdateDevice(t.Context(), device.ID, func(device model.Device) (model.Device, error) {
		device.Name = "Changed"
		return device, editErr
	})
	if !errors.Is(err, editErr) { // Confirm error was returned
		t.Fatalf("UpdateDevice() error = %v, want edit error", err)
	}
	// Confirm that device name was not changed due to error in function value return
	stored, err := s.GetDevice(t.Context(), device.ID)
	if err != nil || stored.Name != device.Name {
		t.Fatalf("GetDevice() = %#v, %v; want unchanged device", stored, err)
	}
}

// Invalid case: UpdateDevice rejects a name or IP used by another device.
func TestUpdateDeviceDuplicate(t *testing.T) {
	// Define two test cases: One modifies device's name, one IP
	tests := []struct {
		name string
		edit func(model.Device) model.Device
	}{
		{
			name: "name",
			edit: func(device model.Device) model.Device {
				device.Name = "Relay B"
				return device
			},
		},
		{
			name: "IP",
			edit: func(device model.Device) model.Device {
				device.IP = "192.0.2.2"
				return device
			},
		},
	}

	// Add two devices to Store in each test
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewStore(t.Context())
			first := testDevice("Relay A", "192.0.2.1")
			second := testDevice("Relay B", "192.0.2.2")
			// Ensure both devices are created in both tests
			for _, device := range []model.Device{first, second} {
				if _, err := s.CreateDevice(t.Context(), device); err != nil {
					t.Fatalf("CreateDevice() error = %v", err)
				}
			}
			// Try to change first's to "Relay B", which already exists. Expect error.
			_, err := s.UpdateDevice(t.Context(), first.ID, func(device model.Device) (model.Device, error) {
				return test.edit(second), nil
			})
			if !errors.Is(err, ErrDuplicateDevice) {
				t.Fatalf("UpdateDevice() error = %v, want ErrDuplicateDevice", err)
			}
			stored, err := s.GetDevice(t.Context(), first.ID)
			if err != nil || stored.Name != first.Name || stored.IP != first.IP {
				t.Fatalf("GetDevice() = %#v, %v; want unchanged device", stored, err)
			}
		})
	}
}

// Invalid case: UpdateDevice rejects an edit that changes the device ID.
func TestUpdateDeviceIDChange(t *testing.T) {
	s := NewStore(t.Context())
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	_, err := s.UpdateDevice(t.Context(), device.ID, func(device model.Device) (model.Device, error) {
		device.ID = uuid.New()
		return device, nil
	})
	if !errors.Is(err, ErrDuringDeviceUpdate) {
		t.Fatalf("UpdateDevice() error = %v, want ErrDuringUpdate", err)
	}
}

// Regular case: ListRules returns rules in name order.
func TestListRules(t *testing.T) {
	s := NewStore(t.Context())
	// New device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// New rules and validate
	for _, rule := range []model.Rule{
		testRule(device.ID, "Voltage"),
		testRule(device.ID, "Current"),
	} {
		if _, err := s.CreateRule(t.Context(), rule); err != nil {
			t.Fatalf("CreateRule() error = %v", err)
		}
	}

	// List rules
	rules, err := s.ListRules(t.Context(), device.ID)
	if err != nil {
		t.Fatalf("ListRules() error = %v", err)
	}
	// Check they are alphabetical
	if len(rules) != 2 || rules[0].Name != "Current" || rules[1].Name != "Voltage" {
		t.Fatalf("ListRules() = %#v, want rules sorted by name", rules)
	}
}

// Edge case: ListRules returns a non-nil empty slice for a device with no rules.
func TestListRulesEdgeCase(t *testing.T) {
	s := NewStore(t.Context())
	// New device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// List rules; validate empty list not error
	rules, err := s.ListRules(t.Context(), device.ID)
	if err != nil || rules == nil || len(rules) != 0 {
		t.Fatalf("ListRules() = %#v, %v; want non-nil empty slice", rules, err)
	}
}

// Invalid case: ListRules rejects an unknown device.
func TestListRulesInvalidCase(t *testing.T) {
	// New store with unknown device
	rules, err := NewStore(t.Context()).ListRules(t.Context(), uuid.New())
	if !errors.Is(err, ErrDeviceNotFound) || rules == nil || len(rules) != 0 {
		t.Fatalf("ListRules() = %#v, %v; want empty slice and ErrDeviceNotFound", rules, err)
	}
}

// Regular case: GetRule returns a rule belonging to the requested device.
func TestGetRule(t *testing.T) {
	s := NewStore(t.Context())
	// New device and rule, and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	rule := testRule(device.ID, "Overvoltage")
	if _, err := s.CreateRule(t.Context(), rule); err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}

	// Get rule
	got, err := s.GetRule(t.Context(), device.ID, rule.ID)
	if err != nil {
		t.Fatalf("GetRule() error = %v", err)
	}
	// Validate rule
	if got.ID != rule.ID || got.Name != rule.Name || got.DeviceID != device.ID {
		t.Fatalf("GetRule() = %#v, want %#v", got, rule)
	}
}

// Edge case: GetRule returns ErrRuleNotFound when a device has no rules.
func TestGetRuleEdgeCase(t *testing.T) {
	s := NewStore(t.Context())
	// New device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Get non-existent rule, validate error
	_, err := s.GetRule(t.Context(), device.ID, uuid.New())
	if !errors.Is(err, ErrRuleNotFound) {
		t.Fatalf("GetRule() error = %v, want ErrRuleNotFound", err)
	}
}

// Invalid case: GetRule rejects an unknown device.
func TestGetRuleInvalidCase(t *testing.T) {
	// Get rule with unknown device and rule, validate error
	_, err := NewStore(t.Context()).GetRule(t.Context(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("GetRule() error = %v, want ErrDeviceNotFound", err)
	}
}

// Regular case: CreateRule stores a rule and sets its timestamps.
func TestCreateRule(t *testing.T) {
	s := NewStore(t.Context())
	// New device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// New rule
	rule := testRule(device.ID, "Overcurrent")

	// Create rule and validate
	created, err := s.CreateRule(t.Context(), rule)
	if err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}
	// Validate values
	if created.ID != rule.ID || created.DeviceID != device.ID || created.Name != rule.Name {
		t.Fatalf("CreateRule() = %#v, want %#v", created, rule)
	}
	// Validate timestamps
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("CreateRule() returned zero timestamps: %#v", created)
	}
}

// Edge case: CreateRule rejects a nil rule UUID.
func TestCreateRuleEdgeCase(t *testing.T) {
	s := NewStore(t.Context())
	// New device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// New rule and set UUID to nil
	rule := testRule(device.ID, "Overcurrent")
	rule.ID = uuid.Nil

	// Try to create Rule with nil UUID, validate error
	if _, err := s.CreateRule(t.Context(), rule); !errors.Is(err, ErrNilRuleUUID) {
		t.Fatalf("CreateRule() error = %v, want ErrNilRuleUUID", err)
	}
}

// Invalid case: CreateRule rejects a duplicate rule name.
func TestCreateRuleInvalidCase(t *testing.T) {
	s := NewStore(t.Context())
	// New device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// New rule "Overcurrent"
	first := testRule(device.ID, "Overcurrent")
	if _, err := s.CreateRule(t.Context(), first); err != nil {
		t.Fatalf("CreateRule(first) error = %v", err)
	}
	// New rule with same name
	duplicate := testRule(device.ID, first.Name)

	// Try to create rule with duplicate Name, validate error
	if _, err := s.CreateRule(t.Context(), duplicate); !errors.Is(err, ErrDuplicateRule) {
		t.Fatalf("CreateRule() error = %v, want ErrDuplicateRule", err)
	}
}

// Regular case: UpdateRule applies an edit and preserves rule identity.
func TestUpdateRule(t *testing.T) {
	s := NewStore(t.Context())
	// Create device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// New rule "Overcurrent"
	rule := testRule(device.ID, "Overcurrent")
	if _, err := s.CreateRule(t.Context(), rule); err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}

	// Update Rule name
	updated, err := s.UpdateRule(t.Context(), device.ID, rule.ID, func(rule model.Rule) (model.Rule, error) {
		rule.Name = "Undervoltage"
		return rule, nil
	})
	// Validate no error after update
	if err != nil {
		t.Fatalf("UpdateRule() error = %v", err)
	}
	// Validate ID persisted, new name
	if updated.ID != rule.ID || updated.DeviceID != device.ID || updated.Name != "Undervoltage" {
		t.Fatalf("UpdateRule() = %#v, want updated rule identity and name", updated)
	}
}

// Edge case: UpdateRule returns the edit error without saving changes.
func TestUpdateRuleEdgeCase(t *testing.T) {
	s := NewStore(t.Context())
	// Create device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// New rule "Overcurrent"
	rule := testRule(device.ID, "Overcurrent")
	if _, err := s.CreateRule(t.Context(), rule); err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}
	editErr := errors.New("edit failed")

	// Define update function value as returning error
	_, err := s.UpdateRule(t.Context(), device.ID, rule.ID, func(rule model.Rule) (model.Rule, error) {
		rule.Name = "Changed"
		return rule, editErr
	})
	// Validate error was returned by UpdateRule
	if !errors.Is(err, editErr) {
		t.Fatalf("UpdateRule() error = %v, want edit error", err)
	}
	// Validate rule was not changed
	stored, err := s.GetRule(t.Context(), device.ID, rule.ID)
	if err != nil || stored.Name != rule.Name {
		t.Fatalf("GetRule() = %#v, %v; want unchanged rule", stored, err)
	}
}

// Invalid case: UpdateRule rejects an edit that changes the rule ID.
func TestUpdateRuleInvalidCase(t *testing.T) {
	s := NewStore(t.Context())
	// Create device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// New rule
	rule := testRule(device.ID, "Overcurrent")
	if _, err := s.CreateRule(t.Context(), rule); err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}

	// Try and update. UpdateRule will reject ID update and not update Rule
	_, err := s.UpdateRule(t.Context(), device.ID, rule.ID, func(rule model.Rule) (model.Rule, error) {
		rule.ID = uuid.New()
		return rule, nil
	})
	// Validate error
	if !errors.Is(err, ErrDuringRuleUpdate) {
		t.Fatalf("UpdateRule() error = %v, want ErrDuringRuleUpdate", err)
	}
}

// Regular case: DeleteRule removes an existing rule.
func TestDeleteRule(t *testing.T) {
	s := NewStore(t.Context())
	// Create device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	// New rule
	rule := testRule(device.ID, "Overcurrent")
	if _, err := s.CreateRule(t.Context(), rule); err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}

	// Delete rule
	if err := s.DeleteRule(t.Context(), device.ID, rule.ID); err != nil {
		t.Fatalf("DeleteRule() error = %v", err)
	}
	// Validate rule was deleted
	if _, err := s.GetRule(t.Context(), device.ID, rule.ID); !errors.Is(err, ErrRuleNotFound) {
		t.Fatalf("GetRule() error = %v, want ErrRuleNotFound", err)
	}
}

// Edge case: DeleteRule rejects deleting a rule that is already absent.
func TestDeleteRuleEdgeCase(t *testing.T) {
	s := NewStore(t.Context())
	// Create device and validate
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	// Try and delete non-existent rule, validate error
	if err := s.DeleteRule(t.Context(), device.ID, uuid.New()); !errors.Is(err, ErrRuleNotFound) {
		t.Fatalf("DeleteRule() error = %v, want ErrRuleNotFound", err)
	}
}

// Invalid case: DeleteRule rejects an unknown device.
func TestDeleteRuleInvalidCase(t *testing.T) {
	// Try and delete unknown rule from unknown device, validate error
	if err := NewStore(t.Context()).DeleteRule(t.Context(), uuid.New(), uuid.New()); !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("DeleteRule() error = %v, want ErrDeviceNotFound", err)
	}
}
