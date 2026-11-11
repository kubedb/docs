---
title: Resetting a MongoDB Branch Root Password
menu:
  docs_{{ .version }}:
    identifier: mongodb-branch-reset-root-password
    name: ResetRootPassword
    parent: guides-mongodb-branch
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Resetting a MongoDB Branch Root Password

Set `spec.resetRootPassword: true` to give the branch a separate root password. Rotating branch credentials reduces the risk of exposing or reusing production credentials in development and test environments. The source retains its original password.

This independent tutorial deploys a source and a same-cluster branch in another namespace, compares their passwords, and verifies which credentials each deployment accepts.

## Execution order

Courier first copies the source credential so the target can start with the cloned data. KubeDB Ops Manager then processes a target `MongoDBOpsRequest` with type `RotateAuth`. Courier waits for rotation to succeed before completing password reset, running any post-actions, and reporting the Branch `Ready`. Different Secret names alone do not establish different passwords.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation) with KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `MongoDBVersion` `7.0.21` is available in your KubeDB catalog.
- Use Ceph RBD storage supporting copy-on-write snapshots and clones. These three-member examples use compatible VolumeSnapshot and VolumeGroupSnapshot classes, with the corresponding CSI APIs, controllers, and driver support. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Install KubeDB Ops Manager and the `MongoDBOpsRequest` CRD; password reset uses a `RotateAuth` request on the target.
- Allow capacity for three source and three target database pods, each requesting 1Gi of memory, plus operators and any action Jobs.
- Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.

The source uses namespace `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the names in this tutorial and create namespaces if they do not already exist.

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `mongodb-reset-source.yaml`. It creates a three-member replica set:

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongodb-reset-source
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
kubectl -n demo apply -f mongodb-reset-source.yaml
kubectl -n demo wait mongodb/mongodb-reset-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Record the source credential

Before creating the Branch, record the decoded source password for comparison and the later rejection check.

Discover the auth Secret referenced by `mongodb-reset-source`:

```shell
kubectl -n demo get mongodb mongodb-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-reset-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mongodb-reset-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mongodb-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

## Seed and inspect the source data

Discover the auth Secret referenced by `mongodb-reset-source`:

```shell
kubectl -n demo get mongodb mongodb-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-reset-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mongodb-reset-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mongodb-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mongodb-reset-source-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-reset-source.demo.svc --port 27017 --username root --authenticationDatabase admin --password
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

## Create a Branch with password reset enabled

Save the following as `mongodb-reset-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mongodb-reset-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MongoDB
      name: mongodb-reset-source
    namespace: demo
  target:
    namespace: dev
    name: mongodb-reset-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
  resetRootPassword: true
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mongodb-reset-branch.yaml
kubectl -n demo wait branch/mongodb-reset-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch mongodb-reset-branch
```

Courier creates `mongodb/mongodb-reset-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

Inspect the initial generation’s rotation request in `dev`:

```shell
kubectl -n dev get mongodbopsrequest mongodb-reset-branch-branch-reset-0 -o custom-columns='NAME:.metadata.name,TYPE:.spec.type,PHASE:.status.phase'
```

```text
NAME                                  TYPE         PHASE
mongodb-reset-branch-branch-reset-0   RotateAuth   Successful
```

Check the Branch conditions:

```shell
kubectl -n demo get branch mongodb-reset-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
```

Expect `RootPasswordReset=True` and `Ready=True`. The following checks establish that the intended credentials actually authenticate.

## Compare source and branch passwords

Discover both Secret names:

```shell
kubectl -n demo get mongodb mongodb-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
kubectl -n dev get mongodb mongodb-reset-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-reset-source-auth
mongodb-reset-branch-auth
```

Use the returned names if they differ. Read each password separately:

```shell
kubectl -n demo get secret mongodb-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
source-example-password
```

```shell
kubectl -n dev get secret mongodb-reset-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
branch-example-password
```

The displayed strings are illustrative. Your source password must match the value recorded before branching, while the branch password must differ.

## Verify authentication through the Services

Open the source pod shell from your terminal:

```shell
kubectl -n demo exec -it mongodb-reset-source-0 -c mongodb -- bash
```

### Connect with the branch credential

Run this command inside the pod and enter the **branch password** at the password prompt:

```shell
mongosh --host mongodb-reset-branch.dev.svc --port 27017 --username root --authenticationDatabase admin --password
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

Run `exit` to leave `mongosh`, keeping the pod shell open.

### Confirm the source credential is rejected by the branch

Repeat the command below, this time entering the **source password recorded before branching** at the prompt. The endpoint is still the branch Service:

```shell
mongosh --host mongodb-reset-branch.dev.svc --port 27017 --username root --authenticationDatabase admin --password
```

```text
MongoServerError: Authentication failed.
```

A timeout or DNS error does not verify credential isolation; expect an authentication rejection from the branch.

### Confirm the source credential still works

Connect to the source Service and enter the **original source password**:

```shell
mongosh --host mongodb-reset-source.demo.svc --port 27017 --username root --authenticationDatabase admin --password
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

## Use password reset with other branch features

On each [refresh](/docs/guides/mongodb/branch/refresh.md), the cloned data brings back the source credential and Courier performs a new rotation before readiness. Re-read the branch Secret after refresh. Configured [post-actions](/docs/guides/mongodb/branch/post-actions.md) run after password reset and receive branch credentials. For another cluster, use the same setting on the [cross-cluster Branch](/docs/guides/mongodb/branch/cross-cluster/index.md) and inspect the rotation request on the target.

See the common [password-reset field reference](/docs/operatormanual/branching/reset-root-password.md#field-reference).

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch mongodb-reset-branch --wait=true --timeout=30m
kubectl -n demo delete mongodb mongodb-reset-source --wait=true --timeout=30m
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.
