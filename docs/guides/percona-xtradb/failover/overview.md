---
title: PerconaXtraDB Failover and DR Scenarios
menu:
docs_{{ .version }}:
identifier: guides-perconaxtradb-failure-and-disaster-recovery-overview
name: Overview
parent: guides-perconaxtradb-failover
weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
-----------------------

> New to KubeDB? Please start [here](/docs/setup/README.md).

# Maximizing PerconaXtraDB Uptime and Reliability

## A Guide to KubeDB's High Availability and Auto-Failover for PerconaXtraDB

For mission-critical workloads, ensuring zero downtime and data consistency is paramount. Percona XtraDB
Cluster (PXC) provides a synchronous, multi-primary clustering solution powered by Galera replication.
When combined with KubeDB, you get automated deployment, health monitoring, failover, and self-healing
all native to Kubernetes.

KubeDB continuously monitors the health of PerconaXtraDB pods. If a node fails, Galera cluster membership
automatically adjusts, and Kubernetes restarts failed pods. Thanks to Galera’s synchronous replication,
all nodes have the same data, enabling transparent failover and continued availability without data loss.

This guide walks you through setting up a PerconaXtraDB HA cluster using KubeDB and simulating different failure scenarios to observe automatic recovery.

> Failover in Galera-based PerconaXtraDB clusters is typically **instantaneous** since any healthy node can accept reads and writes. Application connection handling (through services or proxies) determines how quickly traffic is routed to healthy nodes, usually within seconds.

---

### Before You Start

* You must have a running Kubernetes cluster and a configured `kubectl`.
* Install KubeDB CLI and KubeDB operator following [the setup instructions]( /docs/setup/README.md).
* A valid [StorageClass](https://kubernetes.io /concepts/storage/storage-classes/) is required.

```bash
$ kubectl get storageclasses
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  6h
```

* For isolation, we’ll use a separate namespace called `demo`.

```shell
$ kubectl create ns demo
namespace/demo created
```

### Step 1: Deploy a PerconaXtraDB Cluster

Unlike traditional MySQL, PerconaXtraDB uses a **Galera-based synchronous replication** where each node can serve both reads and writes (multi-primary). A minimum of 3 nodes is recommended for quorum.

Save the following manifest as `pxc-ha.yaml`:

```yaml
apiVersion: kubedb.com/v1
kind: PerconaXtraDB
metadata:
  name: pxc-ha
  namespace: demo
spec:
  version: "8.0.40"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Apply the manifest:

```shell
$ kubectl apply -f pxc-ha.yaml
perconaxtradb.kubedb.com/pxc-ha created
```

Watch resources until ready:

```shell
$ watch kubectl get perconaxtradb,petset,pods -n demo
```

Sample ready output:

```text
NAME                              VERSION   STATUS   AGE
perconaxtradb.kubedb.com/pxc-ha   8.0.40    Ready    16h

NAME                                  AGE
petset.apps.k8s.appscode.com/pxc-ha   16h

NAME           READY   STATUS    RESTARTS   AGE
pod/pxc-ha-0   2/2     Running   0          16h
pod/pxc-ha-1   2/2     Running   0          16h
pod/pxc-ha-2   2/2     Running   0          16h
```

All nodes are part of the Galera cluster and can handle read/write traffic.
Inspect the role/labels of all nodes (note: Galera is multi-primary — labels might show `Primary` for each):

```shell
$ kubectl get pods -n demo --show-labels | grep pxc-ha
```

### Step 2: Verify Cluster Functionality

Let’s connect to one node and create a database:

```shell
$ kubectl exec -it -n demo pxc-ha-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)

bash-5.1$ mysql -u root --password='kpbogmFdR!tXVcaG'
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 17827
Server version: 8.0.40-31.1 Percona XtraDB Cluster (GPL), Release rel31, Revision 4b32153, WSREP version 26.1.4.3

Copyright (c) 2009-2024 Percona LLC and/or its affiliates
Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE DATABASE odissi;
Query OK, 1 row affected (0.05 sec)

mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| odissi             |
| performance_schema |
| sys                |
+--------------------+
6 rows in set (0.00 sec)

mysql> exit
Bye
bash-5.1$ exit
exit
⏎                                   
```

Check the database from another node:

```shell
$kubectl exec -it -n demo pxc-ha-1 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-5.1$ mysql -u root --password='kpbogmFdR!tXVcaG'
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 19410
Server version: 8.0.40-31.1 Percona XtraDB Cluster (GPL), Release rel31, Revision 4b32153, WSREP version 26.1.4.3

Copyright (c) 2009-2024 Percona LLC and/or its affiliates
Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| odissi             |
| performance_schema |
| sys                |
+--------------------+
6 rows in set (0.01 sec)

