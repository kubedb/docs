---
title: YugabyteDB Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-storage-autoscaling-guide
    name: Autoscale Storage
    parent: yugabytedb-storage-autoscaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB Storage Autoscaling

> **Not verified on a cluster.** The autoscaler for YugabyteDB has not been run on a real cluster. What was checked: the controller has unit tests with fake clients (role spec selection, the VerticalScaling and VolumeExpansion requests it builds, the rules for creating ops requests, and a storage autoscaling run against a fake metrics server) and they passed on the machine that built it; the API webhook has a unit test. The API types live on the branch `yugabytedb-autoscaler` of `apimachinery`, which is not yet merged into the branch `yugabytedb`. Nothing here was observed against a real VPA recommender or real volumes.

The storage autoscaler creates `VolumeExpansion` requests, so everything on the [volume expansion](/docs/guides/yugabytedb/volume-expansion/volume-expansion.md) page applies, **including that volume expansion itself is not verified to completion** on a cluster. The StorageClass must allow expansion.

Fields of each role: `trigger`, `usageThreshold` (percent of the volume in use that triggers a request, default 80), `scalingThreshold` (percent by which the volume grows, default 50), `scalingRules` (growth that depends on the size, for example a percentage up to a size and an absolute amount after it), `upperBound` (maximum size) and `expansionMode` (`Online` or `Offline`, required).

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: YugabyteDBAutoscaler
metadata:
  name: yb-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: yb-demo
  opsRequestOptions:
    apply: IfReady
    timeout: 20m
    maxRetries: 1
  storage:
    tserver:
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
      upperBound: 100Gi
      expansionMode: Online
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/autoscaler/yugabytedb-storage-autoscaler.yaml
```

This manifest was written from the API types for this guide; it was not applied. It watches the tservers only. A database with several data volumes per pod (`dataVolumes`) is judged by its fullest volume, and the request expands the role as a whole. YugabyteDB stops accepting writes when a disk has less than 5% free space, so keep `usageThreshold` well below 95.

## Cleaning up

```bash
$ kubectl delete yugabytedbautoscaler -n demo yb-storage-autoscaler
```
