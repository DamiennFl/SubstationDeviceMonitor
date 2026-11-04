### Events Testing

Run ./api, and go run ./cmd/simulator
Run curl -v http://localhost:8080/api/events on new terminal. Will write events as they come through

Create device
curl -v -X POST http://localhost:8080/api/devices \
  -H "Content-Type: application/json" \
  -d '{"name":"Relay 1",
       "ip":"192.168.1.10",
       "type":"relay"
      }'

Get device ID from return and create rule
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

Now in events terminal, events will come through as they are generated, some alarms raised, some cleared, some normal events.