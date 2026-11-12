---
title: Branch MongoDB Across Clusters
menu:
  docs_{{ .version }}:
    identifier: mongodb-branch-cross-cluster
    name: Cross-Cluster
    parent: guides-mongodb-branch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch MongoDB Across Clusters

Cross-cluster branching creates an independent MongoDB deployment on another Kubernetes cluster from a point-in-time snapshot of an existing deployment. Use it to prepare development, testing, or data analysis environments on a separate cluster using existing data.

The branch accepts independent reads and writes without changing the source. Source changes are not continuously replicated; scheduled refreshes can replace branch data with fresh source snapshots.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm) with KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `MongoDBVersion` `7.0.21` is available in your KubeDB catalog.
- Use Ceph RBD storage supporting copy-on-write snapshots and clones. These three-member examples use compatible VolumeSnapshot and VolumeGroupSnapshot classes, with the corresponding CSI APIs, controllers, and driver support. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Have existing OCM hub registration for managed clusters `source` and `target`, with Courier agents enabled. Configure kubectl contexts `hub`, `source`, and `target`.
- Both clusters must access **the same network-reachable Ceph backend** with compatible CSI identities, pool settings, and credentials. Matching StorageClass names alone are insufficient.
- Allow capacity for three source and three target database pods, each requesting 1Gi of memory, plus operators and any action Jobs.
- Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.

The source uses namespace `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the names in this tutorial and create namespaces if they do not already exist.

```shell
kubectl --context source create namespace demo
kubectl --context target create namespace dev
```

Check the existing OCM registration and storage classes:

```shell
kubectl --context hub get managedclusters source target
kubectl --context source get storageclass ceph-rbd
kubectl --context target get storageclass ceph-rbd
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
kubectl --context source -n demo apply -f mongodb-source.yaml
kubectl --context source -n demo wait mongodb/mongodb-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

Discover the auth Secret referenced by `mongodb-source`:

```shell
kubectl --context source -n demo get mongodb mongodb-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl --context source -n demo get secret mongodb-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl --context source -n demo get secret mongodb-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context source -n demo exec -it mongodb-source-0 -c mongodb -- bash
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

## Create the cross-cluster Branch

Save the following as `mongodb-cross.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mongodb-cross
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
    name: mongodb-cross
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
    clusterName: target
  deletionPolicy: Delete
```

```shell
kubectl --context source -n demo apply -f mongodb-cross.yaml
kubectl --context source -n demo wait branch/mongodb-cross --for=condition=Ready --timeout=60m
kubectl --context source -n demo get branch mongodb-cross
```

The Branch is created on `source`. Courier sends the target resource manifests through the OCM hub; the target KubeDB provisioner creates the database using the delivered cloned PVCs. `target.clusterName` is the OCM ManagedCluster name.

```shell
kubectl --context hub -n source get branchworks
kubectl --context hub -n target get manifestworks
kubectl --context target -n dev get mongodb mongodb-cross
```

## Verify the copied data

Discover the auth Secret referenced by `mongodb-cross`:

```shell
kubectl --context target -n dev get mongodb mongodb-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-cross-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl --context target -n dev get secret mongodb-cross-auth -o jsonpath='{.data.username}' | base64 -d
kubectl --context target -n dev get secret mongodb-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it mongodb-cross-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-cross.dev.svc --port 27017 --username root --authenticationDatabase admin --password
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
kubectl --context target -n dev get pods,pvc -l app.kubernetes.io/instance=mongodb-cross
```

The target has three data-bearing members with its own addresses. In the target `mongosh` session, inspect the replica-set members:

Discover the auth Secret referenced by `mongodb-cross`:

```shell
kubectl --context target -n dev get mongodb mongodb-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-cross-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl --context target -n dev get secret mongodb-cross-auth -o jsonpath='{.data.username}' | base64 -d
kubectl --context target -n dev get secret mongodb-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it mongodb-cross-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-cross.dev.svc --port 27017 --username root --authenticationDatabase admin --password
```

Enter the password decoded above. Run JavaScript at the `mongosh` prompt.

```javascript
rs.status().members.map(member => ({host: member.name, state: member.stateStr}))
```

```text
[
  { host: "mongodb-cross-0.mongodb-cross-pods.dev.svc.cluster.local:27017", state: "PRIMARY" },
  { host: "mongodb-cross-1.mongodb-cross-pods.dev.svc.cluster.local:27017", state: "SECONDARY" },
  { host: "mongodb-cross-2.mongodb-cross-pods.dev.svc.cluster.local:27017", state: "SECONDARY" }
]
```

The primary ordinal and DNS suffix can vary; all member addresses must belong to the branch, with one primary and two secondaries.

Run `exit` in `mongosh`, then `exit` in the pod shell to return to your terminal.

## Verify independent writes

### Change the branch

Discover the auth Secret referenced by `mongodb-cross`:

```shell
kubectl --context target -n dev get mongodb mongodb-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-cross-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl --context target -n dev get secret mongodb-cross-auth -o jsonpath='{.data.username}' | base64 -d
kubectl --context target -n dev get secret mongodb-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it mongodb-cross-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-cross.dev.svc --port 27017 --username root --authenticationDatabase admin --password
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
kubectl --context source -n demo get mongodb mongodb-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl --context source -n demo get secret mongodb-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl --context source -n demo get secret mongodb-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context source -n demo exec -it mongodb-source-0 -c mongodb -- bash
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

Discover the auth Secret referenced by `mongodb-cross`:

```shell
kubectl --context target -n dev get mongodb mongodb-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mongodb-cross-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl --context target -n dev get secret mongodb-cross-auth -o jsonpath='{.data.username}' | base64 -d
kubectl --context target -n dev get secret mongodb-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `root`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it mongodb-cross-0 -c mongodb -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
mongosh --host mongodb-cross.dev.svc --port 27017 --username root --authenticationDatabase admin --password
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
kubectl --context source -n demo delete branch mongodb-cross --wait=true --timeout=30m
kubectl --context source -n demo delete mongodb mongodb-source --wait=true --timeout=30m
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.

## Next Steps

- [Refresh the branch](/docs/guides/mongodb/branch/refresh.md).
- [Reset the branch root password](/docs/guides/mongodb/branch/reset-root-password.md).
- [Prepare branch data with post-actions](/docs/guides/mongodb/branch/post-actions.md).
- [Configure snapshots, TLS, and deletion policies](/docs/guides/mongodb/branch/customization/index.md).
