package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

// Simulates power grid devices sending readings, represents devices, occasionally trips relays or satisfies conditions to create alarms.

type Simulator struct {
	Devices    map[uuid.UUID]model.Device
	Rules      map[uuid.UUID][]model.Rule // Cached rules per device, refreshed by the sync loop
	APIBaseURL string                     // Simulator needs an API to GET and POST to
	HTTPClient http.Client                // Simulator needs an HTTP client to send requests
	mu         sync.RWMutex
}

// Define some default values to fall back on in case Rule fetch does not work
var DefaultReadingValues = model.Reading{
	CurrentA:      10,
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

// Main which loops and keeps the simulator running continuously
func main() {
	// Get env variable from Docker. If not present fall back to localhost.
	API_BASE_URL := os.Getenv("API_BASE_URL")
	if API_BASE_URL == "" {
		API_BASE_URL = "http://localhost:8080"
	}
	sim := Simulator{
		Devices:    make(map[uuid.UUID]model.Device),
		Rules:      make(map[uuid.UUID][]model.Rule),
		APIBaseURL: API_BASE_URL,
		HTTPClient: http.Client{Timeout: 10 * time.Second},
	}

	err := sim.syncDevices(context.Background())

	if err != nil {
		log.Fatalf("initial device sync failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Define shutdown OS Signal channel
	shutdown := make(chan os.Signal, 1)
	// Set valid notify signals as interrupt or graceful term
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM) // Notify shutdown via interrupt or graceful termination.

	// Make wait group. Each contains ticker which will loop
	var wg sync.WaitGroup
	wg.Go(func() {
		sim.DeviceSyncLoop(ctx)
	})
	wg.Go(func() {
		sim.DeviceReaderPOSTLoop(ctx)
	})

	// If shutdown comes through, cancel context, then wait for goroutines to finish in WaitGroup.
	<-shutdown
	cancel()
	wg.Wait()
}

func (s *Simulator) DeviceSyncLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C: // Every 15 seconds
			if err := s.syncDevices(ctx); err != nil { // Sync devices
				log.Printf("error while syncing devices: %v", err)
			}
		case <-ctx.Done(): // Context closed, return
			return
		}
	}
}

func (s *Simulator) DeviceReaderPOSTLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C: // Every second
			s.SendReadings() // Send readings
		case <-ctx.Done(): // Context closed, return
			return
		}
	}
}

// To keep this from actually being hooked up to the API, we will just run it every 15 as a GET request to the device list from server.
// This will re-populate the device map and then we can do stuff with them.
func (s *Simulator) syncDevices(ctx context.Context) error {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		strings.TrimRight(s.APIBaseURL, "/")+"/api/devices",
		nil,
	)
	if err != nil {
		return err
	}

	response, err := s.HTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Device sync failed, return error
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("device sync returned status %s", response.Status)
	}

	// Create devices list and decode devices
	var devices []model.Device
	if err := json.NewDecoder(response.Body).Decode(&devices); err != nil {
		return fmt.Errorf("decode device sync response: %w", err)
	}

	// Initialize map of devices, then add all devices in returned list
	devicesByID := make(map[uuid.UUID]model.Device, len(devices))
	for _, device := range devices {
		devicesByID[device.ID] = device
	}

	// Refresh rules for each device here (outside the lock, since these are HTTP calls),
	// so the per-second reading loop never has to hit the API for rules.
	s.mu.RLock()
	previousRules := s.Rules
	s.mu.RUnlock()

	rulesByDevice := make(map[uuid.UUID][]model.Rule, len(devices))
	for deviceID := range devicesByID {
		rules, err := s.fetchDeviceRules(ctx, deviceID)
		if err != nil {
			// Keep last known rules for this device rather than dropping them
			log.Printf("rule sync failed for device %s: %v", deviceID, err)
			if old, ok := previousRules[deviceID]; ok {
				rulesByDevice[deviceID] = old
			}
			continue
		}
		rulesByDevice[deviceID] = rules
	}

	s.mu.Lock()
	s.Devices = devicesByID
	s.Rules = rulesByDevice
	s.mu.Unlock()

	// No error: return nil
	return nil
}

