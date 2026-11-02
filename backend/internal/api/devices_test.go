package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
	"github.com/google/uuid"
)

// Unit tests follow AAA pattern

// Helper functions

// Creates API Store with one Device
func testAPIStore(t *testing.T) (*store.Store, model.Device) {
	t.Helper()

	appStore := store.NewStore()
	device := model.Device{
		ID:         uuid.New(),
		Name:       "Relay A",
		IP:         "192.0.2.1",
		DeviceType: model.DeviceTypeRelay,
	}
	if _, err := appStore.CreateDevice(t.Context(), device); err != nil {
		t.Fatalf("CreateDevice() error = %v", err)
	}

	return appStore, device
}

// Creates a Device GET request
func getDeviceRequest(method string, deviceID string) *http.Request {
	request := httptest.NewRequest(method, "/api/devices/"+deviceID, nil)
	request.SetPathValue("device_id", deviceID)
	return request
}

// Creates a Device List GET request
func getDeviceListRequest(method string) *http.Request {
	request := httptest.NewRequest(method, "/api/devices", nil)
	return request
}

// Creates a JSON request based on inputted body (for CREATE)
func jsonRequest(method, path string, body any) *http.Request {
	payload, err := json.Marshal(body) // Use built-in Marshal to convert Go object to JSON
	if err != nil {
		panic(err)
	}
	// Return new request with io.Reader
	return httptest.NewRequest(method, path, bytes.NewReader(payload))
}

// Creates a JSON request with device ID (for PATCH)
func jsonDeviceRequest(method, deviceID string, body any) *http.Request {
	request := jsonRequest(method, "/api/devices/"+deviceID, body)
	request.SetPathValue("device_id", deviceID)
	return request
}

// Regular case: getDeviceHandler returns an existing device as JSON.
func TestGetDeviceHandler(t *testing.T) {
	// Make new Store, and get device we want
	appStore, want := testAPIStore(t)
	// Make new server
	server := NewServer(appStore)
	// Make new response recorder
	response := httptest.NewRecorder()
	// Make new device GET request
	request := getDeviceRequest(http.MethodGet, want.ID.String())

	// Call getDeviceHandler
	server.getDeviceHandler(response, request)

	// Validate response code
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	// Validate Content-Type header
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	// Decode body and get device
	var got model.Device
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	// Validate values
	if got.ID != want.ID || got.Name != want.Name || got.IP != want.IP {
		t.Fatalf("response device = %#v, want %#v", got, want)
	}
}

// Edge case: getDeviceHandler returns not found for a device that is not stored.
func TestGetDeviceHandlerNotFound(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := getDeviceRequest(http.MethodGet, uuid.New().String())

	server.getDeviceHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if !strings.Contains(response.Body.String(), "Device not found") {
		t.Fatalf("body = %q, want device-not-found message", response.Body.String())
	}
}

// Invalid case: getDeviceHandler rejects malformed device IDs.
func TestGetDeviceHandlerMalformedID(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := getDeviceRequest(http.MethodGet, "not-a-uuid")

	server.getDeviceHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: getDeviceHandler rejects methods other than GET.
func TestGetDeviceHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := getDeviceRequest(http.MethodPost, uuid.New().String())

	server.getDeviceHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// Regular case: listDevicesHandler correctly lists Devices
func TestListDeviceHandler(t *testing.T) {
	// Make new Store, and get device we want
	appStore, want := testAPIStore(t)

	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := getDeviceListRequest(http.MethodGet)

	server.listDevicesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	// Validate Content-Type header
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	// Decode body and get device list
	var got []model.Device
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("response contains %d devices, want 1", len(got))
	}
	// Validate values
	for _, device := range got {
		if device.ID != want.ID || device.Name != want.Name || device.IP != want.IP {
			t.Fatalf("response device = %#v, want %#v", got, want)
		}
	}
}

// Invalid case: listDevicesHandler rejects methods other than GET.
func TestListDeviceHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	// Invalid method
	request := getDeviceListRequest(http.MethodPost)

	server.listDevicesHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// Regular case: createDeviceHandler creates and returns a device.
func TestCreateDeviceHandler(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonRequest(http.MethodPost, "/api/devices", model.CreateDeviceInput{
		Name:       "Relay B",
		IP:         "192.0.2.2",
		DeviceType: model.DeviceTypeRelay,
	})

	server.createDeviceHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", response.Header().Get("Content-Type"))
	}

	var got model.Device
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if got.ID == uuid.Nil || got.Name != "Relay B" || got.IP != "192.0.2.2" {
		t.Fatalf("response device = %#v, want created Relay B", got)
	}
}

