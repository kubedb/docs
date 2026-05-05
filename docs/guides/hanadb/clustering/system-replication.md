---
title: HanaDB System Replication
menu:
  docs_{{ .version }}:
    identifier: hanadb-system-replication-clustering
    name: System Replication
    parent: hanadb-clustering
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDB System Replication

This guide shows how to run SAP HANA system replication using KubeDB. In this mode, KubeDB creates multiple HanaDB pods and wires SAP HANA system replication between them. The primary service routes write traffic to the current primary pod, and the optional secondary service can expose readable secondary pods when read access is enabled.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl`.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- Create a namespace for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/clustering](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/clustering).

## Deploy a System Replication Cluster

The following manifest creates a three-node HanaDB system replication cluster. `spec.topology.mode: SystemReplication` enables SAP HANA system replication.

The important fields are:

| Field | Purpose |
|-------|---------|
| `spec.replicas` | Number of HanaDB data pods. This guide uses `3` replicas. |
| `spec.topology.mode` | Must be `SystemReplication` for SAP HANA system replication. |
| `spec.topology.systemReplication.replicationMode` | Controls transaction commit behavior relative to log shipping. Valid values are `sync`, `syncmem`, `async`, and `fullsync`. If omitted, KubeDB defaults it to `sync`. |
| `spec.topology.systemReplication.operationMode` | Controls how the secondary replays logs. Valid values are `logreplay`, `delta_datashipping`, and `logreplay_readaccess`. If omitted, KubeDB defaults it to `logreplay`. |
| `spec.storage.resources.requests.storage` | Persistent volume size for each HanaDB pod. SAP HANA needs a large volume; the examples use `64Gi`. |

This guide uses `replicationMode: fullsync` and `operationMode: logreplay_readaccess`. With `logreplay_readaccess`, KubeDB creates a secondary service for read-enabled secondaries.

## Choose Replication and Operation Mode

System replication behavior is controlled by two fields under `spec.topology.systemReplication`.

### Replication Mode

`replicationMode` controls how strongly the primary waits for log shipping before committing transactions.

| Mode | Use when |
|------|----------|
| `sync` | You want synchronous replication with the default KubeDB behavior. This is the default when `replicationMode` is omitted. |
| `syncmem` | You want synchronous replication where the secondary acknowledges after receiving logs in memory. |
| `async` | You prefer lower write latency and can tolerate more replication lag. |
| `fullsync` | You want the strictest acknowledgement behavior supported by SAP HANA system replication. |

### Operation Mode

`operationMode` controls how the secondary receives and replays data.
For every system replication cluster, KubeDB creates the primary service and governing headless service. The operation mode determines whether KubeDB also creates a secondary read service.

| Mode | Read access | Secondary service |
|------|-------------|-------------------|
| `logreplay` | Disabled | Not created. This is the default when `operationMode` is omitted. |
| `delta_datashipping` | Disabled | Not created. |
| `logreplay_readaccess` | Enabled | Created as `secondary-<hanadb-name>`. |

Read access is enabled only by setting:

```yaml
topology:
  mode: SystemReplication
  systemReplication:
    operationMode: logreplay_readaccess
```

If you do not need read-only traffic on secondary pods, use `operationMode: logreplay` or omit `spec.topology.systemReplication.operationMode`.

For example, this uses default `sync` + `logreplay` behavior and does not create the secondary read service:

```yaml
topology:
  mode: SystemReplication
```

The following example enables strict replication acknowledgment and read access on secondaries:

```yaml
topology:
  mode: SystemReplication
  systemReplication:
    replicationMode: fullsync
    operationMode: logreplay_readaccess
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 3
  storageType: Durable
  topology:
    mode: SystemReplication
    systemReplication:
      replicationMode: fullsync
      operationMode: logreplay_readaccess
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

Create the database:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/clustering/system-replication.yaml
hanadb.kubedb.com/hanadb-cluster created
```

Wait for the cluster to become ready:

```bash
$ kubectl get hanadb -n demo hanadb-cluster
NAME             VERSION   STATUS   AGE
hanadb-cluster   2.0.82    Ready    8m
```

## Verify System Replication Resources

Check the pods. Each HanaDB pod should have the database container and the coordinator sidecar running.

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=hanadb-cluster"
NAME               READY   STATUS    RESTARTS   AGE
hanadb-cluster-0   2/2     Running   0          8m
hanadb-cluster-1   2/2     Running   0          8m
hanadb-cluster-2   2/2     Running   0          8m
```

Check the services created for the cluster:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=hanadb-cluster"
NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)              AGE
hanadb-cluster             ClusterIP   10.96.100.10    <none>        39017/TCP            8m
secondary-hanadb-cluster   ClusterIP   10.96.100.11    <none>        39017/TCP            8m
hanadb-cluster-pods        ClusterIP   None            <none>        39001/TCP,39017/TCP  8m
```

The services have the following roles:

| Service | Use |
|---------|-----|
| `hanadb-cluster` | Primary service. Applications should use this service for write traffic. |
| `secondary-hanadb-cluster` | Secondary service. Created when `operationMode` is `logreplay_readaccess`; applications can use it for read traffic. |
| `hanadb-cluster-pods` | Headless governing service used for pod DNS and internal cluster coordination. |

If you use `operationMode: logreplay` or `operationMode: delta_datashipping`, the `secondary-hanadb-cluster` service is not created because secondary read access is disabled.

You can also inspect the endpoints to see which pods each service currently selects:

```bash
$ kubectl get endpoints -n demo hanadb-cluster secondary-hanadb-cluster hanadb-cluster-pods
NAME                       ENDPOINTS                                      AGE
hanadb-cluster             10.244.0.12:39017                              8m
secondary-hanadb-cluster   10.244.0.13:39017,10.244.0.14:39017            8m
hanadb-cluster-pods        10.244.0.12:39017,10.244.0.13:39017 + 4 more   8m
```

## Verify Replication Status

KubeDB checks SAP HANA system replication by querying `SYS.M_SERVICE_REPLICATION` from the primary service. You can run the same check manually.

First, read the generated SYSTEM user password:

```bash
$ export HANA_PASSWORD=$(kubectl get secret -n demo hanadb-cluster-auth -o jsonpath='{.data.password}' | base64 -d)
```

Then query the replication status from the primary pod:

```bash
$ kubectl exec -it -n demo hanadb-cluster-0 -c hanadb -- hdbsql \
  -u SYSTEM -p "$HANA_PASSWORD" -d SYSTEMDB \
  "SELECT REPLICATION_STATUS, REPLICATION_STATUS_DETAILS, (LAST_LOG_POSITION - REPLAYED_LOG_POSITION) AS REPLAY_BACKLOG FROM SYS.M_SERVICE_REPLICATION"
REPLICATION_STATUS   REPLICATION_STATUS_DETAILS   REPLAY_BACKLOG
ACTIVE               Connected                     0
ACTIVE               Connected                     0
```

The cluster is healthy when every secondary reports `ACTIVE` and the replay backlog is `0` or close to `0`. During startup, the status may temporarily show `SYNCING` or `INITIALIZING`.

## Connect to HanaDB

Applications should connect to the primary service for writes:

```bash
hanadb-cluster.demo.svc:39017
```

If you selected `operationMode: logreplay_readaccess`, read-only clients can connect to the secondary service:

```bash
secondary-hanadb-cluster.demo.svc:39017
```

## Cleaning up

```bash
kubectl delete hanadb -n demo hanadb-cluster
kubectl delete ns demo
```
