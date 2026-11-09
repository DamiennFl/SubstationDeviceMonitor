package store

import (
	"encoding/json"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

// DBDevice mirrors one row of the devices table, PLUS one row of the readings table,
// which is needed for writing the device's latest reading
type DBDeviceWithReading struct {
	ID            uuid.UUID        `db:"id"`
	Name          string           `db:"name"`
	IP            string           `db:"ip"`
	DeviceType    model.DeviceType `db:"device_type"`
	CreatedAt     time.Time        `db:"created_at"`
	UpdatedAt     time.Time        `db:"updated_at"`
	LastSeenAt    *time.Time       `db:"last_seen_at"`      // Nullable column: nil means the device has never been seen.
	Timestamp     *time.Time       `db:"reading_timestamp"` // Aliased tag for query
	CurrentA      *float64         `db:"current_a"`
	CurrentB      *float64         `db:"current_b"`
	CurrentC      *float64         `db:"current_c"`
	VoltageA      *float64         `db:"voltage_a"`
	VoltageB      *float64         `db:"voltage_b"`
	VoltageC      *float64         `db:"voltage_c"`
	Frequency     *float64         `db:"frequency"`
	RealPower     *float64         `db:"real_power"`
	ReactivePower *float64         `db:"reactive_power"`
	BreakerStatus *model.State     `db:"breaker_status"`
}

// DBRule mirrors one row of the rules table. It matches model.Rule
// exactly, however it defined db column values as struct tags instead
// of json.
type DBRule struct {
	ID          uuid.UUID         `db:"id"`          // Rule ID
	DeviceID    uuid.UUID         `db:"device_id"`   // Mapped Device ID
	Name        string            `db:"name"`        // Rule Name
	Enabled     bool              `db:"enabled"`     // Is Rule Enabled
	Type        model.RuleType    `db:"rule_type"`   // Type of Rule (threshold, band, status)
	Measurement model.Measurement `db:"measurement"` // Measurement Name from Device
	Comparison  *model.Comparison `db:"comparison"`  // Comparison (GT or LT)
	Limit       *float64          `db:"limit_value"` // Limit for threshold
	LowerLimit  *float64          `db:"lower_limit"` // Lower limit for band
	UpperLimit  *float64          `db:"upper_limit"` // Upper limit for band
	Deadband    *float64          `db:"deadband"`    // Deadband
	TripState   *model.State      `db:"trip_state"`  // Trip State (Open or Closed)
	CreatedAt   time.Time         `db:"created_at"`  // Creation Time and Date
	UpdatedAt   time.Time         `db:"updated_at"`  // Update Time and Date
}

// DBReading mirrors one row of the readings table. Every column is NOT NULL, so no pointers.
// Queries must select exactly these 12 columns
type DBReading struct {
	DeviceID      uuid.UUID   `db:"device_id"`
	Timestamp     time.Time   `db:"timestamp"`
	CurrentA      float64     `db:"current_a"`
	CurrentB      float64     `db:"current_b"`
	CurrentC      float64     `db:"current_c"`
	VoltageA      float64     `db:"voltage_a"`
	VoltageB      float64     `db:"voltage_b"`
	VoltageC      float64     `db:"voltage_c"`
	Frequency     float64     `db:"frequency"`
	RealPower     float64     `db:"real_power"`
	ReactivePower float64     `db:"reactive_power"`
	BreakerStatus model.State `db:"breaker_status"`
}

// DBAlarm mirrors one row of the alarms table. trigger_value and rule_snapshot are JSONB, scanned as
// raw JSON and decoded in Go, because trigger_value can be a number or a string
type DBAlarm struct {
	ID           uuid.UUID          `db:"id"`
	RuleID       *uuid.UUID         `db:"rule_id"` // NULL once the rule has been deleted
	DeviceID     uuid.UUID          `db:"device_id"`
	RaisedAt     time.Time          `db:"raised_at"`
	ClearedAt    *time.Time         `db:"cleared_at"`
	ClearReason  *model.ClearReason `db:"clear_reason"`
	TriggerValue json.RawMessage    `db:"trigger_value"`
	RuleSnapshot json.RawMessage    `db:"rule_snapshot"`
}

// DBAlarmWithAck is an alarm LEFT JOINed with its acknowledgement. The last two fields are NULL
// when the alarm has not been acknowledged
type DBAlarmWithAck struct {
	ID             uuid.UUID          `db:"id"`
	RuleID         *uuid.UUID         `db:"rule_id"`
	DeviceID       uuid.UUID          `db:"device_id"`
	RaisedAt       time.Time          `db:"raised_at"`
	ClearedAt      *time.Time         `db:"cleared_at"`
	ClearReason    *model.ClearReason `db:"clear_reason"`
	TriggerValue   json.RawMessage    `db:"trigger_value"`
	RuleSnapshot   json.RawMessage    `db:"rule_snapshot"`
	AcknowledgedBy *string            `db:"acknowledged_by"`
	AcknowledgedAt *time.Time         `db:"acknowledged_at"`
}