// Edge case: createDeviceHandler rejects a duplicate device.
func TestCreateDeviceHandlerDuplicate(t *testing.T) {
	appStore, existing := testAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := jsonRequest(http.MethodPost, "/api/devices", model.CreateDeviceInput{
		Name:       existing.Name,
		IP:         "192.0.2.2",
		DeviceType: model.DeviceTypeRelay,
	})

	server.createDeviceHandler(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
}

// Invalid case: createDeviceHandler rejects malformed input.
func TestCreateDeviceHandlerInvalidInput(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/devices", strings.NewReader(`{"name":""}`))

	server.createDeviceHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: createDeviceHandler rejects methods other than POST.
func TestCreateDeviceHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/devices", nil)

	server.createDeviceHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// Regular case: updateDeviceHandler updates and returns a device.
func TestUpdateDeviceHandler(t *testing.T) {
	appStore, existing := testAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := jsonDeviceRequest(http.MethodPatch, existing.ID.String(), model.UpdateDeviceInput{
		Name: stringPointer("Updated Relay"),
		IP:   stringPointer("192.0.2.20"),
	})

	server.updateDeviceHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got model.Device
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if got.ID != existing.ID || got.Name != "Updated Relay" || got.IP != "192.0.2.20" {
		t.Fatalf("response device = %#v, want updated device", got)
	}
}

// Edge case: updateDeviceHandler leaves omitted fields unchanged.
func TestUpdateDeviceHandlerPartialUpdate(t *testing.T) {
	appStore, existing := testAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := jsonDeviceRequest(http.MethodPatch, existing.ID.String(), model.UpdateDeviceInput{
		Name: stringPointer("Renamed Relay"),
	})

	server.updateDeviceHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	got, err := appStore.GetDevice(t.Context(), existing.ID)
	if err != nil {
		t.Fatalf("GetDevice() error = %v", err)
	}
	if got.Name != "Renamed Relay" || got.IP != existing.IP {
		t.Fatalf("stored device = %#v, want name changed and IP unchanged", got)
	}
}

// Invalid case: updateDeviceHandler returns not found for an unknown device.
func TestUpdateDeviceHandlerNotFound(t *testing.T) {
	server := NewServer(store.NewStore())
	deviceID := uuid.New().String()
	response := httptest.NewRecorder()
	request := jsonDeviceRequest(http.MethodPatch, deviceID, model.UpdateDeviceInput{
		Name: stringPointer("Updated Relay"),
	})

	server.updateDeviceHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

// Invalid case: updateDeviceHandler rejects methods other than PATCH.
func TestUpdateDeviceHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonDeviceRequest(http.MethodPost, uuid.New().String(), model.UpdateDeviceInput{})

	server.updateDeviceHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// Regular case: deleteDeviceHandler deletes an existing device.
func TestDeleteDeviceHandler(t *testing.T) {
	appStore, existing := testAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := getDeviceRequest(http.MethodDelete, existing.ID.String())

	server.deleteDeviceHandler(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if _, err := appStore.GetDevice(t.Context(), existing.ID); err != store.ErrDeviceNotFound {
		t.Fatalf("GetDevice() error = %v, want ErrDeviceNotFound", err)
	}
}

// Edge case: deleteDeviceHandler returns conflict when a device is already deleted.
func TestDeleteDeviceHandlerAlreadyDeleted(t *testing.T) {
	appStore, existing := testAPIStore(t)
	if err := appStore.DeleteDevice(t.Context(), existing.ID); err != nil {
		t.Fatalf("DeleteDevice() error = %v", err)
	}
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := getDeviceRequest(http.MethodDelete, existing.ID.String())

	server.deleteDeviceHandler(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
}

// Invalid case: deleteDeviceHandler rejects malformed device IDs.
func TestDeleteDeviceHandlerMalformedID(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := getDeviceRequest(http.MethodDelete, "not-a-uuid")

	server.deleteDeviceHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: deleteDeviceHandler rejects methods other than DELETE.
func TestDeleteDeviceHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := getDeviceRequest(http.MethodGet, uuid.New().String())

	server.deleteDeviceHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

func stringPointer(value string) *string {
	return &value
}
