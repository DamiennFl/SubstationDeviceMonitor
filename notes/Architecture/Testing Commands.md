If simulator and 

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

  curl -v -X POST http://localhost:8080/api/devices/063b91c5-0374-418b-9a2d-cea8345e618b/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Overvoltage",
    "type": "threshold",
    "measurement": "voltageA",
    "comparison": "GT",
    "limit": 240,
    "deadband": 5
  }'