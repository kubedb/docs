---
title: Refreshing a Neo4j Branch
menu:
  docs_{{ .version }}:
    identifier: neo4j-branch-refresh
    name: Refresh
    parent: guides-neo4j-branch
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Refreshing a Neo4j Branch

Set `spec.schedule.cron` to periodically replace branch data with a fresh source snapshot. This guide deploys its own source and scheduled branch, changes both sides, and verifies what a completed refresh preserves and replaces.

## What happens during refresh

1. Courier advances the refresh generation and takes fresh source snapshots.
2. It imports snapshot references into the target namespace, pauses the target, and replaces its cloned volumes.
3. KubeDB restarts the target using the new volumes.
4. After the target and any configured preparation complete, the Branch returns to `Ready`; Courier releases the previous generation’s snapshots.

**Refresh discards branch-only writes and temporarily interrupts the target.** It replaces data rather than merging changes. The source remains independent. If configured, [password reset](/docs/guides/neo4j/branch/reset-root-password.md) and [post-actions](/docs/guides/neo4j/branch/post-actions.md) run again on each refresh.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `Neo4jVersion` `2025.12.1` is available in your installed catalog.
- Use Ceph RBD storage with copy-on-write cloning, compatible VolumeSnapshot and VolumeGroupSnapshot classes, and the corresponding CSI APIs, controllers, and driver support.
- Review the [snapshot requirements](/docs/operatormanual/branching/snapshots.md). Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.
- Allow capacity for three source and three target database pods, each requesting 2Gi of memory, plus operators and action Jobs.

The source uses `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the tutorial names and create namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `neo4j-refresh-source.yaml`. It creates a three-server cluster:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-refresh-source
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
kubectl -n demo apply -f neo4j-refresh-source.yaml
kubectl -n demo wait neo4j/neo4j-refresh-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

Use the Enterprise server catalog entry for this cluster. The client examples select database `neo4j` and use routing so writes reach its leader.

## Seed and inspect the source data

Discover the auth Secret referenced by `neo4j-refresh-source`:

```shell
kubectl -n demo get neo4j neo4j-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-refresh-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret neo4j-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it neo4j-refresh-source-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-refresh-source.demo.svc:7687 -u neo4j -d neo4j --format plain
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

## Create a scheduled Branch

The example runs on ten-minute clock boundaries, such as `10:00`, `10:10`, and `10:20`; it is not a ten-minute timer from creation.

Save the following as `neo4j-refresh-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: neo4j-refresh-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Neo4j
      name: neo4j-refresh-source
    namespace: demo
  target:
    namespace: dev
    name: neo4j-refresh-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
  schedule:
    cron: '*/10 * * * *'
  historyLimit:
    success: 3
    failed: 2
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f neo4j-refresh-branch.yaml
kubectl -n demo wait branch/neo4j-refresh-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch neo4j-refresh-branch
```

Courier creates `neo4j/neo4j-refresh-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

### Check the initial copy

Discover the auth Secret referenced by `neo4j-refresh-branch`:

```shell
kubectl -n dev get neo4j neo4j-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-refresh-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret neo4j-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it neo4j-refresh-branch-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-refresh-branch.dev.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

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

### Record the current generation and schedule

```shell
kubectl -n demo get branch neo4j-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime,NEXT:.status.nextRefreshTime'
```

Record the generation and successful completion timestamp before making the following edits. Initial generation `0` can appear as `<none>`. Make both edits before the next scheduled run; if one starts while you are editing, finish the edits and wait for a subsequent generation. Runs that are still active can cause scheduled ticks to be skipped.

## Change source and branch data

### Update the source

Discover the auth Secret referenced by `neo4j-refresh-source`:

```shell
kubectl -n demo get neo4j neo4j-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-refresh-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret neo4j-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it neo4j-refresh-source-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-refresh-source.demo.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer {id: 1}) SET c.full_name = "Updated Customer 1";
CALL db.checkpoint();
```

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Updated Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

### Make a branch-only change

Discover the auth Secret referenced by `neo4j-refresh-branch`:

```shell
kubectl -n dev get neo4j neo4j-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-refresh-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret neo4j-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it neo4j-refresh-branch-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-refresh-branch.dev.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer {id: 5}) SET c.full_name = "Branch-only Customer 5";
CALL db.checkpoint();
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
5, "Branch-only Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

## Wait for the next completed refresh

```shell
kubectl -n demo get branch neo4j-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime' --watch
```

Illustrative output:

```text
NAME                   PHASE        GENERATION   LAST-SUCCESS
neo4j-refresh-branch   Ready        <none>       2026-11-11T10:02:00Z
neo4j-refresh-branch   Refreshing   1            2026-11-11T10:02:00Z
neo4j-refresh-branch   Ready        1            2026-11-11T10:18:00Z
```

Wait for **a later generation, phase `Ready`, and an advanced `LAST-SUCCESS` timestamp**. An initial `Ready` row or an increased generation alone does not establish completion. Press **Ctrl+C** to stop watching. Recovery duration depends on the engine and available resources.

```shell
kubectl -n demo get branch neo4j-refresh-branch -o yaml
```

Inspect `status.history`, `lastSuccessfulRefreshTime`, and `snapshot.generation`. With the group-snapshot prerequisite above, `snapshot.strategy` is `VolumeGroupSnapshot`; per-volume fallback can occur if no compatible group class is available. `FRESHNESS` measures time since successful completion, not the source snapshot’s age.

## Verify the refreshed data

### Check the branch

Open a new pod/client session after recovery; the previous session may have been disconnected.

Discover the auth Secret referenced by `neo4j-refresh-branch`:

```shell
kubectl -n dev get neo4j neo4j-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-refresh-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret neo4j-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it neo4j-refresh-branch-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-refresh-branch.dev.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Updated Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

Customer `1` has the source update, customer `5` has its original value, and five records remain. The branch-only edit was discarded.

### Confirm the source remains unchanged by refresh

Discover the auth Secret referenced by `neo4j-refresh-source`:

```shell
kubectl -n demo get neo4j neo4j-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-refresh-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret neo4j-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it neo4j-refresh-source-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-refresh-source.demo.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Updated Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

The source retains its own update and never receives the branch-only change.

## Apply the same strategy across clusters

Use the [cross-cluster tutorial](/docs/guides/neo4j/branch/cross-cluster/index.md) and add the same schedule to its Branch. Manage the original Branch on the source cluster, inspect the database on the target, and wait for a later completed generation before checking data.

See the common [schedule and history fields](/docs/operatormanual/branching/refresh.md#schedule-and-history-field-reference) and [refresh status reference](/docs/operatormanual/branching/refresh.md#refresh-status-reference) for configuration details.

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl -n demo delete branch neo4j-refresh-branch --wait=true --timeout=30m
kubectl -n demo delete neo4j neo4j-refresh-source --wait=true --timeout=30m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.
