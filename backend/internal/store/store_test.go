package store

import (
	"errors"
	"testing"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

func testDevice(name, ip string) model.Device {
	return model.Device{
		ID:   uuid.New(),
		Name: name,
		IP:   ip,
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
