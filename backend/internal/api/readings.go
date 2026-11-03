package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
	"github.com/google/uuid"
)

func (s *Server) createReadingHandler(w http.ResponseWriter, r *http.Request) {

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

	var readingInput model.CreateReadingInput

	// Make new decoder with a 1MB limit
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxBodySize))
	decoder.DisallowUnknownFields()

	// Decode into readingInput, handle error
	if err := decoder.Decode(&readingInput); err != nil {
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

	if readingInput.Timestamp.IsZero() ||
		readingInput.CurrentA == nil ||
		readingInput.CurrentB == nil ||
		readingInput.CurrentC == nil ||
		readingInput.VoltageA == nil ||
		readingInput.VoltageB == nil ||
		readingInput.VoltageC == nil ||
		readingInput.Frequency == nil ||
		readingInput.RealPower == nil ||
		readingInput.ReactivePower == nil ||
		readingInput.BreakerStatus == nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	values := []*float64{
		readingInput.CurrentA,
		readingInput.CurrentB,
		readingInput.CurrentC,
		readingInput.VoltageA,
		readingInput.VoltageB,
		readingInput.VoltageC,
		readingInput.Frequency,
		readingInput.RealPower,
		readingInput.ReactivePower,
	}
	for _, value := range values {
		if math.IsNaN(*value) || math.IsInf(*value, 0) {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	}

	if !readingInput.BreakerStatus.IsValid() {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	reading := model.Reading{
		DeviceID:      deviceID,
		Timestamp:     readingInput.Timestamp,
		CurrentA:      *readingInput.CurrentA,
		CurrentB:      *readingInput.CurrentB,
		CurrentC:      *readingInput.CurrentC,
		VoltageA:      *readingInput.VoltageA,
		VoltageB:      *readingInput.VoltageB,
		VoltageC:      *readingInput.VoltageC,
		Frequency:     *readingInput.Frequency,
		RealPower:     *readingInput.RealPower,
		ReactivePower: *readingInput.ReactivePower,
		BreakerStatus: *readingInput.BreakerStatus,
	}

	readingResult, err := s.store.AddReading(r.Context(), reading)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrDeviceNotFound):
			http.Error(w, "Device not found", http.StatusNotFound)
		case errors.Is(err, store.ErrDuplicateReading):
			http.Error(w, "Duplicate reading", http.StatusConflict)
		default:
			http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Set content type, set status to created.
	w.WriteHeader(http.StatusNoContent)

	// Publish reading event. Comes after HTTP status set because
	// failing a publish event does not fail the underlying store logic, which is
	// more important.
	if err := s.publishReadingResultEvents(readingResult); err != nil {
		log.Printf("failed to publish reading events: %v", err)
		http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		return
	}

	// INTERNAL API CALL, we do not return encoded JSON body.
}

func (s *Server) getReadingsListHandler(w http.ResponseWriter, r *http.Request) {
	// Verify method
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

	// Get from and to parameters
	query := r.URL.Query()
	fromString := query.Get("from")
	toString := query.Get("to")
	if fromString == "" || toString == "" {
		http.Error(w, "Invalid reading range", http.StatusBadRequest)
		return
	}

	// Get from time
	from, err := time.Parse(time.RFC3339, fromString)
	if err != nil {
		http.Error(w, "Invalid reading range", http.StatusBadRequest)
		return
	}

	// Get to time
	to, err := time.Parse(time.RFC3339, toString)
	if err != nil || to.Before(from) {
		http.Error(w, "Invalid reading range", http.StatusBadRequest)
		return
	}

	// List readings of deviceID from -> to
	readingsList, err := s.store.ListReadings(r.Context(), deviceID, from, to)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrDeviceNotFound):
			http.Error(w, "Device not found", http.StatusNotFound)
		default:
			http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Set status OK and encode readingsList
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(readingsList); err != nil {
		log.Printf("failed to encode readings response: %v", err)
	}
}

