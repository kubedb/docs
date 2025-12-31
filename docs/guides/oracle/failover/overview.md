---
title: Oracle Faoilover and Disaster Recovery
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-fdr-overview
    name: Overview
    parent: guides-oracle-fdr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/setup/README.md).

# Maximizing Oracle Uptime and Reliability

## A Guide to KubeDB's Data Guard Based High Availability and Auto-Failover

For mission-critical workloads, Oracle databases are often deployed with `Data Guard`, Oracle’s proven 
technology for disaster recovery and failover. KubeDB extends this capability by natively supporting Data
Guard based replication and automatic failover within Kubernetes clusters.

When the `primary database` becomes unavailable, `KubeDB` together with `Oracle Data Guard` and its `observer process` 
ensures a healthy standby is automatically promoted to primary. This guarantees minimal downtime, strict data 
consistency, and seamless recovery from failures without manual intervention.

This guide demonstrates how to set up an `Oracle HA cluster` with `Data Guard` enabled in `KubeDB`, and how failover works in different scenarios.

---

### Before You Start

* A running Kubernetes cluster with `kubectl` configured.
* KubeDB operator and CLI installed ([instructions](/docs/setup/README.md)).
* A valid `StorageClass` available for persistent volumes.

Check StorageClasses:

```bash
$ kubectl get storageclasses
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  13d

```

* We’ll use the `demo` namespace for isolation:

```bash
$ kubectl create ns demo
```

---

### Step 1: Deploy Oracle with Data Guard Enabled

Save the following YAML as `oracle-dataguard.yaml`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
  namespace: demo
spec:
  version: 21.3.0
  edition: enterprise
  mode: DataGuard
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 30Gi
  deletionPolicy: Delete

  dataGuard:
    protectionMode: MaximumProtection
    standbyType: PHYSICAL
    syncMode: SYNC
    applyLagThreshold: 0
    transportLagThreshold: 0
    fastStartFailover:
      fastStartFailoverThreshold: 15
    observer:
      podTemplate:
        spec:
          containers:
          - name: observer
            resources:
              requests:
                cpu: 500m
                memory: 2Gi
              limits:
                cpu: "1"
                memory: 2Gi
          initContainers:
          - name: observer-init
            resources:
              requests:
                cpu: 200m
                memory: 256Mi
              limits:
                memory: 512Mi

  podTemplate:
    spec:
      serviceAccountName: oracle-sample
      securityContext:
        runAsUser: 54321
        runAsGroup: 54321
        fsGroup: 54321
      containers:
      - name: oracle
        resources:
          requests:
            cpu: "1500m"
            memory: 4Gi
          limits:
            cpu: "4"
            memory: 10Gi
      - name: oracle-coordinator
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            memory: 256Mi
      initContainers:
      - name: oracle-init
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            memory: 512Mi
```
Here,
- `dataGuard.protectionMode` sets the Data Guard protection level. MaximumProtection ensures zero data loss by requiring standby acknowledgment before commit.

- `dataGuard.standbyType` defines the standby as `PHYSICAL`, a block-for-block replica of the primary.

- `dataGuard.syncMode` determines redo log transport. SYNC mode waits for standby confirmation before committing.

- `dataGuard.applyLagThreshold` and `dataGuard.transportLagThreshold` define maximum allowable lag for redo application and transport; 0 enforces immediate synchronization.

- `dataGuard.fastStartFailover.fastStartFailoverThreshold` sets the FSFO trigger time in seconds; here, 15 means automatic failover if the primary is unresponsive for 15s.

- `dataGuard.observer.podTemplate` configures the observer pod, which monitors the primary and triggers FSFO, with CPU and memory resources specified.


Apply the manifest:

```bash
$ kubectl apply -f oracle-dataguard.yaml
```

```shell
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/dataguard/dataguard.yaml
oracle.kubedb.com/oracle created