mysql> exit
Bye
bash-5.1$ exit
exit

# odissi is present instantly (Galera synchronous replication)
```

### Step 3: Simulate Node Failure

> **Important correction:** earlier this document used the phrase **"Delete the Primary Node"**. That phrasing is misleading for Percona XtraDB (Galera) because **every node is capable of serving reads and writes** (multi-primary). To simulate a single-node failure, delete one *node/pod* (for example `pxc-ha-0`) — not a conceptual "primary" node.

Below are corrected, explicit steps and expected outputs for the failure scenarios.

#### Case 1: Delete a single node (simulate node failure)

1. **Observe current pod state** (open a separate terminal and run):

```bash
$ kubectl get pods -n demo --show-labels | grep pxc-ha
pxc-ha-0 Primary
pxc-ha-1 Primary
pxc-ha-2 Primary
```

2. **Delete the node (pod)**:

```bash
$ kubectl delete pod -n demo pxc-ha-0
pod "pxc-ha-0" deleted
```
we can see output similar to:
```shell


```

3. **Watch cluster state while the pod is terminated and recreated**:

```bash
$ watch -n 2 "kubectl get pods -n demo --show-labels | grep pxc-ha"
```

You will see a sequence similar to:

* Immediately after deletion (brief):

```
pxc-ha-0 clusterStatusCheckFailure
pxc-ha-1 Primary
pxc-ha-2 Primary
```

* During recreation:

```
pxc-ha-0 Non-primary
pxc-ha-1 Primary
pxc-ha-2 Primary
```

* After readiness (pod becomes Ready again):

```
pxc-ha-0 Primary
pxc-ha-1 Primary
pxc-ha-2 Primary

```

4. **Verify Galera cluster membership and status while the node is down** (check from any remaining node, e.g., `pxc-ha-1`):

```bash
$ kubectl exec -it -n demo pxc-ha-1 -- mysql -uroot --password='kpbogmFdR!tXVcaG' -e "SHOW STATUS LIKE 'wsrep_cluster_size';"

+---------------------+-------+
| Variable_name       | Value |
+---------------------+-------+
| wsrep_cluster_size  | 2     |
+---------------------+-------+
```

When `pxc-ha-0` is back:

```shell
$ kubectl exec -it -n demo pxc-ha-1 -- mysql -uroot --password='kpbogmFdR!tXVcaG' -e "SHOW STATUS LIKE 'wsrep_cluster_size';"
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
mysql: [Warning] Using a password on the command line interface can be insecure.
+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| wsrep_cluster_size | 3     |
+--------------------+-------+

```

**Why this fixes the mistake:**

* We replaced the misleading heading "Delete the Primary Node" with an explicit and accurate description: "Delete a single node".
* We added the exact `kubectl delete pod` command so the step actually performs the deletion.
* We included the expected `kubectl get pods` outputs **immediately after deletion**, **during recreation**, and **after readiness** so readers can confirm the cluster state.
* We added `wsrep_cluster_size` checks to show Galera membership changes while a node is down and after it rejoins.

#### Case 2: Delete two nodes

(Kept and clarified from the original doc.) Example commands and expected behavior:

```bash
$ kubectl delete pod -n demo pxc-ha-0 pxc-ha-1
pod "pxc-ha-0" deleted
pod "pxc-ha-1" deleted

```

* Immediately after deletion you'll see two pods terminating; the remaining pod will keep serving traffic as long as it can (quorum depends on how your cluster is configured and how many nodes are healthy).
* Monitor `wsrep_cluster_size` on the remaining node — it will show `1` while two nodes are down and return to `3` after the others rejoin.

#### Case 3: Delete all nodes

```bash
$ kubectl delete pod -n demo pxc-ha-0 pxc-ha-1 pxc-ha-2
pod "pxc-ha-0" deleted
pod "pxc-ha-1" deleted
pod "pxc-ha-2" deleted

```

KubeDB will recreate all pods. After they are all `Running` and Galera finishes SST/IST, cluster size will return to `3` and data will be available.

### Cleanup

To clean up resources created in this tutorial:

```bash
$ kubectl delete perconaxtradb -n demo pxc-ha
$ kubectl delete ns demo
```

### Next Steps


* Monitor your Percona XtraDB database with KubeDB using [built-in Prometheus](/docs/guides/percona-xtradb/monitoring/builtin-prometheus/index.md)
* Monitor with [Prometheus operator](/docs/guides/percona-xtradb/monitoring/prometheus-operator/index.md)
* Use [private Docker registry](/docs/guides/percona-xtradb/private-registry/quickstart/index.md) to deploy Percona XtraDB with KubeDB
* Detail concepts of [Percona XtraDB object]( /docs/guides/percona-xtradb/concepts/perconaxtradb/index.md)


---
