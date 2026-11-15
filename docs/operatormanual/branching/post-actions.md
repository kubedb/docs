---
title: Running Branch Post-Actions
menu:
  docs_{{ .version }}:
    identifier: branching-post-actions
    name: Post-Actions
    parent: branching
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Branch Post-Actions

Post-actions prepare cloned data for the people and applications that will use a branch. Define containers in `spec.postActions`, and Courier runs them as Kubernetes Jobs against the target database. For example, you can:

- Mask names, email addresses, and other personal information before giving developers or AI agents access.
- Drop tables containing sensitive information that the branch does not need.
- Delete confidential rows or data outside the scope of a test.
- Insert test fixtures or application-specific seed data.

These changes run against the branch database and leave the source data unchanged. The feature applies to all supported database engines; the client image and preparation commands depend on the engine.

This tutorial uses PostgreSQL to run one post-action Job that anonymizes five customer records, then drops a table containing sensitive tokens. You will deploy and seed the source, create the Branch, inspect the Job, and compare the source and branch data. No database from another tutorial is required.

## Execution and readiness

Courier runs the actions **in list order**, creating a separate Kubernetes Job for each action in the target namespace. The next action starts only after the preceding Job succeeds. Actions start after the target database is ready and requested credential isolation is complete. The Branch stays in `ActionsRunning` until all actions succeed, then becomes `Ready`.

The target database Service can exist before the actions finish. Applications and automation consuming prepared data should wait for the **Branch** to become `Ready`. Branch readiness provides a completion signal; it does not restrict connections to the Service while data preparation is running.

## Before You Begin

