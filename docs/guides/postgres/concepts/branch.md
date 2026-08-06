---
title: Branch CRD
menu:
  docs_{{ .version }}:
    identifier: pg-branch-concepts
    name: Branch
    parent: pg-concepts-postgres
    weight: 27
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch

## What is Branch

`Branch` is a Kubernetes `Custom Resource Definition` (CRD). It provides a declarative way to create an
instant, writable copy of a running KubeDB-managed database — the way you would branch a Git
repository. You describe the source database and the target you want, and the kubedb-courier operator
snapshots the source's volumes, clones them copy-on-write, and provisions a second database on top of
the clone.

One `Branch` object is one branch. It owns the database it creates: the branch's lifecycle, its optional
refresh schedule, and its teardown are all driven from this single object. The branched database itself
is an ordinary KubeDB database that can be scaled, reconfigured, backed up, and monitored like any
other.

`Branch` lives in the `courier.kubedb.com/v1alpha1` API group. Branching currently supports KubeDB
`PostgreSQL`.

## Branch Spec

As with all other Kubernetes objects, a `Branch` needs `apiVersion`, `kind`, and `metadata` fields. It
also needs a `.spec` section. Below is an example `Branch` object exercising every optional field.

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: dev-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: sample-postgres
    namespace: demo
  target:
    # clusterName omitted — same-cluster branch
    namespace: demo
    name: dev-postgres
    storageClassName: gp3
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: postgres-ca-issuer
  resetRootPassword: true
  postActions:
    - image: ghcr.io/appscode-images/postgres:16.1-alpine
      jobDefaults:
        imagePullPolicy: IfNotPresent
        backoffLimit: 3
        ttlSecondsAfterFinished: 600
        activeDeadlineSeconds: 900
      jobTemplate:
        metadata:
          labels:
            purpose: anonymize
        spec:
          args:
            - psql
            - -c
            - "UPDATE customers SET email = 'redacted';"
  schedule:
    cron: "0 2 * * *"
  historyLimit:
    success: 3
    failed: 2
  volumeSnapshotClassName: topolvm-vsc-explicit
  deletionPolicy: Delete
