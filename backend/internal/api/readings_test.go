package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
	"github.com/google/uuid"
)

func testReadingAPIStore(t *testing.T) (*store.Store, model.Device) {
	t.Helper()

	appStore, device := testAPIStore(t)
	return appStore, device
}

func testReadingInput(deviceID uuid.UUID, timestamp time.Time) model.CreateReadingInput {
	currentA := 10.0
	currentB := 11.0
	currentC := 12.0
	voltageA := 230.0
	voltageB := 231.0
	voltageC := 232.0
	frequency := 60.0
	realPower := 100.0
	reactivePower := 20.0
	breakerStatus := model.StateClosed

	return model.CreateReadingInput{
		DeviceID:      deviceID,
		Timestamp:     timestamp,
		CurrentA:      &currentA,
		CurrentB:      &currentB,
		CurrentC:      &currentC,
		VoltageA:      &voltageA,
		VoltageB:      &voltageB,
		VoltageC:      &voltageC,
		Frequency:     &frequency,
		RealPower:     &realPower,
		ReactivePower: &reactivePower,
		BreakerStatus: &breakerStatus,
	}
}

func jsonReadingCreateRequest(method string, body any) *http.Request {
	return jsonRequest(method, "/api/readings", body)
}

func getReadingsRequest(method, deviceID string, from, to time.Time) *http.Request {
	path := "/api/devices/" + deviceID + "/readings?from=" +
		from.Format(time.RFC3339) + "&to=" + to.Format(time.RFC3339)
	request := httptest.NewRequest(method, path, nil)
	request.SetPathValue("device_id", deviceID)
	return request
}

func testReading(deviceID uuid.UUID, timestamp time.Time, currentA float64) model.Reading {
	return model.Reading{
		DeviceID:      deviceID,
		Timestamp:     timestamp,
		CurrentA:      currentA,
		CurrentB:      11,
		CurrentC:      12,
		VoltageA:      230,
		VoltageB:      231,
		VoltageC:      232,
		Frequency:     60,
		RealPower:     100,
		ReactivePower: 20,
		BreakerStatus: model.StateClosed,
	}
}

