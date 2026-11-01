Use Go for concurrency capabilities, channels for passing info. Maps directly to what this project does.

Two important links in the system:

Simulator -> API Server: How devices report readings
API Server -> Browser: How the readings are updated and displayed, live updates

### Components List
 - React + Typescript frontend, served by nginx
 - Go API and backend
 - Go simulator
 - Postgres database
 - Ingress routing

### API Endpoints

Devices
| Type | Endpoint | Purpose | Success | Errors |
| ---- | -------- | ------- | ------- | ------ |
| GET    | `/api/devices` | Get the list of all devices | 200 Success | 503 Service Unavailable |
| POST   | `/api/devices` | Create new device | 201 Created | 400 Bad Setting |
| GET    | `/api/devices/{device_id}` | Get specific device | 200 | 404 Unknown ID |
| PATCH  | `/api/devices/{device_id}` | Update specific device | 200 | 400 Bad Setting, 404 Unknown ID |
| DELETE | `/api/devices/{device_id}` | Delete specific device | 204 No Content | 404 Unknown ID |

Rules
| Type | Endpoint | Purpose | Success | Errors |
| ---- | -------- | ------- | ------- | ------ |
| GET    | `/api/devices/{device_id}/rules` | Rules for a device | 200 Success | 404 Unknown ID (Device) |
| GET    | `/api/devices/{device_id}/rules/rule` | Specific rule | 200 Success | 404 Unknown ID (Device or Rule) |
| PATCH  | `/api/devices/{device_id}/rules/rule` | Modify rule  | 200 Success | 404 Unknown ID, 400 Bad Setting |
| DELETE | `/api/devices/{device_id}/rules/rule` | Delete rule  | 204 Success | 404 Unknown ID, 400 Bad Setting, 409 Conflict Error |
| POST   | `/api/devices/{device_id}/rules/new_rule` | New rule  | 200 Success | 404 Unknown ID, 400 Bad Setting |

Readings
| Type | Endpoint | Purpose | Success | Errors |
| ---- | -------- | ------- | ------- | ------ |
| POST | `/api/devices/{device_id}/readings` | Report latest reading, called by Simulator | 204 Success, simulator needs nothing back | 400 Bad Body, 404 Unknown ID |
| GET  | `/api/devices/{device_id}/readings?from=...&to=...` | Reports readings from x to y | 200 Success | 400 Bad Range, 404 Unknown ID |

Alarms
| Type | Endpoint | Purpose | Success | Errors |
| ---- | -------- | ------- | ------- | ------ |
| GET  | `/api/devices/{device_id}/alarms` | Alarm list on device page | 200 Success | 404 Unknown ID |
| POST | `/api/devices/{device_id}/{alarm_id}/acknowledge` | Acknowledge Alarm | 200 Success | 404 Unknown ID (Device or Alarm)

Live Updates
| Type | Endpoint | Purpose |
| ---- | -------- | ------- |
| GET | `/api/events` | SSE Stream that stays open. Reading and alarm updates come through here |

### Server side rules
 - When a reading arrives, server updates that device's latest values, checks alarm rules, raises or clears alarms, and pushes SSE events. All side effects happen on server in one place.
 - Only server can clear alarm based on conditions and state of relay, never user, the user can only acknowledge.
 - Device is the source of truth. API only shows what devices report.
 - At most one active alarm per rule at a time, and a new one only after the old one clears and the rule trips again.


### Notes and Design Decisions
 - Health endpoints for Kubernetes service (liveliness and readiness). Readiness will check for database, since if it is down no reads or writes can occur.
 - Simulator will poll API about device changes (settings, new devices, deletions, etc.) Pull model instead of API pushing.
 - Each device will have rules, and rules will generate alarms based on what occurs in the events if necessary.
 - Acknowledgements will record who acknowledged and when via POST (requires some kind of login, for now all acknowledgements will be done by `admin`).
 - Relays will hold: Per phase currents and voltages (A, B, C), frequency, real and reactive power, and breaker status (open or closed). Inputs and measurements matching rules. Rules will have configurable deadband.
 - One API replica is the limit now, since I am using SSE and no shared event bus. We cannot have two pods running the services simultaneously without a shared event bus, a Redis pub/sub service, or Postgres LISTEN/NOTIFY.
 - "Last seen" for each device, so once stale requests are detected the device can be labelled as so.
 - Simulator is down: Reads keep working in case it comes back, devices marked as stale and banner notifying user that the device the dashboard is connected to is disconnected.
 - Database down: 503 on all services, different banner saying database is no longer accessible (look more into this).
 - Offline detection after 30 seconds for devices, 60 seconds for the simulator. 
 - GET devices includes each devices' latest values.
 - Error shape: JSON body with readable message, no internal details.
 - Event timestamps come from devices, to ensure proper logging of events and alarms.
 - When a rule is modified, disabled, or deleted, any alarms currently triggered by those rules will clear with a message stating which occurred. If it cleared due to regular conditions, e.g. the relay reported normal values, no clear reason message will be displayed.


 - Deploy order matters if data models are updated in the package, since the API will have strict input checking. Update API first then simulator.
 - When using pointers for input, a PATCH will never be able to clear a value like removing a deadband, since "not sent" and "send as null" both arrive as nil.
 - Rules are stored as a list per device. It is slower to find a single rule, since we have to iterate through the list, however editing is rarer than fetching the list of rules, which we need when readings come through to potentially trigger alarms.
 - Store contains a clock to keep devices, rules, and alarms synchronized.

### Future improvements:
 - More SSE events (device created or deleted, alarm acknowledged)
 - SSE Reconnect with resync
 - Device commands like resetting relays
 - MQTT Broker for readings instead of simple HTTP POST
 - Dedicated alarm panel for seeing all alarms
 - Dedicated logging for internal errors, accessible only by admin