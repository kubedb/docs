---
title: Expose With Cluster Announce
menu:
docs_{{ .version }}:
identifier: rd-update-announce
name: Update Announce
parent: rd-announce
weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Redis External Connections Outside Kubernetes using Redis Announce

Redis Announce is a feature in Redis that enables external connections to Redis replica sets deployed within Kubernetes. It allows applications or clients outside the Kubernetes cluster to connect to individual replica set members by mapping internal Kubernetes DNS names to externally accessible hostnames or IP addresses. This is useful for scenarios where external access is needed, such as hybrid deployments or connecting from outside the cluster.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: YAML files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prerequisites

We need to have the following prerequisites to run this tutorial:

### Install Voyager Gateway

Install voyager gateway using the following command:
```bash
helm install ace oci://ghcr.io/appscode-charts/voyager-gateway \
  --version v2025.6.30 \
  -n ace-gw --create-namespace \
  --set gateway-converter.enabled=false \
  --wait --burst-limit=10000 --debug
```

### Create EnvoyProxy and GatewayClass
We need to setup `EnvoyProxy` and `GatewayClass` to use voyager gateway.

Create `EnvoyProxy` using the following command:
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: ace
  namespace: ace-gw
spec:
  logging:
    level:
      default: warn
  mergeGateways: true
  provider:
    kubernetes:
      envoyDeployment:
        container:
          image: ghcr.io/voyagermesh/envoy:v1.34.1-ac
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
              - ALL
            privileged: false
            runAsNonRoot: true
            runAsUser: 65534
            seccompProfile:
              type: RuntimeDefault
        patch:
          value:
            spec:
              template:
                spec:
                  containers:
                  - name: shutdown-manager
                    securityContext:
                      allowPrivilegeEscalation: false
                      capabilities:
                        drop:
                        - ALL
                      privileged: false
                      runAsNonRoot: true
                      runAsUser: 65534
                      seccompProfile:
                        type: RuntimeDefault
      envoyService:
        externalTrafficPolicy: Cluster
        type: LoadBalancer
    type: Kubernetes
```


> If you want to use `NodePort` service. Update `.spec.provider.kubernetes.envoyService.type` to `NodePort` in the above YAML.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/announce/envoyproxy.yaml
envoyproxy.gateway.envoyproxy.io/ace created
```

Create `GatewayClass` using the following command:
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  annotations:
    catalog.appscode.com/gateway-config: |-
      service:
        externalTrafficPolicy: Cluster
        nodeportRange: 30000-32767
        portRange: 10000-12767
        seedBackendPort: 8080
        type: LoadBalancer
      vaultServer:
        name: vault
        namespace: ace
    catalog.appscode.com/is-default-gatewayclass: "true"
  name: ace
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  description: Default Service GatewayClass
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: ace
    namespace: ace-gw
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/announce/gatewayclass.yaml
gatewayclass.gateway.networking.k8s.io/ace created
```

Check the `GatewayClass` status `True`.
```bash
$ kubectl get gatewayclass 
NAME   CONTROLLER                                      ACCEPTED   AGE
ace    gateway.envoyproxy.io/gatewayclass-controller   True       16s
```

### Install `FluxCD` in your cluster
Install `FluxCD` in your cluster using the following command:
```bash
helm upgrade -i flux2 \
  oci://ghcr.io/appscode-charts/flux2 \
  --version 2.15.0 \
  --namespace flux-system --create-namespace \
  --wait --debug --burst-limit=10000
```

###  Install Keda

Install `Keda` in your cluster using the following command:
```bash
$ kubectl create ns kubeops
namespace/kubeops created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/announce/helmrepo.yaml
helmrepository.source.toolkit.fluxcd.io/appscode-charts-oci created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/announce/keda.yaml
helmrelease.helm.toolkit.fluxcd.io/keda created
helmrelease.helm.toolkit.fluxcd.io/keda-add-ons-http created
```

### Install `Catalog Manager`

Install `Catalog Manager` in your cluster using the following command:
```bash
helm install catalog-manager oci://ghcr.io/appscode-charts/catalog-manager \
  --version=v2025.6.30 \
  -n ace --create-namespace \
  --set helmrepo.name=appscode-charts-oci \
  --set helmrepo.namespace=kubeops \
  --wait --burst-limit=10000 --debug
