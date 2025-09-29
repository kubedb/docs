---

title: Oracle Failover and DR Scenarios
menu:
docs_{{ .version }}:
identifier: guides-oracle-failure-and-disaster-recovery-overview
name: Guide
parent: guides-oracle-failure-and-disaster-recovery
weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
-----------------------

> New to KubeDB? Please start [here](/docs/README.md).

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
* KubeDB operator and CLI installed ([instructions](/setup/README.md)).
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

Apply the manifest:

```bash
$ kubectl apply -f oracle-dataguard.yaml
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

* **Primary**: Accepts all writes.
* **Physical Standby**: Exact replica kept in sync with redo logs (WAL-like mechanism).
* **Observer**: External monitoring process that detects failures and triggers `Fast-Start Failover (FSFO)`.
* **Maximum Protection mode**: Ensures **zero data loss** by requiring at least one synchronous standby to acknowledge transactions before commit.
* **FastStartFailoverThreshold**: Defines how quickly failover should happen (e.g., 15 seconds).

When failure is detected, the observer promotes a standby to primary. The cluster is then automatically reconfigured so all remaining replicas realign under the new primary.



### Step 3: Simulating Failover Scenarios

You can check current roles:

```bash
$ kubectl get pods -n demo --show-labels | grep role
oracle-sample-0            2/2     Running   0          49m   app.kubernetes.io/component=database,app.kubernetes.io/instance=oracle-sample,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=oracles.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=oracle-sample-6d6fdb69ff,kubedb.com/role=primary,oracle.db/role=instance,statefulset.kubernetes.io/pod-name=oracle-sample-0
oracle-sample-1            2/2     Running   0          49m   app.kubernetes.io/component=database,app.kubernetes.io/instance=oracle-sample,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=oracles.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=oracle-sample-6d6fdb69ff,kubedb.com/role=standby,oracle.db/role=instance,statefulset.kubernetes.io/pod-name=oracle-sample-1
oracle-sample-2            2/2     Running   0          48m   app.kubernetes.io/component=database,app.kubernetes.io/instance=oracle-sample,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=oracles.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=oracle-sample-6d6fdb69ff,kubedb.com/role=standby,oracle.db/role=instance,statefulset.kubernetes.io/pod-name=oracle-sample-2
oracle-sample-observer-0   1/1     Running   0          49m   app.kubernetes.io/component=database,app.kubernetes.io/instance=oracle-sample,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=oracles.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=oracle-sample-observer-68648c7957,oracle.db/role=observer,statefulset.kubernetes.io/pod-name=oracle-sample-observer-0

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

Within ~15 seconds (defined by `fastStartFailoverThreshold`), a standby is promoted:

```
oracle-sample-1   kubedb.com/role=primary
oracle-sample-2   kubedb.com/role=standby
```

The deleted pod comes back as a **standby** and automatically resynchronizes.

---

#### Case 2: Delete Primary and One Standby

```bash
kubectl delete pod -n demo oracle-sample-0 oracle-sample-1
```

The remaining standby (`oracle-sample-2`) is promoted to primary. The deleted pods return and rejoin as standbys.

---

#### Case 3: Delete All Standbys

```bash
kubectl delete pod -n demo oracle-sample-1 oracle-sample-2
```

The primary (`oracle-sample-0`) continues serving traffic. Once the standbys are recreated, they rejoin the Data Guard configuration and catch up from archived redo logs.

---

#### Case 4: Delete All Pods

```bash
kubectl delete pod -n demo oracle-sample-0 oracle-sample-1 oracle-sample-2
```

After restart, the cluster automatically re-establishes Data Guard roles:

```
oracle-sample-0   kubedb.com/role=primary
oracle-sample-1   kubedb.com/role=standby
oracle-sample-2   kubedb.com/role=standby
```

---

### Step 4: Disaster Recovery with Volume Expansion

If storage becomes full, the database enters a `Not Ready` state. KubeDB allows recovery via **VolumeExpansion OpsRequest**:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: oracle-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: oracle-sample
  volumeExpansion:
    mode: "Offline"
    oracle: 50Gi
```

Apply the manifest:

```bash
kubectl apply -f oracle-volume-expansion.yaml
```

KubeDB expands the volume and recovers the database automatically.

> **Note:** Use `Online` mode if your storage class supports it; otherwise choose `Offline`.

---

### CleanUp

To delete resources:

```bash
kubectl delete oracle -n demo oracle-sample
kubectl delete ns demo
```

---

