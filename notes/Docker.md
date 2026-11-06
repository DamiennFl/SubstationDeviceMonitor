### Docker Notes

### Docker commands

Run simulator container locally connected to simulator. API_BASE_URL is Docker hostname which resolves to host machine (my computer), as seen from within a container.
```
docker run --rm -e API_BASE_URL=http://host.docker.internal:8080 simulator
```