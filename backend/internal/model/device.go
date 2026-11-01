package model

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a monitored device.
type Device struct {
	ID            uuid.UUID    `json:"deviceId"`                // Unique UUID
	Name          string       `json:"name"`                    // Name
	IP            string       `json:"ip"`                      // IP Address
	DeviceType    DeviceType   `json:"type"`                    // Device Type
	CreatedAt     time.Time    `json:"createdAt"`               // Creation Time and Date
	UpdatedAt     time.Time    `json:"updatedAt"`               // Update Time and Date
	DeletedAt     *time.Time   `json:"-"`                       // Deletion Time and Date, skipped for API
	LastSeen      *time.Time   `json:"lastSeen,omitempty"`      // Last time data was updated
	Status        DeviceStatus `json:"status"`                  // Status of device
	LatestReading *Reading     `json:"latestReading,omitempty"` // Latest data reading
}
