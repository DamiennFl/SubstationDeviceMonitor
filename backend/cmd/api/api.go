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
	// Create Server with new DeviceService
	server := api.NewServer(appStore)

	// Listen and serve server.Routes, right now at localhost. Later will be at home page URL.
	log.Fatal(http.ListenAndServe(":8080", server.Routes()))
}