```

### spec.source

`spec.source` is a required field identifying the KubeDB database to branch **from**. The source is
never paused, halted, or modified — it keeps serving traffic for the whole operation.

- `databaseRef` — a reference to the source KubeDB database.
  - `apiGroup` — the API group of the source database, `kubedb.com`.
  - `kind` — the kind of the source database, `Postgres`.
  - `name` — the name of the source database.
- `namespace` — the namespace of the source database. Defaults to the `Branch`'s own namespace when
  omitted.

### spec.target

`spec.target` is a required field describing the database to create. It carries **only what differs from
the source** — everything else (database version, replica count, custom configuration, TLS settings, pod
template) is copied from the source database.

- `name` — the name of the branched database. Required.
- `namespace` — the namespace for the branched database. Required; it may differ from the source's, in
  which case courier mirrors the source snapshot into the target namespace so the clone can reference
  it.
- `clusterName` — the target cluster. Leave it empty (or omit it) for a same-cluster branch; a different
  cluster name selects a cross-cluster branch.
- `storageClassName` — the `StorageClass` for the branch's volumes. It must be backed by the **same CSI
  driver** as the source, because a snapshot can only be restored by the driver that created it.
- `resources` — cpu/memory requests and limits for the branched database, typically smaller than the
  source's.
- `issuerRef` — a reference to a cert-manager `Issuer` or `ClusterIssuer` in the target namespace. TLS
  secrets are namespace-scoped, so a branch cannot reuse the source's certificates; when the source has
  TLS enabled, KubeDB mints fresh certificates for the branch from this issuer. Required for a
  TLS-enabled source, ignored otherwise.

### spec.resetRootPassword

`spec.resetRootPassword` decides which credential the branch uses. Default `false`.

- `false` — the source's auth Secret is copied to the branch, so the same password opens both.
- `true` — the branch gets a freshly generated credential in its own `<target>-auth` Secret, and the
  source's password no longer authenticates against the branch. Use this whenever the branch is handed
  to someone who should not hold production credentials.

### spec.postActions

`spec.postActions` is an optional list of containers run as Jobs against the branched database **after
it is up but before the `Branch` is marked `Ready`** — the place to anonymize, redact, or trim cloned
data. Actions run in order, and the branch only becomes `Ready` once every one of them has succeeded.

courier injects the branch's connection details into each Job as `PGHOST`, `PGPORT`, `PGUSER`, and
`PGPASSWORD`, so a stock `postgres` image running `psql` is usually enough.

- `image` — the container image to run.
- `jobDefaults` — Job-level settings.
  - `imagePullPolicy` — image pull policy for the Job's pod. Defaults to `IfNotPresent`.
  - `backoffLimit` — number of retries before the Job is marked failed. Defaults to `6`.
  - `ttlSecondsAfterFinished` — how long the finished Job is retained. Unset means the Job is kept until
    the branch is torn down.
  - `activeDeadlineSeconds` — how long the Job may run before it is terminated. Unset means no deadline.
- `jobTemplate` — pod-level customization: `metadata.labels`, `metadata.annotations`, and a pod spec
  supporting `args`, `env`, `resources`, `securityContext`, scheduling, volumes, and so on.

> `jobTemplate.spec` exposes `args` but **not** `command`, so the values in `args` are passed to the
> image's own entrypoint. For an image with no entrypoint, `args` becomes the whole command line and its
> first element must be the binary.

On a scheduled branch every refresh re-clones the data, so every post-action runs again on the new copy —
each generation gets its own Job.

### spec.schedule

`spec.schedule` is optional. Omit it for a one-shot branch that is never refreshed.

- `cron` — a standard cron expression. At each tick courier takes a fresh source snapshot **first**,
  then halts the branch, discards its cloned volumes, re-clones from the new snapshot, brings the branch
  back up, and re-runs any post-actions. Snapshotting before halting anything means a failure at that
  stage leaves the branch running on its previous copy.

A refresh replaces the branch's data: anything written to the branch since the last refresh is
discarded, and the branch is briefly unavailable while it is re-cloned.

### spec.historyLimit

`spec.historyLimit` bounds the run history kept in `status.history`.

- `success` — number of successful runs to retain. Defaults to `3`.
- `failed` — number of failed runs to retain. Defaults to `2`.

### spec.volumeSnapshotClassName

`spec.volumeSnapshotClassName` pins the `VolumeSnapshotClass` used wherever courier creates a
`VolumeSnapshot` — snapshotting the source volumes and, for a cross-namespace branch, the mirrored
snapshot in the target namespace. It must belong to the CSI driver backing the source volumes. When
empty, courier resolves the driver's default class.

### spec.deletionPolicy

`spec.deletionPolicy` decides the branched database's fate when the `Branch` object is deleted. Defaults
to `Delete`.

- `Delete` — tears the branch down completely: the target database, its cloned volumes, its auth Secret,
  the snapshots, and any post-action Jobs. The source is untouched.
- `Orphan` — keeps the branched database as a standalone KubeDB database. Its branch ownership metadata
  is stripped and its auth Secret is retained, so it keeps working — it is simply no longer managed by a
  `Branch`.

## Branch Status

`status` describes the current state of the branch.

- `phase` — the lifecycle phase: `Pending`, `Snapshotting`, `Cloning`, `Provisioning`, `ActionsRunning`,
  `Ready`, `Refreshing`, `Deleting`, or `Failed`.
- `mode` — how this operator participates: `Local` for a same-cluster branch, `Initiator`/`Creator` for
  the two sides of a cross-cluster branch.
- `conditions` — the milestones reached, in order: `SnapshotReady`, `TargetCreated`, `TargetReady`,
  `RootPasswordReset` (when `spec.resetRootPassword` is set), `PostActionsCompleted` (when
  `spec.postActions` is set), and `Ready`. A `Failed` condition carries the reason —
  `PostActionFailed` for a post-action failure, which is terminal until the spec changes.
- `resources` — the objects the branch owns, which is exactly what teardown removes:
  `authSecret`, `clonedPVCs`, `configSecret`, and the current generation's `postActionJob`.
- `snapshot` — the snapshot set backing the current copy.
  - `strategy` — `VolumeGroupSnapshot` when the CSI driver offers a `VolumeGroupSnapshotClass`
    (group-consistent across all volumes), otherwise `VolumeSnapshot` (per-volume fallback).
  - `generation` — the refresh generation these snapshots belong to.
  - `groupRef` — the `VolumeGroupSnapshot` name, set only for the `VolumeGroupSnapshot` strategy.
  - `ready` — true once every member snapshot is `readyToUse`.
  - `members` — one entry per source data volume, ordered by ordinal and aligned to the cloned target
    volumes, each with `name`, `sourcePVC`, `readyToUse`, `restoreSize`, and `creationTime`.
- `refreshGeneration` — increments on each scheduled refresh.
- `lastRefreshTime` — the time of the last refresh **attempt**, whatever its outcome.
- `lastSuccessfulRefreshTime` — the time of the last **successful** refresh. This is how current the
  branch data is, and it drives the `FRESHNESS` print column.
- `nextRefreshTime` — the next scheduled refresh, computed from `spec.schedule.cron`. Empty for a
  one-shot branch.
- `history` — the bounded run history, each entry carrying `at`, `result`, and a `message` for a failed
  run.

## Next Steps

- Learn how branching works in the [PostgreSQL Branching Overview](/docs/guides/postgres/branch/overview/index.md).
- Follow the hands-on walkthrough: [Branch a PostgreSQL Database in the Same Cluster](/docs/guides/postgres/branch/same-cluster/index.md).
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