```

## Overview

KubeDB uses following crd fields to enable Redis Announce:

- `spec:`
    - `announce:`
        - `type`
        - `shards`
            - `endpoints`

Read about the fields in details in [redis concept](/docs/guides/redis/concepts/redis.md)

## Redis Cluster with Announce

### Create DNS Records
Create dns `A`/`CNAME` records for redis cluster pods, let's say, `Redis` has `2` replicas and `3` shards.

Example:
- `DNS`: `kubedb.appscode`, this will be used to connect to the Redis replica set using `redis+srv`.
- `A/CNAME Record` for each Redis replicas with exposed Envoy Gateway `LoadBalancer/NodePort` IP/Host:
    - "rd0-0.kubedb.appscode"
    - "rd0-1.kubedb.appscode"
    - "rd1-0.kubedb.appscode"
    - "rd1-1.kubedb.appscode"
    - "rd2-0.kubedb.appscode"
    - "rd2-1.kubedb.appscode"


Below is the YAML for Redis Announce. 

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-announce
  namespace: demo
spec:
  version: 7.4.0
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
    announce:
      type: hostname
      shards:
        - endpoints:
            - "rd0-0.kubedb.appscode"
            - "rd0-1.kubedb.appscode"
        - endpoints:
            - "rd1-0.kubedb.appscode"
            - "rd1-1.kubedb.appscode"
        - endpoints:
            - "rd2-0.kubedb.appscode"
            - "rd2-1.kubedb.appscode"
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 20M
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

Here,
- `.spec.cluster.announce.type` specifies preferred dns type. It can be hostname or ip.
- `.spec.cluster.announce.shards` specifies the DNS names for each shards in the replica set.
- `.spec.cluster.announce.shards.endpoints`  specifies the DNS names for each pod in the specific shard.

>> **Note**: If you don't want to use `redis+srv` connection string, you can connect to the Redis replica set using the individual pod DNS names (e.g., `rd0-0.kubedb.appscode:10000`, `rd0-1.kubedb.appscode:10001`, etc.).

### Deploy Redis Cluster Announce

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/announce/redis.yaml
redis.kubedb.com/redis-announce created
```

Now, wait until `redis-announce` has status `Ready`. i.e,

```bash
$ watch kubectl get rd -n demo
Every 2.0s: kubectl get rd -n demo
NAME            VERSION   STATUS   AGE
redis-announce   7.4.0     Ready    6m56s
```


Now, create `RedisBinding` object to configure the whole process.
```yaml                                                                                                                           
apiVersion: catalog.appscode.com/v1alpha1                                                                                         
kind: RedisBinding                                                                                                              
metadata:                                                                                                                         
  name: redis-bind                                                                                                              
  namespace: demo                                                                                                                 
spec:                                                                                                                             
  sourceRef:                                                                                                                      
    name: redis-announce                                                                                                        
    namespace: demo                                                                                                               
```                                                                                                                               

```bash                                                                                                                           
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/announce/binding.yaml 
redisbinding.catalog.appscode.com/redis-bind created                                                                          
```                                                                                                                               
Now, check the status of `redisbinding` objects and ops requests.

```bash
$ kubectl get redisbinding,rdopsrequest -n demo
NAME                                               SRC_NS   SRC_NAME           STATUS   AGE
redisbinding.catalog.appscode.com/redis-bind       demo     redis-announce      Current  3m28s

NAME                                                       TYPE       STATUS       AGE
redisopsrequest.ops.kubedb.com/redis-announce-jddiql        Announce   Successful   2m58s
```

### Connect to Redis as Cluster

To connect to the Redis replica set, you can use the following command:

Collect the replicas from the `redis-announce` object:
```bash
$ kubectl get redis -n demo redis-announce -ojson | jq .spec.cluster.announce.shards[0]
{
  "endpoints": [
    "rd0-0.kubedb.appscode",
    "rd0-1.kubedb.appscode"
  ]
}

Connect with the database:

$ redis-cli -h rd0-0.kubedb.appscode -p 10050 -a <password> -c ping
PONG
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete redisbinding -n demo redis-bind
kubectl delete rd -n demo redis-announce

kubectl delete gatewayclass ace
kubectl delete -n ace-gw envoyproxy ace

helm uninstall -n ace catalog-manager
```


If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Backup and Restore](/docs/guides/redis/backup/kubestash/overview/index.md) Redis databases using KubeStash.
- Initialize [Redis with Script](/docs/guides/redis/initialization/using-script.md).
- Monitor your Redis database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Detail concepts of [redisVersion object](/docs/guides/redis/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).