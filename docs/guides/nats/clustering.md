---
title: NATS Clustering and Configuration
menu:
  docs_{{ .version }}:
    identifier: nats-clustering
    name: Clustering
    parent: nats-guides
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Clustering and Configuration

A `NATS` object is one cluster. The servers form a full mesh over routes on port 6222. A server needs only one seed route because the servers gossip, but KubeDB renders routes to every current pod, so a seed that is down at boot does not matter. Routes are authenticated with a generated credential stored in `<name>-cluster-auth`.

> Note: The yaml files used in this tutorial are stored in [docs/examples/nats/cluster](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/nats/cluster) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not have one, create it with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Install KubeDB following the steps [here](/docs/setup/README.md), with `--set global.featureGates.NATS=true` so the NATS CRDs are installed.
- This tutorial uses the `demo` namespace:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Three-server cluster with JetStream

```yaml
apiVersion: kubedb.com/v1alpha2
kind: NATS
metadata:
  name: nats-cluster
  namespace: demo
spec:
  version: "2.15.1"
  replicas: 3
  jetStream:
    storageType: Durable
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 10Gi
      storageClassName: standard
    memoryStoreSize: 256Mi
  serverConfig:
    maxConnections: 10000
    maxPayload: 1048576
  lameDuck:
    gracePeriod: 10s
    duration: 30s
  podTemplate:
    spec:
      containers:
        - name: nats
          resources:
            requests:
              cpu: "500m"
              memory: 1Gi
            limits:
              cpu: "1"
              memory: 1Gi
  serviceTemplates:
    - alias: primary
      spec:
        type: LoadBalancer
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/nats/cluster/nats-cluster.yaml
nats.kubedb.com/nats-cluster created
```

- `jetStream.memoryStoreSize` caps the JetStream memory store. If unset, memory streams are disabled. It must be at most 50% of the container memory limit.
- `jetStream.storage` sets the file store PVC. KubeDB writes `max_file_store` to the config explicitly so that [volume expansion](/docs/guides/nats/ops-requests.md#volumeexpansion) can raise it without a restart.
- `serverConfig` holds typed limits: `maxConnections`, `maxPayload`, `maxPending`, `maxSubscriptions`, `maxControlLine`, `writeDeadline`, `authTimeout`, `debug`, `trace`. `maxPayload` must not exceed `maxPending`.
- `lameDuck` sets how long a pod keeps serving while it asks clients to move before it stops. `duration` must be at least `30s` (the server minimum) and `gracePeriod` must be smaller than `duration`. Changing it rolls the pods, since it also changes `terminationGracePeriodSeconds`.

## Core NATS without JetStream

```yaml
apiVersion: kubedb.com/v1alpha2
kind: NATS
metadata:
  name: nats-core
  namespace: demo
spec:
  version: "2.15.1"
  replicas: 2
  jetStream:
    enabled: false
  deletionPolicy: WipeOut
```

No PVC is created and any replica count `>= 1` is valid. `jetStream.enabled` cannot be changed after creation.

## Custom configuration

Ordinary server settings go in a Secret referenced by `spec.configuration.secretName`. Two keys are supported:

- `nats.conf`: extra top-level server settings.
- `accounts.conf`: extra accounts, in `Config` auth mode only. The `SYS` and `APP` account names are reserved.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: nats-custom-config
  namespace: demo
stringData:
  nats.conf: |
    ping_interval: "30s"
    max_connections: 5000
  accounts.conf: |
    ORDERS {
      jetstream: enabled
      users: [{user: orders, password: orders-secret}]
    }
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/nats/cluster/custom-config.yaml
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/nats/cluster/nats-custom-config.yaml
```

KubeDB owns `server_name`, `listen`, `jetstream`, `cluster`, `accounts` and `authorization`; the webhook rejects them in `nats.conf`. Most settings reload in place when changed through a [Reconfigure](/docs/guides/nats/ops-requests.md#reconfigure) ops request.

## Resources KubeDB creates

```bash
$ kubectl get svc -n demo -l app.kubernetes.io/instance=nats-cluster
NAME                   TYPE           PORT(S)
nats-cluster           LoadBalancer   4222/TCP
nats-cluster-pods      ClusterIP      6222/TCP,8222/TCP
nats-cluster-stats     ClusterIP      7777/TCP
```

`spec.serviceTemplates` accepts the aliases `primary`, `stats` and `leafnodes`.

## Cleanup

```bash
$ kubectl delete -n demo nats/nats-cluster nats/nats-core nats/nats-custom
```
