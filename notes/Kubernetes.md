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