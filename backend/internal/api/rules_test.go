package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
	"github.com/google/uuid"
)

// Creates an API store with one device and one rule.
func testRuleAPIStore(t *testing.T) (*store.Store, model.Device, model.Rule) {
	t.Helper()

	// Call from device_test, and add Rule on top
	appStore, device := testAPIStore(t)
	limit := 240.0
	comparison := model.ComparisonGT
	rule := model.Rule{
		ID:          uuid.New(),
		DeviceID:    device.ID,
		Name:        "Overvoltage",
		Enabled:     true,
		Type:        model.RuleTypeThreshold,
		Measurement: model.MeasurementVoltageA,
		Comparison:  &comparison,
		Limit:       &limit,
	}
	created, err := appStore.CreateRule(t.Context(), rule)
	if err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}

	return appStore, device, created
}

// Create GET request for Rule
func getRuleRequest(method, deviceID, ruleID string) *http.Request {
	request := httptest.NewRequest(method, "/api/devices/"+deviceID+"/rules/"+ruleID, nil)
	request.SetPathValue("device_id", deviceID)
	request.SetPathValue("rule_id", ruleID)
	return request
}

// Create GET request for list of Rules
func getRuleListRequest(method, deviceID string) *http.Request {
	request := httptest.NewRequest(method, "/api/devices/"+deviceID+"/rules", nil)
	request.SetPathValue("device_id", deviceID)
	return request
}

// Create JSON Rule request (for PATCH/Update)
func jsonRuleRequest(method, deviceID, ruleID string, body any) *http.Request {
	request := jsonRequest(method, "/api/devices/"+deviceID+"/rules/"+ruleID, body)
	request.SetPathValue("device_id", deviceID)
	request.SetPathValue("rule_id", ruleID)
	return request
}

// Create JSON new Rule request (for POST/Create)
func jsonRuleCreateRequest(method, deviceID string, body any) *http.Request {
	request := jsonRequest(method, "/api/devices/"+deviceID+"/rules", body)
	request.SetPathValue("device_id", deviceID)
	return request
}

// Regular case: getRuleHandler returns an existing rule as JSON.
func TestGetRuleHandler(t *testing.T) {
	appStore, device, want := testRuleAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := getRuleRequest(http.MethodGet, device.ID.String(), want.ID.String())

	server.getRuleHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", response.Header().Get("Content-Type"))
	}

	var got model.Rule
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if got.ID != want.ID || got.DeviceID != device.ID || got.Name != want.Name {
		t.Fatalf("response rule = %#v, want %#v", got, want)
	}
}

// Edge case: getRuleHandler returns not found for an unknown rule.
func TestGetRuleHandlerNotFound(t *testing.T) {
	appStore, device, _ := testRuleAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := getRuleRequest(http.MethodGet, device.ID.String(), uuid.New().String())

	server.getRuleHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if !strings.Contains(response.Body.String(), "Rule not found") {
		t.Fatalf("body = %q, want rule-not-found message", response.Body.String())
	}
}

