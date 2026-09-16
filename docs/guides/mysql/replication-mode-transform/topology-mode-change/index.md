---
title: MySQL Topology Mode Change
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-topology-mode-change
    name: Topology Mode Change
    parent: guides-mysql-mode-transform
    weight: 13
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

## MySQL Topology Mode Change

This guide shows how to change the **mode (topology)** of an existing MySQL database with a
`ReplicationModeTransformation` `MySQLOpsRequest` — promoting a **standalone** MySQL into a
clustered topology.

The target topology is selected with `spec.replicationModeTransformation.targetTopologyMode`:
`GroupReplication`, `InnoDBCluster` or `SemiSync`.

> Transforming a **Remote Replica** is covered separately in
> [Remote/Read Only Replica Mode Transfer](/docs/guides/mysql/replication-mode-transform/remote-replica-mode-transfer/index.md).

### Supported Mode Changes

The source database must be **`Standalone`** (no `spec.topology`) or a **`RemoteReplica`**.
Changing one clustered topology into another is not supported yet.

| From (source) | → `GroupReplication` | → `InnoDBCluster` | → `SemiSync` |
|---------------|:--------------------:|:-----------------:|:------------:|
| **Standalone** (no `spec.topology`) | ✅ | ✅ | ✅ |
| **RemoteReplica** | ✅ | ✅ | ✅ |
| **GroupReplication** | — | ❌ not supported yet | ❌ not supported yet |
| **InnoDBCluster** | ❌ not supported yet | — | ❌ not supported yet |
| **SemiSync** | ❌ not supported yet | ❌ not supported yet | — |

Key guarantees:

- **Your data is preserved.** A mode change never deletes a volume. When a new member has to be
  seeded it is seeded in place with MySQL's `CLONE INSTANCE`, which overwrites the data directory
  while the `PersistentVolumeClaim` is retained.
- A standalone database is scaled up to at least **3 members**, since a clustered topology needs a
  quorum.
- Requires MySQL **8.4.2 or newer**.

### Before You Begin

- You need a Kubernetes cluster with the KubeDB operator installed — see [here](/docs/setup/README.md).
- This tutorial uses the `demo` namespace:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Promote a Standalone MySQL

### Deploy a standalone MySQL

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: my-standalone
  namespace: demo
spec:
  version: "8.4.8"
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/replication-mode-transform/topology-mode-change/examples/my-standalone.yaml
mysql.kubedb.com/my-standalone created

$ kubectl get mysql -n demo my-standalone
NAME            VERSION   STATUS   AGE
my-standalone   8.4.8     Ready    2m
```

Insert some data so you can confirm it survives the mode change:

```bash
$ kubectl exec -it -n demo my-standalone-0 -c mysql -- mysql -uroot -p'pass' \
    -e "CREATE DATABASE playground; CREATE TABLE playground.t(id INT PRIMARY KEY); INSERT INTO playground.t VALUES(1),(2),(3);"
```

### Standalone → GroupReplication

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: promote-to-gr
  namespace: demo
spec:
  type: ReplicationModeTransformation
  databaseRef:
    name: my-standalone
  replicationModeTransformation:
    targetTopologyMode: GroupReplication
    mode: Single-Primary
  timeout: 15m
  apply: Always
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/replication-mode-transform/topology-mode-change/examples/promote-to-group-replication.yaml
mysqlopsrequest.ops.kubedb.com/promote-to-gr created

$ kubectl get mysqlopsrequest -n demo promote-to-gr
NAME            TYPE                            STATUS       AGE
promote-to-gr   ReplicationModeTransformation   Successful   3m
```

The standalone is scaled to a 3-member group and the pre-existing data is on every member:

```bash
$ kubectl exec -it -n demo my-standalone-0 -c mysql -- mysql -uroot -p'pass' \
    -e "SELECT MEMBER_HOST, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;"
+-----------------------------------------+--------------+-------------+
| MEMBER_HOST                             | MEMBER_STATE | MEMBER_ROLE |
+-----------------------------------------+--------------+-------------+
| my-standalone-0.my-standalone-pods.demo | ONLINE       | PRIMARY     |
| my-standalone-1.my-standalone-pods.demo | ONLINE       | SECONDARY   |
| my-standalone-2.my-standalone-pods.demo | ONLINE       | SECONDARY   |
+-----------------------------------------+--------------+-------------+
```

