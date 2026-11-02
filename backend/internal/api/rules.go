package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
	"github.com/google/uuid"
)

func (s *Server) getRuleHandler(w http.ResponseWriter, r *http.Request) {

	// Error out if not GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get deviceID from path string
	deviceIDString := r.PathValue("device_id")
	// Parse the UUID, return if it is invalid
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	// Get rule from path string
	ruleIDString := r.PathValue("rule_id")
	// Parse the UUID, return if it is invalid
	ruleID, err := uuid.Parse(ruleIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	rule, err := s.store.GetRule(r.Context(), deviceID, ruleID)

	// Return resource-specific errors without exposing internal details.
	if errors.Is(err, store.ErrDeviceNotFound) {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	} else if errors.Is(err, store.ErrRuleNotFound) {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
		// If any other error, return unexpected error
	} else if err != nil {
		http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		return
	}

	// Set return header to JSON content-type
	w.Header().Set("Content-Type", "application/json")

	// Try and encode rule return to writer. Error out if error occurs, otherwise the rule JSON will be written out.
	if err := json.NewEncoder(w).Encode(rule); err != nil {
		log.Printf("failed to encode rule response: %v", err)
		return
	}
}

func (s *Server) getRulesListHandler(w http.ResponseWriter, r *http.Request) {
	// Error out if not GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get deviceID from path string
	deviceIDString := r.PathValue("device_id")
	// Parse the UUID, return if it is invalid
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	// ListRules through the store
	ruleList, err := s.store.ListRules(r.Context(), deviceID)

	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		return
	}

	// Set return header to JSON content-type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Try and encode rule list return to writer. Error out if error occurs, otherwise the JSON will be written out.
	if err := json.NewEncoder(w).Encode(ruleList); err != nil {
		log.Printf("failed to encode device response: %v", err)
		return
	}
}

