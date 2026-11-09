package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	deviceListQuery := `
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
	rows, err := db.pool.Query(ctx, deviceListQuery)
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

	// Loop through rows, build each device (with its latest reading) via helper
	now := time.Now().UTC()
	for _, deviceWithReading := range deviceWithReadingRows {
		deviceList = append(deviceList, deviceFromRow(deviceWithReading, now))
	}

	return deviceList, nil
}

func (db *DatabaseStruct) GetDevice(ctx context.Context, deviceID uuid.UUID) (model.Device, error) {
	deviceGetQuery := `
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
	rows, err := db.pool.Query(ctx, deviceGetQuery, deviceID)
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

	// Build and return device via helper
	return deviceFromRow(deviceWithReading, time.Now().UTC()), nil
}

func (db *DatabaseStruct) CreateDevice(ctx context.Context, device model.Device) (model.Device, error) {
	// Null UUID check
	if device.ID == uuid.Nil {
		return model.Device{}, ErrNilDeviceUUID
	}

	// Define query with placeholders
	deviceCreateQuery := `
		INSERT INTO devices (id, name, ip, device_type)
		VALUES($1, $2, $3, $4)
		RETURNING created_at, updated_at;`

	// Query row with placeholders, scan return values into Device
	err := db.pool.QueryRow(
		ctx,
		deviceCreateQuery,
		device.ID,
		device.Name,
		device.IP,
		device.DeviceType,
	).Scan(&device.CreatedAt, &device.UpdatedAt)

	// Check for duplicate error returned from Postgres. Other errors return themselves.
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Device{}, ErrDuplicateDevice
		}
		return model.Device{}, err
	}

	// Set status before return
	device.Status = currentDeviceStatus(device, time.Now().UTC())

	return device, nil
}

// UpdateDevice reads the device, runs edit on it in Go, and writes the result back, all in one
// transaction. The row is locked with FOR UPDATE so concurrent updates can't overwrite each other
// (the in-memory Store gets this from its mutex).
func (db *DatabaseStruct) UpdateDevice(ctx context.Context, deviceID uuid.UUID, edit func(model.Device) (model.Device, error)) (model.Device, error) {
	if edit == nil {
		return model.Device{}, ErrDuringDeviceUpdate
	}

	// Select the device's own columns and lock the row until this transaction ends.
	// Soft-deleted devices are treated as not found
	deviceLockQuery := `
		SELECT id, name, ip, device_type, created_at, updated_at, last_seen_at
		FROM devices
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE`

	// Only the fields an edit is allowed to change. created_at, last_seen_at and deleted_at are left alone
	deviceUpdateQuery := `
		UPDATE devices
		SET name = $2, ip = $3, device_type = $4, updated_at = NOW()
		WHERE id = $1`

	// BeginFunc commits if the function returns nil, and rolls back if it returns an error.
	// Errors returned from inside come back unchanged, so the sentinel errors reach the handler
	err := pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		// Read and lock current device, scanning straight into a model.Device
		var device model.Device
		err := tx.QueryRow(ctx, deviceLockQuery, deviceID).Scan(
			&device.ID,
			&device.Name,
			&device.IP,
			&device.DeviceType,
			&device.CreatedAt,
			&device.UpdatedAt,
			&device.LastSeen,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrDeviceNotFound
			}
			return err
		}

		// Run the caller's edit in Go. If it fails, return its error to roll back
		newDevice, err := edit(device)
		if err != nil {
			return err
		}

		// An edit must not change the device's ID
		if newDevice.ID != device.ID {
			return ErrDuringDeviceUpdate
		}

		// Write the edited fields back. A duplicate name or IP raises a unique violation
		_, err = tx.Exec(ctx, deviceUpdateQuery, deviceID, newDevice.Name, newDevice.IP, newDevice.DeviceType)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrDuplicateDevice
			}
			return err
		}

		// Function param returns nothing
		return nil
	})
	if err != nil {
		return model.Device{}, err
	}

	// Transaction is committed. Read the device back so the response has the new updated_at,
	// Status and LatestReading, same as the in-memory Store returns
	return db.GetDevice(ctx, deviceID)
}

