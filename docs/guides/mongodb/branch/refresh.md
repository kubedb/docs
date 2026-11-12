---
title: Refreshing a MongoDB Branch
menu:
  docs_{{ .version }}:
    identifier: mongodb-branch-refresh
    name: Refresh
    parent: guides-mongodb-branch
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Refreshing a MongoDB Branch

Set `spec.schedule.cron` to periodically replace branch data with a fresh source snapshot. This guide deploys its own source and scheduled branch, changes both sides, and verifies what a completed refresh preserves and replaces.

## What happens during refresh

1. Courier advances the refresh generation and takes fresh source snapshots.
2. It imports snapshot references into the target namespace, pauses the target, and replaces its cloned volumes.
3. KubeDB restarts the target using the new volumes.
4. After the target and any configured preparation complete, the Branch returns to `Ready`; Courier releases the previous generation’s snapshots.

**Refresh discards branch-only writes and temporarily interrupts the target.** It replaces data rather than merging changes. The source remains independent. If configured, [password reset](/docs/guides/mongodb/branch/reset-root-password.md) and [post-actions](/docs/guides/mongodb/branch/post-actions.md) run again on each refresh.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation) with KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `MongoDBVersion` `7.0.21` is available in your KubeDB catalog.
- Use Ceph RBD storage supporting copy-on-write snapshots and clones. These three-member examples use compatible VolumeSnapshot and VolumeGroupSnapshot classes, with the corresponding CSI APIs, controllers, and driver support. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Allow capacity for three source and three target database pods, each requesting 1Gi of memory, plus operators and any action Jobs.
- Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.

The source uses namespace `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the names in this tutorial and create namespaces if they do not already exist.

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `mongodb-refresh-source.yaml`. It creates a three-member replica set:

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongodb-refresh-source
  namespace: demo
spec:
  version: 7.0.21
  replicas: 3
  replicaSet:
    name: rs0
  podTemplate:
    spec:
      containers:
      - name: mongodb
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            memory: 1Gi
  storageType: Durable
  storage:
    storageClassName: ceph-rbd
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```shell
kubectl -n demo apply -f mongodb-refresh-source.yaml
kubectl -n demo wait mongodb/mongodb-refresh-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

Discover the auth Secret referenced by `mongodb-refresh-source`:

