### Docker Run
To run, simply clone the repo and run `docker compose up`. It will be available at http://localhost:3000.

### Kubernetes Run

Now can be run using Kubernetes pods locally:

#### Bash
```
// Make a new cluster using kind (Kubernetes in Docker)
kind create cluster

// Build docker images
docker build -t damienflutre/sdm-api:v1 -f Dockerfile.api .  // From inside backend/
docker build -t damienflutre/sdm-sim:v1 -f Dockerfile.simulator . // From inside backend/
docker build -t damienflutre/sdm-frontend:v1 -f Dockerfile .  // From inside frontend/substation-device-monitoring

// Load docker images into container
kind load docker-image damienflutre/sdm-api:v1
kind load docker-image damienflutre/sdm-sim:v1
kind load docker-image damienflutre/sdm-frontend:v1

// Launch pods for each piece
kubectl apply -f k8s/simulator.yaml
kubectl apply -f k8s/api.yaml
kubectl apply -f k8s/frontend.yaml

// Forward traffic from localhost:3000 to frontend-service
kubectl port-forward svc/frontend-service 3000

// App is now accessible from localhost:3000.
```