package model

// DeviceType identifies a supported device kind.
type DeviceType string

const (
	DeviceTypeRelay DeviceType = "relay"
)

// UPDATE THIS WITH NEW DEVICE TYPES AS NECESSARY
func (d DeviceType) IsValid() bool {
	switch d {
	case DeviceTypeRelay:
		return true
	default:
		return false
	}
}

// DeviceStatus describes the current availability of a device.
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"  // Device online
	DeviceStatusPending DeviceStatus = "pending" // Device pending, likely newly created
	DeviceStatusStale   DeviceStatus = "stale"   // Device stale, haven't heard from it in 60 seconds
)

// RuleType identifies how a rule evaluates a measurement.
type RuleType string

const (
	RuleTypeThreshold RuleType = "threshold"
	RuleTypeBand      RuleType = "band"
	RuleTypeStatus    RuleType = "status"
)

// Comparison identifies a threshold comparison.
type Comparison string

const (
	ComparisonGT Comparison = "GT"
	ComparisonLT Comparison = "LT"
)

// State represents an open or closed breaker state.
type State string

const (
	StateOpen   State = "open"
	StateClosed State = "closed"
)

// ClearReason identifies why an alarm was cleared.
type ClearReason string

const (
	ClearReasonNormal        ClearReason = "normal"
	ClearReasonRuleDisabled  ClearReason = "ruleDisabled"
	ClearReasonRuleDeleted   ClearReason = "ruleDeleted"
	ClearReasonDeviceDeleted ClearReason = "deviceDeleted"
)

// Measurement identifies a supported device measurement.
type Measurement string

const (
	MeasurementCurrentA      Measurement = "currentA"
	MeasurementCurrentB      Measurement = "currentB"
	MeasurementCurrentC      Measurement = "currentC"
	MeasurementVoltageA      Measurement = "voltageA"
	MeasurementVoltageB      Measurement = "voltageB"
	MeasurementVoltageC      Measurement = "voltageC"
	MeasurementFrequency     Measurement = "frequency"
	MeasurementRealPower     Measurement = "realPower"
	MeasurementReactivePower Measurement = "reactivePower"
	MeasurementBreakerStatus Measurement = "breakerStatus"
)
