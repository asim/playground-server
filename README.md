# Playground Server

The playground server is an experimental system for building and deploying microservices as docker containers on kubernetes. 

## Prerequisites 

Requires Kubernetes

### Run Kubernetes locally with Vagrant

```
git clone https://github.com/GoogleCloudPlatform/kubernetes
cd kubernetes
export KUBERNETES_PROVIDER=vagrant
cluster/kube-up.sh
```

## Getting Started

Start dependencies

### Start Redis

```
cluster/kubectl.sh create -f github.com/myodc/playground/playground-redis.json
cluster/kubectl.sh create -f github.com/myodc/playground/playground-redis-service.json
```

### Start Registry

```
cluster/kubectl.sh create -f github.com/myodc/playground/playground-registry.json
cluster/kubectl.sh create -f github.com/myodc/playground/playground-registry-service.json
```

### Start Server

Set PLAYGROUND_KUBE_HOST, PLAYGROUND_KUBE_USER and PLAYGROUND_KUBE_PASS to the kubernetes master api in playground-server.json

```
cluster/kubectl.sh create -f playground-server.json
cluster/kubectl.sh create -f playground-server-service.json
```

Browse to http://HOST_MACHINE:8081 or if using the [playground-proxy](https://github.com/myodc/playground-proxy), http://playground-server.PROXY_DOMAIN
