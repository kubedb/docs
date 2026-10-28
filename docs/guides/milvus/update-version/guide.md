---
title: Update Version of Milvus
menu:
  docs_{{ .version }}:
    identifier: milvus-update-version-guide
    name: Guide
    parent: milvus-update-version
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update Version of Milvus

This guide will show you how to use the `KubeDB` Ops-manager operator to update the version of a Milvus database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)
  - [Update Version Overview](/docs/guides/milvus/update-version/overview.md)

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO and creates the object-storage secrets (`milvus-storage-config` for `milvus-standalone`, `milvus-cluster-storage-config` for `milvus-cluster`).

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/update-version/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/update-version/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Available Versions

```bash
$ kubectl get milvusversions
NAME         VERSION   DB_IMAGE                                    DEPRECATED   AGE
2.6.11       2.6.11    ghcr.io/appscode-images/milvus:2.6.11                    21h
2.6.15       2.6.15    ghcr.io/appscode-images/milvus:2.6.15                    21h
2.6.15-gpu   2.6.15    ghcr.io/appscode-images/milvus:2.6.15-gpu                21h
2.6.7        2.6.7     ghcr.io/appscode-images/milvus:2.6.7                     21h
2.6.9        2.6.9     ghcr.io/appscode-images/milvus:2.6.9                     21h
3.0.0        3.0.0     ghcr.io/appscode-images/milvus:3.0.0                     21h
3.0.0-gpu    3.0.0     ghcr.io/appscode-images/milvus:3.0.0-gpu                 21h
```

## Update Version of Standalone Milvus

Deploy a standalone Milvus at version `2.6.9` and wait until it is `Ready`. `update-version/yamls/standalone.yaml` is the [standalone quickstart](/docs/guides/milvus/quickstart/standalone.md) manifest with `spec.version: "2.6.9"` (the quickstart itself uses `2.6.11`, which would make the update below a no-op):

> If you deleted a Milvus of the same name in an earlier guide, delete its object-storage prefix first, as shown in [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#delete-a-milvus-instances-object-storage-data). Otherwise the new database can hang in `Provisioning`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/update-version/yamls/standalone.yaml
milvus.kubedb.com/milvus-standalone created

$ kubectl get milvuses.kubedb.com milvus-standalone -n demo
NAME                VERSION   STATUS   AGE
milvus-standalone   2.6.9     Ready    46s
```

### Apply the UpdateVersion OpsRequest

`update-version-standalone.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: milvus-standalone
  updateVersion:
    targetVersion: 2.6.11
  timeout: 5m
  apply: IfReady
```

Here, `spec.updateVersion.targetVersion` is the name of the target `MilvusVersion` (`2.6.11`).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/update-version/yamls/update-version-standalone.yaml
milvusopsrequest.ops.kubedb.com/milvus-update-version created
```

### Watch Progress

```bash
$ kubectl get milvusopsrequest milvus-update-version -n demo
NAME                    TYPE            STATUS       AGE
milvus-update-version   UpdateVersion   Successful   77s
```

```bash
$ kubectl describe milvusopsrequest milvus-update-version -n demo
...
Status:
  Conditions:
    Message:  Milvus ops-request has started to update version
    Reason:   UpdateVersion
    Type:     UpdateVersion
    Message:  successfully reconciled the Milvus with updated version
    Reason:   UpdatePetSets
    Type:     UpdatePetSets
    Message:  check pod running; ConditionStatus:True; PodName:milvus-standalone-0
    Type:     CheckPodRunning--milvus-standalone-0
    Message:  Successfully Restarted Milvus nodes
    Reason:   RestartPods
    Type:     RestartPods
    Message:  Successfully completed update ... version
    Reason:   Successful
    Type:     Successful
  Phase:      Successful
```

### Verify the New Version

```bash
$ kubectl get milvuses.kubedb.com milvus-standalone -n demo
NAME                VERSION   STATUS   AGE
milvus-standalone   2.6.11    Ready    2m3s
```

The pod now runs the `2.6.11` image:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=milvus-standalone -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image
```

## Update Version of Distributed Milvus

For a distributed Milvus, point `spec.databaseRef.name` at the distributed database (`milvus-cluster`). The operator updates the image of every distributed role and restarts each one.

Deploy a distributed Milvus at version `2.6.9` (the [distributed quickstart](/docs/guides/milvus/quickstart/distributed.md) manifest with `spec.version: "2.6.9"`) and wait until it is `Ready`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/update-version/yamls/distributed.yaml
milvus.kubedb.com/milvus-cluster created

$ kubectl get milvuses.kubedb.com milvus-cluster -n demo -w
```

The distributed sample reuses the OpsRequest name of the standalone one, so delete the standalone OpsRequest first:

```bash
$ kubectl delete milvusopsrequest -n demo milvus-update-version
```

`update-version-distributed.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: milvus-cluster
  updateVersion:
    targetVersion: 2.6.11
  timeout: 5m
  apply: IfReady
```

Apply it the same way, pointing at the distributed database:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/update-version/yamls/update-version-distributed.yaml
milvusopsrequest.ops.kubedb.com/milvus-update-version created
```

The distributed flow is mechanically identical to the standalone flow shown above: the operator validates the target `MilvusVersion`, pauses the database, updates the container image of **each** distributed role (`mixcoord`, `datanode`, `querynode`, `streamingnode`, `proxy`), restarts them one workload at a time, and resumes the database. Because `apply: IfReady` is set, the ops request runs only once the database is `Ready`, after which it reports `Successful`.

Afterwards, every role pod runs the `2.6.11` image:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=milvus-cluster -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image
```

## Cleaning up

```bash
$ kubectl delete milvusopsrequest -n demo milvus-update-version
$ kubectl patch -n demo milvus.kubedb.com milvus-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl patch -n demo milvus.kubedb.com milvus-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone milvus-cluster
$ kubectl delete ns demo
```

## Next Steps

- Learn about [vertical scaling](/docs/guides/milvus/scaling/vertical-scaling/guide.md) of a Milvus database.
- Learn how the [Recommendation Engine](/docs/guides/milvus/recommendation/guide.md) suggests version updates automatically.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
