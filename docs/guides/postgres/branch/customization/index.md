---
title: PostgreSQL Branch Customization | KubeDB
description: Customizing a PostgreSQL branch — credentials, post-actions, scheduled refresh, target namespace, HA, snapshot class
menu:
  docs_{{ .version }}:
    identifier: guides-pg-branch-customization
    name: Customizing Branches
    parent: guides-pg-branch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Customizing a PostgreSQL Branch

The [same-cluster walkthrough](/docs/guides/postgres/branch/same-cluster/index.md) creates the simplest possible branch: a standalone copy in the source's own namespace, reusing the source credential, taken once. Every section below is one optional field you add to that same `Branch` to change what you get — a separate credential, anonymized data, a branch that refreshes itself, a different namespace, a replicated cluster, or a specific snapshot class.

Each field is independent — apply only the ones you need, or combine several in a single `Branch`. The sections below are written as one continuous session, though, run in order against the `sample-postgres` source left behind by the walkthrough. That source starts at **1001 rows**, and the scheduled-refresh section adds one more, so the counts you see grow as you go; each section states the number it expects.

## Before You Begin

Everything here assumes the setup from the [same-cluster walkthrough](/docs/guides/postgres/branch/same-cluster/index.md): the KubeDB Courier operator installed, a CSI driver with volume-snapshot support, and the `sample-postgres` source database seeded in the `demo` namespace. Start there if you have not already.

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/branch/customization/examples](/docs/guides/postgres/branch/customization/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Give the Branch Its Own Credential

By default the branch reuses the source's root credential, so anyone who knows the production password can log into the branch. Set `spec.resetRootPassword: true` and KubeDB gives the branch a freshly generated password instead:

```yaml
spec:
  resetRootPassword: true
```

The new password is written to the branch's own `<target>-auth` Secret, and the source's password no longer works against the branch. We use it in the next section.

## Mask Sensitive Data Before Handing the Branch Out

A clone of production carries production's personal data. `spec.postActions` runs one or more containers as Jobs against the branch **after it comes up but before the `Branch` is marked `Ready`**, which is the right place to anonymize, redact, or trim data. Because the `Branch` only reports `Ready` once every post-action has succeeded, anything that gates on that — a pipeline step, a GitOps sync, a person watching `kubectl get branch` — never sees the branch before it has been sanitized.

Courier injects the branch's connection details into each Job as `PGHOST`, `PGPORT`, `PGUSER`, and `PGPASSWORD`, so a plain `postgres` image with `psql` is enough — you do not need to build a custom image.

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: dev-branch-masked
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: sample-postgres
    namespace: demo
  target:
    namespace: demo
    name: dev-postgres-masked
    storageClassName: topolvm-provisioner-thin
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  resetRootPassword: true
  postActions:
    - image: ghcr.io/appscode-images/postgres:16.1-alpine
      jobDefaults:
        backoffLimit: 3
        ttlSecondsAfterFinished: 600
        activeDeadlineSeconds: 900
        imagePullPolicy: IfNotPresent
      jobTemplate:
        metadata:
          labels:
            purpose: anonymize
        spec:
          args:
            - psql
            - -v
            - ON_ERROR_STOP=1
            - -c
            - |
              UPDATE customers
                 SET full_name = 'Masked Customer',
                     email = 'user' || id || '@example.invalid',
                     phone = '+1-555-0000';
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
  deletionPolicy: Delete
```

Here,

- `postActions[].image` is the container to run. Actions run in order, and the branch becomes `Ready` only once all of them have succeeded.
- `postActions[].jobDefaults` sets Job-level behaviour: `backoffLimit`, `ttlSecondsAfterFinished`, `activeDeadlineSeconds`, and `imagePullPolicy`.
- `postActions[].jobTemplate` customizes the Pod: labels, annotations, `args`, `env`, `resources`, node selectors, security context, volumes, and so on.

{{< notice type="warning" message="`jobTemplate.spec` accepts `args` but not `command`, so the values in `args` are passed to the image's own entrypoint. The `postgres` image's entrypoint runs whatever command it is handed, which is why `args` starts with `psql` here. For an image with no entrypoint, `args` becomes the whole command line and its first element must be the binary — for a shell script that means `sh`, then `-c`, then the script." >}}

Create it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/branch/customization/examples/dev-branch-masked.yaml
branch.courier.kubedb.com/dev-branch-masked created
```

This branch passes through an extra `ActionsRunning` phase while the Job runs. Once it is `Ready`:

```bash
$ kubectl get branch -n demo
NAME                PHASE   MODE    TARGET                FRESHNESS   AGE
dev-branch-masked   Ready   Local   dev-postgres-masked   2s          52s

$ kubectl get job -n demo
NAME                              STATUS     COMPLETIONS   DURATION   AGE
dev-postgres-masked-post-action   Complete   1/1           3s         12s

$ kubectl logs -n demo job/dev-postgres-masked-post-action
UPDATE 1001
```

The conditions show the extra milestones:

```bash
$ kubectl get branch -n demo dev-branch-masked -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
SnapshotReady           True
TargetCreated           True
TargetReady             True
RootPasswordReset       True
PostActionsCompleted    True
Ready                   True
```

And `status.resources` names the Job that ran:

```bash
$ kubectl get branch -n demo dev-branch-masked -o jsonpath='{.status.resources}'
{"authSecret":"dev-postgres-masked-auth","clonedPVCs":["data-dev-postgres-masked-0"],"postActionJob":"dev-postgres-masked-post-action"}
```

The branch's data is anonymized:

```bash
$ kubectl exec -it -n demo dev-postgres-masked-0 -c postgres -- psql -U postgres \
    -c "SELECT id, full_name, email, phone FROM customers ORDER BY id LIMIT 3;"
 id |    full_name    |         email         |    phone
----+-----------------+-----------------------+-------------
  1 | Masked Customer | user1@example.invalid | +1-555-0000
  2 | Masked Customer | user2@example.invalid | +1-555-0000
  3 | Masked Customer | user3@example.invalid | +1-555-0000
(3 rows)
```

…while the source is untouched:

```bash
$ kubectl exec -it -n demo sample-postgres-0 -c postgres -- psql -U postgres \
    -c "SELECT id, full_name, email, phone FROM customers ORDER BY id LIMIT 3;"
 id | full_name  |         email         |    phone
----+------------+-----------------------+-------------
  1 | Customer 1 | customer1@example.com | +1-555-0001
  2 | Customer 2 | customer2@example.com | +1-555-0002
  3 | Customer 3 | customer3@example.com | +1-555-0003
(3 rows)
```

Because we also set `resetRootPassword: true`, the production password does not open this branch. Verify it from another Pod, over the branch's Service — a check run *inside* the database Pod would not prove anything, since PostgreSQL trusts local connections:

```bash
$ SRC_PW=$(kubectl get secret -n demo sample-postgres-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -n demo sample-postgres-0 -c postgres -- env PGPASSWORD="$SRC_PW" \
    psql -h dev-postgres-masked.demo.svc -U postgres -c "SELECT 1;"
psql: error: connection to server at "dev-postgres-masked.demo.svc" (10.43.250.6), port 5432 failed: FATAL:  password authentication failed for user "postgres"
command terminated with exit code 2

$ BRANCH_PW=$(kubectl get secret -n demo dev-postgres-masked-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -n demo sample-postgres-0 -c postgres -- env PGPASSWORD="$BRANCH_PW" \
    psql -h dev-postgres-masked.demo.svc -U postgres -c "SELECT count(*) FROM customers;"
 count
-------
  1001
(1 row)
```

## Keep the Branch Fresh

A branch is a point-in-time copy — it does not follow the source. For a long-lived dev environment, add a cron schedule and KubeDB re-clones the branch from a fresh source snapshot at every tick.

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: dev-branch-refresh
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: sample-postgres
    namespace: demo
  target:
    namespace: demo
    name: dev-postgres-refresh
    storageClassName: topolvm-provisioner-thin
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  schedule:
    cron: "*/5 * * * *"
  historyLimit:
    success: 3
    failed: 2
  deletionPolicy: Delete
```

Here,

- `spec.schedule.cron` is a standard cron expression. We use `*/5 * * * *` so the tutorial does not take all day; in practice a nightly `0 2 * * *` is more typical.
- `spec.historyLimit` bounds how many run outcomes are kept in `status.history` (3 successful and 2 failed by default).

{{< notice type="warning" message="A refresh discards everything written to the branch since the last refresh, and the branch is unavailable for the duration of the re-clone. Use a schedule only for branches whose contents are disposable." >}}

Create it and wait for the first copy:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/branch/customization/examples/dev-branch-refresh.yaml
branch.courier.kubedb.com/dev-branch-refresh created

$ kubectl wait --for=condition=Ready branch/dev-branch-refresh -n demo --timeout=10m
branch.courier.kubedb.com/dev-branch-refresh condition met

$ kubectl exec -n demo dev-postgres-refresh-0 -c postgres -- psql -U postgres -tAc "SELECT count(*) FROM customers;"
1001
```

Now add a row to the **source** that the branch has not seen:

```bash
$ kubectl exec -n demo sample-postgres-0 -c postgres -- psql -U postgres \
    -c "INSERT INTO customers (full_name, email, phone) VALUES ('Added After Branch', 'after@example.com', '+1-555-7777');"
INSERT 0 1

$ kubectl exec -n demo dev-postgres-refresh-0 -c postgres -- psql -U postgres \
    -tAc "SELECT count(*) FROM customers WHERE full_name = 'Added After Branch';"
0
```

As expected, the branch does not have it. Wait for the next cron tick and watch the branch cycle:

```bash
$ kubectl get branch -n demo dev-branch-refresh -w
NAME                 PHASE        MODE    TARGET                 FRESHNESS   AGE
dev-branch-refresh   Ready        Local   dev-postgres-refresh   3m30s       10m
dev-branch-refresh   Refreshing   Local   dev-postgres-refresh   3m30s       10m
dev-branch-refresh   Refreshing   Local   dev-postgres-refresh   3m40s       10m
dev-branch-refresh   Refreshing   Local   dev-postgres-refresh   5m          11m
dev-branch-refresh   Refreshing   Local   dev-postgres-refresh   0s          11m
dev-branch-refresh   Ready        Local   dev-postgres-refresh   0s          11m
```

`FRESHNESS` climbs while the branch sits on its old copy and resets to `0s` the moment the new one is live.

Behind that transition, Courier took a new source snapshot, halted the branch, dropped its old volume, cloned a new one, and brought the database back up. The status shows the bookkeeping:

```bash
$ kubectl get branch -n demo dev-branch-refresh -o yaml
```

```yaml
status:
  history:
  - at: "2026-08-06T10:10:22Z"
    result: Succeeded
  - at: "2026-08-06T10:16:30Z"
    result: Succeeded
  - at: "2026-08-06T10:21:30Z"
    result: Succeeded
  lastRefreshTime: "2026-08-06T10:21:30Z"
  lastSuccessfulRefreshTime: "2026-08-06T10:21:30Z"
  mode: Local
  nextRefreshTime: "2026-08-06T10:25:00Z"
  phase: Ready
  refreshGeneration: 2
  resources:
    authSecret: dev-postgres-refresh-auth
    clonedPVCs:
    - data-dev-postgres-refresh-0
  snapshot:
    generation: 2
    members:
    - creationTime: "2026-08-06T10:20:00Z"
      name: br-dev-branch-refresh-data-sample-postgres-0-2
      readyToUse: true
      restoreSize: 1Gi
      sourcePVC: data-sample-postgres-0
    ready: true
    strategy: VolumeSnapshot
```

- `refreshGeneration` increments on every tick, and the snapshot backing the current copy is named after that generation (`…-0-2`). The previous generation's snapshot is released once the new copy is `Ready`.
- `nextRefreshTime` is the next cron deadline, and `lastSuccessfulRefreshTime` drives the `FRESHNESS` column.
- `history` holds the three successful runs so far — the first copy plus two refreshes. With `historyLimit.success: 3`, a fourth would evict the oldest entry.

The row we added to the source after the first copy is now on the branch:

```bash
$ kubectl exec -n demo dev-postgres-refresh-0 -c postgres -- psql -U postgres -c "SELECT count(*) FROM customers;"
 count
-------
  1002
(1 row)

$ kubectl exec -n demo dev-postgres-refresh-0 -c postgres -- psql -U postgres \
    -c "SELECT count(*) AS visible_on_branch FROM customers WHERE full_name = 'Added After Branch';"
 visible_on_branch
-------------------
                 1
(1 row)
```

## Branch into a Different Namespace

A same-cluster branch does not have to live in the source's namespace. Point `spec.target.namespace` somewhere else and Courier mirrors the source snapshot into that namespace so the clone can reference it:

```yaml
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: sample-postgres
    namespace: demo
  target:
    namespace: dev          # different namespace from the source
    name: dev-postgres-xns
```

The `Branch` object itself still lives next to the source, but the database it creates lands in `dev`:

```bash
$ kubectl create ns dev
namespace/dev created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/branch/customization/examples/dev-branch-xns.yaml
branch.courier.kubedb.com/dev-branch-xns created

$ kubectl wait --for=condition=Ready branch/dev-branch-xns -n demo --timeout=10m
branch.courier.kubedb.com/dev-branch-xns condition met

$ kubectl get branch -n demo dev-branch-xns
NAME             PHASE   MODE    TARGET             FRESHNESS   AGE
dev-branch-xns   Ready   Local   dev-postgres-xns   1s          61s

$ kubectl get pg -n dev
NAME               VERSION   STATUS   AGE
dev-postgres-xns   16.1      Ready    31s

$ kubectl exec -n dev dev-postgres-xns-0 -c postgres -- psql -U postgres -tAc "SELECT count(*) FROM customers;"
1002
```

A `VolumeSnapshot` is namespaced, so a PVC in `dev` cannot reference one in `demo`. Courier handles this by mirroring the source snapshot into the target namespace, and the clone is taken from the mirror:

```bash
$ kubectl get volumesnapshot -n dev
NAME                                         READYTOUSE   SOURCESNAPSHOTCONTENT                        CREATIONTIME   AGE
br-dev-branch-xns-data-sample-postgres-0-0   true         br-dev-branch-xns-data-sample-postgres-0-0   51s            51s

$ kubectl get pvc -n dev data-dev-postgres-xns-0 -o jsonpath='{.spec.dataSource}'
{"apiGroup":"snapshot.storage.k8s.io","kind":"VolumeSnapshot","name":"br-dev-branch-xns-data-sample-postgres-0-0"}
```

Deleting the `Branch` cleans up the mirrored snapshot along with everything else.

## Branch an HA Cluster

Branching a replicated `PostgreSQL` needs no extra configuration — the branch inherits the source's replica count. Courier snapshots every data volume, clones each one to its matching ordinal (`0 → 0`, `1 → 1`, `2 → 2`), and the branched cluster re-forms as a primary with streaming standbys.

Deploy a 3-replica source and branch it exactly as before:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/branch/customization/examples/sample-postgres-ha.yaml
postgres.kubedb.com/sample-postgres-ha created

$ kubectl wait --for=condition=Ready pg/sample-postgres-ha -n demo --timeout=10m
postgres.kubedb.com/sample-postgres-ha condition met

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/branch/customization/examples/dev-branch-ha.yaml
branch.courier.kubedb.com/dev-branch-ha created

$ kubectl wait --for=condition=Ready branch/dev-branch-ha -n demo --timeout=10m
branch.courier.kubedb.com/dev-branch-ha condition met

$ kubectl get branch -n demo dev-branch-ha
NAME            PHASE   MODE    TARGET            FRESHNESS   AGE
dev-branch-ha   Ready   Local   dev-postgres-ha   1s          81s
```

`status.snapshot.members` has one entry per source volume:

```bash
$ kubectl get branch -n demo dev-branch-ha -o jsonpath='{.status.snapshot}'
```

```json
{
    "members": [
        {
            "creationTime": "2026-08-06T10:17:13Z",
            "name": "br-dev-branch-ha-data-sample-postgres-ha-0-0",
            "readyToUse": true,
            "restoreSize": "1Gi",
            "sourcePVC": "data-sample-postgres-ha-0"
        },
        {
            "creationTime": "2026-08-06T10:17:13Z",
            "name": "br-dev-branch-ha-data-sample-postgres-ha-1-0",
            "readyToUse": true,
            "restoreSize": "1Gi",
            "sourcePVC": "data-sample-postgres-ha-1"
        },
        {
            "creationTime": "2026-08-06T10:17:13Z",
            "name": "br-dev-branch-ha-data-sample-postgres-ha-2-0",
            "readyToUse": true,
            "restoreSize": "1Gi",
            "sourcePVC": "data-sample-postgres-ha-2"
        }
    ],
    "ready": true,
    "strategy": "VolumeSnapshot"
}
```

Each branch volume is cloned from its ordinal-matched snapshot:

```bash
$ for i in 0 1 2; do
    echo -n "data-dev-postgres-ha-$i <- "
    kubectl get pvc -n demo data-dev-postgres-ha-$i -o jsonpath='{.spec.dataSource.name}{"\n"}'
  done
data-dev-postgres-ha-0 <- br-dev-branch-ha-data-sample-postgres-ha-0-0
data-dev-postgres-ha-1 <- br-dev-branch-ha-data-sample-postgres-ha-1-0
data-dev-postgres-ha-2 <- br-dev-branch-ha-data-sample-postgres-ha-2-0
```

And the branched cluster comes up as a working replication group:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=dev-postgres-ha -L kubedb.com/role
NAME                READY   STATUS    RESTARTS   AGE   ROLE
dev-postgres-ha-0   2/2     Running   0          45s   primary
dev-postgres-ha-1   2/2     Running   0          37s   standby
dev-postgres-ha-2   2/2     Running   0          30s   standby

$ kubectl exec -n demo dev-postgres-ha-0 -c postgres -- psql -U postgres \
    -c "SELECT client_addr, state, sync_state FROM pg_stat_replication;"
 client_addr |   state   | sync_state
-------------+-----------+------------
 10.42.0.28  | streaming | async
 10.42.0.29  | streaming | async
(2 rows)
```

If the CSI driver provides a `VolumeGroupSnapshotClass`, all volumes are captured as one crash-consistent group and `status.snapshot.strategy` reads `VolumeGroupSnapshot`. Without one, Courier falls back to per-volume snapshots (`strategy: VolumeSnapshot`), as above.

## Pin a VolumeSnapshotClass

By default Courier picks the default `VolumeSnapshotClass` for the source volume's CSI driver. Set `spec.volumeSnapshotClassName` to use a specific one — for example a class with different deletion or secret settings:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: dev-branch-vsc
  namespace: demo
spec:
  volumeSnapshotClassName: topolvm-vsc-explicit   # instead of the driver's default class
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: sample-postgres
    namespace: demo
  target:
    namespace: demo
    name: dev-postgres-vsc
    storageClassName: topolvm-provisioner-thin
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Delete
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/branch/customization/examples/dev-branch-vsc.yaml
branch.courier.kubedb.com/dev-branch-vsc created

$ kubectl wait --for=condition=Ready branch/dev-branch-vsc -n demo --timeout=10m
branch.courier.kubedb.com/dev-branch-vsc condition met
```

Only the pinned branch uses that class; every other branch created in this guide left the field unset and took the driver's default, `topolvm-provisioner-thin`:

```bash
$ kubectl get volumesnapshot -n demo
NAME                                             READYTOUSE   SOURCEPVC                   RESTORESIZE   SNAPSHOTCLASS              AGE
br-dev-branch-ha-data-sample-postgres-ha-0-0     true         data-sample-postgres-ha-0   1Gi           topolvm-provisioner-thin   6m9s
br-dev-branch-ha-data-sample-postgres-ha-1-0     true         data-sample-postgres-ha-1   1Gi           topolvm-provisioner-thin   6m9s
br-dev-branch-ha-data-sample-postgres-ha-2-0     true         data-sample-postgres-ha-2   1Gi           topolvm-provisioner-thin   6m9s
br-dev-branch-masked-data-sample-postgres-0-0    true         data-sample-postgres-0      1Gi           topolvm-provisioner-thin   15m
br-dev-branch-refresh-data-sample-postgres-0-2   true         data-sample-postgres-0      1Gi           topolvm-provisioner-thin   3m21s
br-dev-branch-vsc-data-sample-postgres-0-0       true         data-sample-postgres-0      1Gi           topolvm-vsc-explicit       43s
br-dev-branch-xns-data-sample-postgres-0-0       true         data-sample-postgres-0      1Gi           topolvm-provisioner-thin   8m14s
```

The class must belong to the same CSI driver that provisioned the source volumes.

## Troubleshooting

| Symptom | Cause and fix |
|---|---|
| `Branch` is `Failed` with `reason: PostActionFailed` | A post-action container exited non-zero. Courier suspends the Job so the failed Pod and its logs survive — read them with `kubectl logs -n <ns> job/<target>-post-action`, then fix `spec.postActions` and re-apply. The edit triggers a retry. |
| A post-action Pod fails to start with an exec error | `jobTemplate.spec` has no `command` field, so `args` are passed to the image's entrypoint. For an entrypoint-less image, `args` must start with the binary, e.g. `["sh", "-c", "…"]`. |
| A post-action succeeded but the branch still shows raw data after a refresh | Each refresh generation runs its own Job (`<target>-post-action-gen<N>`). Check that the latest generation's Job completed — Courier labels the Jobs it owns with the branch name: `kubectl get job -n <ns> -l courier.kubedb.com/branch=<branch>`. |
| `FRESHNESS` is not advancing on a scheduled branch | Check `status.history` and `status.nextRefreshTime`, and the Courier logs: `kubectl logs -n <ns> deploy/kubedb-courier`. |
| A cross-namespace branch's PVC stays `Pending` | The mirrored snapshot in the target namespace is not ready yet, or the target StorageClass uses a different CSI driver than the source. Check `kubectl get volumesnapshot -n <target-ns>`. |

For problems with the core branch flow — snapshotting, cloning, provisioning — see the [walkthrough's troubleshooting table](/docs/guides/postgres/branch/same-cluster/index.md#troubleshooting).

## Cleaning up

This page created five branches alongside the source in `demo`, plus one extra database and one extra namespace. Delete each branch by name rather than `--all`, so an unrelated branch left over from another walkthrough in the same `demo` namespace is not swept up. `kubectl delete` on a `Branch` blocks until Courier's cleanup finalizer clears, so there is nothing further to wait for:

```bash
$ kubectl delete branch -n demo dev-branch-masked dev-branch-refresh dev-branch-xns dev-branch-ha dev-branch-vsc
$ kubectl delete pg -n demo sample-postgres-ha
$ kubectl delete ns dev
```

Deleting a `Branch` with the default `Delete` policy removes the branched database, its cloned volumes, its credential, the snapshots, and any post-action Jobs — including, for a cross-namespace branch, the snapshot mirrored into the target namespace. The `sample-postgres-ha` source created for the HA section is an ordinary database and has to be deleted directly.

## Next Steps

- Go back to the [same-cluster walkthrough](/docs/guides/postgres/branch/same-cluster/index.md).
- See every field of the [Branch CRD](/docs/guides/postgres/concepts/branch.md).
- Read the [PostgreSQL Branching Overview](/docs/guides/postgres/branch/overview/index.md) for how branching works under the hood.
