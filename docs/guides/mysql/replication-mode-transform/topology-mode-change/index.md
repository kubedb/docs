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
`ReplicationModeTransformation` `MySQLOpsRequest` — promoting a **standalone** MySQL into a cluster,
and switching an existing cluster between clustered topologies.

The target topology is selected with `spec.replicationModeTransformation.targetMode`:
`GroupReplication`, `InnoDBCluster` or `SemiSync`.

> Transforming a **Remote Replica** is covered separately in
> [Remote/Read Only Replica Mode Transfer](/docs/guides/mysql/replication-mode-transform/remote-replica-mode-transfer/index.md).

### Supported Mode Changes

| From (source) | → `GroupReplication` | → `InnoDBCluster` | → `SemiSync` |
|---------------|:--------------------:|:-----------------:|:------------:|
| **Standalone** (no `spec.topology`) | ✅ | ✅ | ✅ |
| **GroupReplication** | — | ✅ | ✅ |
| **InnoDBCluster** | ✅ | — | ❌ not supported yet |
| **SemiSync** | ❌ not supported yet | ❌ not supported yet | — |

Key guarantees:

- **Your data is preserved.** A mode change never deletes a volume. When a new member has to be
  seeded it is seeded in place with MySQL's `CLONE INSTANCE`, which overwrites the data directory
  while the `PersistentVolumeClaim` is retained.
- **Cluster → cluster happens in place.** `GroupReplication` ⇄ `InnoDBCluster` keeps the running
  group and only hands over management — no teardown, no re-clone.
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
    targetMode: GroupReplication
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
    targetMode: GroupReplication
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
    targetMode: InnoDBCluster
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
    targetMode: SemiSync
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

## Change the Mode of an Existing Cluster

Cluster-to-cluster changes are performed **in place**: the running group is kept and only its
management changes, so there is no teardown, no re-clone and no data movement.

Deploy a 3-member Group Replication cluster to work with:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: my-cluster
  namespace: demo
spec:
  version: "8.4.8"
  replicas: 3
  topology:
    mode: GroupReplication
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/replication-mode-transform/topology-mode-change/examples/my-cluster.yaml
mysql.kubedb.com/my-cluster created
```

### GroupReplication → InnoDBCluster

The live Group Replication group is **adopted** into an InnoDB Cluster and a MySQL Router is added.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: gr-to-innodb
  namespace: demo
spec:
  type: ReplicationModeTransformation
  databaseRef:
    name: my-cluster
  replicationModeTransformation:
    targetMode: InnoDBCluster
    mode: Single-Primary
  timeout: 20m
  apply: Always
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/replication-mode-transform/topology-mode-change/examples/gr-to-innodb-cluster.yaml
mysqlopsrequest.ops.kubedb.com/gr-to-innodb created

$ kubectl get mysqlopsrequest -n demo gr-to-innodb
NAME           TYPE                            STATUS       AGE
gr-to-innodb   ReplicationModeTransformation   Successful   1m
```

Because the group is adopted rather than rebuilt, the members never restart and the change completes
in well under a minute.

### InnoDBCluster → GroupReplication

The InnoDB Cluster management and its Router are removed, and the same group continues as plain
Group Replication.

```yaml
  replicationModeTransformation:
    targetMode: GroupReplication
    mode: Single-Primary
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/replication-mode-transform/topology-mode-change/examples/innodb-cluster-to-gr.yaml
mysqlopsrequest.ops.kubedb.com/innodb-to-gr created
```

> **Note:** removing the Router requires the KubeDB ops-manager ServiceAccount to be able to delete
> `apps/deployments`. If that permission is missing the mode change still succeeds and the Router is
> simply left behind, to be removed manually.

### GroupReplication → SemiSync

```yaml
  replicationModeTransformation:
    targetMode: SemiSync
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/replication-mode-transform/topology-mode-change/examples/gr-to-semisync.yaml
mysqlopsrequest.ops.kubedb.com/gr-to-semisync created
```

### Not supported yet

`SemiSync` → `GroupReplication`/`InnoDBCluster` and `InnoDBCluster` → `SemiSync` are not supported
yet. SemiSync is asynchronous replication rather than a group, so those directions require forming
or tearing down a real group and are still being worked on.

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
kubectl delete -n demo my/my-standalone my/my-cluster
kubectl delete -n demo myops/promote-to-gr myops/promote-to-multi-primary
kubectl delete -n demo myops/promote-to-innodb myops/promote-to-semisync
kubectl delete -n demo myops/gr-to-innodb myops/innodb-to-gr myops/gr-to-semisync
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLOpsRequest object](/docs/guides/mysql/concepts/opsrequest.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
