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

	// Listen and serve server.Routes, from home page of app.
	log.Fatal(http.ListenAndServe("/", server.Routes()))
}
