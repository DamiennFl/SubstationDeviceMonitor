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