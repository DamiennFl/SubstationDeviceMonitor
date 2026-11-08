package store

import (
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

// DBDevice mirrors one row of the devices table, PLUS one row of the readings table,
// which is needed for writing the device's latest reading.
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