// DeleteDevice soft-deletes a device. In one transaction it: refuses if any active alarm is
// unacknowledged, clears the remaining active alarms with reason "deviceDeleted", and sets deleted_at
func (db *DatabaseStruct) DeleteDevice(ctx context.Context, deviceID uuid.UUID) error {
	// Lock the device row until this transaction ends. Unlike UpdateDevice, soft-deleted devices are
	// selected too, so "already deleted" can be told apart from "never existed".
	deviceLockQuery := `
		SELECT deleted_at
		FROM devices
		WHERE id = $1
		FOR UPDATE`

	// An alarm blocks deletion if it is active (not cleared) and has no acknowledgement row
	// Just get first one if any, no need for more
	unacknowledgedAlarmQuery := `
		SELECT EXISTS (
			SELECT 1
			FROM alarms a
			WHERE a.device_id = $1
			  AND a.cleared_at IS NULL
			  AND NOT EXISTS (SELECT 1 FROM acknowledgements k WHERE k.alarm_id = a.id)
		)`

	// Clear whatever active alarms remain (they are all acknowledged at this point)
	clearAlarmsQuery := `
		UPDATE alarms
		SET cleared_at = NOW(), clear_reason = $2
		WHERE device_id = $1 AND cleared_at IS NULL`

	// Soft delete. NOW() is the transaction's start time, so alarms and device get the same timestamp
	deleteDeviceQuery := `
		UPDATE devices
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1`

	// BeginFunc commits if the function returns nil, and rolls back if it returns an error.
	// Errors returned from inside come back unchanged, so the sentinel errors reach the handler
	return pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		// Lock the device and read its deleted_at. It is NULL for an active device.
		var deletedAt *time.Time
		err := tx.QueryRow(ctx, deviceLockQuery, deviceID).Scan(&deletedAt)
		// Error occurred during row query. Check device not found before returning other errors
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrDeviceNotFound
			}
			return err
		}
		// Device already deleted
		if deletedAt != nil {
			return ErrDeviceAlreadyDeleted
		}

		// Refuse if an active alarm has not been acknowledged
		var hasUnacknowledged bool
		err = tx.QueryRow(ctx, unacknowledgedAlarmQuery, deviceID).Scan(&hasUnacknowledged)
		if err != nil {
			return err
		}
		if hasUnacknowledged {
			return ErrUnacknowledgedAlarm
		}

		// Clear the remaining active alarms
		_, err = tx.Exec(ctx, clearAlarmsQuery, deviceID, model.ClearReasonDeviceDeleted)
		if err != nil {
			return err
		}

		// Soft delete the device
		_, err = tx.Exec(ctx, deleteDeviceQuery, deviceID)
		return err
	})
}

func (db *DatabaseStruct) ListRules(ctx context.Context, deviceID uuid.UUID) ([]model.Rule, error) {
	ruleListQuery := `
		SELECT r.id, r.device_id, r.name, r.enabled, r.rule_type, r.measurement, r.comparison,
			   r.limit_value, r.lower_limit, r.upper_limit, r.deadband, r.trip_state, r.created_at,
			   r.updated_at
		FROM rules r
		WHERE r.device_id = $1
		ORDER BY r.name`

	// Get rules list for device
	rows, err := db.pool.Query(ctx, ruleListQuery, deviceID)
	if err != nil {
		return []model.Rule{}, err
	}

	// Scan rules into list of DBRules
	rules, err := pgx.CollectRows(rows, pgx.RowToStructByName[DBRule])
	if err != nil {
		return []model.Rule{}, err
	}

	// Convert each row struct into a model.Rule
	ruleList := make([]model.Rule, 0, len(rules))

	for _, rule := range rules {
		ruleList = append(ruleList, ruleFromRow(rule))
	}

	return ruleList, nil
}

func (db *DatabaseStruct) GetRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID) (model.Rule, error) {
	ruleGetQuery := `
		SELECT r.id, r.device_id, r.name, r.enabled, r.rule_type, r.measurement, r.comparison,
			   r.limit_value, r.lower_limit, r.upper_limit, r.deadband, r.trip_state, r.created_at,
			   r.updated_at
		FROM rules r
		WHERE r.id = $1 AND r.device_id = $2
		ORDER BY r.name`

	// Get one or no row with Query
	rows, err := db.pool.Query(ctx, ruleGetQuery, ruleID, deviceID)
	if err != nil {
		return model.Rule{}, err
	}

	// Get single rule row, or no row.
	ruleRow, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DBRule])
	if err != nil {
		// Catch no row and return device not found.
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Rule{}, ErrRuleNotFound
		}
		return model.Rule{}, err
	}

	// Build return Rule
	foundRule := model.Rule(ruleRow)

	return foundRule, nil
}

