package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
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
	s := store.NewStore()
	if s == nil {
		t.Fatal("NewStore returned nil")
	}
	if devices, err := s.ListDevices(); devices == nil || len(devices) != 0 || err != nil {
		t.Fatalf("ListDevices() = %#v, want a non-nil empty slice", devices)
	}
}

// Edge case: NewStore calls create independent stores with independent maps.
func TestNewStoreEdgeCase(t *testing.T) {
	first := store.NewStore()
	second := store.NewStore()
	device := testDevice("Relay A", "192.0.2.1")

	if _, err := first.CreateDevice(device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	if devices, err := second.ListDevices(); len(devices) != 0 || err != nil {
		t.Fatalf("second store contains devices = %#v, want empty", devices)
	}
}

// Regular case: ListDevices returns active devices in name order.
func TestListDevices(t *testing.T) {
	s := store.NewStore()
	for _, device := range []model.Device{
		testDevice("Relay C", "192.0.2.3"),
		testDevice("Relay A", "192.0.2.1"),
		testDevice("Relay B", "192.0.2.2"),
	} {
		if _, err := s.CreateDevice(device); err != nil {
			t.Fatalf("CreateDevice() error = %v", err)
		}
	}

	devices, _ := s.ListDevices()
	if len(devices) != 3 {
		t.Fatalf("ListDevices() returned %d devices, want 3", len(devices))
	}
	for i, want := range []string{"Relay A", "Relay B", "Relay C"} {
		if devices[i].Name != want {
			t.Errorf("ListDevices()[%d].Name = %q, want %q", i, devices[i].Name, want)
		}
	}
}

// Edge case: ListDevices returns a non-nil empty slice for an empty store.
func TestListDevicesEdgeCase(t *testing.T) {
	devices, err := store.NewStore().ListDevices()
	if devices == nil || len(devices) != 0 || err != nil {
		t.Fatalf("ListDevices() = %#v, want a non-nil empty slice", devices)
	}
}

// Invalid case: ListDevices excludes soft-deleted devices.
func TestListDevicesInvalidCase(t *testing.T) {
	s := store.NewStore()
	deletedAt := time.Now()
	device := testDevice("Deleted Relay", "192.0.2.10")
	device.DeletedAt = &deletedAt
	saved, err := s.CreateDevice(device)
	if err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	if devices, _ := s.ListDevices(); len(devices) != 0 {
		t.Fatalf("ListDevices() = %#v, want no active devices", devices)
	}
	if saved.ID != device.ID {
		t.Fatalf("CreateDevice() returned ID %v, want %v", saved.ID, device.ID)
	}
}

// Regular case: GetDevice returns an existing active device.
func TestGetDevice(t *testing.T) {
	s := store.NewStore()
	device := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	got, err := s.GetDevice(device.ID)
	if err != nil {
		t.Fatalf("GetDevice() error = %v", err)
	}
	if got.ID != device.ID || got.Name != device.Name {
		t.Fatalf("GetDevice() = %#v, want %#v", got, device)
	}
}

// Edge case: GetDevice can retrieve a device whose ID is the zero UUID.
func TestGetDeviceEdgeCase(t *testing.T) {
	s := store.NewStore()
	device := testDevice("Zero ID Relay", "192.0.2.2")
	device.ID = uuid.Nil
	if _, err := s.CreateDevice(device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	got, err := s.GetDevice(uuid.Nil)
	if err != nil {
		t.Fatalf("GetDevice() error = %v", err)
	}
	if got.ID != uuid.Nil {
		t.Fatalf("GetDevice().ID = %v, want zero UUID", got.ID)
	}
}

// Invalid case: GetDevice returns ErrDeviceNotFound for an unknown ID.
func TestGetDeviceInvalidCase(t *testing.T) {
	s := store.NewStore()
	_, err := s.GetDevice(uuid.New())
	if !errors.Is(err, store.ErrDeviceNotFound) {
		t.Fatalf("GetDevice() error = %v, want ErrDeviceNotFound", err)
	}
}

// Regular case: CreateDevice adds a device and returns the created value.
func TestCreateDevice(t *testing.T) {
	s := store.NewStore()
	device := testDevice("Relay A", "192.0.2.1")

	got, err := s.CreateDevice(device)
	if err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}
	if got != device {
		t.Fatalf("CreateDevice() = %#v, want %#v", got, device)
	}
	if stored, err := s.GetDevice(device.ID); err != nil || stored != device {
		t.Fatalf("GetDevice() = %#v, %v; want %#v, nil", stored, err, device)
	}
}

// Edge case: CreateDevice allows a new device to reuse values from a soft-deleted device.
func TestCreateDeviceEdgeCase(t *testing.T) {
	s := store.NewStore()
	deletedAt := time.Now()
	deleted := testDevice("Relay A", "192.0.2.1")
	deleted.DeletedAt = &deletedAt
	if _, err := s.CreateDevice(deleted); err != nil {
		t.Fatalf("CreateDevice(deleted) error = %v", err)
	}

	active := testDevice("Relay A", "192.0.2.1")
	if _, err := s.CreateDevice(active); err != nil {
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
			s := store.NewStore()
			existing := testDevice("Relay A", "192.0.2.1")
			if _, err := s.CreateDevice(existing); err != nil {
				t.Fatalf("CreateDevice(existing) error = %v", err)
			}

			_, err := s.CreateDevice(test.build(testDevice("Relay A", "192.0.2.1")))
			if !errors.Is(err, store.ErrDuplicateDevice) {
				t.Fatalf("CreateDevice() error = %v, want ErrDuplicateDevice", err)
			}
			if devices, _ := s.ListDevices(); len(devices) != 1 {
				t.Fatalf("ListDevices() returned %d devices, want 1", len(devices))
			}
		})
	}
}
