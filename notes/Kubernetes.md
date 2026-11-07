### Cluster

Group of nodes that run containers, plus control plane that decides what runs where

#### Control Plane

API Server: Front door, kubectl talks only to this. Every create, read and update goes through it
etcd: Database that stores cluster's desired and current state
Scheduler: Picks which node a new pod runs on
Controller Manager: Runs the loops that compare desired state with actual state and fix differences.

#### Worker Nodes

kubelet: Agent on each node. Starts and watches the containers the control plane assigned
Container runtime: Runs the containers
kube-proxy: Sets up networking that makes Services work

#### Notes

Kubernetes is configured declaratively.

Self healing: Crashed pods restart. If a node dies, pods are rescheduled elsewhere.
Service discovery: Service gives a group of pods a stable DNS name. Pods change but the name stays valid.


#### Objects

Pod: One or more containers that share a network and lifecycle
Deployment: Manages a set of identical pods. Handles restarts and updates
Service: Stable name and IP in front of a set of pods
ConfigMap/Secret: Config and secret values like env vars
Namespace: Virtual partition inside a cluster

#### Plug into current project

Docker images stay the same, cluster replaces Compose as the thing that runs them.

kind cluster (local) -> Deployment: api <- Service "api"
                     -> Deployment: simulator
                     -> Deployment: frontend <- Service (external)

#### API Deployment and Service

Service is deployed. General flow is as so:

Kubernetes maintains 1 replica (pod) that runs a container created from the docker image of the API. The docker image comes from building the image from the dockerfile. Docker image is an immutable read-only template that contains stuff like source code, runtime, system libraries, deps, and configs for the whole application. Think of it like the image describes how the container should be configured and ran, the pod takes that information and builds the container.
```
docker build -t damienflutre/sdm-api:v1 -f Dockerfile.api .  // From inside backend/

kind load docker-image damienflutre/sdm-api:v1

kubectl apply -f k8s/api.yaml // Starts a pod via api.yaml config file
```

#### Simulator Deployment
No service for simulator, because it is a client. It does not need an ingest port because it is not
listening.
```
docker build -t damienflutre/sdm-sim:v1 -f Dockerfile.simulator . // From inside backend/

kind load docker-image damienflutre/sdm-sim:v1

kubectl apply -f k8s/simulator.yaml
```

#### Frontend Deployment and Service
```
docker build -t damienflutre/sdm-frontend:v1 -f Dockerfile .  // From inside frontend/substation-device-monitoring

kind load docker-image damienflutre/sdm-frontend:v1

kubectl apply -f k8s/simulator.yaml
```