// Handle logic for tripping random alarm in here.
func (s *Simulator) sendDeviceReading(deviceID uuid.UUID) {
	// Validte device based on device ID
	s.mu.RLock()
	_, ok := s.Devices[deviceID]
	s.mu.RUnlock()
	if !ok {
		log.Printf("cannot send reading: device %s is not in simulator state", deviceID)
		return
	}

	// For each device, randomly choose whether or not it will be tripped.
	var reading model.Reading
	var err error
	// Generate reading, either within or out of rule bounds.
	if rand.Intn(10) == (0 | 1) {
		reading, err = s.generateTrippingReading(deviceID)
	} else {
		reading, err = s.generateNominalReading(deviceID)
	}
	if err != nil {
		log.Printf("cannot generate reading for device %s: %v", deviceID, err)
		return
	}

	// Create Reading JSON
	payload, err := json.Marshal(reading)
	if err != nil {
		log.Printf("cannot encode reading for device %s: %v", deviceID, err)
		return
	}

	// Create new POST request, to POST device's new Reading, simulating data being sent from device.
	// Send POST request
	response, err := s.HTTPClient.Post(
		strings.TrimRight(s.APIBaseURL, "/")+"/api/readings",
		"application/json",
		bytes.NewReader(payload),
	)

	if err != nil {
		log.Printf("cannot send reading for device %s: %v", deviceID, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		log.Printf("reading for device %s returned status %s", deviceID, response.Status)
	}
}

// This one will be hooked up to the server and send via the below POST for each device:
//
//	mux.HandleFunc("POST /api/readings", server.createReadingHandler) // Only for Simulator
//
// This function should just iterate over the list of devices and call sendDeviceReading for each.
func (s *Simulator) SendReadings() {
	s.mu.RLock()
	deviceIDs := make([]uuid.UUID, 0, len(s.Devices))
	for deviceID := range s.Devices {
		deviceIDs = append(deviceIDs, deviceID)
	}
	s.mu.RUnlock()

	for _, deviceID := range deviceIDs {
		s.sendDeviceReading(deviceID)
	}
}

// Look up the device's cached rules (populated by syncDevices). Returns an error if none are cached yet,
// so callers fall back to default readings.
func (s *Simulator) getDeviceRules(deviceID uuid.UUID) ([]model.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rules, ok := s.Rules[deviceID]
	if !ok {
		return nil, fmt.Errorf("no cached rules for device %s", deviceID)
	}
	return rules, nil
}

// Get the specified device's rules from the server, so we can create readings that violate its rules
func (s *Simulator) fetchDeviceRules(ctx context.Context, deviceID uuid.UUID) ([]model.Rule, error) {
	// Create GET request. TrimRight removes trailing "/"
	url := strings.TrimRight(s.APIBaseURL, "/") + "/api/devices/" + deviceID.String() + "/rules"

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Send GET request
	response, err := s.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	// Make sure to close body since response is data usstream, once we are done with it
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rule fetch returned status %s", response.Status)
	}

	// Decode rules into list and return if no errors occur
	var rules []model.Rule
	if err := json.NewDecoder(response.Body).Decode(&rules); err != nil {
		return nil, fmt.Errorf("decode rule response: %w", err)
	}
	return rules, nil
}

// Create basic reading with default vals and now timestamp, with deviceID.
// Backup if getting rules fails, so the device still sends something.
// Subject to change if during testing this generates a ton of unnecessary alarms.
func baseReading() model.Reading {
	reading := DefaultReadingValues
	reading.Timestamp = time.Now().UTC()
	return reading
}

// Update measurement on a reading.
// Used to set reading vals in or out of bounds.
func setMeasurement(reading *model.Reading, measurement model.Measurement, value any) {
	switch measurement {
	case model.MeasurementCurrentA:
		reading.CurrentA = value.(float64)
	case model.MeasurementCurrentB:
		reading.CurrentB = value.(float64)
	case model.MeasurementCurrentC:
		reading.CurrentC = value.(float64)
	case model.MeasurementVoltageA:
		reading.VoltageA = value.(float64)
	case model.MeasurementVoltageB:
		reading.VoltageB = value.(float64)
	case model.MeasurementVoltageC:
		reading.VoltageC = value.(float64)
	case model.MeasurementFrequency:
		reading.Frequency = value.(float64)
	case model.MeasurementRealPower:
		reading.RealPower = value.(float64)
	case model.MeasurementReactivePower:
		reading.ReactivePower = value.(float64)
	case model.MeasurementBreakerStatus:
		reading.BreakerStatus = value.(model.State)
	}
}