```
Monitor status until all pods are ready:

```bash
$ watch kubectl get oracle -n demo
NAME            VERSION   MODE        STATUS   AGE
oracle-sample   21.3.0    DataGuard   Ready    25m

```

---

### Step 2: Understanding Oracle Data Guard Failover

Oracle Data Guard in KubeDB works by maintaining **synchronous replication** between a `primary` and `standby`
databases. Key concepts:

* **Primary**: Accepts all writes (read/write).
* **Physical Standby**: Exact replica kept in sync by continuously applying **redo logs**.
* **Redo Logs**: Every change in Oracle (INSERT/UPDATE/DELETE) generates redo entries. These are written
  to the primary’s redo logs, then shipped and applied to the standby. This ensures durability and that
  the standby is always consistent with the primary.
* **Observer**: External monitoring process that detects failures and coordinates **Fast-Start Failover (FSFO)**.
* **Maximum Protection mode**: Ensures **zero data loss** by requiring at least one synchronous standby to
  acknowledge receipt of redo logs before a transaction is committed.
* **FastStartFailover (FSFO)**: Automates failover by monitoring the primary. If the primary is unavailable
  beyond the configured `fastStartFailoverThreshold` (e.g., 15s), the observer promotes a healthy standby
  to primary without human intervention.

---

### How Redo Logs Work (Oracle’s Write-Ahead Mechanism)

1. When a transaction is issued on the primary (e.g., an `INSERT`), Oracle writes the changes into `redo logs`.
2. The `redo logs` are then transported to `standby` databases.
    * **SYNC mode** → Primary waits for standby confirmation before commit (`zero data loss`).
    * **ASYNC mode** → Primary does not wait; minimal latency but potential small data loss.
3. The standby continuously applies redo logs so its data stays in sync with the primary.
4. During failover, the new primary already has all the applied redo logs and can immediately continue serving traffic.

---

### How FSFO(FastStartFailover) Works

* `Observer process` continuously monitors the primary and standbys.
* If the primary is unavailable for longer than `fastStartFailoverThreshold`, FSFO triggers an `automatic failover`.
* A `standby` with the most recent redo logs is promoted to `primary`.
* The failed primary, once recovered, rejoins as a `standby` and `resynchronizes`.

Together, **Redo Logs** and **FSFO** guarantee that Oracle databases deployed with KubeDB remain highly available,
consistent, and resilient against node or pod failures.


### Step 3: Simulating Failover Scenarios

You can check current roles:

```bash
$ kubectl get pods -n demo --show-labels | grep role
oracle-sample-0            2/2     Running   0          49m   app.kubernetes.io/component=database,app.kubernetes.io/instance=oracle-sample,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=oracles.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=oracle-sample-6d6fdb69ff,kubedb.com/role=primary,oracle.db/role=instance,statefulset.kubernetes.io/pod-name=oracle-sample-0
oracle-sample-1            2/2     Running   0          49m   app.kubernetes.io/component=database,app.kubernetes.io/instance=oracle-sample,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=oracles.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=oracle-sample-6d6fdb69ff,kubedb.com/role=standby,oracle.db/role=instance,statefulset.kubernetes.io/pod-name=oracle-sample-1
oracle-sample-2            2/2     Running   0          48m   app.kubernetes.io/component=database,app.kubernetes.io/instance=oracle-sample,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=oracles.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=oracle-sample-6d6fdb69ff,kubedb.com/role=standby,oracle.db/role=instance,statefulset.kubernetes.io/pod-name=oracle-sample-2
oracle-sample-observer-0   1/1     Running   0          49m   app.kubernetes.io/component=database,app.kubernetes.io/instance=oracle-sample,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=oracles.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=oracle-sample-observer-68648c7957,oracle.db/role=observer,statefulset.kubernetes.io/pod-name=oracle-sample-observer-0