// CreateRule inserts a rule for a device. In one transaction it checks that the device exists and
// is not soft-deleted (locking it so DeleteDevice can't run in between), then inserts the rule
func (db *DatabaseStruct) CreateRule(ctx context.Context, rule model.Rule) (model.Rule, error) {
	// Null UUID check. A nil DeviceID needs no check of its own: no device has that ID, so the
	// device lookup below returns ErrDeviceNotFound, which is what the handler expects
	if rule.ID == uuid.Nil {
		return model.Rule{}, ErrNilRuleUUID
	}

	// Lock the device row in share mode. Other readers and share locks are fine, but DeleteDevice
	// (which uses FOR UPDATE) has to wait until this transaction ends
	// Soft-deleted devices still exist as rows, so they are filtered out here, not by a foreign key
	deviceLockQuery := `
		SELECT 1
		FROM devices
		WHERE id = $1 AND deleted_at IS NULL
		FOR SHARE`

	// Define query with placeholders
	ruleCreateQuery := `
		INSERT INTO rules (id, device_id, name, enabled, rule_type, measurement, comparison,
							limit_value, lower_limit, upper_limit, deadband, trip_state)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at;
	`

	// BeginFunc commits if the function returns nil, and rolls back if it returns an error
	// Errors returned from inside come back unchanged, so the sentinel errors reach the handler
	err := pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		// Make sure the device exists and is not deleted
		var one int
		err := tx.QueryRow(ctx, deviceLockQuery, rule.DeviceID).Scan(&one)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrDeviceNotFound
			}
			return err
		}

		// Insert the rule, scan the database-generated created_at and updated_at into the Rule
		err = tx.QueryRow(
			ctx,
			ruleCreateQuery,
			rule.ID,
			rule.DeviceID,
			rule.Name,
			rule.Enabled,
			rule.Type,
			rule.Measurement,
			rule.Comparison,
			rule.Limit,
			rule.LowerLimit,
			rule.UpperLimit,
			rule.Deadband,
			rule.TripState,
		).Scan(&rule.CreatedAt, &rule.UpdatedAt)

		// Duplicate name on this device, or duplicate ID. Other errors return themselves
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrDuplicateRule
			}
			return err
		}

		return nil
	})
	if err != nil {
		return model.Rule{}, err
	}

	return rule, nil
}

func (db *DatabaseStruct) UpdateRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID, edit func(rule model.Rule) (model.Rule, error)) (model.Rule, error) {
	// No edit function, nothing to run
	if edit == nil {
		return model.Rule{}, ErrDuringRuleUpdate
	}

	// Lock the device row in share mode: a concurrent DeleteDevice (FOR UPDATE) has to wait for this
	// transaction. Soft-deleted devices are treated as not found
	deviceLockQuery := `
		SELECT 1
		FROM devices
		WHERE id = $1 AND deleted_at IS NULL
		FOR SHARE`

	// Lock the rule row so two updates of the same rule can't overwrite each other.
	// Matching device_id too means a rule ID from another device counts as not found.
	ruleLockQuery := `
		SELECT id, device_id, name, enabled, rule_type, measurement, comparison,
		       limit_value, lower_limit, upper_limit, deadband, trip_state, created_at, updated_at
		FROM rules
		WHERE id = $1 AND device_id = $2
		FOR UPDATE`

	// Write back every field an edit may change. id, device_id and created_at are never changed.
	ruleUpdateQuery := `
		UPDATE rules
		SET name = $3, enabled = $4, rule_type = $5, measurement = $6, comparison = $7,
		    limit_value = $8, lower_limit = $9, upper_limit = $10, deadband = $11, trip_state = $12,
		    updated_at = NOW()
		WHERE id = $1 AND device_id = $2
		RETURNING updated_at`

	// A rule that was disabled can no longer be alarming, so clear its active alarm
	clearAlarmQuery := `
		UPDATE alarms
		SET cleared_at = NOW(), clear_reason = $3
		WHERE device_id = $1 AND rule_id = $2 AND cleared_at IS NULL`

	var updatedRule model.Rule

	// BeginFunc commits if the function returns nil, and rolls back if it returns an error.
	// Errors returned from inside come back unchanged, so the sentinel errors reach the handler
	err := pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		// Check the device exists and is not deleted
		var one int
		err := tx.QueryRow(ctx, deviceLockQuery, deviceID).Scan(&one)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrDeviceNotFound
			}
			return err
		}

		// Read and lock the current rule, scanning straight into a model.Rule
		var rule model.Rule
		err = tx.QueryRow(ctx, ruleLockQuery, ruleID, deviceID).Scan(
			&rule.ID,
			&rule.DeviceID,
			&rule.Name,
			&rule.Enabled,
			&rule.Type,
			&rule.Measurement,
			&rule.Comparison,
			&rule.Limit,
			&rule.LowerLimit,
			&rule.UpperLimit,
			&rule.Deadband,
			&rule.TripState,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRuleNotFound
			}
			return err
		}

		// Run the caller's edit in Go. If it fails, return its error to roll back everything
		newRule, err := edit(rule)
		if err != nil {
			return err
		}

		// An edit must not change the rule's identity
		if newRule.ID != rule.ID || newRule.DeviceID != rule.DeviceID {
			return ErrDuringRuleUpdate
		}

		// Write the edited rule back
		err = tx.QueryRow(
			ctx,
			ruleUpdateQuery,
			ruleID,
			deviceID,
			newRule.Name,
			newRule.Enabled,
			newRule.Type,
			newRule.Measurement,
			newRule.Comparison,
			newRule.Limit,
			newRule.LowerLimit,
			newRule.UpperLimit,
			newRule.Deadband,
			newRule.TripState,
		).Scan(&newRule.UpdatedAt)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "23505": // Another rule on this device already has that name
					return ErrDuplicateRule
				case "23514": // The edited fields break a CHECK constraint (e.g. a threshold rule with no limit)
					return ErrDuringRuleUpdate
				}
			}
			return err
		}

		// Rule was disabled, clear its active alarm if present
		if !newRule.Enabled {
			_, err = tx.Exec(ctx, clearAlarmQuery, deviceID, ruleID, model.ClearReasonRuleDisabled)
			if err != nil {
				return err
			}
		}

		updatedRule = newRule
		return nil
	})
	if err != nil {
		return model.Rule{}, err
	}

	return updatedRule, nil
}