// publishReadingResultEvents publishes the accepted reading first, followed by
// alarms raised and cleared while evaluating the reading.
func (s *Server) publishReadingResultEvents(readingResult model.ReadingResult) error {
	// Call payload creation helper
	readingPayload, err := createReadingPayload(readingResult.Reading)
	if err != nil {
		return err
	}
	// Publish new reading event first through server's event broker.
	if err := s.eventBroker.Publish(model.Event{
		EventID:        uuid.New(),
		DeviceID:       readingResult.Reading.DeviceID,
		EventType:      model.ReadingEvent,
		EventTimestamp: readingResult.Reading.Timestamp,
		Payload:        readingPayload,
	}); err != nil {
		return err
	}

	// Iterate raised alarms, and publish new alarm raised events through server's event broker.
	for _, alarm := range readingResult.RuleEvaluationResult.RaisedAlarms {
		alarmPayload, err := createRaisedAlarmPayload(alarm)
		if err != nil {
			return err
		}
		if err := s.eventBroker.Publish(model.Event{
			EventID:        uuid.New(),
			DeviceID:       alarm.DeviceID,
			EventType:      model.AlarmRaisedEvent,
			EventTimestamp: alarm.RaisedAt,
			Payload:        alarmPayload,
		}); err != nil {
			return err
		}
	}

	// Iterate cleared alarms, and publish new alarm cleared events through server's event broker.
	for _, alarm := range readingResult.RuleEvaluationResult.ClearedAlarms {
		alarmPayload, err := createClearedAlarmPayload(alarm)
		if err != nil {
			return err
		}
		if alarm.ClearedAt == nil {
			return errors.New("cleared alarm is missing cleared timestamp")
		}
		if err := s.eventBroker.Publish(model.Event{
			EventID:        uuid.New(),
			DeviceID:       alarm.DeviceID,
			EventType:      model.AlarmClearedEvent,
			EventTimestamp: *alarm.ClearedAt,
			Payload:        alarmPayload,
		}); err != nil {
			return err
		}
	}

	return nil
}

// createReadingPayload serializes a reading event payload.
//
// Example payload:
//
//	{"deviceId":"00000000-0000-0000-0000-000000000001","timestamp":"2026-11-03T20:00:00Z","currentA":10,"currentB":11,"currentC":12,"voltageA":120,"voltageB":121,"voltageC":122,"frequency":60,"realPower":100,"reactivePower":20,"breakerStatus":"closed"}
func createReadingPayload(reading model.Reading) (string, error) {
	payload, err := json.Marshal(reading)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// createRaisedAlarmPayload serializes an alarm-raised event payload.
//
// Example payload:
//
//	{"id":"00000000-0000-0000-0000-000000000002","ruleId":"00000000-0000-0000-0000-000000000003","deviceId":"00000000-0000-0000-0000-000000000001","raisedAt":"2026-11-03T20:00:00Z","triggerValue":401,"ruleSnapshot":{"name":"High current","type":"threshold","measurement":"currentA","comparison":"GT","limit":400}}
func createRaisedAlarmPayload(alarm model.Alarm) (string, error) {
	payload, err := json.Marshal(alarm)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// createClearedAlarmPayload serializes an alarm-cleared event payload.
//
// Example payload:
//
//	{"id":"00000000-0000-0000-0000-000000000002","ruleId":"00000000-0000-0000-0000-000000000003","deviceId":"00000000-0000-0000-0000-000000000001","raisedAt":"2026-11-03T20:00:00Z","clearedAt":"2026-11-03T20:05:00Z","clearReason":"normal","triggerValue":401,"ruleSnapshot":{"name":"High current","type":"threshold","measurement":"currentA","comparison":"GT","limit":400}}
func createClearedAlarmPayload(alarm model.Alarm) (string, error) {
	payload, err := json.Marshal(alarm)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
