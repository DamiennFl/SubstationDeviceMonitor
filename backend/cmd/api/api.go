package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/api"
	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// newDataStore picks the data store from the environment. If DATABASE_URL is set it connects
// to Postgres (and exits if it can't), otherwise it uses the in-memory Store.
// The returned func releases the store's resources and should be called on shutdown.
func newDataStore() (api.DataStore, func()) {
	// Get DB URL from env var
	dbURL := os.Getenv("DATABASE_URL")

	// No env var, default to in-memory store
	if dbURL == "" {
		log.Printf("No database URL found, defaulting to in-memory Store.")
		return store.NewStore(), func() {}
	}

	// Create new Postgres connection pool
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Error during database initialization: %s", err)
	}

	// pgxpool.New does not connect, so ping to confirm the database is reachable.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Pinging database failed: %s", err)
	}

	log.Printf("Connected to database, using Postgres store.")
	// Return db struct with pool, and pool close function
	return store.NewDatabaseStruct(pool), pool.Close
}

// Main function for API
func main() {
	// Make data store
	appStore, closeStore := newDataStore()

	// Create Server with the chosen data store
	server := api.NewServer(appStore)

	// Listen and serve server.Routes, right now at localhost. Later will be at home page URL.
	// ListenAndServe only returns on error, so release the store before exiting.
	err := http.ListenAndServe(":8080", server.Routes())
	closeStore()
	log.Fatal(err)
}