```
The pod having `kubedb.com/role=primary` is the primary, `kubedb.com/role=standby` are the standby's and 
`oracle-sample-observer-0` is the observer.


Let's create a table in the primary.
```shell
kubectl exec -it -n demo oracle-sample-0 -- bash
Defaulted container "oracle" out of: oracle, oracle-coordinator, oracle-init (init)
bash-4.2$ sqlplus / as sysdba

SQL*Plus: Release 21.0.0.0.0 - Production on Tue Sep 30 06:02:05 2025
Version 21.3.0.0.0

Copyright (c) 1982, 2021, Oracle.  All rights reserved.


Connected to:
Oracle Database 21c Enterprise Edition Release 21.0.0.0.0 - Production
Version 21.3.0.0.0

SQL> CREATE TABLE kathak (
    id      NUMBER PRIMARY KEY,
    name    VARCHAR2(100),
    age     NUMBER
);
  2    3    4    5  
Table created.

SQL> INSERT INTO kathak (id, name, age) VALUES (1, 'Radha', 25);
1 row created.

SQL> INSERT INTO kathak (id, name, age) VALUES (2, 'Gopal', 28);

1 row created.

SQL> commit;

Commit complete.

# The following commands help format the table for better readability.
SQL> COLUMN id FORMAT 9999
COLUMN name FORMAT A20
COLUMN age FORMAT 999

SQL> SELECT * FROM kathak;

   ID NAME		    AGE
----- -------------------- ----
    1 Radha		     25
    2 Gopal		     28


SQL> exit
Disconnected from Oracle Database 21c Enterprise Edition Release 21.0.0.0.0 - Production
Version 21.3.0.0.0
bash-4.2$ exit
exit

```
Verify that the table has been created on the standby nodes. Note that standby pods have read-only access,
so you won't be able to perform any write operations.
```shell
kubectl exec -it -n demo oracle-sample-0 -- bash
Defaulted container "oracle" out of: oracle, oracle-coordinator, oracle-init (init)
bash-4.2$ sqlplus / as sysdba

SQL*Plus: Release 21.0.0.0.0 - Production on Tue Sep 30 06:50:32 2025
Version 21.3.0.0.0

Copyright (c) 1982, 2021, Oracle.  All rights reserved.


Connected to:
Oracle Database 21c Enterprise Edition Release 21.0.0.0.0 - Production
Version 21.3.0.0.0

SQL> COLUMN id FORMAT 9999
COLUMN name FORMAT A20
COLUMN age FORMAT 999

SQL> SELECT database_role, open_mode FROM v$database;

DATABASE_ROLE	 OPEN_MODE
---------------- --------------------
PHYSICAL STANDBY MOUNTED

SQL> ALTER DATABASE OPEN READ ONLY;

Database altered.

SQL> SELECT * FROM kathak;  

   ID NAME		    AGE
----- -------------------- ----
    1 Radha		     25
    2 Gopal		     28


```
Typical output:

```shell
$ watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```

```bash
oracle-sample-0 primary
oracle-sample-1 standby
oracle-sample-2 standby
oracle-sample-observer-0
```

---

#### Case 1: Delete the Primary

```bash
$ kubectl delete pod -n demo oracle-sample-0
```

Within  few minutes (defined by `fastStartFailoverThreshold`), a standby is promoted:

```
oracle-sample-0 standby
oracle-sample-1 primary
oracle-sample-2 standby
oracle-sample-observer-0

```

The deleted pod comes back as a **standby** and automatically resynchronizes.
Now we know how failover is done, let's check if the new primary is working.
```shell
kubectl exec -it -n demo oracle-sample-1 -- bash
Defaulted container "oracle" out of: oracle, oracle-coordinator, oracle-init (init)
bash-4.2$ sqlplus / as sysdba

SQL*Plus: Release 21.0.0.0.0 - Production on Tue Sep 30 06:46:31 2025
Version 21.3.0.0.0

Copyright (c) 1982, 2021, Oracle.  All rights reserved.