// Invalid case: getRuleHandler rejects malformed IDs.
func TestGetRuleHandlerMalformedID(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := getRuleRequest(http.MethodGet, "not-a-uuid", uuid.New().String())

	server.getRuleHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: getRuleHandler rejects methods other than GET.
func TestGetRuleHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := getRuleRequest(http.MethodPost, uuid.New().String(), uuid.New().String())

	server.getRuleHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// Regular case: getRulesListHandler returns the rules for a device.
func TestGetRulesListHandler(t *testing.T) {
	appStore, device, want := testRuleAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := getRuleListRequest(http.MethodGet, device.ID.String())

	server.getRulesListHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got []model.Rule
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(got) != 1 || got[0].ID != want.ID {
		t.Fatalf("response rules = %#v, want one rule %#v", got, want)
	}
}

// Invalid case: getRulesListHandler rejects methods other than GET.
func TestGetRulesListHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := getRuleListRequest(http.MethodPost, uuid.New().String())

	server.getRulesListHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// Regular case: createRuleHandler creates and returns a rule.
func TestCreateRuleHandler(t *testing.T) {
	appStore, device := testAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	Comparison := model.ComparisonLT
	Limit := 59.0
	request := jsonRuleCreateRequest(http.MethodPost, device.ID.String(), model.CreateRuleInput{
		Name:        "Underfrequency",
		Type:        model.RuleTypeThreshold,
		Measurement: model.MeasurementFrequency,
		Comparison:  &Comparison,
		Limit:       &Limit,
	})

	server.createRuleHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	var got model.Rule
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if got.ID == uuid.Nil || got.DeviceID != device.ID || got.Name != "Underfrequency" || !got.Enabled {
		t.Fatalf("response rule = %#v, want created enabled rule", got)
	}
}

// Edge case: createRuleHandler rejects a duplicate rule.
func TestCreateRuleHandlerDuplicate(t *testing.T) {
	appStore, device, existing := testRuleAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := jsonRuleCreateRequest(http.MethodPost, device.ID.String(), model.CreateRuleInput{
		Name:        existing.Name,
		Type:        existing.Type,
		Measurement: existing.Measurement,
		Comparison:  existing.Comparison,
		Limit:       existing.Limit,
	})

	server.createRuleHandler(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
}

// Invalid case: createRuleHandler rejects malformed input.
func TestCreateRuleHandlerInvalidInput(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/devices/"+uuid.New().String()+"/rules", strings.NewReader(`{"name":"","type":"invalid"}`))

	server.createRuleHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: createRuleHandler rejects methods other than POST.
func TestCreateRuleHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonRuleCreateRequest(http.MethodGet, uuid.New().String(), model.CreateRuleInput{})

	server.createRuleHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// Regular case: updateRuleHandler updates all supplied fields.
func TestUpdateRuleHandler(t *testing.T) {
	appStore, device, existing := testRuleAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()

	Name := "Updated Rule"
	Enabled := false
	Measurement := model.MeasurementCurrentA
	Comparison := model.ComparisonLT
	Limit := 10.0
	request := jsonRuleRequest(http.MethodPatch, device.ID.String(), existing.ID.String(), model.UpdateRuleInput{
		Name:        &Name,
		Enabled:     &Enabled,
		Measurement: &Measurement,
		Comparison:  &Comparison,
		Limit:       &Limit,
	})

	server.updateRuleHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got model.Rule
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if got.Name != "Updated Rule" || got.Enabled || got.Measurement != model.MeasurementCurrentA ||
		got.Comparison == nil || *got.Comparison != model.ComparisonLT || got.Limit == nil || *got.Limit != 10 {
		t.Fatalf("response rule = %#v, want updated fields", got)
	}
}

// Edge case: updateRuleHandler leaves omitted fields unchanged.
func TestUpdateRuleHandlerPartialUpdate(t *testing.T) {
	appStore, device, existing := testRuleAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	Name := "Renamed Rule"
	request := jsonRuleRequest(http.MethodPatch, device.ID.String(), existing.ID.String(), model.UpdateRuleInput{
		Name: &Name,
	})

	server.updateRuleHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	got, err := appStore.GetRule(t.Context(), device.ID, existing.ID)
	if err != nil {
		t.Fatalf("GetRule() error = %v", err)
	}
	if got.Name != "Renamed Rule" || got.Measurement != existing.Measurement || got.Limit == nil || *got.Limit != *existing.Limit {
		t.Fatalf("stored rule = %#v, want name changed and other fields unchanged", got)
	}
}

// Invalid case: updateRuleHandler returns not found for an unknown rule.
func TestUpdateRuleHandlerNotFound(t *testing.T) {
	appStore, device, _ := testRuleAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	Name := "Updated Rule"
	request := jsonRuleRequest(http.MethodPatch, device.ID.String(), uuid.New().String(), model.UpdateRuleInput{
		Name: &Name,
	})

	server.updateRuleHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

// Invalid case: updateRuleHandler rejects invalid rule values.
func TestUpdateRuleHandlerInvalidInput(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonRuleRequest(http.MethodPatch, uuid.New().String(), uuid.New().String(), map[string]any{
		"measurement": "invalid",
	})

	server.updateRuleHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: updateRuleHandler rejects methods other than PATCH.
func TestUpdateRuleHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := jsonRuleRequest(http.MethodPost, uuid.New().String(), uuid.New().String(), model.UpdateRuleInput{})

	server.updateRuleHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// Regular case: deleteRuleHandler deletes an existing rule.
func TestDeleteRuleHandler(t *testing.T) {
	appStore, device, existing := testRuleAPIStore(t)
	server := NewServer(appStore)
	response := httptest.NewRecorder()
	request := getRuleRequest(http.MethodDelete, device.ID.String(), existing.ID.String())

	server.deleteRuleHandler(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if _, err := appStore.GetRule(t.Context(), device.ID, existing.ID); err != store.ErrRuleNotFound {
		t.Fatalf("GetRule() error = %v, want ErrRuleNotFound", err)
	}
}

// Invalid case: deleteRuleHandler rejects malformed rule IDs.
func TestDeleteRuleHandlerMalformedID(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := getRuleRequest(http.MethodDelete, uuid.New().String(), "not-a-uuid")

	server.deleteRuleHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// Invalid case: deleteRuleHandler rejects methods other than DELETE.
func TestDeleteRuleHandlerWrongMethod(t *testing.T) {
	server := NewServer(store.NewStore())
	response := httptest.NewRecorder()
	request := getRuleRequest(http.MethodGet, uuid.New().String(), uuid.New().String())

	server.deleteRuleHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}
