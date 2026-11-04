If simulator and 

### Create Device
curl -v -X POST http://localhost:8080/api/devices \
  -H "Content-Type: application/json" \
  -d '{"name":"Relay 1","ip":"192.168.1.10","type":"relay"}'

