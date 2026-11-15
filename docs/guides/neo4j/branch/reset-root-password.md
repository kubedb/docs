---
title: Resetting a Neo4j Branch Root Password
menu:
  docs_{{ .version }}:
    identifier: neo4j-branch-reset-root-password
    name: ResetRootPassword
    parent: guides-neo4j-branch
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Resetting a Neo4j Branch Root Password

Set `spec.resetRootPassword: true` to give the branch its own administrator password. Separating branch credentials reduces the risk of exposing or reusing production credentials in development and test environments. The source keeps its original credential.

This tutorial deploys its own source and branch, compares their passwords, and verifies authentication through their Services.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `Neo4jVersion` `2025.12.1` is available in your installed catalog.
- Use Ceph RBD storage with copy-on-write cloning, compatible VolumeSnapshot and VolumeGroupSnapshot classes, and the corresponding CSI APIs, controllers, and driver support.
- Review the [snapshot requirements](/docs/operatormanual/branching/snapshots.md). Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.
- Allow capacity for three source and three target database pods, each requesting 2Gi of memory, plus operators and action Jobs.
- Install KubeDB Ops Manager and the `Neo4jOpsRequest` CRD; credential reset uses a target `RotateAuth` request.

The source uses `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the tutorial names and create namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `neo4j-reset-source.yaml`. It creates a three-server cluster:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-reset-source
  namespace: demo
spec:
  version: 2025.12.1
  replicas: 3
  podTemplate:
    spec:
      containers:
      - name: neo4j
        resources:
          requests:
            cpu: 500m
            memory: 2Gi
          limits:
            memory: 2Gi
  storageType: Durable
  storage:
    storageClassName: ceph-rbd
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 4Gi
  deletionPolicy: WipeOut
```

```shell
kubectl -n demo apply -f neo4j-reset-source.yaml
kubectl -n demo wait neo4j/neo4j-reset-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

Use the Enterprise server catalog entry for this cluster. The client examples select database `neo4j` and use routing so writes reach its leader.

## Record the source password

Record the source password before creating the Branch. You will compare it with the source value after branching and use it for the rejection check.

Discover the auth Secret referenced by `neo4j-reset-source`:

```shell
kubectl -n demo get neo4j neo4j-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-reset-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret neo4j-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

## Seed and inspect the source data

Use the source password recorded above.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it neo4j-reset-source-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-reset-source.demo.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

Create five customers. Expected results below omit connection banners and are formatted for readability.

```cypher
CREATE (c1:Customer {id: 1, full_name: "Customer 1", email: "customer1@example.com"}),
       (c2:Customer {id: 2, full_name: "Customer 2", email: "customer2@example.com"}),
       (c3:Customer {id: 3, full_name: "Customer 3", email: "customer3@example.com"}),
       (c4:Customer {id: 4, full_name: "Customer 4", email: "customer4@example.com"}),
       (c5:Customer {id: 5, full_name: "Customer 5", email: "customer5@example.com"});
CALL db.checkpoint();
```

`db.checkpoint()` is provided by the Enterprise server used here. Committed transaction logs still support recovery when the cloned database starts.

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

## Create a Branch with password reset enabled

Save the following as `neo4j-reset-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: neo4j-reset-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Neo4j
      name: neo4j-reset-source
    namespace: demo
  target:
    namespace: dev
    name: neo4j-reset-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
  resetRootPassword: true
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f neo4j-reset-branch.yaml
kubectl -n demo wait branch/neo4j-reset-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch neo4j-reset-branch
```

Courier creates `neo4j/neo4j-reset-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

Inspect the target rotation request:

```shell
kubectl -n dev get neo4jopsrequest neo4j-reset-branch-branch-reset-0 -o custom-columns='NAME:.metadata.name,TYPE:.spec.type,PHASE:.status.phase'
```

```text
NAME                                  TYPE         PHASE
neo4j-reset-branch-branch-reset-0   RotateAuth   Successful
```

Check the Branch conditions:

```shell
kubectl -n demo get branch neo4j-reset-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
```

Expect `RootPasswordReset=True` and `Ready=True`.

## Compare source and branch passwords

Discover the Secret names:

```shell
kubectl -n demo get neo4j neo4j-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
kubectl -n dev get neo4j neo4j-reset-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-reset-source-auth
neo4j-reset-branch-auth
```

Use the returned names if they differ. Read each password separately:

```shell
kubectl -n demo get secret neo4j-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
source-example-password
```

```shell
kubectl -n dev get secret neo4j-reset-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
branch-example-password
```

These strings are illustrative. The source password must match the value recorded before branching; the branch password must differ.

## Verify authentication through the Services

Open the source pod’s Bash shell from your terminal:

```shell
kubectl -n demo exec -it neo4j-reset-source-0 -c neo4j -- bash
```

### Authenticate with the branch password

Inside the pod, connect to the **branch** Service and enter the decoded **branch password** when prompted:

```shell
cypher-shell -a neo4j://neo4j-reset-branch.dev.svc:7687 -u neo4j -d neo4j --format plain
```

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Customer 5", "customer5@example.com"
```

Run `:exit` to return to the pod shell.

### Confirm the source password is rejected by the branch

Repeat this connection to the **same branch Service**, entering the **original source password** at the prompt:

```shell
cypher-shell -a neo4j://neo4j-reset-branch.dev.svc:7687 -u neo4j -d neo4j --format plain
```

```text
The client is unauthorized due to authentication failure.
```

Expect an authentication rejection; a DNS failure or connection timeout does not establish password isolation.

### Confirm the source credential still works

Connect to the **source** Service and enter the **original source password**:

```shell
cypher-shell -a neo4j://neo4j-reset-source.demo.svc:7687 -u neo4j -d neo4j --format plain
```

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

## Use password reset with other branch features

On refresh, the source credential returns with the freshly cloned data, and Courier runs another target rotation before readiness. Re-read the target Secret after completion.

Configured [post-actions](/docs/guides/neo4j/branch/post-actions.md) receive target credentials after password preparation. For [cross-cluster branching](/docs/guides/neo4j/branch/cross-cluster/index.md), inspect credentials and any rotation requests on the target cluster. See the common [password-reset fields](/docs/operatormanual/branching/reset-root-password.md#field-reference) and [refresh tutorial](/docs/guides/neo4j/branch/refresh.md).

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl -n demo delete branch neo4j-reset-branch --wait=true --timeout=30m
kubectl -n demo delete neo4j neo4j-reset-source --wait=true --timeout=30m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.