- Complete the [single-cluster Courier installation](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD.
- Have a working Ceph RBD StorageClass with copy-on-write cloning and a compatible snapshot class. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Ensure PostgreSQL version `18.3` is available in the installed KubeDB catalog, and use branching-capable KubeDB and database images.
- Run the commands below against the same Kubernetes context. Your account must be able to create databases and Branches, inspect Jobs and pods, and open a shell in the database pods.

The examples use the StorageClass name `ceph-rbd`. Substitute your Ceph RBD class consistently in the source and Branch manifests. The namespaces are `demo` for the source and `dev` for the target; use the database names below for this tutorial only.

Create the namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `post-actions-source.yaml`. It creates a three-replica PostgreSQL source named `pa-source`:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pa-source
  namespace: demo
spec:
  version: "18.3"
  replicas: 3
  standbyMode: Hot
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

Apply the manifest:

```shell
kubectl apply -f post-actions-source.yaml
```

Once the source is ready, find its primary pod:

```shell
kubectl -n demo get pods -l app.kubernetes.io/instance=pa-source,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Example output:

```text
pa-source-0
```

Use the returned pod name in the commands below; this example uses `pa-source-0`.

## Seed and inspect the source data

From your terminal, open a shell in the source primary:

```shell
kubectl -n demo exec -it pa-source-0 -c postgres -- sh
```

Inside the pod, connect to database `postgres`. Local socket connections in this setup use trust authentication, so no password is required:

```shell
psql -U postgres -d postgres
```

At the `postgres=#` prompt, create five customer records and a table of sample sensitive tokens:

```sql
CREATE TABLE customers (
    id integer PRIMARY KEY,
    full_name text NOT NULL,
    email text NOT NULL,
    phone text NOT NULL
);
INSERT INTO customers (id, full_name, email, phone) VALUES
    (1, 'Customer 1', 'customer1@example.com', '+1-555-0001'),
    (2, 'Customer 2', 'customer2@example.com', '+1-555-0002'),
    (3, 'Customer 3', 'customer3@example.com', '+1-555-0003'),
    (4, 'Customer 4', 'customer4@example.com', '+1-555-0004'),
    (5, 'Customer 5', 'customer5@example.com', '+1-555-0005');

CREATE TABLE sensitive_tokens (
    id integer PRIMARY KEY,
    token text NOT NULL
);
INSERT INTO sensitive_tokens (id, token) VALUES
    (1, 'sample-token-1'),
    (2, 'sample-token-2');
```

In the same `psql` session, inspect the original data:

```sql
SELECT * FROM customers ORDER BY id;
SELECT * FROM sensitive_tokens ORDER BY id;
```

Expected results:

```text
 id | full_name  |         email         |    phone
----+------------+-----------------------+-------------
  1 | Customer 1 | customer1@example.com | +1-555-0001
  2 | Customer 2 | customer2@example.com | +1-555-0002
  3 | Customer 3 | customer3@example.com | +1-555-0003
  4 | Customer 4 | customer4@example.com | +1-555-0004
  5 | Customer 5 | customer5@example.com | +1-555-0005
(5 rows)

 id |     token
----+----------------
  1 | sample-token-1
  2 | sample-token-2
(2 rows)
```

Leave `psql` with `\q`, then run `exit` in the pod shell to return to your terminal.

## Create a Branch with post-actions

Save the following as `branch-post-actions.yaml`. The Branch lives in `demo` and creates the target database `dev/pa-branch`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pa-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: pa-source
    namespace: demo
  target:
    namespace: dev
    name: pa-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  postActions:
    - image: postgres:18.3-alpine
      jobDefaults:
        imagePullPolicy: IfNotPresent
        backoffLimit: 3
        activeDeadlineSeconds: 900
        ttlSecondsAfterFinished: 3600
      jobTemplate:
        metadata:
          labels:
            branch-action: pa-anonymize
        spec:
          # First -c: anonymize customer data.
          # Second -c: drop the sensitive table only if anonymization succeeds.
          args: [psql, -w, -v, ON_ERROR_STOP=1, -d, postgres,
                 -c, "UPDATE customers SET full_name = 'Masked Customer', email = 'user' || id || '@example.invalid', phone = '+1-555-0000';",
                 -c, "DROP TABLE IF EXISTS sensitive_tokens;"]
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
  deletionPolicy: Delete
```

The single `postActions` entry creates one Job. Its `psql` process executes the two `-c` commands in order: anonymize the customer data, then drop the sensitive table. Both operations are safe to repeat: the update produces the same values on each run, and the drop allows the table to be absent.

The image contains `psql`, which reads the connection environment supplied by Courier. `-d postgres` selects the database, `-w` prevents an interactive password prompt, and `ON_ERROR_STOP=1` stops execution on a SQL error. If the update fails, the drop command does not run and the container exits unsuccessfully.

Apply the Branch and wait for all preparation to finish:

```shell
kubectl apply -f branch-post-actions.yaml
kubectl -n demo wait branch/pa-branch \
  --for=condition=Ready --timeout=20m
kubectl -n demo get branch pa-branch
```

Representative output after successful completion:

```text
NAME        PHASE   MODE    TARGET      FRESHNESS   AGE
pa-branch   Ready   Local   pa-branch   2s          5m
```

The Branch passes through `ActionsRunning` while the Job executes. It may finish too quickly for you to see that phase in an interactive watch. Names and data results below follow the manifests; ages, durations, and pod names vary between runs.

## Inspect the action Job

Courier labels its Jobs with the Branch name and source namespace. List this Branch's Jobs in the target namespace:

```shell
kubectl -n dev get jobs \
  -l courier.kubedb.com/branch=pa-branch,courier.kubedb.com/branch-namespace=demo
```

Representative output:

```text
NAME                    STATUS     COMPLETIONS   DURATION   AGE
pa-branch-post-action   Complete   1/1           13s         30s
```

The Job contains one container named `actions`. Inspect its logs:

```shell
kubectl -n dev logs job/pa-branch-post-action -c actions
```

The same Job logs both SQL results:

```text
UPDATE 5
DROP TABLE
```

The custom label in `jobTemplate.metadata` identifies the action pod:

```shell
kubectl -n dev get pods -l branch-action=pa-anonymize
```

Check the Branch's conditions:

```shell
kubectl -n demo get branch pa-branch \
  -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
```

Successful completion includes these conditions; their display order may vary:

```text
SnapshotReady           True
TargetCreated           True
TargetReady             True
PostActionsCompleted    True
Ready                   True
```

Inspect the resource references recorded by Courier:

```shell
kubectl -n demo get branch pa-branch -o jsonpath='{.status.resources}{"\n"}'
```

The `authSecret` and `clonedPVCs` fields identify target resources. `postActionJob` identifies `pa-branch-post-action` here. For Branches with multiple `postActions` entries, this field references the first Job; use the Job labels to find all of them.

## Verify that only the branch was changed

From your terminal, find the branch primary:

```shell
kubectl -n dev get pods -l app.kubernetes.io/instance=pa-branch,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Use the returned name below; this example uses `pa-branch-0`:

```shell
kubectl -n dev exec -it pa-branch-0 -c postgres -- bash
```

Inside the pod, start `psql`:

```shell
psql -U postgres -d postgres -v ON_ERROR_STOP=1
```

At the `postgres=#` prompt, inspect all five customers and check that the sensitive table was removed:

```sql
SELECT * FROM customers ORDER BY id;
SELECT to_regclass('public.sensitive_tokens') IS NULL AS sensitive_table_removed;
```

Expected results:

```text
 id |    full_name    |         email         |    phone
----+-----------------+-----------------------+-------------
  1 | Masked Customer | user1@example.invalid | +1-555-0000
  2 | Masked Customer | user2@example.invalid | +1-555-0000
  3 | Masked Customer | user3@example.invalid | +1-555-0000
  4 | Masked Customer | user4@example.invalid | +1-555-0000
  5 | Masked Customer | user5@example.invalid | +1-555-0000
(5 rows)

 sensitive_table_removed
-------------------------
 t
(1 row)
```

The five customer records are preserved with anonymized values, and the sensitive table no longer exists on the branch. Leave `psql` with `\q`, then run `exit` in the pod shell to return to your terminal.

Find the source primary again in case a failover occurred:

```shell
kubectl -n demo get pods -l app.kubernetes.io/instance=pa-source,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Open a shell in that pod, substituting its name if different:

```shell
kubectl -n demo exec -it pa-source-0 -c postgres -- sh
```

Inside the pod, start `psql`:

```shell
psql -U postgres -d postgres -v ON_ERROR_STOP=1
```

Query the original tables:

```sql
SELECT * FROM customers ORDER BY id;
SELECT * FROM sensitive_tokens ORDER BY id;
```

The source returns the same five original customer records and two sample tokens shown in [Seed and inspect the source data](#seed-and-inspect-the-source-data). This verifies both that anonymization was limited to the branch and that dropping its sensitive table did not remove the source table.

Leave `psql` with `\q`, then run `exit` in the pod shell to return to your terminal.

## Post-action field reference

### Action configuration

The fields below are relative to the Branch specification:

| Field | Type | Default or omitted behavior | Details |
|---|---|---|---|
| `spec.postActions` | Array of action objects | Omitted or empty: no action Jobs run. | Ordered preparation steps against the target database. Each entry runs as its own Job; all must succeed before Branch readiness. |
| `spec.postActions[].image` | String | No default; supply an image for every action. | Image containing the database client or application tooling and an entrypoint capable of running the supplied arguments. Courier does not install a client into the image. |
| `spec.postActions[].jobDefaults` | Object | Omitted: Kubernetes defaults apply to the generated Job and container. | Per-action execution settings described below. One action's settings do not configure the other actions. |
| `spec.postActions[].jobTemplate` | Object | Omitted: no user template overrides; the image's default command runs with Courier's connection environment. | Configures the action pod and container arguments, resources, labels, and other supported runtime settings. Supply arguments that cause your image to perform the intended preparation and exit. |

`jobTemplate` uses KubeDB's pod-template shape. Container options such as `args` and `resources` go directly under `jobTemplate.spec`; there is no `containers` list or configurable `command` field there. Arguments are passed to the image entrypoint. In this example, the `postgres` image entrypoint executes `psql` because it is the first argument. The resource requests reserve CPU and memory for each action container. Courier sets the action pod's restart policy to `Never`.

The labels under `jobTemplate.metadata` identify the pods in the inspection commands above. For the template structure available in your installed CRD, use:

```shell
kubectl explain branch.spec.postActions.jobTemplate --recursive
```

### Job defaults

These fields are relative to each `spec.postActions[]` entry's `jobDefaults` object:

| Field | Type | Default or omitted behavior | Details |
|---|---|---|---|
| `imagePullPolicy` | String | `IfNotPresent` when `jobDefaults` is present and this field is omitted. If the whole object is omitted, Kubernetes chooses the pull policy from the image reference. | Accepts `Always`, `IfNotPresent`, or `Never`. Controls whether the action image is pulled. The example uses a tagged image with `IfNotPresent`. |
| `backoffLimit` | Integer | `6`. | Kubernetes Job retry budget. The example sets `3`. Courier can suspend an action Job when its first failed Pod is observed, so this value does not guarantee that all retries will run. |
| `activeDeadlineSeconds` | Integer, seconds | Unset: no Job-level execution deadline. | Maximum elapsed time the Job may remain active, including retries. Use a positive duration that accommodates the SQL or script. The example allows `900` seconds (15 minutes) per action. |
| `ttlSecondsAfterFinished` | Integer, seconds | Unset: no automatic cleanup through the Job TTL mechanism. | Time to retain a finished Job and its pods before TTL cleanup. `0` makes it eligible for immediate cleanup. The example retains them for `3600` seconds (one hour), allowing Courier to observe completion and you to inspect logs. |

With `jobDefaults: {}`, the Branch CRD supplies `imagePullPolicy: IfNotPresent` and `backoffLimit: 6`. If `jobDefaults` is absent, Kubernetes supplies the Job backoff default of `6`; the image pull policy is normally `IfNotPresent` for a non-`latest` tag and `Always` for `latest` or an omitted tag. Set it explicitly when the distinction matters.

Choose a TTL long enough for Courier to observe Job completion and for readers or automation to collect logs. Courier also removes action Jobs during Branch cleanup and removes stale generations after a successful refresh, so the TTL does not guarantee a minimum retention period.

## Connection environment

Courier injects engine-specific connection variables into each action container. For this PostgreSQL example:

| Variable | Value or source | Use |
|---|---|---|
| `PGHOST` | Target Service name, `pa-branch`. | Connects to the branch in the Job's target namespace. |
| `PGPORT` | `5432`. | PostgreSQL Service port. |
| `PGUSER` | `username` key of the target auth Secret. | Authenticates as the branch's database user. |
| `PGPASSWORD` | `password` key of the target auth Secret. | Supplies the branch's current password through a Secret reference. |

`PGDATABASE` is not injected. This tutorial passes `-d postgres` explicitly. Other database engines receive their own client connection environment; the `PG*` names are specific to PostgreSQL. Keep passwords out of image layers, SQL text, and logs.

## Using post-actions with refresh or another cluster

A [scheduled refresh](/docs/operatormanual/branching/refresh.md) replaces branch data from a fresh source snapshot and reruns the full action list before the Branch returns to `Ready`. Write actions that can be repeated against freshly cloned data and retried without producing unwanted changes. The deterministic update and `DROP TABLE IF EXISTS` in this tutorial meet that requirement.

Refresh generations have separate Jobs. For example, generation `1` uses `pa-branch-post-action-gen1` to run both commands. Inspect the current generation rather than assuming the initial Job name remains current. Courier removes older action Jobs after successful replacement.

The same feature applies to [cross-cluster branching](/docs/guides/postgres/branch/cross-cluster/index.md). Apply the original Branch on the source cluster; action Jobs execute in the target namespace on the **target** cluster, where the target Service, auth Secret, and any referenced ServiceAccounts or image pull Secrets must be available. Inspect Jobs and logs using the target cluster context. The source Branch reports readiness after target-side preparation completes.

## Cleanup

Delete this tutorial's Branch and wait for Courier to finish cleanup:

```shell
kubectl -n demo delete branch pa-branch --wait=true --timeout=15m
```

With `deletionPolicy: Delete`, Courier removes the target database, cloned storage, auth Secret, snapshots, and action Jobs. The source remains available.

Remove the tutorial source when you are finished. Its `WipeOut` policy deletes its data:

```shell
kubectl -n demo delete postgres pa-source --wait=true --timeout=15m
```

These commands leave the `demo` and `dev` namespaces available for other resources.