func (s *Server) createRuleHandler(w http.ResponseWriter, r *http.Request) {
	// Close body since it is data stream once function is finished, to free system resources and prevent data leak.
	defer r.Body.Close()

	// Validate method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get deviceID from path string
	deviceIDString := r.PathValue("device_id")
	// Parse the UUID, return if it is invalid
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	/* type CreateRuleInput struct {
		Name        string      `json:"name"`
		Enabled     *bool       `json:"enabled"`
		Type        RuleType    `json:"type"`
		Measurement Measurement `json:"measurement"`
		Comparison  *Comparison `json:"comparison"`
		Limit       *float64    `json:"limit"`
		LowerLimit  *float64    `json:"lowerLimit"`
		UpperLimit  *float64    `json:"upperLimit"`
		Deadband    *float64    `json:"deadband"`
		TripState   *State      `json:"tripState"`
	}
	*/
	var ruleInput model.CreateRuleInput

	// Make new decoder with a 1MB limit
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxBodySize))
	decoder.DisallowUnknownFields()

	// Decode into ruleInput, handle error
	if err := decoder.Decode(&ruleInput); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Using &struct{}{} checks that only one JSON value was decoded.
	// If more than one JSON value is present, Decode will return nil, instead of
	// io.EOF.
	if decoder.Decode(&struct{}{}) != io.EOF {
		http.Error(w, "Request body must contain one JSON object", http.StatusBadRequest)
		return
	}

	// Validate input fields
	if strings.TrimSpace(ruleInput.Name) == "" {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if !ruleInput.Type.IsValid() {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if ruleInput.Enabled == nil {
		enabled := true
		ruleInput.Enabled = &enabled
	}

	if !ruleInput.Measurement.IsValid() {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if ruleInput.Comparison != nil && !ruleInput.Comparison.IsValid() {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if ruleInput.Limit != nil {
		if math.IsNaN(*ruleInput.Limit) || math.IsInf(*ruleInput.Limit, 0) {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	}

	if ruleInput.LowerLimit != nil {
		if math.IsNaN(*ruleInput.LowerLimit) || math.IsInf(*ruleInput.LowerLimit, 0) {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	}

	if ruleInput.UpperLimit != nil {
		if math.IsNaN(*ruleInput.UpperLimit) || math.IsInf(*ruleInput.UpperLimit, 0) {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	}

	if ruleInput.Deadband != nil {
		if math.IsNaN(*ruleInput.Deadband) || math.IsInf(*ruleInput.Deadband, 0) {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	}

	if ruleInput.TripState != nil && !ruleInput.TripState.IsValid() {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Define new Rule with ruleInput fields
	rule := model.Rule{
		ID:          uuid.New(),
		DeviceID:    deviceID,
		Name:        strings.TrimSpace(ruleInput.Name),
		Enabled:     *ruleInput.Enabled,
		Type:        ruleInput.Type,
		Measurement: ruleInput.Measurement,
		Comparison:  ruleInput.Comparison,
		Limit:       ruleInput.Limit,
		LowerLimit:  ruleInput.LowerLimit,
		UpperLimit:  ruleInput.UpperLimit,
		Deadband:    ruleInput.Deadband,
		TripState:   ruleInput.TripState,
	}

	// Create Rule in Store
	createdRule, err := s.store.CreateRule(r.Context(), rule)

	// If error during Rule creation, print corresponding error
	if err != nil {
		switch {
		case errors.Is(err, store.ErrDeviceNotFound):
			http.Error(w, "Device not found", http.StatusNotFound)
		case errors.Is(err, store.ErrDuplicateRule):
			http.Error(w, "Duplicate Rule already exists", http.StatusConflict)
		case errors.Is(err, store.ErrNilRuleUUID):
			http.Error(w, "Could not create rule ID", http.StatusInternalServerError)
		default:
			http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return newly created Rule encoded into JSON; will be used by frontend to display
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdRule); err != nil {
		log.Printf("failed to encode created rule response: %v", err)
	}
}

func (s *Server) updateRuleHandler(w http.ResponseWriter, r *http.Request) {
	// Close body since it is data stream once function is finished, to free system resources and prevent data leak.
	defer r.Body.Close()

	// Validate method
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get deviceID from path string
	deviceIDString := r.PathValue("device_id")
	// Parse the UUID, return if it is invalid
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	// Get ruleID from path string
	ruleIDString := r.PathValue("rule_id")
	// Parse the UUID, return if it is invalid
	ruleID, err := uuid.Parse(ruleIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	/* type UpdateRuleInput struct {
		Name        *string      `json:"name"`
		Enabled     *bool        `json:"enabled"`
		Measurement *Measurement `json:"measurement"`
		Comparison  *Comparison  `json:"comparison"`
		Limit       *float64     `json:"limit"`
		LowerLimit  *float64     `json:"lowerLimit"`
		UpperLimit  *float64     `json:"upperLimit"`
		Deadband    *float64     `json:"deadband"`
		TripState   *State       `json:"tripState"`
	}
	*/
	var ruleInput model.UpdateRuleInput
	// Make new decoder with a 1MB limit
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxBodySize))
	decoder.DisallowUnknownFields()

	// Decode into ruleInput, handle error
	if err := decoder.Decode(&ruleInput); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Using &struct{}{} checks that only one JSON value was decoded.
	// If more than one JSON value is present, Decode will return nil, instead of
	// io.EOF.
	if decoder.Decode(&struct{}{}) != io.EOF {
		http.Error(w, "Request body must contain one JSON object", http.StatusBadRequest)
		return
	}

	// Validate input fields
	if ruleInput.Name != nil && strings.TrimSpace(*ruleInput.Name) == "" {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if ruleInput.Measurement != nil && !ruleInput.Measurement.IsValid() {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if ruleInput.Comparison != nil && !ruleInput.Comparison.IsValid() {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if ruleInput.Limit != nil {
		if math.IsNaN(*ruleInput.Limit) || math.IsInf(*ruleInput.Limit, 0) {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	}

	if ruleInput.LowerLimit != nil {
		if math.IsNaN(*ruleInput.LowerLimit) || math.IsInf(*ruleInput.LowerLimit, 0) {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	}

	if ruleInput.UpperLimit != nil {
		if math.IsNaN(*ruleInput.UpperLimit) || math.IsInf(*ruleInput.UpperLimit, 0) {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	}

	if ruleInput.Deadband != nil {
		if math.IsNaN(*ruleInput.Deadband) || math.IsInf(*ruleInput.Deadband, 0) {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	}

	if ruleInput.TripState != nil && !ruleInput.TripState.IsValid() {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	updatedRule, err := s.store.UpdateRule(
		r.Context(),
		deviceID,
		ruleID,
		func(rule model.Rule) (model.Rule, error) {
			ruleInput.UpdateFields(&rule) // Call helper
			return rule, nil
		})

	if err != nil {
		switch {
		case errors.Is(err, store.ErrDeviceNotFound):
			http.Error(w, "Device not found", http.StatusNotFound)
		case errors.Is(err, store.ErrRuleNotFound):
			http.Error(w, "Rule not found", http.StatusNotFound)
		case errors.Is(err, store.ErrDuplicateRule):
			http.Error(w, "Rule already exists", http.StatusConflict)
		case errors.Is(err, store.ErrDuringRuleUpdate):
			http.Error(w, "Invalid rule update", http.StatusBadRequest)
		default:
			http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return the updated device encoded into JSON.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updatedRule); err != nil {
		log.Printf("failed to encode updated device response: %v", err)
	}
}

func (s *Server) deleteRuleHandler(w http.ResponseWriter, r *http.Request) {
	// Validate method
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get deviceID from path string
	deviceIDString := r.PathValue("device_id")
	// Parse the UUID, return if it is invalid
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	// Get ruleID from path string
	ruleIDString := r.PathValue("rule_id")
	// Parse the UUID, return if it is invalid
	ruleID, err := uuid.Parse(ruleIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	err = s.store.DeleteRule(r.Context(), deviceID, ruleID)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrDeviceNotFound):
			http.Error(w, "Device not found", http.StatusNotFound)
		case errors.Is(err, store.ErrRuleNotFound):
			http.Error(w, "Rule not found", http.StatusNotFound)
		case errors.Is(err, store.ErrUnacknowledgedAlarm):
			http.Error(w, "Unacknowledged alarm for rule", http.StatusConflict)
		default:
			http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return 204
	w.WriteHeader(http.StatusNoContent)
}
