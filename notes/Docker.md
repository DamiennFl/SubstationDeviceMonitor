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
Build Frontend
```
docker build -t frontend frontend/
```

Run simulator container locally connected to simulator. API_BASE_URL is Docker hostname which resolves to host machine (my computer), as seen from within a container.
```
docker run --rm -e API_BASE_URL=http://host.docker.internal:8080 simulator
```
Run api while mapping container port to host port.
```
docker run --rm -p 8080:8080 api
```

#### Full flow with all three containers
```
// The three containers must share a Docker network. This lets Docker's built-in DNS reach each other by container name, so the frontend's nginx server can proxy to http://api:8080 and the simulator can call http://api:8080. Docker's default bridge network does not provide name-based DNS. The container's name must match the hostname the others use.

docker network create substation
docker run --network substation --name api api
docker run --network substation --name simulator simulator
docker run --network substation --name -p 3000:8080 frontend frontend

We map 3000:8080 so traffic on 3000 maps to nginx's listening port 8080.
```