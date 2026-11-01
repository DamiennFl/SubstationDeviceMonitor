package cmd

import (
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/api"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
)

func main() {
	store := store.NewStore()
	server := api.NewServer(store)

}
