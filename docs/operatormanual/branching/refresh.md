---
title: Refreshing Database Branches
menu:
  docs_{{ .version }}:
    identifier: branching-refresh
    name: Refresh
    parent: branching
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Refreshing Database Branches

A scheduled Branch periodically replaces its target data with a new snapshot of the source. Use refresh when development or test data should follow production on a predictable cadence.

## Schedule and history

Configure `spec.schedule.cron` using standard five-field cron syntax. For example, `0 2 * * *` requests a daily refresh at 02:00 in the Courier controller's timezone. Omit the schedule for a one-time branch. Use `spec.historyLimit.success` and `spec.historyLimit.failed` to bound run history; their defaults are `3` and `2`.

History limits bound status entries, rather than preserving that many restorable data copies.

## What happens during refresh

1. Courier takes a fresh snapshot of the source while retaining the previous snapshot generation.
2. Courier pauses the target and replaces its volumes with clones of the new snapshots.
3. KubeDB restarts the target database.
4. Courier completes password isolation when requested and reruns all post-actions against the refreshed data.
5. The Branch becomes `Ready`; previous generation resources are released after successful replacement.

**Refresh discards writes made only to the branch.** The target is unavailable while its storage is replaced and the database restarts. Choose a schedule that fits the target application's maintenance requirements. The source continues to serve its own workload.

## Cross-cluster refresh

The source agent owns the refresh schedule. It snapshots the source and delivers the next generation through the OCM hub. The target agent performs the target replacement and reports completion. Both agents and the hub delivery path must remain available.

## Observe freshness

Inspect the Branch on the source cluster:

```bash
kubectl -n demo get branch
kubectl -n demo get branch pg-dev -o yaml
```

`status.lastRefreshTime` records an attempt. `status.lastSuccessfulRefreshTime` records a completed run and drives the `Freshness` column. `status.nextRefreshTime` shows the next scheduled refresh, and `status.refreshGeneration` identifies the current generation. A failed run must not be interpreted as fresh data.

Post-action Job retries follow their configured Job limits. Once a post-action Job fails terminally, the Branch remains failed until its specification is corrected. Other transient failures of scheduled branches are retried with backoff; inspect status conditions before changing the schedule.

For runnable manifests and data checks, follow [Refreshing a PostgreSQL Branch](/docs/guides/postgres/branch/refresh.md).
