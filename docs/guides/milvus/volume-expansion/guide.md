---
title: Milvus Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: milvus-volume-expansion-guide
    name: Guide
    parent: milvus-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus Volume Expansion

This guide will show you how to use the `KubeDB` Ops-manager operator to expand the persistent volumes of a Milvus database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)
  - [Volume Expansion Overview](/docs/guides/milvus/volume-expansion/overview.md)

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO and creates the object-storage secrets (`milvus-storage-config` for `milvus-standalone`, `milvus-cluster-storage-config` for `milvus-cluster`).

- The PVC's `StorageClass` **must** support volume expansion (`allowVolumeExpansion: true`). The base examples use `local-path`, which does **not** support expansion, so this guide uses `longhorn-custom`:

  ```bash
  $ kubectl get sc longhorn-custom -o jsonpath='{.allowVolumeExpansion}'
  true
  ```

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/volume-expansion/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/volume-expansion/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Volume Expansion — Standalone Milvus

Deploy a standalone Milvus on an expansion-capable `StorageClass` (here `longhorn-custom`) with a `1Gi` volume. `volume-expansion/yamls/milvus-standalone.yaml` is the [standalone quickstart](/docs/guides/milvus/quickstart/standalone.md) manifest with `storageClassName: longhorn-custom`:

> If you deleted a Milvus of the same name in an earlier guide, delete its object-storage prefix first, as shown in [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#delete-a-milvus-instances-object-storage-data). Otherwise the new database can hang in `Provisioning`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/volume-expansion/yamls/milvus-standalone.yaml
milvus.kubedb.com/milvus-standalone created

$ kubectl get milvuses.kubedb.com -n demo milvus-standalone -w

$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-standalone -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage,SC:.spec.storageClassName
NAME                       SIZE   SC
data-milvus-standalone-0   1Gi    longhorn-custom
```

### Offline Volume Expansion

`volume-expansion-offline-standalone.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: volume-expansion-offline
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: milvus-standalone
  volumeExpansion:
    node: 4Gi
    mode: Offline
```

Here, `spec.volumeExpansion.node` is the standalone target size, and `mode: Offline` takes the pod down while the volume is resized.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/volume-expansion/yamls/volume-expansion-offline-standalone.yaml
milvusopsrequest.ops.kubedb.com/volume-expansion-offline created

$ kubectl get milvusopsrequest volume-expansion-offline -n demo
NAME                       TYPE              STATUS       AGE
volume-expansion-offline   VolumeExpansion   Successful   2m
```

The volume is now `4Gi`:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-standalone -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage,SC:.spec.storageClassName
NAME                       SIZE   SC
data-milvus-standalone-0   4Gi    longhorn-custom
```

### Online Volume Expansion

Online expansion grows the volume without taking the pod down. The volume is already `4Gi` after the offline step above, so this request grows it to `6Gi`. (A request for the current size is accepted and reported `Successful` without changing anything.)

`volume-expansion-online-standalone.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: volume-expansion-online
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: milvus-standalone
  volumeExpansion:
    node: 6Gi
    mode: Online
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/volume-expansion/yamls/volume-expansion-online-standalone.yaml
milvusopsrequest.ops.kubedb.com/volume-expansion-online created

$ kubectl get milvusopsrequest volume-expansion-online -n demo
NAME                      TYPE              STATUS       AGE
volume-expansion-online   VolumeExpansion   Successful   2m4s
```

The volume has grown to `6Gi`, and the database spec reflects the new size:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-standalone -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage,SC:.spec.storageClassName
NAME                       SIZE   SC
data-milvus-standalone-0   6Gi    longhorn-custom

$ kubectl get milvuses.kubedb.com milvus-standalone -n demo -o jsonpath='{.spec.storage.resources.requests.storage}'
6Gi
```

## Volume Expansion — Distributed Milvus

For a distributed Milvus, volume expansion targets `streamingnode` — the only distributed role that carries a persistent volume. Point `spec.databaseRef.name` at `milvus-cluster` and use the `streamingnode` key.

Deploy a distributed Milvus with two `streamingnode` replicas on `longhorn-custom`, each with a `1Gi` volume. `volume-expansion/yamls/milvus-distributed.yaml` is the [distributed quickstart](/docs/guides/milvus/quickstart/distributed.md) manifest with `storageClassName: longhorn-custom` and `streamingnode.replicas: 2`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/volume-expansion/yamls/milvus-distributed.yaml
milvus.kubedb.com/milvus-cluster created

$ kubectl get milvuses.kubedb.com -n demo milvus-cluster -w
```

The distributed OpsRequests reuse the names of the standalone ones, so delete the standalone OpsRequests first:

```bash
$ kubectl delete milvusopsrequest -n demo volume-expansion-offline volume-expansion-online
```

`volume-expansion-online-distributed.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: volume-expansion-online
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: milvus-cluster
  volumeExpansion:
    streamingnode: 4Gi
    mode: Online
```

(An `Offline` variant — `volume-expansion-offline-distributed.yaml` — uses `streamingnode: 3Gi` with `mode: Offline`.)

The operator expands the PVC of every `streamingnode` replica, one replica at a time. Starting from `1Gi` on `longhorn-custom`, an **offline** expansion to `3Gi` followed by an **online** expansion to `4Gi`. Apply the online request only after the offline one is `Successful`:

```bash
# offline: streamingnode 1Gi -> 3Gi
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/volume-expansion/yamls/volume-expansion-offline-distributed.yaml
milvusopsrequest.ops.kubedb.com/volume-expansion-offline created

$ kubectl get milvusopsrequest volume-expansion-offline -n demo
NAME                       TYPE              STATUS       AGE
volume-expansion-offline   VolumeExpansion   Successful   ...

# online: streamingnode 3Gi -> 4Gi
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/volume-expansion/yamls/volume-expansion-online-distributed.yaml
milvusopsrequest.ops.kubedb.com/volume-expansion-online created

$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-cluster -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage
NAME                                  SIZE
data-milvus-cluster-streamingnode-0   4Gi
data-milvus-cluster-streamingnode-1   4Gi
```

Both `streamingnode` replicas are expanded. The stateless roles (`mixcoord`, `datanode`, `querynode`, `proxy`) have no persistent volume and are unaffected.

> On a single-node test cluster, offline expansion of a distributed role can take a few minutes because the pods are restarted after the resize.

## Cleaning up

```bash
$ kubectl delete milvusopsrequest -n demo volume-expansion-offline volume-expansion-online
$ kubectl patch -n demo milvus.kubedb.com milvus-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl patch -n demo milvus.kubedb.com milvus-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone milvus-cluster
$ kubectl delete ns demo
```

## Next Steps

- Learn about [storage migration](/docs/guides/milvus/storage-migration/guide.md) of a Milvus database.
- Learn about [storage autoscaling](/docs/guides/milvus/autoscaler/storage/guide.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