// DeleteRule deletes a rule. In one transaction it: refuses if the rule has an unacknowledged active
// alarm, clears its remaining active alarms with reason "ruleDeleted", and deletes the rule.
// The alarms table has ON DELETE SET NULL on rule_id, so alarm history stays and its rule_id becomes
// NULL (the alarm's rule_snapshot still records what the rule was).
func (db *DatabaseStruct) DeleteRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID) error {
	// Share lock on the device so a concurrent DeleteDevice waits. Soft-deleted devices are not found
	// so rules must be deletd before device.
	deviceLockQuery := `
		SELECT 1
		FROM devices
		WHERE id = $1 AND deleted_at IS NULL
		FOR SHARE`

	// Lock the rule row. Matching device_id too means another device's rule counts as not found.
	ruleLockQuery := `
		SELECT 1
		FROM rules
		WHERE id = $1 AND device_id = $2
		FOR UPDATE`

	// Active (not cleared) and not acknowledged blocks deletion
	unacknowledgedAlarmQuery := `
		SELECT EXISTS (
			SELECT 1
			FROM alarms a
			WHERE a.device_id = $1
			  AND a.rule_id = $2
			  AND a.cleared_at IS NULL
			  AND NOT EXISTS (SELECT 1 FROM acknowledgements k WHERE k.alarm_id = a.id)
		)`

	clearAlarmsQuery := `
		UPDATE alarms
		SET cleared_at = NOW(), clear_reason = $3
		WHERE device_id = $1 AND rule_id = $2 AND cleared_at IS NULL`

	deleteRuleQuery := `
		DELETE FROM rules
		WHERE id = $1 AND device_id = $2`

	return pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		// Check device exists and lock it. Use throwaway var.
		// QueryRow requires .Scan to read and error out, and .Scan requires a destination.
		// deviceLockQuery returns 1 if device exists, and ErrNoRows if it doesn't
		var one int
		err := tx.QueryRow(ctx, deviceLockQuery, deviceID).Scan(&one)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrDeviceNotFound
			}
			return err
		}

		// Check rule exists and lock it
		err = tx.QueryRow(ctx, ruleLockQuery, ruleID, deviceID).Scan(&one)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRuleNotFound
			}
			return err
		}

		// Check for unacknowledged alarms under rule and device
		var hasUnacknowledged bool
		err = tx.QueryRow(ctx, unacknowledgedAlarmQuery, deviceID, ruleID).Scan(&hasUnacknowledged)
		if err != nil {
			return err
		}
		if hasUnacknowledged {
			return ErrUnacknowledgedAlarm
		}

		// Remaining active alarms are acknowledged. Clear them, then delete the rule.
		_, err = tx.Exec(ctx, clearAlarmsQuery, deviceID, ruleID, model.ClearReasonRuleDeleted)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, deleteRuleQuery, ruleID, deviceID)
		return err
	})
}

// TODO: Add ListAllAlarms. Right now there is no way to list alarms from all devices regardless of device deleted_at
// 		 status, due to the DB schema + methods. This was not a problem before the DB because device list could just be iterated.
// 		 Here we will need a separate function with a joined call from devices and alarms
//		 without filtering out deleted_at devices.

