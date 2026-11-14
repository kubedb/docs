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

- Install the **KubeDB Autoscaler** operator.

- During KubeDB installation, enable the KubeDB storage metrics server by passing the following Helm flag:

  ```bash
  --set kubedb-autoscaler.storage-metrics-server.enabled=true
  ```

  It provides the **custom metrics API** (`custom.metrics.k8s.io`) backed by the KubeDB storage-metrics apiserver. The storage autoscaler reads PVC usage from this API. It is disabled by default. To enable it on an existing installation:

  ```bash
  $ helm upgrade kubedb oci://ghcr.io/appscode-charts/kubedb \
      --namespace kubedb --reuse-values \
      --set kubedb-autoscaler.storage-metrics-server.enabled=true
  ```

  Check that the API is served. The autoscaler queries `custom.metrics.k8s.io/v1beta2`; a `NotFound` error here means the storage metrics server is not running:

  ```bash
  $ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta2
  ```

  Without this API the storage autoscaler never triggers. The autoscaler's status and events stay empty; only the KubeDB autoscaler operator's log shows `custom metrics API returned 404 Not Found`.

- The PVC's `StorageClass` must support volume expansion (`allowVolumeExpansion: true`) — e.g. `longhorn-custom`.

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO and creates the object-storage secrets (`milvus-storage-config` for `milvus-standalone`, `milvus-cluster-storage-config` for `milvus-cluster`).

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/autoscaler/storage/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/autoscaler/storage/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Storage Autoscaling — Standalone Milvus

Deploy a standalone Milvus on an expansion-capable `StorageClass`. The [volume expansion guide](/docs/guides/milvus/volume-expansion/guide.md)'s `milvus-standalone.yaml` is the standalone quickstart manifest with `storageClassName: longhorn-custom` and a `1Gi` volume:

> If you deleted a Milvus of the same name in an earlier guide, delete its object-storage prefix first, as shown in [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#delete-a-milvus-instances-object-storage-data). Otherwise the new database can hang in `Provisioning`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/volume-expansion/yamls/milvus-standalone.yaml
milvus.kubedb.com/milvus-standalone created

$ kubectl get milvuses.kubedb.com -n demo milvus-standalone -w
```

Then create a `MilvusAutoscaler` targeting the standalone `node`:

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
NAME                        AGE
milvus-storage-autoscaler   93s
```

You can query the same metric the autoscaler uses for the PVC:

```bash
$ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta2/namespaces/demo/persistentvolumeclaims/data-milvus-standalone-0/volume_used_percentage
```

The storage autoscaler watches the `streamingnode`/`node` PVC usage through the `custom.metrics.k8s.io` API backed by the KubeDB storage-metrics apiserver. When usage crosses `usageThreshold`, it creates a `VolumeExpansion` `MilvusOpsRequest`. In this walkthrough the volume stayed well below the threshold (a freshly-created, near-empty `1Gi` volume), so no expansion was triggered:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-standalone -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage
NAME                       SIZE
data-milvus-standalone-0   1Gi
```

> To see the expansion fire, write enough data to push PVC usage past `usageThreshold` (30% here). For a quick test, write a 400MiB file into the data volume (about 39% of `1Gi`), and remove it afterwards:
>
> ```bash
> $ kubectl exec -n demo milvus-standalone-0 -c milvus -- fallocate -l 400M /var/lib/milvus/fill.bin
> ```
> When it does, the autoscaler creates a `VolumeExpansion` `MilvusOpsRequest` (with `expansionMode` as configured), which the Ops-manager applies — see the [volume expansion guide](/docs/guides/milvus/volume-expansion/guide.md) for the resulting flow and output.

## Storage Autoscaling — Distributed Milvus

For a distributed Milvus, storage autoscaling targets **`streamingnode`** — the only distributed role with a persistent volume.

Deploy a distributed Milvus on `longhorn-custom`, for example with the volume expansion guide's `milvus-distributed.yaml` (two `streamingnode` replicas with `1Gi` each). The distributed autoscaler reuses the name `milvus-storage-autoscaler`, so delete the standalone one first:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/volume-expansion/yamls/milvus-distributed.yaml
milvus.kubedb.com/milvus-cluster created

$ kubectl get milvuses.kubedb.com -n demo milvus-cluster -w

$ kubectl delete milvusautoscaler -n demo milvus-storage-autoscaler
```

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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/autoscaler/storage/yamls/storage-distributed.yaml
milvusautoscaler.autoscaling.kubedb.com/milvus-storage-autoscaler created
```

Because only `streamingnode` carries a persistent volume among the distributed roles, storage autoscaling targets `streamingnode` exclusively. When the `streamingnode` PVC usage crosses `usageThreshold` (34% here), the autoscaler creates a `VolumeExpansion` `MilvusOpsRequest` against `streamingnode` (`expansionMode: Online`), which the Ops-manager applies as shown in the [volume expansion guide](/docs/guides/milvus/volume-expansion/guide.md).

## Cleaning up

```bash
$ kubectl get milvusopsrequest -n demo   # wait until none is Progressing
$ kubectl delete milvusautoscaler -n demo --all
$ kubectl patch -n demo milvus.kubedb.com milvus-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl patch -n demo milvus.kubedb.com milvus-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone milvus-cluster
$ kubectl delete ns demo
```

## Next Steps

- Learn about [compute autoscaling](/docs/guides/milvus/autoscaler/compute/guide.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
