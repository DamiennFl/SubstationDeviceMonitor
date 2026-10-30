package model

import (
	"time"

	"github.com/google/uuid"
)

type Rule struct {
	ID          uuid.UUID `json:"id"`                   // Rule ID
	DeviceID    uuid.UUID `json:"deviceId"`             // Mapped Device ID
	Name        string    `json:"name"`                 // Rule Name
	Enabled     bool      `json:"enabled"`              // Is Rule Enabled
	Type        string    `json:"type"`                 // Type of Rule (threshold, band, status)
	Measurement string    `json:"measurement"`          // Measurement Name from Device
	Comparison  *string   `json:"comparison,omitempty"` // Comparison (GT or LT)
	Limit       *float64  `json:"limit,omitempty"`      // Limit for threshold
	LowerLimit  *float64  `json:"lowerLimit,omitempty"` // Lower limit for band
	UpperLimit  *float64  `json:"upperLimit,omitempty"` // Upper limit for band
	Deadband    *float64  `json:"deadband,omitempty"`   // Deadband
	TripState   *string   `json:"tripState,omitempty"`  // Trip State (Open or Closed)
	CreatedAt   time.Time `json:"createdAt"`            // Creation Time and Date
	UpdatedAt   time.Time `json:"updatedAt"`            // Update Time and Date
}
