---
title: Running Neo4j Branch Post-Actions
menu:
  docs_{{ .version }}:
    identifier: neo4j-branch-post-actions
    name: Post-Actions
    parent: guides-neo4j-branch
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Neo4j Branch Post-Actions

Post-actions prepare branch data before the Branch becomes ready. They can mask personal information, remove sensitive data, or create test fixtures. This example masks five customer nodes and deletes sensitive token nodes, leaving the source unchanged.

## Execution and readiness

Each `postActions` entry creates a Job in the target namespace. This tutorial uses one entry and one Job, running masking first and sensitive-data removal second. Courier waits for target readiness and any configured credential reset before running the action; the Branch becomes `Ready` only after the Job succeeds.

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

Save the following as `neo4j-pa-source.yaml`. It creates a three-server cluster:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-pa-source
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
kubectl -n demo apply -f neo4j-pa-source.yaml
kubectl -n demo wait neo4j/neo4j-pa-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

Use the Enterprise server catalog entry for this cluster. The client examples select database `neo4j` and use routing so writes reach its leader.

## Seed and inspect the source data

Discover the auth Secret referenced by `neo4j-pa-source`:

```shell
kubectl -n demo get neo4j neo4j-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-pa-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret neo4j-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it neo4j-pa-source-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-pa-source.demo.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

Create five customers. Expected results below omit connection banners and are formatted for readability.

```cypher
CREATE (c1:Customer {id: 1, full_name: "Customer 1", email: "customer1@example.com", phone: "+1-555-0001"}),
       (c2:Customer {id: 2, full_name: "Customer 2", email: "customer2@example.com", phone: "+1-555-0002"}),
       (c3:Customer {id: 3, full_name: "Customer 3", email: "customer3@example.com", phone: "+1-555-0003"}),
       (c4:Customer {id: 4, full_name: "Customer 4", email: "customer4@example.com", phone: "+1-555-0004"}),
       (c5:Customer {id: 5, full_name: "Customer 5", email: "customer5@example.com", phone: "+1-555-0005"});
CREATE (:SensitiveToken {id: 1, token: 'sample-token-1'}),
       (:SensitiveToken {id: 2, token: 'sample-token-2'});
CALL db.checkpoint();
```

`db.checkpoint()` is provided by the Enterprise server used here. Committed transaction logs still support recovery when the cloned database starts.

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email, c.phone AS phone ORDER BY id;
```

```text
id, full_name, email, phone
1, "Customer 1", "customer1@example.com", "+1-555-0001"
2, "Customer 2", "customer2@example.com", "+1-555-0002"
3, "Customer 3", "customer3@example.com", "+1-555-0003"
4, "Customer 4", "customer4@example.com", "+1-555-0004"
5, "Customer 5", "customer5@example.com", "+1-555-0005"
```

```cypher
MATCH (t:SensitiveToken) RETURN count(t) AS tokens;
```

```text
tokens
2
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

## Create a Branch with post-actions

Save the following as `neo4j-pa-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: neo4j-pa-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Neo4j
      name: neo4j-pa-source
    namespace: demo
  target:
    namespace: dev
    name: neo4j-pa-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
  postActions:
  - image: neo4j:2025.12.1
    jobDefaults:
      imagePullPolicy: IfNotPresent
      backoffLimit: 3
      activeDeadlineSeconds: 900
      ttlSecondsAfterFinished: 3600
    jobTemplate:
      metadata:
        labels:
          branch-action: neo4j-pa-branch-prepare
      spec:
        args:
        - bash
        - -ec
        - |
          # First operation: mask the five customer nodes through leader routing.
          cypher-shell -a "neo4j://${NEO4J_HOST}:${NEO4J_PORT}" \
            -u "$NEO4J_USERNAME" -p "$NEO4J_PASSWORD" --format plain --fail-fast \
            "MATCH (c:Customer) SET c.full_name = 'Masked Customer', c.email = 'user@example.invalid', c.phone = '+1-555-0000' RETURN count(c) AS masked;"
          # Second operation: remove sensitive nodes after masking succeeds.
          cypher-shell -a "neo4j://${NEO4J_HOST}:${NEO4J_PORT}" \
            -u "$NEO4J_USERNAME" -p "$NEO4J_PASSWORD" --format plain --fail-fast \
            "MATCH (t:SensitiveToken) DETACH DELETE t;"
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f neo4j-pa-branch.yaml
kubectl -n demo wait branch/neo4j-pa-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch neo4j-pa-branch
```

Courier creates `neo4j/neo4j-pa-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

