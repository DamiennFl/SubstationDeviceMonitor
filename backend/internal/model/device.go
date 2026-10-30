package model

import (
	"time"
	"uuid"
)

// Device struct with necessary fields
type Device struct {
	Id            uuid.UUID // Unique UUID
	Name          string    // Name
	Ip            string    // IP Address
	DeviceType    string    // Device Type
	CreatedAt     time.Time // Creation Time and Date
	UpdatedAt     time.Time // Update Time and Date
	DeletedAt     time.Time // Deletion Time and Date
	LastSeen      time.Time // Last time data was updated
	Status        string    // Status of device
	LatestReading Reading   // Latest data reading
}
