// Types.tsx contains matching types from the Store backend.
// Input types match the backend input types, allowing us to create JSON objects while guaranteeing the correct format.
// Regular types are used for the opposite; decoding JSON sent from the backend into a type.

export const DeviceType = {
  Relay: 'relay',
} as const

// Use value union to define one type cleanly within other types while still matching enums.
export type DeviceType = typeof DeviceType[keyof typeof DeviceType]

export const RuleType = {
  Status: 'status',
  Threshold: 'threshold',
  Band: 'band',
} as const

export type RuleType = typeof RuleType[keyof typeof RuleType]

export const Measurement = {
  CurrentA: 'currentA',
  CurrentB: 'currentB',
  CurrentC: 'currentC',
  VoltageA: 'voltageA',
  VoltageB: 'voltageB',
  VoltageC: 'voltageC',
  Frequency: 'frequency',
  RealPower: 'realPower',
  ReactivePower: 'reactivePower',
  BreakerStatus: 'breakerStatus',
} as const

export type Measurement = typeof Measurement[keyof typeof Measurement]

export const Comparison = {
  GreaterThan: 'GT',
  LessThan: 'LT',
} as const

export type Comparison = typeof Comparison[keyof typeof Comparison]

export const State = {
  Open: 'open',
  Closed: 'closed',
} as const

export type State = typeof State[keyof typeof State]

export const DeviceStatus = {
  DeviceStatusOnline: "online",
	DeviceStatusPending: "pending",
	DeviceStatusStale: "stale"
} as const

export type DeviceStatus = typeof DeviceStatus[keyof typeof DeviceStatus]

// Device Input. Specifically match just the input type, not the full Device.
export type CreateDeviceInput =  {
    name: string
    ip: string
    type: DeviceType
}

export type UpdateDeviceInput = {
  name?: string
  ip?: string
}

export type Device = {
  deviceId: string
  name: string
  ip: string
  type: DeviceType
  status: DeviceStatus
  createdAt: string
  updatedAt: string
  lastSeen?: string
}

// Rule Creation Input.
export type CreateRuleInput = {
    name: string
    enabled?: boolean
    type: RuleType
    measurement: Measurement
    comparison?: Comparison
    limit?: number
    lowerLimit?: number
    upperLimit?: number
    deadband?: number
    tripState?: State
}

// Rule Update Input. All optional so we can form updates of all kinds.
export type UpdateRuleInput = {
    name?: string
    enabled?: boolean
    type?: RuleType
    measurement?: Measurement
    comparison?: Comparison
    limit?: number
    lowerLimit?: number
    upperLimit?: number
    deadband?: number
    tripState?: State
}