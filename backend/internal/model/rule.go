package model

import (
	"time"
	"uuid"
)

type Rule struct {
	Id          uuid.UUID // Rule ID
	DeviceID    uuid.UUID // Mapped Device ID
	Name        string    // Rule NAme
	Enabled     bool      // Is Rule Enabled
	Type        string    // Type of Rule (threshold, band, status)
	Measurement string    // Measurement Name from Device
	Comparison  string    // Comparison (GT or LT)
	Limit       float64   // Limit for threshold
	LowerLimit  float64   // Lower limit for band
	UpperLimit  float64   // Upper limit for band
	Deadband    float64   // Deadband
	TripState   string    // Trip State (Open or Closed)
	CreatedAt   time.Time // Creation Time and Date
	UpdatedAt   time.Time // Update Time and Date
}