Connected to:
Oracle Database 21c Enterprise Edition Release 21.0.0.0.0 - Production
Version 21.3.0.0.0


SQL> COLUMN id FORMAT 9999
COLUMN name FORMAT A20
COLUMN age FORMAT 999
SQL> SQL> SQL> 
SQL> SELECT * FROM kathak;

   ID NAME		    AGE
----- -------------------- ----
    1 Radha		     25
    2 Gopal		     28

SQL> INSERT INTO kathak (id, name, age) VALUES (3, 'Mohan', 30);

1 row created.

SQL> commit;

Commit complete.

SQL> SELECT * FROM kathak;

   ID NAME		    AGE
----- -------------------- ----
    1 Radha		     25
    2 Gopal		     28
    3 Mohan		     30

SQL> exit
Disconnected from Oracle Database 21c Enterprise Edition Release 21.0.0.0.0 - Production
Version 21.3.0.0.0
bash-4.2$ exit

```

Let's verify the new `standby` is working perfectly
```shell
kubectl exec -it -n demo oracle-sample-0 -- bash
Defaulted container "oracle" out of: oracle, oracle-coordinator, oracle-init (init)
bash-4.2$ sqlplus / as sysdba

SQL*Plus: Release 21.0.0.0.0 - Production on Tue Sep 30 06:50:32 2025
Version 21.3.0.0.0

Copyright (c) 1982, 2021, Oracle.  All rights reserved.


Connected to:
Oracle Database 21c Enterprise Edition Release 21.0.0.0.0 - Production
Version 21.3.0.0.0

SQL> COLUMN id FORMAT 9999
COLUMN name FORMAT A20
COLUMN age FORMAT 999

SQL> SELECT database_role, open_mode FROM v$database;

DATABASE_ROLE	 OPEN_MODE
---------------- --------------------
PHYSICAL STANDBY MOUNTED

SQL> ALTER DATABASE OPEN READ ONLY;

Database altered.

SQL> SELECT * FROM kathak;  

   ID NAME		    AGE
----- -------------------- ----
    1 Radha		     25
    3 Gopal		     28
    2 Mohan		     30

SQL> exit
Disconnected from Oracle Database 21c Enterprise Edition Release 21.0.0.0.0 - Production
Version 21.3.0.0.0
bash-4.2$ command terminated with exit code 137

```
#### Case 2: Delete Primary and One Standby

```bash
$ kubectl delete pod -n demo oracle-sample-0 oracle-sample-1
pod "oracle-sample-0" deleted
pod "oracle-sample-1" deleted

```

The remaining standby (`oracle-sample-2`) is promoted to primary. The deleted pods return and rejoin as standbys.

```shell
oracle-sample-0 primary
oracle-sample-1 standby
oracle-sample-2 standby
oracle-sample-observer-0

```

#### Case 3: Delete All Standbys

```bash
$ kubectl delete pod -n demo oracle-sample-1 oracle-sample-2
```

The primary (`oracle-sample-0`) continues serving traffic. Once the standbys are recreated, they rejoin the Data Guard configuration and catch up from archived redo logs.

```bash
oracle-sample-0 primary
oracle-sample-1 standby
oracle-sample-2 standby
oracle-sample-observer-0

```

#### Case 4: Delete All Pods

```bash
$ kubectl delete pod -n demo oracle-sample-0 oracle-sample-1 oracle-sample-2
pod "oracle-sample-0" deleted
pod "oracle-sample-1" deleted
pod "oracle-sample-2" deleted
```
```shell
oracle-sample-0 
oracle-sample-1 
oracle-sample-observer-0
```
After restart, the cluster automatically re-establishes Data Guard roles:

```bash
oracle-sample-0 primary
oracle-sample-1 standby
oracle-sample-2 standby
oracle-sample-observer-0

```


### CleanUp

To delete resources run:

```bash
kubectl delete oracle -n demo oracle-sample
kubectl delete ns demo
```