// ListAlarms returns a device's alarms, newest first, each paired with its acknowledgement (nil if
// the alarm has not been acknowledged).
func (db *DatabaseStruct) ListAlarms(ctx context.Context, deviceID uuid.UUID) ([]model.AlarmWithAcknowledgement, error) {
	// Like the in-memory Store, only existence is checked here (soft-deleted devices still list alarms)
	deviceExistsQuery := `SELECT EXISTS (SELECT 1 FROM devices WHERE id = $1)`

	// Get alarms and their acknowledgements
	alarmListQuery := `
		SELECT a.id, a.rule_id, a.device_id, a.raised_at, a.cleared_at, a.clear_reason,
		       a.trigger_value, a.rule_snapshot, k.acknowledged_by, k.acknowledged_at
		FROM alarms a
		LEFT JOIN acknowledgements k ON k.alarm_id = a.id
		WHERE a.device_id = $1
		ORDER BY a.raised_at DESC, a.id`

	// Check existence of Device, deleted or not
	var exists bool
	if err := db.pool.QueryRow(ctx, deviceExistsQuery, deviceID).Scan(&exists); err != nil {
		return []model.AlarmWithAcknowledgement{}, err
	}
	if !exists {
		return []model.AlarmWithAcknowledgement{}, ErrDeviceNotFound
	}

	// Get alarms and their acknowledgements from device
	rows, err := db.pool.Query(ctx, alarmListQuery, deviceID)
	if err != nil {
		return []model.AlarmWithAcknowledgement{}, err
	}

	// Create list of Alarms with Acknowledgements
	dbAlarms, err := pgx.CollectRows(rows, pgx.RowToStructByName[DBAlarmWithAck])
	if err != nil {
		return []model.AlarmWithAcknowledgement{}, err
	}

	// Iterate list, use helper to build AlarmWithAcknowledgement object
	alarmList := make([]model.AlarmWithAcknowledgement, 0, len(dbAlarms))
	for _, dbAlarm := range dbAlarms {
		alarmWithAck, err := alarmWithAckFromRow(dbAlarm)
		if err != nil {
			return []model.AlarmWithAcknowledgement{}, err
		}
		alarmList = append(alarmList, alarmWithAck)
	}

	return alarmList, nil
}

// AcknowledgeAlarm records an acknowledgement for one of a device's alarms. Cleared alarms can still be
// acknowledged. A second acknowledgement of the same alarm fails (alarm_id is the primary key)
func (db *DatabaseStruct) AcknowledgeAlarm(ctx context.Context, deviceID uuid.UUID, alarmID uuid.UUID, ackBy string) (model.AlarmWithAcknowledgement, error) {
	// INSERT ... SELECT only inserts if the alarm exists AND belongs to this device, so no row back
	// means the alarm was not found. The ::text cast gives the parameter a type inside the SELECT list
	// Acknowledgement is thus only added if the alarm exists on the specified device
	ackInsertQuery := `
		INSERT INTO acknowledgements (alarm_id, acknowledged_by, acknowledged_at)
		SELECT a.id, $3::text, NOW()
		FROM alarms a
		WHERE a.id = $1 AND a.device_id = $2
		RETURNING acknowledged_at`

	// Read the alarm back with its new acknowledgement
	alarmQuery := `
		SELECT a.id, a.rule_id, a.device_id, a.raised_at, a.cleared_at, a.clear_reason,
		       a.trigger_value, a.rule_snapshot, k.acknowledged_by, k.acknowledged_at
		FROM alarms a
		LEFT JOIN acknowledgements k ON k.alarm_id = a.id
		WHERE a.id = $1 AND a.device_id = $2`

	var result model.AlarmWithAcknowledgement

	// Begin transaction
	err := pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		var acknowledgedAt time.Time
		// Insert acknowledgement if alarm exists on device
		err := tx.QueryRow(ctx, ackInsertQuery, alarmID, deviceID, ackBy).Scan(&acknowledgedAt)
		// Alarm was not found
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrAlarmNotFound
			}
			// Alarm was already acknowledged
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrAlarmAlreadyAcknowledged
			}
			return err
		}

		// Read alarm with new acknowledgement
		rows, err := tx.Query(ctx, alarmQuery, alarmID, deviceID)
		if err != nil {
			return err
		}
		// Collect the one expected row
		dbAlarm, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DBAlarmWithAck])
		if err != nil {
			return err
		}

		// Build AlarmWithAcknowledgement (union of both data types)
		result, err = alarmWithAckFromRow(dbAlarm)
		return err
	})
	if err != nil {
		return model.AlarmWithAcknowledgement{}, err
	}

	return result, nil
}