// Regular case: createReadingHandler creates a reading.
func TestCreateReadingHandler(t *testing.T) {
	appStore, device := testReadingAPIStore(t)
	server := NewServer(appStore)
	timestamp := time.Date(2026, 11, 2, 20, 0, 0, 0, time.UTC)
	response := httptest.NewRecorder()
	request := jsonReadingCreateRequest(
		http.MethodPost,
		testReadingInput(device.ID, timestamp),
	)

	server.createReadingHandler(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

// Edge case: createReadingHandler rejects a duplicate timestamp for a device.
func TestCreateReadingHandlerDuplicate(t *testing.T) {
	appStore, device := testReadingAPIStore(t)
	timestamp := time.Date(2026, 11, 2, 20, 0, 0, 0, time.UTC)
	if _, err := appStore.AddReading(t.Context(), testReading(device.ID, timestamp, 10)); err != nil {
		t.Fatalf("AddReading() error = %v", err)
	}

	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := jsonReadingCreateRequest(
		http.MethodPost,
		testReadingInput(device.ID, timestamp),
	)

	server.createReadingHandler(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
}

// Invalid case: createReadingHandler rejects a request missing a required field.
func TestCreateReadingHandlerMissingField(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonReadingCreateRequest(
		http.MethodPost,
		map[string]any{
			"deviceId":      uuid.New().String(),
			"timestamp":     "2026-11-02T20:00:00Z",
			"currentA":      10,
			"currentB":      11,
			"currentC":      12,
			"voltageA":      230,
			"voltageB":      231,
			"voltageC":      232,
			"frequency":     60,
			"realPower":     100,
			"breakerStatus": "closed",
		},
	)

	server.createReadingHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: createReadingHandler rejects an invalid breaker state.
func TestCreateReadingHandlerInvalidState(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonReadingCreateRequest(
		http.MethodPost,
		map[string]any{
			"deviceId":      uuid.New().String(),
			"timestamp":     "2026-11-02T20:00:00Z",
			"currentA":      10,
			"currentB":      11,
			"currentC":      12,
			"voltageA":      230,
			"voltageB":      231,
			"voltageC":      232,
			"frequency":     60,
			"realPower":     100,
			"reactivePower": 20,
			"breakerStatus": "invalid",
		},
	)

	server.createReadingHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: createReadingHandler rejects an unknown device.
func TestCreateReadingHandlerUnknownDevice(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonReadingCreateRequest(
		http.MethodPost,
		testReadingInput(uuid.New(), time.Now().UTC()),
	)

	server.createReadingHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if !strings.Contains(response.Body.String(), "Device not found") {
		t.Fatalf("body = %q, want device-not-found message", response.Body.String())
	}
}

// Invalid case: createReadingHandler rejects malformed device IDs.
func TestCreateReadingHandlerMalformedID(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonReadingCreateRequest(
		http.MethodPost,
		map[string]any{"deviceId": "not-a-uuid", "timestamp": "2026-11-02T20:00:00Z"},
	)

	server.createReadingHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: createReadingHandler rejects methods other than POST.
func TestCreateReadingHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonReadingCreateRequest(
		http.MethodGet,
		testReadingInput(uuid.New(), time.Now().UTC()),
	)

	server.createReadingHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// Regular case: getReadingsListHandler returns readings in the requested range.
func TestGetReadingsListHandler(t *testing.T) {
	appStore, device := testReadingAPIStore(t)
	from := time.Date(2026, 11, 2, 20, 0, 0, 0, time.UTC)
	inRange := testReading(device.ID, from.Add(time.Minute), 20)
	outOfRange := testReading(device.ID, from.Add(2*time.Hour), 30)
	for _, reading := range []model.Reading{inRange, outOfRange} {
		if _, err := appStore.AddReading(t.Context(), reading); err != nil {
			t.Fatalf("AddReading() error = %v", err)
		}
	}

	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := getReadingsRequest(http.MethodGet, device.ID.String(), from, from.Add(time.Hour))

	server.getReadingsListHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got []model.Reading
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(got) != 1 || got[0] != inRange {
		t.Fatalf("response readings = %#v, want [%#v]", got, inRange)
	}
}

// Edge case: getReadingsListHandler returns an empty list for a valid range.
func TestGetReadingsListHandlerEmpty(t *testing.T) {
	appStore, device := testReadingAPIStore(t)
	server := NewServer(appStore)
	from := time.Date(2026, 11, 2, 20, 0, 0, 0, time.UTC)
	response := httptest.NewRecorder()
	request := getReadingsRequest(http.MethodGet, device.ID.String(), from, from.Add(time.Hour))

	server.getReadingsListHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got []model.Reading
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("response readings = %#v, want non-nil empty list", got)
	}
}

// Invalid case: getReadingsListHandler rejects a missing range parameter.
func TestGetReadingsListHandlerMissingRange(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/devices/"+uuid.New().String()+"/readings?from=2026-11-02T20:00:00Z",
		nil,
	)
	request.SetPathValue("device_id", uuid.New().String())

	server.getReadingsListHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: getReadingsListHandler rejects reversed ranges.
func TestGetReadingsListHandlerReversedRange(t *testing.T) {
	server := NewServer(store.NewStore())
	from := time.Date(2026, 11, 2, 21, 0, 0, 0, time.UTC)
	to := from.Add(-time.Hour)
	response := httptest.NewRecorder()
	request := getReadingsRequest(http.MethodGet, uuid.New().String(), from, to)

	server.getReadingsListHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: getReadingsListHandler returns not found for an unknown device.
func TestGetReadingsListHandlerUnknownDevice(t *testing.T) {
	server := NewServer(store.NewStore())
	from := time.Date(2026, 11, 2, 20, 0, 0, 0, time.UTC)
	response := httptest.NewRecorder()
	request := getReadingsRequest(http.MethodGet, uuid.New().String(), from, from.Add(time.Hour))

	server.getReadingsListHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

// Invalid case: getReadingsListHandler rejects malformed device IDs.
func TestGetReadingsListHandlerMalformedID(t *testing.T) {
	server := NewServer(store.NewStore())
	from := time.Date(2026, 11, 2, 20, 0, 0, 0, time.UTC)
	response := httptest.NewRecorder()
	request := getReadingsRequest(http.MethodGet, "not-a-uuid", from, from.Add(time.Hour))

	server.getReadingsListHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: getReadingsListHandler rejects methods other than GET.
func TestGetReadingsListHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	from := time.Date(2026, 11, 2, 20, 0, 0, 0, time.UTC)
	response := httptest.NewRecorder()
	request := getReadingsRequest(http.MethodPost, uuid.New().String(), from, from.Add(time.Hour))

	server.getReadingsListHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// Invalid case: createReadingHandler rejects a reading without a device ID.
func TestCreateReadingHandlerMissingDeviceID(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonReadingCreateRequest(
		http.MethodPost,
		testReadingInput(uuid.Nil, time.Now().UTC()),
	)

	server.createReadingHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
