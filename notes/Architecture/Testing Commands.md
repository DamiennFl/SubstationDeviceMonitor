### Create Device
curl -v -X POST http://localhost:8080/api/devices \
  -H "Content-Type: application/json" \
  -d '{"name":"Relay 1",
       "ip":"192.168.1.10",
       "type":"relay"
      }'

curl -v -X POST http://localhost:8080/api/devices/{device_id}/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Overvoltage",
    "type": "threshold",
    "measurement": "voltageA",
    "comparison": "GT",
    "limit": 240,
    "deadband": 5
  }'

  curl -v -X POST http://localhost:8080/api/devices/655da4eb-ceb6-4ebb-8598-058453d92c75/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Overvoltage",
    "type": "threshold",
    "measurement": "voltageA",
    "comparison": "GT",
    "limit": 240,
    "deadband": 5
  }'

Define a FROM and TO, then get readings for a device

FROM=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)
TO=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8080/api/devices/<device_id>/readings?from=$FROM&to=$TO"


### Postgres DB

Postgres DB Setup Commands

```
// Define goose commands in terminal to test DB
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="postgres://postgres:password@localhost:5432/dev_db?sslmode=disable"
GOOSE_MIGRATION_DIR=backend/migrations

goose status
goose up
goose down
goose reset


// Open SQL Shell in container
docker exec -it local-postgres psql -U postgres -d sdm
```

### Postgres SQL Shell Commands
```
\dt: Check tables
\d <table>: Show columns, constraints, indexes
\di: Lists indexes
\q: quit

INSERT INTO devices (id, name, ip, device_type)
VALUES(gen_random_uuid(), 'Relay 1', '127.0.0.1', 'relay');

SELECT * FROM devices;

INSERT INTO rules (id, device_id, name, enabled, rule_type, measurement, trip_state)
VALUES (gen_random_uuid(), '', 'Test', true, 'status', 'breakerStatus', 'closed');

SELECT * FROM rules;

INSERT INTO alarms (id, rule_id, device_id, trigger_value)
VALUES (gen_random_uuid(), '', '', '100');

SELECT * FROM alarms;

INSERT INTO acknowledgements (alarm_id, acknowledged_by, acknowledged_at)
VALUES ('', 'admin', '2026-11-07 22:06:00-08');

INSERT INTO readings (device_id, timestamp, current_a, current_b, current_c, voltage_a, voltage_b, voltage_c, frequency, real_power, reactive_power, breaker_status)
VALUES ('', '2026-11-07 22:06:00-08', 10, 11, 12, 230, 231, 232, 15, 12, 2, 'open');
```