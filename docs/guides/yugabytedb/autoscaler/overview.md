---
title: YugabyteDB Autoscaler
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-autoscaler-overview
    name: Overview
    parent: yugabytedb-autoscaler
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB Autoscaler

> **Not verified on a cluster.** The autoscaler for YugabyteDB has not been run on a real cluster. What was checked: the controller has unit tests with fake clients (role spec selection, the VerticalScaling and VolumeExpansion requests it builds, the rules for creating ops requests, and a storage autoscaling run against a fake metrics server) and they passed on the machine that built it; the API webhook has a unit test. The API types live on the branch `yugabytedb-autoscaler` of `apimachinery`, which is not yet merged into the branch `yugabytedb`. Nothing here was observed against a real VPA recommender or real volumes.

`YugabyteDBAutoscaler` (`autoscaling.kubedb.com/v1alpha1`, short name `ybscaler`) adjusts the CPU and memory, and the volume size, of a YugabyteDB universe by creating `YugabyteDBOpsRequest` objects. Compute and storage are configured **per role**.

```
spec:
  databaseRef: {name: ...}
  opsRequestOptions: {apply, timeout, maxRetries}
  compute:
    nodeTopology: ...
    master:      {trigger, minAllowed, maxAllowed, controlledResources, ...}
    tserver:     {...}
    readReplica: {...}
  storage:
    master:      {trigger, usageThreshold, scalingThreshold, scalingRules, upperBound, expansionMode}
    tserver:     {...}
    readReplica: {...}
```

The role names are `master` (yb-master), `tserver` (yb-tserver) and `readReplica`. **The controller currently handles only `master` and `tserver`**; the `readReplica` field exists in the API but the controller does not act on it yet.

## How the requests are made

- **Compute.** The autoscaler creates one VPA per role, named after the PetSet, and turns its recommendation into a `VerticalScaling` request. The request always uses `mode: Restart`, never `InPlace`, because the memory of YugabyteDB cannot be resized without a restart (see [in-place vertical scaling](/docs/guides/yugabytedb/scaling/vertical-scaling/in-place.md)). Every resize therefore rolls the pods of the role one at a time.
- **Storage.** The autoscaler reads the volume usage of the data volumes of a role (one PVC per data volume), takes the highest usage, and creates a `VolumeExpansion` request for that role only when it reaches `usageThreshold`. The PVCs of a role are found by the role label of the database; this was not checked on a real cluster.
- **Guard rails.** No new request is created while a `Pending` or `Progressing` request of the database exists, and a request identical to the last one the autoscaler created is skipped. `opsRequestOptions` (`apply`, `timeout`, `maxRetries`) is copied into the requests it creates.

See [compute](/docs/guides/yugabytedb/autoscaler/compute/autoscale.md) and [storage](/docs/guides/yugabytedb/autoscaler/storage/autoscale.md) autoscaling.
