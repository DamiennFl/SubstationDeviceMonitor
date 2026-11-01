package main

import (
	"log"
	"net/http"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/api"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
)

// Main function for API
func main() {
	// Create new Store
	appStore := store.NewStore()
	// Create new DeviceService
	deviceService := api.NewDeviceService(appStore)
	// Create Server with new DeviceService
	server := api.NewServer(deviceService)

	// Listen and serve server.Routes, right now at localhost. Later will be at home page URL.
	log.Fatal(http.ListenAndServe("localhost:8080", server.Routes()))
}