The image executes the Bash arguments and uses `cypher-shell` only as a client. Both commands use `neo4j://` routing to reach the writer. `--fail-fast` and `bash -ec` stop the Job on an error, so sensitive-node removal follows successful masking. Repeating the fixed update and deleting already-absent nodes are harmless.

## Inspect the action Job

The Branch passes through `ActionsRunning`; the phase may finish before you see it. Find the Job and inspect its logs:

```shell
kubectl -n dev get jobs -l courier.kubedb.com/branch=neo4j-pa-branch,courier.kubedb.com/branch-namespace=demo
kubectl -n dev logs job/neo4j-pa-branch-post-action -c actions
kubectl -n dev get pods -l branch-action=neo4j-pa-branch-prepare
```

Expect Job status `Complete` with `1/1` completions. The first query reports `masked` as `5`.

```shell
kubectl -n demo get branch neo4j-pa-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
kubectl -n demo get branch neo4j-pa-branch -o jsonpath='{.status.resources}{"\n"}'
```

Expect `PostActionsCompleted=True` and `Ready=True`. Resource references identify the target auth Secret, cloned PVCs, and action Job.

## Verify that only the branch was changed

### Inspect the branch

Discover the auth Secret referenced by `neo4j-pa-branch`:

```shell
kubectl -n dev get neo4j neo4j-pa-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-pa-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret neo4j-pa-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it neo4j-pa-branch-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-pa-branch.dev.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email, c.phone AS phone ORDER BY id;
```

```text
id, full_name, email, phone
1, "Masked Customer", "user@example.invalid", "+1-555-0000"
2, "Masked Customer", "user@example.invalid", "+1-555-0000"
3, "Masked Customer", "user@example.invalid", "+1-555-0000"
4, "Masked Customer", "user@example.invalid", "+1-555-0000"
5, "Masked Customer", "user@example.invalid", "+1-555-0000"
```

```cypher
MATCH (t:SensitiveToken) RETURN count(t) AS tokens;
```

```text
tokens
0
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

### Inspect the source

Discover the auth Secret referenced by `neo4j-pa-source`:

```shell
kubectl -n demo get neo4j neo4j-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-pa-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret neo4j-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it neo4j-pa-source-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-pa-source.demo.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email, c.phone AS phone ORDER BY id;
```

```text
id, full_name, email, phone
1, "Customer 1", "customer1@example.com", "+1-555-0001"
2, "Customer 2", "customer2@example.com", "+1-555-0002"
3, "Customer 3", "customer3@example.com", "+1-555-0003"
4, "Customer 4", "customer4@example.com", "+1-555-0004"
5, "Customer 5", "customer5@example.com", "+1-555-0005"
```

```cypher
MATCH (t:SensitiveToken) RETURN count(t) AS tokens;
```

```text
tokens
2
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

The source keeps all five original customer records and both sensitive tokens.

## Connection environment

| Variable | Value used by the action |
|---|---|
| `NEO4J_HOST` | Target Service |
| `NEO4J_PORT` | 7687 |
| `NEO4J_USERNAME` | Username from the target auth Secret |
| `NEO4J_PASSWORD` | Password from the target auth Secret |

See the common [action fields and Job defaults](/docs/operatormanual/branching/post-actions.md#post-action-field-reference). The action image must execute its supplied arguments; the Branch template does not expose a container `command` override.

## Use post-actions with refresh or another cluster

[Refresh](/docs/guides/neo4j/branch/refresh.md) runs the action against every new copy before readiness. Inspect each generation’s Job using the Courier labels above; older Jobs can be removed during cleanup. For [cross-cluster branching](/docs/guides/neo4j/branch/cross-cluster/index.md), the Job runs on the target cluster and receives the target’s credentials.

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl -n demo delete branch neo4j-pa-branch --wait=true --timeout=30m
kubectl -n demo delete neo4j neo4j-pa-source --wait=true --timeout=30m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.