// AddReading stores a reading and, if it is newer than the device's latest one, evaluates the device's
// rules and raises or clears alarms. It all happens in one transaction, with the device row locked so
// readings for a device are processed one at a time (the in-memory Store gets this from its mutex).
// The raised and cleared alarms are returned so the handler can publish events.
// TODO: Investigate if we should process readings first, then add to the database afterwards/
//
//	Might be a scaling problem I don't need to deal with right now
func (db *DatabaseStruct) AddReading(ctx context.Context, reading model.Reading) (model.ReadingResult, error) {
	// FOR UPDATE: serializes readings per device, and makes DeleteDevice / CreateRule / UpdateRule
	// wait. Soft-deleted devices are not found.
	deviceLockQuery := `
		SELECT 1
		FROM devices
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE`

	// Newest reading before this one. Served by the (device_id, timestamp) primary key.
	latestReadingQuery := `SELECT MAX(timestamp) FROM readings WHERE device_id = $1`

	readingInsertQuery := `
		INSERT INTO readings (device_id, timestamp, current_a, current_b, current_c,
		                      voltage_a, voltage_b, voltage_c, frequency, real_power,
		                      reactive_power, breaker_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	// Like the in-memory Store, last seen is the time the server processed the reading
	lastSeenQuery := `UPDATE devices SET last_seen_at = NOW() WHERE id = $1`

	// Define empty result
	result := model.ReadingResult{
		Reading: reading,
		RuleEvaluationResult: model.RuleEvaluationResult{
			RaisedAlarms:  make([]model.Alarm, 0),
			ClearedAlarms: make([]model.Alarm, 0),
		},
	}

	// Begin transaction
	err := pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		var one int
		// Check device exists
		err := tx.QueryRow(ctx, deviceLockQuery, reading.DeviceID).Scan(&one)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrDeviceNotFound
			}
			return err
		}

		// Get the newest stored timestamp. MAX returns NULL (nil here) when the device has no readings yet
		var latest *time.Time
		if err := tx.QueryRow(ctx, latestReadingQuery, reading.DeviceID).Scan(&latest); err != nil {
			return err
		}

		// A reading is "newer" if it is the device's first, or its timestamp is after the latest stored one
		// The || short-circuits, so *latest is only dereferenced when latest is not nil
		isFirstReading := latest == nil
		isNewerReading := isFirstReading || reading.Timestamp.After(*latest)

		// Insert the reading. The primary key (device_id, timestamp) rejects duplicates.
		_, err = tx.Exec(ctx, readingInsertQuery,
			reading.DeviceID,
			reading.Timestamp,
			reading.CurrentA,
			reading.CurrentB,
			reading.CurrentC,
			reading.VoltageA,
			reading.VoltageB,
			reading.VoltageC,
			reading.Frequency,
			reading.RealPower,
			reading.ReactivePower,
			reading.BreakerStatus,
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrDuplicateReading
			}
			return err
		}

		// Get when device was last seen in readings
		if _, err := tx.Exec(ctx, lastSeenQuery, reading.DeviceID); err != nil {
			return err
		}

		// Only a newer reading can change alarm state. Late or out-of-order readings are just stored.
		if !isNewerReading {
			return nil
		}

		// Evaluate rules with helper
		return evaluateRules(ctx, tx, &result)
	})
	if err != nil {
		return model.ReadingResult{}, err
	}

	return result, nil
}

// evaluateRules raises or clears alarms for a reading, inside the caller's transaction. The caller must
// hold the device row lock. It fills result.RuleEvaluationResult in place. Same logic as the in-memory
// Store's evaluateRulesLocked, reusing readingMeasurement, ruleShouldTrip and ruleSnapshot
func evaluateRules(ctx context.Context, tx pgx.Tx, result *model.ReadingResult) error {
	deviceID := result.Reading.DeviceID

	// Get list of rules for device
	ruleListQuery := `
		SELECT id, device_id, name, enabled, rule_type, measurement, comparison,
		       limit_value, lower_limit, upper_limit, deadband, trip_state, created_at, updated_at
		FROM rules
		WHERE device_id = $1
		ORDER BY created_at, id`

	// The rules that currently have an active alarm
	activeRulesQuery := `
		SELECT rule_id
		FROM alarms
		WHERE device_id = $1 AND cleared_at IS NULL AND rule_id IS NOT NULL`

	alarmInsertQuery := `
		INSERT INTO alarms (id, rule_id, device_id, raised_at, trigger_value, rule_snapshot)
		VALUES ($1, $2, $3, $4, $5, $6)`

	// Alarm is cleared at the reading's timestamp, not the server's time
	alarmClearQuery := `
		UPDATE alarms
		SET cleared_at = $3, clear_reason = $4
		WHERE device_id = $1 AND rule_id = $2 AND cleared_at IS NULL
		RETURNING id, rule_id, device_id, raised_at, cleared_at, clear_reason, trigger_value, rule_snapshot`

	// Load the device's rules. CollectRows reads and closes the result before the next query.
	rows, err := tx.Query(ctx, ruleListQuery, deviceID)
	if err != nil {
		return err
	}
	dbRules, err := pgx.CollectRows(rows, pgx.RowToStructByName[DBRule])
	if err != nil {
		return err
	}

	// Load which rules already have an active alarm
	rows, err = tx.Query(ctx, activeRulesQuery, deviceID)
	if err != nil {
		return err
	}
	activeRuleIDs, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return err
	}
	activeRules := make(map[uuid.UUID]bool, len(activeRuleIDs))
	for _, id := range activeRuleIDs {
		activeRules[id] = true
	}

	// Iterate rules
	for _, dbRule := range dbRules {
		rule := ruleFromRow(dbRule)

		// Rule not enabled, skip
		if !rule.Enabled {
			continue
		}

		// Get the rule's measurement from the reading. Skip if there is none.
		value, ok := readingMeasurement(result.Reading, rule.Measurement)
		if !ok {
			continue
		}

		// Is rule active?
		active := activeRules[rule.ID]
		// Should rule be tripped (generate alarm)?
		shouldTrip := ruleShouldTrip(rule, value, active)

		if shouldTrip && !active {
			// Raise a new alarm, timestamped by the reading (NOT the server)
			ruleID := rule.ID
			alarm := model.Alarm{
				ID:           uuid.New(),
				RuleID:       &ruleID,
				DeviceID:     deviceID,
				RaisedAt:     result.Reading.Timestamp,
				TriggerValue: value,
				RuleSnapshot: ruleSnapshot(rule),
			}

			// JSONB parameters are encoded to JSON here. pgx would send a Go string as raw text,
			// which is not valid JSON, so a string trigger value (breaker "open") must be marshaled.
			// This is necessary because a trigger value may be a float or string.
			triggerJSON, err := json.Marshal(alarm.TriggerValue)
			if err != nil {
				return err
			}
			// Make rule snapshot JSON. This IS repeated data however it is guaranteed to always
			// correlate. Would be worse to try and map a rule ID. Also protects history by showing
			// what rule was when alarm was triggered.
			snapshotJSON, err := json.Marshal(alarm.RuleSnapshot)
			if err != nil {
				return err
			}

			// Insert alarm
			_, err = tx.Exec(ctx, alarmInsertQuery, alarm.ID, ruleID, deviceID, alarm.RaisedAt, triggerJSON, snapshotJSON)
			if err != nil {
				return err
			}
			// Append raised alarm
			result.RuleEvaluationResult.RaisedAlarms = append(result.RuleEvaluationResult.RaisedAlarms, alarm)

		} else if !shouldTrip && active {
			// Back within limits, clear the active alarm and return it for the event
			rows, err := tx.Query(ctx, alarmClearQuery, deviceID, rule.ID, result.Reading.Timestamp, model.ClearReasonNormal)
			if err != nil {
				return err
			}
			// Get the cleared alarm
			dbAlarm, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DBAlarm])
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return ErrAlarmNotFound
				}
				return err
			}
			// Created model.Alarm from DBAlarm
			clearedAlarm, err := alarmFromRow(dbAlarm)
			if err != nil {
				return err
			}
			// Append cleared alarm
			result.RuleEvaluationResult.ClearedAlarms = append(result.RuleEvaluationResult.ClearedAlarms, clearedAlarm)
		}
	}

	return nil
}

// ListReadings returns a device's readings between start and end (inclusive), newest first.
func (db *DatabaseStruct) ListReadings(ctx context.Context, deviceID uuid.UUID, start time.Time, end time.Time) ([]model.Reading, error) {
	// TIME VALIDATION IS IN THE API, like the in-memory Store
	// Soft-deleted devices are not found
	deviceExistsQuery := `SELECT EXISTS (SELECT 1 FROM devices WHERE id = $1 AND deleted_at IS NULL)`

	// The (device_id, timestamp) primary key serves the filter and the newest-first order
	readingListQuery := `
		SELECT device_id, timestamp, current_a, current_b, current_c,
		       voltage_a, voltage_b, voltage_c, frequency, real_power,
		       reactive_power, breaker_status
		FROM readings
		WHERE device_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp DESC`

	// Check device existence
	var exists bool
	if err := db.pool.QueryRow(ctx, deviceExistsQuery, deviceID).Scan(&exists); err != nil {
		return []model.Reading{}, err
	}
	if !exists {
		return []model.Reading{}, ErrDeviceNotFound
	}

	// Get reading rows within range
	rows, err := db.pool.Query(ctx, readingListQuery, deviceID, start, end)
	if err != nil {
		return []model.Reading{}, err
	}

	// Collect rows into []DBReading
	dbReadings, err := pgx.CollectRows(rows, pgx.RowToStructByName[DBReading])
	if err != nil {
		return []model.Reading{}, err
	}

	// Build readings list and return
	readingList := make([]model.Reading, 0, len(dbReadings))
	for _, r := range dbReadings {
		readingList = append(readingList, readingFromRow(r))
	}

	return readingList, nil
}

// deviceFromRow converts a device LEFT JOIN latest reading row into a model.Device. Status is derived
// from now. LatestReading stays nil when the device has no readings (the reading columns are NULL).
func deviceFromRow(r DBDeviceWithReading, now time.Time) model.Device {
	device := model.Device{
		ID:         r.ID,
		Name:       r.Name,
		IP:         r.IP,
		DeviceType: r.DeviceType, // Already validated in DB
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
		LastSeen:   r.LastSeenAt,
	}
	device.Status = currentDeviceStatus(device, now)

	// The reading columns are all NULL or all set, and the timestamp is the marker
	if r.Timestamp != nil {
		device.LatestReading = &model.Reading{
			DeviceID:      r.ID,
			Timestamp:     *r.Timestamp,
			CurrentA:      *r.CurrentA,
			CurrentB:      *r.CurrentB,
			CurrentC:      *r.CurrentC,
			VoltageA:      *r.VoltageA,
			VoltageB:      *r.VoltageB,
			VoltageC:      *r.VoltageC,
			Frequency:     *r.Frequency,
			RealPower:     *r.RealPower,
			ReactivePower: *r.ReactivePower,
			BreakerStatus: *r.BreakerStatus,
		}
	}

	return device
}

// readingFromRow converts a readings table row into a model.Reading
func readingFromRow(r DBReading) model.Reading {
	return model.Reading{
		DeviceID:      r.DeviceID,
		Timestamp:     r.Timestamp,
		CurrentA:      r.CurrentA,
		CurrentB:      r.CurrentB,
		CurrentC:      r.CurrentC,
		VoltageA:      r.VoltageA,
		VoltageB:      r.VoltageB,
		VoltageC:      r.VoltageC,
		Frequency:     r.Frequency,
		RealPower:     r.RealPower,
		ReactivePower: r.ReactivePower,
		BreakerStatus: r.BreakerStatus,
	}
}

// ruleFromRow converts a rules table row into a model.Rule
func ruleFromRow(r DBRule) model.Rule {
	return model.Rule{
		ID:          r.ID,
		DeviceID:    r.DeviceID,
		Name:        r.Name,
		Enabled:     r.Enabled,
		Type:        r.Type,
		Measurement: r.Measurement,
		Comparison:  r.Comparison,
		Limit:       r.Limit,
		LowerLimit:  r.LowerLimit,
		UpperLimit:  r.UpperLimit,
		Deadband:    r.Deadband,
		TripState:   r.TripState,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// alarmFromRow converts an alarms table row into a model.Alarm, decoding the two JSONB columns.
// trigger_value decodes to a float64 (number) or a string, depending on the rule type.
func alarmFromRow(r DBAlarm) (model.Alarm, error) {
	var triggerValue any
	if err := json.Unmarshal(r.TriggerValue, &triggerValue); err != nil {
		return model.Alarm{}, err
	}

	var snapshot model.RuleSnapshot
	if err := json.Unmarshal(r.RuleSnapshot, &snapshot); err != nil {
		return model.Alarm{}, err
	}

	return model.Alarm{
		ID:           r.ID,
		RuleID:       r.RuleID,
		DeviceID:     r.DeviceID,
		RaisedAt:     r.RaisedAt,
		ClearedAt:    r.ClearedAt,
		ClearReason:  r.ClearReason,
		TriggerValue: triggerValue,
		RuleSnapshot: snapshot,
	}, nil
}

// alarmWithAckFromRow converts an alarm LEFT JOIN acknowledgement row into a model.AlarmWithAcknowledgement.
// Acknowledgement stays nil when the alarm has not been acknowledged
func alarmWithAckFromRow(r DBAlarmWithAck) (model.AlarmWithAcknowledgement, error) {
	alarm, err := alarmFromRow(DBAlarm{
		ID:           r.ID,
		RuleID:       r.RuleID,
		DeviceID:     r.DeviceID,
		RaisedAt:     r.RaisedAt,
		ClearedAt:    r.ClearedAt,
		ClearReason:  r.ClearReason,
		TriggerValue: r.TriggerValue,
		RuleSnapshot: r.RuleSnapshot,
	})
	if err != nil {
		return model.AlarmWithAcknowledgement{}, err
	}

	// Always needs Alarm
	result := model.AlarmWithAcknowledgement{Alarm: alarm}

	// Conditionally add Acknowledgement
	if r.AcknowledgedBy != nil && r.AcknowledgedAt != nil {
		result.Acknowledgement = &model.Acknowledgement{
			AlarmID:        r.ID,
			AcknowledgedBy: *r.AcknowledgedBy,
			AcknowledgedAt: *r.AcknowledgedAt,
		}
	}

	return result, nil
}

// currentDeviceStatus checks a device's status and updates accordingly.
func currentDeviceStatus(device model.Device, now time.Time) model.DeviceStatus {
	// No need to update device.Status in Store, because status is derived data.
	// Also allows us to keep read lock with ListDevice and GetDevice, instead of having write lock.
	if device.LastSeen == nil {
		return model.DeviceStatusPending
	}
	if !now.Before(device.LastSeen.Add(time.Minute)) {
		return model.DeviceStatusStale
	}
	return model.DeviceStatusOnline
}