// generateNominalReading gets the device's list of rules via GET, then generates a
// reading which will fit within all Rule parameters. Any values that are not
// needed for a rule will be default values.
func (s *Simulator) generateNominalReading(deviceID uuid.UUID) (model.Reading, error) {
	// Start with base reading
	reading := baseReading()
	// Set Device ID
	reading.DeviceID = deviceID
	// Get list of rules
	rules, err := s.getDeviceRules(deviceID)
	// Getting rules failed, use backup defaults
	if err != nil {
		log.Printf("using default nominal reading for device %s: %v", deviceID, err)
		return reading, nil
	}
	// Iterate rules
	for _, rule := range rules {
		// Skip rule if it isn't enabled
		if !rule.Enabled {
			continue
		}
		switch rule.Type {
		// Check relay trip status
		case model.RuleTypeStatus:
			if rule.TripState != nil {
				status := model.StateClosed
				if *rule.TripState == status {
					status = model.StateOpen // Open relay
				}
				setMeasurement(&reading, rule.Measurement, status) // Update measurement in reading
			}
		// If rule is threshold rule
		case model.RuleTypeThreshold:
			// If rule does not define limit and comparison, skip it and don't set measurement values
			if rule.Limit == nil || rule.Comparison == nil {
				continue
			}
			// If comparison is a GT comparison, set value less than limit. Opposite for LT.
			if *rule.Comparison == model.ComparisonGT {
				setMeasurement(&reading, rule.Measurement, *rule.Limit-5)
			} else {
				setMeasurement(&reading, rule.Measurement, *rule.Limit+5)
			}
		// If rule is a band rule
		case model.RuleTypeBand:
			// Set value right between lower and upper limit
			if rule.LowerLimit != nil && rule.UpperLimit != nil {
				setMeasurement(&reading, rule.Measurement, (*rule.LowerLimit+*rule.UpperLimit)/2)
			}
		}
	}
	return reading, nil
}

// generateTrippingReading gets the device's list of rules via GET, then generates a
// reading which will trip a Rule. The rule is randomly chosen.
func (s *Simulator) generateTrippingReading(deviceID uuid.UUID) (model.Reading, error) {
	// Start with base reading
	reading := baseReading()
	reading.DeviceID = deviceID
	// Get device's rules.
	rules, err := s.getDeviceRules(deviceID)
	// If GET rules fails, then fall back to default values.
	if err != nil {
		log.Printf("using default reading for device %s: %v", deviceID, err)
		return reading, nil
	}
	// Count enabled rules. Used to randomly pick them
	enabledCount := 0
	for _, rule := range rules {
		if rule.Enabled {
			enabledCount++
		}
	}
	// If no rules enabled, return default reading.
	if enabledCount == 0 {
		return reading, nil
	}

	// Get a random enabled rule index
	enabledIndex := rand.Intn(enabledCount)
	// Set count of enabled rules we have traversed to 0
	enabledCount = 0
	var rule model.Rule
	// Iterate rules
	for _, candidate := range rules {
		// If rule is not enabled, continue
		if !candidate.Enabled {
			continue
		}
		// If we found enabled rule equal to the random enabled index, set rule and exit loop
		if enabledCount == enabledIndex {
			rule = candidate
			break
		}
		// Iterate
		enabledCount++
	}
	switch rule.Type {
	// If rule is the status,
	case model.RuleTypeStatus:
		// If the rule does have a defined TripState, set measurement to it, to trip.
		if rule.TripState != nil {
			setMeasurement(&reading, rule.Measurement, *rule.TripState)
		}
	// If rule is a threshold
	case model.RuleTypeThreshold:
		// Check that there is limit and comparison.
		if rule.Limit != nil && rule.Comparison != nil {
			// Set measurement above or below threshold, depending on GT or LT.
			if *rule.Comparison == model.ComparisonGT {
				setMeasurement(&reading, rule.Measurement, *rule.Limit+5)
			} else {
				setMeasurement(&reading, rule.Measurement, *rule.Limit-5)
			}
		}
	// If rule is a band
	case model.RuleTypeBand:
		// Check it has limits and randomly set value above or below band
		if rule.LowerLimit != nil && rule.UpperLimit != nil {
			value := *rule.LowerLimit - 5
			if rand.Intn(2) == 1 {
				value = *rule.UpperLimit + 5
			}
			setMeasurement(&reading, rule.Measurement, value)
		}
	}
	return reading, nil
}
