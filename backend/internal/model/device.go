package model

import (
	"time"
	"uuid"
)

// Device struct with necessary fields
type Device struct {
	Id            uuid.UUID  `json:"id"`                      // Unique UUID
	Name          string     `json:"name"`                    // Name
	Ip            string     `json:"ip"`                      // IP Address
	DeviceType    string     `json:"type"`                    // Device Type
	CreatedAt     time.Time  `json:"createdAt"`               // Creation Time and Date
	UpdatedAt     time.Time  `json:"updatedAt"`               // Update Time and Date
	DeletedAt     *time.Time `json:"deletedAt,omitempty"`     // Deletion Time and Date
	LastSeen      *time.Time `json:"lastSeen,omitempty"`      // Last time data was updated
	Status        string     `json:"status"`                  // Status of device
	LatestReading *Reading   `json:"latestReading,omitempty"` // Latest data reading
}
