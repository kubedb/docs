---
title: Milvus Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: milvus-autoscaler-storage-guide
    name: Guide
    parent: milvus-autoscaler-storage
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus Storage Autoscaling

This guide will show you how to use the `KubeDB` Autoscaler operator to autoscale the persistent storage of a Milvus database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusAutoscaler](/docs/guides/milvus/concepts/milvusautoscaler.md)
  - [Storage Autoscaling Overview](/docs/guides/milvus/autoscaler/storage/overview.md)

- Install the **KubeDB Autoscaler** operator and **Prometheus** (storage autoscaling reads PVC usage from Prometheus).

- The PVC's `StorageClass` must support volume expansion (`allowVolumeExpansion: true`) — e.g. `longhorn-custom`.

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO, creates the `my-release-minio` secret, and installs the etcd operator required by Milvus.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/autoscaler/storage/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/autoscaler/storage/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Storage Autoscaling — Standalone Milvus

Deploy a standalone Milvus on an expansion-capable `StorageClass` and create a `MilvusAutoscaler` targeting the standalone `node`:

`storage-standalone.yaml`

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MilvusAutoscaler
metadata:
  name: milvus-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: milvus-standalone
  storage:
    node:
      trigger: "On"
      usageThreshold: 30
      expansionMode: "Offline"
      scalingRules:
        - appliesUpto: "100Ti"
          threshold: "50%"
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/autoscaler/storage/yamls/storage-standalone.yaml
milvusautoscaler.autoscaling.kubedb.com/milvus-storage-autoscaler created
```

When the volume usage crosses `usageThreshold` (30%), the autoscaler creates a `VolumeExpansion` `MilvusOpsRequest` sized per `scalingRules`.

```bash
$ kubectl get milvusautoscaler -n demo
NAME                                   AGE
milvus-standalone-compute-autoscaler   94s
milvus-storage-autoscaler              93s
```

The storage autoscaler watches the `streamingnode`/`node` PVC usage (read from Prometheus). When usage crosses `usageThreshold`, it creates a `VolumeExpansion` `MilvusOpsRequest`. In this walkthrough the volume stayed well below the threshold (a freshly-created, near-empty `1Gi` volume), so no expansion was triggered:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-standalone -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage
NAME                       SIZE
data-milvus-standalone-0   1Gi
```

> To see the expansion fire, write enough data to push PVC usage past `usageThreshold` (30% here). When it does, the autoscaler creates a `VolumeExpansion` `MilvusOpsRequest` (with `expansionMode` as configured), which the Ops-manager applies — see the [volume expansion guide](/docs/guides/milvus/volume-expansion/guide.md) for the resulting flow and output.

## Storage Autoscaling — Distributed Milvus

For a distributed Milvus, storage autoscaling targets **`streamingnode`** — the only distributed role with a persistent volume:

`storage-distributed.yaml`

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MilvusAutoscaler
metadata:
  name: milvus-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: milvus-cluster
  storage:
    streamingnode:
      trigger: "On"
      usageThreshold: 34
      expansionMode: "Online"
      scalingRules:
        - appliesUpto: "100Ti"
          threshold: "50%"
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
```

Because only `streamingnode` carries a persistent volume among the distributed roles, storage autoscaling targets `streamingnode` exclusively. When the `streamingnode` PVC usage crosses `usageThreshold` (34% here), the autoscaler creates a `VolumeExpansion` `MilvusOpsRequest` against `streamingnode` (`expansionMode: Online`), which the Ops-manager applies as shown in the [volume expansion guide](/docs/guides/milvus/volume-expansion/guide.md).

## Cleaning up

```bash
$ kubectl delete milvusautoscaler -n demo --all
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone
$ kubectl delete ns demo
```

## Next Steps

- Learn about [compute autoscaling](/docs/guides/milvus/autoscaler/compute/guide.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
