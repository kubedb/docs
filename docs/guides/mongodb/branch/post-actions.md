---
title: Running MongoDB Branch Post-Actions
menu:
  docs_{{ .version }}:
    identifier: mongodb-branch-post-actions
    name: Post-Actions
    parent: guides-mongodb-branch
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running MongoDB Branch Post-Actions

Post-actions prepare a branch before it becomes ready for use. For MongoDB, they can mask personal fields, remove sensitive collections, or create test fixtures. This guide masks five customer documents and removes a sensitive collection without changing the source.

## Execution and readiness

Each `postActions` entry creates a Job in the target namespace. This example uses one entry and one Job. Courier waits for target database readiness, then runs the action; the Branch becomes `Ready` only after it succeeds. A configured password reset would run before the action.

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

Save the following as `mongodb-pa-source.yaml`. It creates a three-member replica set:

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongodb-pa-source
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
kubectl -n demo apply -f mongodb-pa-source.yaml
kubectl -n demo wait mongodb/mongodb-pa-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

Discover the auth Secret referenced by `mongodb-pa-source`:

```shell
kubectl -n demo get mongodb mongodb-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-pa-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mongodb-pa-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mongodb-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mongodb-pa-source-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-pa-source.demo.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

Result examples below are formatted for readability. Insert five customers. Majority acknowledgment and journaling are requested for the replica set writes; the final command flushes the primary’s pending writes.

```javascript
use branchdb
db.customers.insertMany([
  {"_id": 1, "full_name": "Customer 1", "email": "customer1@example.com", "phone": "+1-555-0001"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com", "phone": "+1-555-0002"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com", "phone": "+1-555-0003"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com", "phone": "+1-555-0004"},
  {"_id": 5, "full_name": "Customer 5", "email": "customer5@example.com", "phone": "+1-555-0005"}
], {writeConcern: {w: "majority", j: true}})
db.sensitive_tokens.insertMany([{_id: 1, token: "sample-token-1"}, {_id: 2, token: "sample-token-2"}], {writeConcern: {w: "majority", j: true}})
db.getSiblingDB("admin").runCommand({fsync: 1})
```

```javascript
use branchdb
db.customers.find().sort({_id: 1}).toArray()
```

```text
[
  {"_id": 1, "full_name": "Customer 1", "email": "customer1@example.com", "phone": "+1-555-0001"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com", "phone": "+1-555-0002"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com", "phone": "+1-555-0003"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com", "phone": "+1-555-0004"},
  {"_id": 5, "full_name": "Customer 5", "email": "customer5@example.com", "phone": "+1-555-0005"}
]
```

The source also contains two sensitive tokens:

```javascript
db.sensitive_tokens.find().sort({_id: 1}).toArray()
```

```text
[{ _id: 1, token: "sample-token-1" }, { _id: 2, token: "sample-token-2" }]
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

## Create a Branch with post-actions

Save the following as `mongodb-pa-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mongodb-pa-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MongoDB
      name: mongodb-pa-source
    namespace: demo
  target:
    namespace: dev
    name: mongodb-pa-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
  postActions:
  - image: ghcr.io/appscode-images/mongo:7.0.21
    jobDefaults:
      imagePullPolicy: IfNotPresent
      backoffLimit: 3
      activeDeadlineSeconds: 900
      ttlSecondsAfterFinished: 3600
    jobTemplate:
      metadata:
        labels:
          branch-action: mongodb-pa-branch-prepare
      spec:
        # First operation: mask customer data.
        # Second operation: drop the sensitive collection after masking succeeds.
        args: [mongosh, --quiet, "mongodb://$(MONGO_HOST):$(MONGO_PORT)/branchdb",
               --username, "$(MONGO_USERNAME)", --password, "$(MONGO_PASSWORD)",
               --authenticationDatabase, admin, --eval,
               "db.customers.updateMany({}, {$set: {full_name: \"Masked Customer\", email: \"user@example.invalid\", phone: \"+1-555-0000\"}}); db.sensitive_tokens.drop(); print(\"Prepared customers: \" + db.customers.countDocuments({}));"]
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mongodb-pa-branch.yaml
kubectl -n demo wait branch/mongodb-pa-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch mongodb-pa-branch
```

Courier creates `mongodb/mongodb-pa-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

The image entrypoint executes `mongosh`. Kubernetes expands the `$(MONGO_...)` references from Courier’s injected connection environment. The JavaScript updates customers first, then drops the collection; an uncaught command error stops evaluation and exits unsuccessfully. Repeated updates produce the same masked values, and dropping an absent collection is harmless.

## Inspect the action Job

The Branch passes through `ActionsRunning`; the phase can finish before you observe it. Find its Job by Courier’s labels:

```shell
kubectl -n dev get jobs -l courier.kubedb.com/branch=mongodb-pa-branch,courier.kubedb.com/branch-namespace=demo
kubectl -n dev logs job/mongodb-pa-branch-post-action -c actions
kubectl -n dev get pods -l branch-action=mongodb-pa-branch-prepare
```

Expect Job status `Complete` with `1/1` completions. The log includes `Prepared customers: 5`.

```shell
kubectl -n demo get branch mongodb-pa-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
kubectl -n demo get branch mongodb-pa-branch -o jsonpath='{.status.resources}{"\n"}'
```

Expect `PostActionsCompleted=True` and `Ready=True`. The resource references identify the target auth Secret, cloned PVCs, and the action Job.

## Verify that only the branch was changed

### Inspect the branch

Discover the auth Secret referenced by `mongodb-pa-branch`:

```shell
kubectl -n dev get mongodb mongodb-pa-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-pa-branch-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mongodb-pa-branch-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mongodb-pa-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mongodb-pa-branch-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-pa-branch.dev.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

```javascript
use branchdb
db.customers.find().sort({_id: 1}).toArray()
```

```text
[
  {"_id": 1, "full_name": "Masked Customer", "email": "user@example.invalid", "phone": "+1-555-0000"},
  {"_id": 2, "full_name": "Masked Customer", "email": "user@example.invalid", "phone": "+1-555-0000"},
  {"_id": 3, "full_name": "Masked Customer", "email": "user@example.invalid", "phone": "+1-555-0000"},
  {"_id": 4, "full_name": "Masked Customer", "email": "user@example.invalid", "phone": "+1-555-0000"},
  {"_id": 5, "full_name": "Masked Customer", "email": "user@example.invalid", "phone": "+1-555-0000"}
]
```

```javascript
db.getCollectionNames().includes("sensitive_tokens")
```

```text
false
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

### Inspect the source

Discover the auth Secret referenced by `mongodb-pa-source`:

```shell
kubectl -n demo get mongodb mongodb-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-pa-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mongodb-pa-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mongodb-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mongodb-pa-source-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-pa-source.demo.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

```javascript
use branchdb
db.customers.find().sort({_id: 1}).toArray()
```

```text
[
  {"_id": 1, "full_name": "Customer 1", "email": "customer1@example.com", "phone": "+1-555-0001"},
  {"_id": 2, "full_name": "Customer 2", "email": "customer2@example.com", "phone": "+1-555-0002"},
  {"_id": 3, "full_name": "Customer 3", "email": "customer3@example.com", "phone": "+1-555-0003"},
  {"_id": 4, "full_name": "Customer 4", "email": "customer4@example.com", "phone": "+1-555-0004"},
  {"_id": 5, "full_name": "Customer 5", "email": "customer5@example.com", "phone": "+1-555-0005"}
]
```

```javascript
db.sensitive_tokens.find().sort({_id: 1}).toArray()
```

```text
[{ _id: 1, token: "sample-token-1" }, { _id: 2, token: "sample-token-2" }]
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

## Connection environment

| Variable | Value used by the action |
|---|---|
| `MONGO_HOST` | Target database Service |
| `MONGO_PORT` | 27017 |
| `MONGO_USERNAME` | Username from the target auth Secret |
| `MONGO_PASSWORD` | Password from the target auth Secret |

See the common [action configuration and Job defaults](/docs/operatormanual/branching/post-actions.md#post-action-field-reference) for the feature fields. The image must execute its supplied arguments; the Branch action template does not expose a container `command` override.

## Use post-actions with refresh or another cluster

With [refresh](/docs/guides/mongodb/branch/refresh.md), actions run again against each new copy before readiness. Generation `1` uses a Job named `mongodb-pa-branch-post-action-1`; older generation Jobs can be removed by Courier. Design actions for repetition and use logs before cleanup. For [cross-cluster branching](/docs/guides/mongodb/branch/cross-cluster/index.md), the Job runs on the target cluster; inspect it there and use the target namespace in any explicit connection address.

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch mongodb-pa-branch --wait=true --timeout=30m
kubectl -n demo delete mongodb mongodb-pa-source --wait=true --timeout=30m
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.
