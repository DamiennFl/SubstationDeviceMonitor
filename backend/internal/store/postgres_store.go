package store

import (
	"context"
	"errors"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// errNotImplemented is returned by methods that are still stubs, so an unfinished
// method fails loudly (handlers turn it into a 500) instead of silently succeeding.
var errNotImplemented = errors.New("postgres store: not implemented")

// This file is an expansion on store.go, dealing with the Postgres DB instead
// of an in-memory Store. It is defined matching the methods in store.go exactly,
// so the handlers in /internal/api do not have to change anything. Instead of
// a function call like `s.store.ListDevices` accessing the in-memory store,
// it would hit the database. The choice of whether to use the in-memory store
// or the Postgres DB would be set at server.go startup.

type DatabaseStruct struct {
	pool *pgxpool.Pool
}

func NewDatabaseStruct(pool *pgxpool.Pool) *DatabaseStruct {
	return &DatabaseStruct{
		pool: pool,
	}
}

func (db *DatabaseStruct) ListDevices(ctx context.Context) ([]model.Device, error) {

	// Query can be read as:
	// 	- SELECT (all device and reading columns, r.timestamp is aliased reading_timestamp for clarity), from devices table.
	//  - LEFT JOIN LATERAL allows reference to the previous table (the table of devices which we just got using SELECT).
	//  - We SELECT all the reading columns FROM readings WHERE the device_id matches, ordered by timestamp. We limit to 1,
	// 	  and because readings have a (device_id, timestamp) index defined, we can immediately jump to the latest reading. Limit
	//    1 allows return immediately after first reading is found.
	//  - WHERE d.deleted is NULL meaning not soft-deleted devices
	//  - Sort the returned devices by name
	DeviceListQuery := `
		SELECT d.id, d.name, d.ip, d.device_type, d.created_at, d.updated_at, d.last_seen_at,
		       r.timestamp AS reading_timestamp,
		       r.current_a, r.current_b, r.current_c,
		       r.voltage_a, r.voltage_b, r.voltage_c,
		       r.frequency, r.real_power, r.reactive_power, r.breaker_status
		FROM devices d
		LEFT JOIN LATERAL (
			SELECT timestamp, current_a, current_b, current_c, voltage_a, voltage_b, voltage_c,
			       frequency, real_power, reactive_power, breaker_status
			FROM readings
			WHERE device_id = d.id
			ORDER BY timestamp DESC
			LIMIT 1
		) r ON true
		WHERE d.deleted_at IS NULL
		ORDER BY d.name`

	// Get rows with query
	rows, err := db.pool.Query(ctx, DeviceListQuery)
	if err != nil {
		return []model.Device{}, err
	}

	// Scan rows into list of DBDeviceWithReading
	deviceWithReadingRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[DBDeviceWithReading])
	if err != nil {
		return []model.Device{}, err
	}

	// Init device list
	deviceList := make([]model.Device, 0)

	// Loop through rows, define and add devices with their current reading
	for _, deviceWithReading := range deviceWithReadingRows {
		currentDevice := model.Device{
			ID:         deviceWithReading.ID,
			Name:       deviceWithReading.Name,
			IP:         deviceWithReading.IP,
			DeviceType: deviceWithReading.DeviceType, // Already validated in DB
			CreatedAt:  deviceWithReading.CreatedAt,
			UpdatedAt:  deviceWithReading.UpdatedAt,
			LastSeen:   deviceWithReading.LastSeenAt,
		}

		// Set status via helper
		currentDevice.Status = currentDeviceStatus(currentDevice, time.Now().UTC())

		// Create reading
		if deviceWithReading.Timestamp != nil {
			currentReading := model.Reading{
				DeviceID:      deviceWithReading.ID,
				Timestamp:     *deviceWithReading.Timestamp,
				CurrentA:      *deviceWithReading.CurrentA,
				CurrentB:      *deviceWithReading.CurrentB,
				CurrentC:      *deviceWithReading.CurrentC,
				VoltageA:      *deviceWithReading.VoltageA,
				VoltageB:      *deviceWithReading.VoltageB,
				VoltageC:      *deviceWithReading.VoltageC,
				Frequency:     *deviceWithReading.Frequency,
				RealPower:     *deviceWithReading.RealPower,
				ReactivePower: *deviceWithReading.ReactivePower,
				BreakerStatus: *deviceWithReading.BreakerStatus,
			}
			// Set reading
			currentDevice.LatestReading = &currentReading
		}
		// Add to device list
		deviceList = append(deviceList, currentDevice)
	}

	return deviceList, nil
}

