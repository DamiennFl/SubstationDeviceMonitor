package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
	"github.com/google/uuid"
)

func (s *Server) getDeviceAlarmListHandler(w http.ResponseWriter, r *http.Request) {
	// Validate method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get device ID
	deviceIDString := r.PathValue("device_id")
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	alarmList, err := s.store.ListAlarms(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			http.Error(w, "Device not found", http.StatusNotFound)
		} else {
			http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(alarmList); err != nil {
		log.Printf("failed to encode alarm list response: %v", err)
		return
	}
}

func (s *Server) acknowledgeAlarmHandler(w http.ResponseWriter, r *http.Request) {
	// Validate method
	// Get deviceId and ruleID
	// Create acknowledgement
	// Acknowledge Alarm
	// Return acknowledged alarm

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deviceIDString := r.PathValue("device_id")
	deviceID, err := uuid.Parse(deviceIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	alarmIDString := r.PathValue("alarm_id")
	alarmID, err := uuid.Parse(alarmIDString)
	if err != nil {
		http.Error(w, "Malformed request or missing parameters", http.StatusBadRequest)
		return
	}

	acknowledgedBy := "admin" // TODO: PLACEHOLDER, UPDATE IF/WHEN ACCOUNTS EXIST

	alarmWithAck, err := s.store.AcknowledgeAlarm(r.Context(), deviceID, alarmID, acknowledgedBy)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrAlarmNotFound):
			http.Error(w, "Alarm not found", http.StatusNotFound)
		case errors.Is(err, store.ErrAlarmAlreadyAcknowledged):
			http.Error(w, "Alarm already acknowledged", http.StatusConflict)
		default:
			http.Error(w, "Unexpected internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(alarmWithAck); err != nil {
		log.Printf("failed to encode acknowledged alarm response: %v", err)
		return
	}

	if err := s.publishAcknowledgeAlarmEvent(alarmWithAck); err != nil {
		log.Printf("failed to publish alarm ack event: %v", err)
		return
	}
}

// Sends an alarm acknowledgement event through the server's event broker.
func (s *Server) publishAcknowledgeAlarmEvent(alarmWithAck model.AlarmWithAcknowledgement) error {
	acknowledgementPayload, err := createAckPayload(alarmWithAck)
	if err != nil {
		return err
	}

	// Publish event with necessary values
	if err := s.eventBroker.Publish(model.Event{
		EventID:        uuid.New(),
		DeviceID:       alarmWithAck.Alarm.DeviceID,
		EventType:      model.AlarmAckedEvent,
		EventTimestamp: alarmWithAck.Acknowledgement.AcknowledgedAt,
		Payload:        acknowledgementPayload,
	}); err != nil {
		return err
	}

	return nil
}

func createAckPayload(alarmWithAck model.AlarmWithAcknowledgement) (string, error) {
	// Create JSON for alarm and acknowledgment
	payload, err := json.Marshal(alarmWithAck)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}
