package store

import (
	"context"
	"errors"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
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
	return []model.Device{}, errNotImplemented
}

func (db *DatabaseStruct) GetDevice(ctx context.Context, deviceID uuid.UUID) (model.Device, error) {
	return model.Device{}, errNotImplemented
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