func (db *DatabaseStruct) GetDevice(ctx context.Context, deviceID uuid.UUID) (model.Device, error) {
	DeviceListQuery := `
		SELECT d.id, d.name, d.ip, d.device_type, d.created_at, d.updated_at, d.last_seen_at,
		       r.timestamp AS reading_timestamp,
		       r.current_a, r.current_b, r.current_c,
		       r.voltage_a, r.voltage_b, r.voltage_c,
		       r.frequency, r.real_power, r.reactive_power, r.breaker_status
		FROM devices d
		LEFT JOIN LATERAL (
			SELECT timestamp, current_a, current_b, current_c, voltage_a, voltage_b, voltage_c,
			       frequency, real_power, reactive_power, breaker_status
			FROM readings
			WHERE device_id = d.id
			ORDER BY timestamp DESC
			LIMIT 1
		) r ON true
		WHERE d.id = $1 AND d.deleted_at IS NULL`

	// Get one or no row with Query
	rows, err := db.pool.Query(ctx, DeviceListQuery, deviceID)
	if err != nil {
		return model.Device{}, err
	}

	// Get single device row, or no row.
	deviceWithReading, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DBDeviceWithReading])
	if err != nil {
		// Catch no row and return device not found.
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Device{}, ErrDeviceNotFound
		}
		return model.Device{}, err
	}

	// Build device
	device := model.Device{
		ID:         deviceWithReading.ID,
		Name:       deviceWithReading.Name,
		IP:         deviceWithReading.IP,
		DeviceType: deviceWithReading.DeviceType, // Already validated in DB
		CreatedAt:  deviceWithReading.CreatedAt,
		UpdatedAt:  deviceWithReading.UpdatedAt,
		LastSeen:   deviceWithReading.LastSeenAt,
	}

	// Set status via helper
	device.Status = currentDeviceStatus(device, time.Now().UTC())

	// Create reading
	if deviceWithReading.Timestamp != nil {
		currentReading := model.Reading{
			DeviceID:      deviceWithReading.ID,
			Timestamp:     *deviceWithReading.Timestamp,
			CurrentA:      *deviceWithReading.CurrentA,
			CurrentB:      *deviceWithReading.CurrentB,
			CurrentC:      *deviceWithReading.CurrentC,
			VoltageA:      *deviceWithReading.VoltageA,
			VoltageB:      *deviceWithReading.VoltageB,
			VoltageC:      *deviceWithReading.VoltageC,
			Frequency:     *deviceWithReading.Frequency,
			RealPower:     *deviceWithReading.RealPower,
			ReactivePower: *deviceWithReading.ReactivePower,
			BreakerStatus: *deviceWithReading.BreakerStatus,
		}
		// Set reading
		device.LatestReading = &currentReading
	}

	// Return device
	return device, nil
}

func (db *DatabaseStruct) CreateDevice(ctx context.Context, device model.Device) (model.Device, error) {
	return model.Device{}, errNotImplemented
}

func (db *DatabaseStruct) UpdateDevice(ctx context.Context, deviceID uuid.UUID, edit func(model.Device) (model.Device, error)) (model.Device, error) {
	return model.Device{}, errNotImplemented
}

func (db *DatabaseStruct) DeleteDevice(ctx context.Context, deviceID uuid.UUID) error {
	return errNotImplemented
}

func (db *DatabaseStruct) ListRules(ctx context.Context, deviceID uuid.UUID) ([]model.Rule, error) {
	return nil, errNotImplemented
}

func (db *DatabaseStruct) GetRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID) (model.Rule, error) {
	return model.Rule{}, errNotImplemented
}

func (db *DatabaseStruct) CreateRule(ctx context.Context, rule model.Rule) (model.Rule, error) {
	return model.Rule{}, errNotImplemented
}

func (db *DatabaseStruct) UpdateRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID, edit func(rule model.Rule) (model.Rule, error)) (model.Rule, error) {
	return model.Rule{}, errNotImplemented
}

func (db *DatabaseStruct) DeleteRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID) error {
	return errNotImplemented
}

func (db *DatabaseStruct) ListAlarms(ctx context.Context, deviceID uuid.UUID) ([]model.AlarmWithAcknowledgement, error) {
	return nil, errNotImplemented
}

func (db *DatabaseStruct) AcknowledgeAlarm(ctx context.Context, deviceID uuid.UUID, alarmID uuid.UUID, ackBy string) (model.AlarmWithAcknowledgement, error) {
	return model.AlarmWithAcknowledgement{}, errNotImplemented
}

func (db *DatabaseStruct) AddReading(ctx context.Context, reading model.Reading) (model.ReadingResult, error) {
	return model.ReadingResult{}, errNotImplemented
}

func (db *DatabaseStruct) ListReadings(ctx context.Context, deviceID uuid.UUID, start time.Time, end time.Time) ([]model.Reading, error) {
	return nil, errNotImplemented
}

// currentDeviceStatus checks a device's status and updates accordingly.
func currentDeviceStatus(device model.Device, now time.Time) model.DeviceStatus {
	// No need to update device.Status in Store, because status is derived data.
	// Also allows us to keep read lock with ListDevice and GetDevice, instead of having write lock.
	if device.LastSeen == nil {
		return model.DeviceStatusPending
	}
	if now.Before(device.LastSeen.Add(time.Minute)) {
		return model.DeviceStatusStale
	}
	return model.DeviceStatusOnline
}
