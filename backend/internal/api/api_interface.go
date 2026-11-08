package api

import (
	"context"
	"time"

	"github.com/DamiennFl/SubstationDeviceMonitor/backend/internal/model"
	"github.com/google/uuid"
)

type DataStore interface {
	ListDevices(ctx context.Context) ([]model.Device, error)
	GetDevice(ctx context.Context, deviceID uuid.UUID) (model.Device, error)
	CreateDevice(ctx context.Context, device model.Device) (model.Device, error)
	UpdateDevice(ctx context.Context, deviceID uuid.UUID, edit func(model.Device) (model.Device, error)) (model.Device, error)
	DeleteDevice(ctx context.Context, deviceID uuid.UUID) error
	ListRules(ctx context.Context, deviceID uuid.UUID) ([]model.Rule, error)
	GetRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID) (model.Rule, error)
	CreateRule(ctx context.Context, rule model.Rule) (model.Rule, error)
	UpdateRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID, edit func(rule model.Rule) (model.Rule, error)) (model.Rule, error)
	DeleteRule(ctx context.Context, deviceID uuid.UUID, ruleID uuid.UUID) error
	ListAlarms(ctx context.Context, deviceID uuid.UUID) ([]model.AlarmWithAcknowledgement, error)
	AcknowledgeAlarm(ctx context.Context, deviceID uuid.UUID, alarmID uuid.UUID, ackBy string) (model.AlarmWithAcknowledgement, error)
	AddReading(ctx context.Context, reading model.Reading) (model.ReadingResult, error)
	ListReadings(ctx context.Context, deviceID uuid.UUID, start time.Time, end time.Time) ([]model.Reading, error)
}