#### Multi-Primary (multi-master)

Set `mode: Multi-Primary` to get a multi-master group where **every member accepts writes**:

```yaml
  replicationModeTransformation:
    targetTopologyMode: GroupReplication
    mode: Multi-Primary
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/replication-mode-transform/topology-mode-change/examples/promote-to-multi-primary.yaml
mysqlopsrequest.ops.kubedb.com/promote-to-multi-primary created

$ kubectl exec -it -n demo my-standalone-0 -c mysql -- mysql -uroot -p'pass' \
    -e "SELECT MEMBER_HOST, MEMBER_ROLE FROM performance_schema.replication_group_members;"
+-----------------------------------------+-------------+
| MEMBER_HOST                             | MEMBER_ROLE |
+-----------------------------------------+-------------+
| my-standalone-0.my-standalone-pods.demo | PRIMARY     |
| my-standalone-1.my-standalone-pods.demo | PRIMARY     |
| my-standalone-2.my-standalone-pods.demo | PRIMARY     |
+-----------------------------------------+-------------+
```

All three members report `PRIMARY` and have `super_read_only=0`, so writes issued on any member are
accepted and replicated to the rest.

### Standalone → InnoDBCluster

```yaml
  replicationModeTransformation:
    targetTopologyMode: InnoDBCluster
    mode: Single-Primary
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/replication-mode-transform/topology-mode-change/examples/promote-to-innodb-cluster.yaml
mysqlopsrequest.ops.kubedb.com/promote-to-innodb created
```

In addition to the 3 database members, a **MySQL Router** is provisioned:

```bash
$ kubectl get pods -n demo | grep my-standalone
my-standalone-0          2/2     Running   0   4m
my-standalone-1          2/2     Running   0   4m
my-standalone-2          2/2     Running   0   4m
my-standalone-router-0   1/1     Running   0   4m
```

### Standalone → SemiSync

```yaml
  replicationModeTransformation:
    targetTopologyMode: SemiSync
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/replication-mode-transform/topology-mode-change/examples/promote-to-semisync.yaml
mysqlopsrequest.ops.kubedb.com/promote-to-semisync created
```

This produces a semi-synchronous **primary** with standby replicas. The `mode` field does not apply
here (SemiSync has no group), and the pod holding the existing data is elected as the primary:

```bash
$ kubectl get pods -n demo -L kubedb.com/role | grep my-standalone
my-standalone-0   2/2   Running   0   3m   primary
my-standalone-1   2/2   Running   0   3m   standby
my-standalone-2   2/2   Running   0   3m   standby
```

## Changing the Mode of an Existing Cluster

**Not supported yet.** A database that already runs a clustered topology
(`GroupReplication`, `InnoDBCluster` or `SemiSync`) cannot be transformed into a different one.
`ReplicationModeTransformation` applies to a **`Standalone`** or **`RemoteReplica`** source only.

To move an existing cluster to another topology, take a backup and restore it into a new database
provisioned with the topology you want.

## Verify

After any mode change, confirm the new topology and that your data is intact on every member:

```bash
$ kubectl get mysql -n demo my-standalone -o jsonpath='{.spec.topology.mode}{"\n"}'
GroupReplication

$ for i in 0 1 2; do
    kubectl exec -n demo my-standalone-$i -c mysql -- \
      mysql -uroot -p'pass' -N -e "SELECT COUNT(*) FROM playground.t;"
  done
3
3
3
```

## Cleaning up

```bash
kubectl delete -n demo my/my-standalone
kubectl delete -n demo myops/promote-to-gr myops/promote-to-multi-primary
kubectl delete -n demo myops/promote-to-innodb myops/promote-to-semisync
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLOpsRequest object](/docs/guides/mysql/concepts/opsrequest.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
