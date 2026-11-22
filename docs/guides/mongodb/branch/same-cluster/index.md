---
title: Branch MongoDB into Another Namespace
menu:
  docs_{{ .version }}:
    identifier: mongodb-branch-same-cluster
    name: Same-Cluster
    parent: guides-mongodb-branch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch MongoDB into Another Namespace

Same-cluster branching creates an independent MongoDB deployment from a point-in-time snapshot within the same Kubernetes cluster. Place the branch in another namespace to prepare development, testing, or data analysis environments using existing data.

The branch accepts independent reads and writes without changing the source. Source changes are not continuously replicated; scheduled refreshes can replace branch data with fresh source snapshots.

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

Save the following as `mongodb-source.yaml`. It creates a three-member replica set:

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongodb-source
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
kubectl -n demo apply -f mongodb-source.yaml
kubectl -n demo wait mongodb/mongodb-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

Discover the auth Secret referenced by `mongodb-source`:

```shell
kubectl -n demo get mongodb mongodb-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mongodb-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mongodb-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mongodb-source-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-source.demo.svc --port 27017 --username root --authenticationDatabase admin --password
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

## Create the Branch

Save the following as `mongodb-dev.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mongodb-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MongoDB
      name: mongodb-source
    namespace: demo
  target:
    namespace: dev
    name: mongodb-dev
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mongodb-dev.yaml
kubectl -n demo wait branch/mongodb-dev --for=condition=Ready --timeout=60m
kubectl -n demo get branch mongodb-dev
```

Courier creates `mongodb/mongodb-dev` and its cloned PVCs in `dev`; KubeDB provisions the target database.

## Verify the copied data

Discover the auth Secret referenced by `mongodb-dev`:

```shell
kubectl -n dev get mongodb mongodb-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-dev-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mongodb-dev-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mongodb-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mongodb-dev-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-dev.dev.svc --port 27017 --username root --authenticationDatabase admin --password
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

### Check the target cluster

```shell
kubectl -n dev get pods,pvc -l app.kubernetes.io/instance=mongodb-dev
```

The target has three data-bearing members with its own addresses. In the target `mongosh` session, inspect the replica-set members:

Discover the auth Secret referenced by `mongodb-dev`:

```shell
kubectl -n dev get mongodb mongodb-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-dev-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mongodb-dev-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mongodb-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mongodb-dev-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-dev.dev.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

```javascript
rs.status().members.map(member => ({host: member.name, state: member.stateStr}))
```

```text
[
  { host: "mongodb-dev-0.mongodb-dev-pods.dev.svc.cluster.local:27017", state: "PRIMARY" },
  { host: "mongodb-dev-1.mongodb-dev-pods.dev.svc.cluster.local:27017", state: "SECONDARY" },
  { host: "mongodb-dev-2.mongodb-dev-pods.dev.svc.cluster.local:27017", state: "SECONDARY" }
]
```

The primary ordinal and DNS suffix can vary; all member addresses must belong to the branch, with one primary and two secondaries.

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

## Verify independent writes

### Change the branch

Discover the auth Secret referenced by `mongodb-dev`:

```shell
kubectl -n dev get mongodb mongodb-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-dev-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mongodb-dev-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mongodb-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mongodb-dev-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-dev.dev.svc --port 27017 --username root --authenticationDatabase admin --password
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

### Change and inspect the source

Discover the auth Secret referenced by `mongodb-source`:

```shell
kubectl -n demo get mongodb mongodb-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mongodb-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mongodb-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mongodb-source-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-source.demo.svc --port 27017 --username root --authenticationDatabase admin --password
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

The source retains the original customer `5`; the branch edit did not reach it.

### Check the branch again

Discover the auth Secret referenced by `mongodb-dev`:

```shell
kubectl -n dev get mongodb mongodb-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-dev-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mongodb-dev-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mongodb-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mongodb-dev-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-dev.dev.svc --port 27017 --username root --authenticationDatabase admin --password
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
  {"_id": 5, "full_name": "Branch-only Customer 5", "email": "customer5@example.com"}
]
```

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

The branch retains its own customer `5` and the original customer `1`. The source update did not reach it.

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch mongodb-dev --wait=true --timeout=30m
kubectl -n demo delete mongodb mongodb-source --wait=true --timeout=30m
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.

## Next Steps

- [Refresh the branch](/docs/guides/mongodb/branch/refresh.md).
- [Reset the branch root password](/docs/guides/mongodb/branch/reset-root-password.md).
- [Prepare branch data with post-actions](/docs/guides/mongodb/branch/post-actions.md).
- [Configure snapshots, TLS, and deletion policies](/docs/guides/mongodb/branch/customization/index.md).
