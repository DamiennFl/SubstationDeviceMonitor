# Data Model

Types used below: `uuid` (string), `string`, `number` (float), `bool`,
`timestamp` (RFC 3339 string in JSON; real time types in Go and Postgres).
JSON field names are camelCase. Fields marked "API only" are computed for
responses and not stored.

## Globals

### Device types
| Value   | Description             |
| ------- | ----------------------- |
| `relay` | Protective relay        |

### Measurements by device type

**relay**
| Name            | Kind   | Unit / values      |
| --------------- | ------ | ------------------ |
| `currentA`      | analog | A                  |
| `currentB`      | analog | A                  |
| `currentC`      | analog | A                  |
| `voltageA`      | analog | V                  |
| `voltageB`      | analog | V                  |
| `voltageC`      | analog | V                  |
| `frequency`     | analog | Hz                 |
| `realPower`     | analog | kW                 |
| `reactivePower` | analog | kvar               |
| `breakerStatus` | status | `open` or `closed` |

Analog measurements can use `threshold` and `band` rules.
Status measurements can only use `status` rules.

## Device
| Field          | Type                  | Notes |
| -------------- | --------------------- | ----- |
| `id`           | uuid                  | Never changes |
| `name`         | string                | Unique among non-deleted devices |
| `ip`           | string                | Valid IPv4; unique among non-deleted devices |
| `type`         | device type           | Cannot change after creation |
| `createdAt`    | timestamp             | Server time |
| `updatedAt`    | timestamp             | Server time |
| `deletedAt`    | timestamp or null     | Soft delete; hidden from all lists when set |
| `lastSeen`     | timestamp or null     | Server time the last reading was received; null if never |
| `status`       | `online` or `stale`   | API only: `stale` if no reading for 60 s or never reported |
| `latestReading`| Reading or null       | API only: newest reading by device timestamp |

## Reading
| Field           | Type      | Notes |
| --------------- | --------- | ----- |
| `deviceId`      | uuid      | |
| `timestamp`     | timestamp | Device time; must not be missing or more than 5 s in the future |
| `currentA`      | number    | A |
| `currentB`      | number    | A |
| `currentC`      | number    | A |
| `voltageA`      | number    | V |
| `voltageB`      | number    | V |
| `voltageC`      | number    | V |
| `frequency`     | number    | Hz |
| `realPower`     | number    | kW |
| `reactivePower` | number    | kvar |
| `breakerStatus` | `open` or `closed` | |

Key: (`deviceId`, `timestamp`) is unique. A duplicate is rejected with 409.

## Rule
| Field         | Type                           | Used by          | Notes |
| ------------- | ------------------------------ | ---------------- | ----- |
| `id`          | uuid                           | all              | |
| `deviceId`    | uuid                           | all              | |
| `name`        | string                         | all              | |
| `enabled`     | bool                           | all              | Disabled rules are not evaluated |
| `type`        | `threshold`, `band`, `status`  | all              | Cannot change after creation |
| `measurement` | measurement name               | all              | Must exist for the device's type and match the rule type's kind |
| `comparison`  | `GT` or `LT`                   | threshold        | |
| `limit`       | number                         | threshold        | |
| `lowerLimit`  | number                         | band             | |
| `upperLimit`  | number                         | band             | Must be greater than `lowerLimit` |
| `deadband`    | number                         | threshold, band  | 0 or more |
| `tripState`   | `open` or `closed`             | status           | Breaker state that trips the rule |
| `createdAt`   | timestamp                      | all              | |
| `updatedAt`   | timestamp                      | all              | |

Fields not used by a rule's type must be absent (null). Sending them is a 400.

### Trip and clear conditions
| Type              | Trips when                         | Clears when |
| ----------------- | ---------------------------------- | ----------- |
| threshold, `GT`   | value > limit                      | value < limit − deadband |
| threshold, `LT`   | value < limit                      | value > limit + deadband |
| band              | value < lowerLimit or > upperLimit | lowerLimit + deadband ≤ value ≤ upperLimit − deadband |
| status            | value = tripState                  | value ≠ tripState |

For a band, `lowerLimit + deadband` must stay below `upperLimit − deadband`.

## Alarm
An alarm is one occurrence of a rule tripping. It is never edited by users.
| Field          | Type               | Notes |
| -------------- | ------------------ | ----- |
| `id`           | uuid               | |
| `ruleId`       | uuid               | |
| `deviceId`     | uuid               | Copied from the rule for easier queries |
| `raisedAt`     | timestamp          | Device timestamp of the reading that tripped it |
| `clearedAt`    | timestamp or null  | Device timestamp of the reading that cleared it; null while active |
| `clearReason`  | string             | Reason for which the alarm was cleared: ruleUpdated, ruleDisabled, ruleDeleted |
| `triggerValue` | number or string   | The value that tripped it |
| `ruleSnapshot` | object             | Copy of the rule's `name`, `type`, `measurement`, `comparison`, `limit`, `lowerLimit`, `upperLimit`, `deadband`, `tripState` at the moment it tripped |
| `acknowledgement` | Acknowledgement or null | API only: joined from Acknowledgements |

At most one alarm per rule has `clearedAt = null` at any time.
A rule can only trip again after its active alarm clears.

## Acknowledgement
| Field            | Type      | Notes |
| ---------------- | --------- | ----- |
| `alarmId`        | uuid      | Also the key: one acknowledgement per alarm; a second is 409 |
| `acknowledgedBy` | string    | `admin` until logins exist |
| `acknowledgedAt` | timestamp | Server time |

## Live events (SSE, `/api/events`)
Each event has a type name and a JSON payload.
| Event                | Payload |
| -------------------- | ------- |
| `reading`            | Reading |
| `alarmRaised`        | Alarm |
| `alarmCleared`       | `id`, `ruleId`, `deviceId`, `clearedAt` |
| `alarmAcknowledged`  | Acknowledgement |

## Errors
Every error response has the body `{ "message": "<readable text>" }`.
Messages never include internal details (database errors, stack traces).