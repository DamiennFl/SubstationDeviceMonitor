### Docker Notes

### Simulator Dockerfile

Simulator dockerfile does this:
 - Build stage: Get golang image, set directory, copy go.mod for deps, download deps, copy source code, builds binary for just simulator.
 - Run stage: Create alpine docker image, define non-root user and user group, set directory, copy binary as non-root user, run ./main which will run simulator.

### API Dockerfile


### Docker commands
Build simulator
```
docker build -f backend/Dockerfile.simulator -t simulator backend/
```
Build API
```
docker build -f backend/Dockerfile.api -t api backend/
```

Run simulator container locally connected to simulator. API_BASE_URL is Docker hostname which resolves to host machine (my computer), as seen from within a container.
```
docker run --rm -e API_BASE_URL=http://host.docker.internal:8080 simulator
```