```shell
kubectl -n demo get mongodb mongodb-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-refresh-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mongodb-refresh-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mongodb-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mongodb-refresh-source-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-refresh-source.demo.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

Result examples below are formatted for readability. Insert five customers. Majority acknowledgment and journaling are requested for the replica set writes; the final command flushes the primary’s pending writes.

```javascript
use branchdb
db.customers.insertMany([
  {"_id": 1, "full_name": "Customer 1", "email": "customer1@example.com"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com"},
  {"_id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
], {writeConcern: {w: "majority", j: true}})
db.getSiblingDB("admin").runCommand({fsync: 1})
```

```javascript
use branchdb
db.customers.find().sort({_id: 1}).toArray()
```

```text
[
  {"_id": 1, "full_name": "Customer 1", "email": "customer1@example.com"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com"},
  {"_id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
]
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

## Create a scheduled Branch

The example runs on ten-minute clock boundaries, such as `10:00`, `10:10`, and `10:20`; it is not a ten-minute timer from creation.

Save the following as `mongodb-refresh-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mongodb-refresh-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MongoDB
      name: mongodb-refresh-source
    namespace: demo
  target:
    namespace: dev
    name: mongodb-refresh-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
  schedule:
    cron: '*/10 * * * *'
  historyLimit:
    success: 3
    failed: 2
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mongodb-refresh-branch.yaml
kubectl -n demo wait branch/mongodb-refresh-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch mongodb-refresh-branch
```

Courier creates `mongodb/mongodb-refresh-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

### Check the initial copy

Discover the auth Secret referenced by `mongodb-refresh-branch`:

```shell
kubectl -n dev get mongodb mongodb-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-refresh-branch-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mongodb-refresh-branch-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mongodb-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mongodb-refresh-branch-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-refresh-branch.dev.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

```javascript
use branchdb
db.customers.find().sort({_id: 1}).toArray()
```

```text
[
  {"_id": 1, "full_name": "Customer 1", "email": "customer1@example.com"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com"},
  {"_id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
]
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

### Record the current generation and schedule

```shell
kubectl -n demo get branch mongodb-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime,NEXT:.status.nextRefreshTime'
```

Record the generation and successful completion timestamp before making the following edits. Initial generation `0` can appear as `<none>`. Make both edits before the next scheduled run; if one starts while you are editing, finish the edits and wait for a subsequent generation. Runs that are still active can cause scheduled ticks to be skipped.

## Change source and branch data

### Update the source

Discover the auth Secret referenced by `mongodb-refresh-source`:

```shell
kubectl -n demo get mongodb mongodb-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-refresh-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mongodb-refresh-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mongodb-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mongodb-refresh-source-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-refresh-source.demo.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

```javascript
use branchdb
db.customers.updateOne({_id: 1}, {$set: {full_name: "Updated Customer 1"}}, {writeConcern: {w: "majority", j: true}})
db.getSiblingDB("admin").runCommand({fsync: 1})
```

```javascript
use branchdb
db.customers.find().sort({_id: 1}).toArray()
```

```text
[
  {"_id": 1, "full_name": "Updated Customer 1", "email": "customer1@example.com"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com"},
  {"_id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
]
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

### Make a branch-only change

Discover the auth Secret referenced by `mongodb-refresh-branch`:

```shell
kubectl -n dev get mongodb mongodb-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-refresh-branch-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mongodb-refresh-branch-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mongodb-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mongodb-refresh-branch-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-refresh-branch.dev.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

```javascript
use branchdb
db.customers.updateOne({_id: 5}, {$set: {full_name: "Branch-only Customer 5"}}, {writeConcern: {w: "majority", j: true}})
db.getSiblingDB("admin").runCommand({fsync: 1})
```

```javascript
use branchdb
db.customers.find().sort({_id: 1}).toArray()
```

```text
[
  {"_id": 1, "full_name": "Customer 1", "email": "customer1@example.com"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com"},
  {"_id": 5, "full_name": "Branch-only Customer 5", "email": "customer5@example.com"}
]
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

## Wait for the next completed refresh

```shell
kubectl -n demo get branch mongodb-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime' --watch
```

Illustrative output:

```text
NAME                   PHASE        GENERATION   LAST-SUCCESS
mongodb-refresh-branch Ready        <none>       2026-11-11T10:02:00Z
mongodb-refresh-branch Refreshing   1            2026-11-11T10:02:00Z
mongodb-refresh-branch Ready        1            2026-11-11T10:18:00Z
```

Wait for **a later generation, phase `Ready`, and an advanced `LAST-SUCCESS` timestamp**. An initial `Ready` row or an increased generation alone does not establish completion. Press **Ctrl+C** to stop watching. Recovery duration depends on the engine and available resources.

```shell
kubectl -n demo get branch mongodb-refresh-branch -o yaml
```

Inspect `status.history`, `lastSuccessfulRefreshTime`, and `snapshot.generation`. With the group-snapshot prerequisite above, `snapshot.strategy` is `VolumeGroupSnapshot`; per-volume fallback can occur if no compatible group class is available. `FRESHNESS` measures time since successful completion, not the source snapshot’s age.

## Verify the refreshed data

### Check the branch

Open a new pod/client session after recovery; the previous session may have been disconnected.

Discover the auth Secret referenced by `mongodb-refresh-branch`:

```shell
kubectl -n dev get mongodb mongodb-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-refresh-branch-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mongodb-refresh-branch-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mongodb-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mongodb-refresh-branch-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-refresh-branch.dev.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

```javascript
use branchdb
db.customers.find().sort({_id: 1}).toArray()
```

```text
[
  {"_id": 1, "full_name": "Updated Customer 1", "email": "customer1@example.com"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com"},
  {"_id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
]
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

Customer `1` has the source update, customer `5` has its original value, and five records remain. The branch-only edit was discarded.

### Confirm the source remains unchanged by refresh

Discover the auth Secret referenced by `mongodb-refresh-source`:

```shell
kubectl -n demo get mongodb mongodb-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-refresh-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mongodb-refresh-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mongodb-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mongodb-refresh-source-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-refresh-source.demo.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

```javascript
use branchdb
db.customers.find().sort({_id: 1}).toArray()
```

```text
[
  {"_id": 1, "full_name": "Updated Customer 1", "email": "customer1@example.com"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com"},
  {"_id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
]
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

The source retains its own update and never receives the branch-only change.

## Apply the same strategy across clusters

Use the [cross-cluster tutorial](/docs/guides/mongodb/branch/cross-cluster/index.md) and add the same schedule to its Branch. Manage the original Branch on the source cluster, inspect the database on the target, and wait for a later completed generation before checking data.

See the common [schedule and history fields](/docs/operatormanual/branching/refresh.md#schedule-and-history-field-reference) and [refresh status reference](/docs/operatormanual/branching/refresh.md#refresh-status-reference) for configuration details.

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch mongodb-refresh-branch --wait=true --timeout=30m
kubectl -n demo delete mongodb mongodb-refresh-source --wait=true --timeout=30m
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.
