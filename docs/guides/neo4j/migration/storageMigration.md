---
title: Neo4j StorageClass Migration Guide
menu:
  docs_{{ .version }}:
    identifier: neo4j-migration-storageclass
    name: StorageClass Migration
    parent: neo4j-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j StorageClass Migration

This guide shows how to migrate the `StorageClass` of a KubeDB-managed Neo4j cluster using `Neo4jOpsRequest` with `type: StorageMigration`.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured.
- Install KubeDB operator following [setup guide](/docs/setup/README.md).
- Ensure at least two `StorageClass` resources are available in your cluster.

Use a dedicated namespace for this walkthrough:

```bash
kubectl create ns demo
```
namespace/demo created

## Prepare Neo4j Database

First, verify available storage classes:

```bash
kubectl get sc
```
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
custom-longhorn        driver.longhorn.io      Delete          WaitForFirstConsumer   true                   3h38m
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  4h26m
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   3h43m
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   3h43m

We will deploy Neo4j with `local-path`, then migrate to `custom-longhorn`.

> Both old and new PVCs should stay on the same node. If the old class uses `WaitForFirstConsumer`, use a new class with `WaitForFirstConsumer` as well.

Apply the Neo4j database manifest:

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  replicas: 3
  deletionPolicy: WipeOut
  version: "2025.12.1"
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
EOF
```
neo4j.kubedb.com/neo4j-test created

```bash
kubectl get neo4j,pvc -n demo
```
NAME                      VERSION     STATUS   AGE
neo4j.kubedb.com/neo4j-test   2025.12.1   Ready    2m

NAME                               STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-neo4j-test-0   Bound    ...      2Gi        RWO            local-path     2m
persistentvolumeclaim/data-neo4j-test-1   Bound    ...      2Gi        RWO            local-path     2m
persistentvolumeclaim/data-neo4j-test-2   Bound    ...      2Gi        RWO            local-path     2m

The database is `Ready` and all the `PersistentVolumeClaim` uses `local-path`  StorageClass, Let's create a database and seed some data.

## Step 3 — Create a Database and Seed Data

Open a shell into pod `neo4j-test-0` and run the following Cypher commands to:
- Create a new database called `appdb`
- Insert 2,000 test `User` nodes
- Verify the count

```bash
# Retrieve the admin password
PASS=$(kubectl get secret -n demo neo4j-test-auth \
  -o jsonpath='{.data.password}' | base64 -d)

# Create the database
$ kubectl exec -n demo neo4j-test-0 -- \
  cypher-shell -u neo4j -p "$PASS" \
  "CREATE DATABASE appdb IF NOT EXISTS WAIT"

# Seed 2,000 User nodes
$ kubectl exec -n demo neo4j-test-0 -- \
  cypher-shell -d appdb -u neo4j -p "$PASS" \
  "UNWIND range(1,2000) AS i CREATE (:User {id:i, name:'user-'+toString(i)})"

# Confirm the count
$ kubectl exec -n demo neo4j-test-0 -- \
  cypher-shell -d appdb -u neo4j -p "$PASS" \
  "MATCH (u:User) RETURN count(u) AS totalUsers"
```

Expected output:

```
totalUsers
2000
```

## Apply StorageMigration OpsRequest

To migrate `StorageClass`, create a `Neo4jOpsRequest`:

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: neo4j-test
  migration:
    storageClassName: custom-longhorn
    oldPVReclaimPolicy: Delete
  timeout: 3000s
EOF
```
neo4jopsrequest.ops.kubedb.com/storage-migration created

Here,

- `spec.type` must be `StorageMigration`.
- `spec.databaseRef.name` points to target Neo4j database.
- `spec.migration.storageClassName` is the destination `StorageClass`.
- `spec.migration.oldPVReclaimPolicy` controls old PV reclaim policy.

> To retain old PVs after migration, use `oldPVReclaimPolicy: Retain`.

## Verify StorageClass Migration

Watch the OpsRequest status:

```bash
kubectl get neo4jopsrequest -n demo -w
```
NAME                TYPE               STATUS       AGE
storage-migration   StorageMigration   Successful   8m

Check PVC storage class after migration:

```bash
kubectl get pvc -n demo
```
NAME                STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS      AGE
data-neo4j-test-0   Bound    ...      2Gi        RWO            custom-longhorn   14m
data-neo4j-test-1   Bound    ...      2Gi        RWO            custom-longhorn   14m
data-neo4j-test-2   Bound    ...      2Gi        RWO            custom-longhorn   14m

The PVCs now use `custom-longhorn`, which confirms successful StorageClass migration.

```bash
PASS=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}' | base64 -d)
```

```bash
kubectl exec -n demo neo4j-test-0 -- \
  cypher-shell -d appdb -u neo4j -p "$PASS" \
  "MATCH (u:User) RETURN count(u) AS totalUsers"
```
totalUsers
2000

From the above output we can verify that data remains intact after the `StorageMigration` operation.


## Cleanup

```bash
kubectl delete neo4jopsrequest -n demo storage-migration
```

```bash
kubectl delete neo4j -n demo neo4j-test
```

```bash
kubectl delete ns demo
```
