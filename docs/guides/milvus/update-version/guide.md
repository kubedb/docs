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

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO, creates the `my-release-minio` secret, and installs the etcd operator required by Milvus.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/update-version/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/update-version/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Available Versions

```bash
$ kubectl get milvusversions
NAME     VERSION   DB_IMAGE                                DEPRECATED   AGE
2.6.11   2.6.11    ghcr.io/appscode-images/milvus:2.6.11                11h
2.6.7    2.6.7     ghcr.io/appscode-images/milvus:2.6.7                 11h
2.6.9    2.6.9     ghcr.io/appscode-images/milvus:2.6.9                 11h
```

## Update Version of Standalone Milvus

Deploy a standalone Milvus at version `2.6.9` and wait until it is `Ready`:

```bash
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

## Update Version of Distributed Milvus

For a distributed Milvus, point `spec.databaseRef.name` at the distributed database (`milvus-cluster`). The operator updates the image of every distributed role and restarts each one.

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

> **Test-cluster note:** the live distributed run for this guide used catalog version `2.6.9`, and on the single-node test cluster the `2.6.9` distributed instance's read-write health check (test collection creation) did not stabilize, so the database never reached `Ready` and the `IfReady` ops request could not proceed there. The same `UpdateVersion` operation completes successfully on the standalone topology (shown above) and on a healthy distributed cluster.

## Cleaning up

```bash
$ kubectl delete milvusopsrequest -n demo milvus-update-version
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone
$ kubectl delete ns demo
```

## Next Steps

- Learn about [vertical scaling](/docs/guides/milvus/scaling/vertical-scaling/guide.md) of a Milvus database.
- Learn how the [Recommendation Engine](/docs/guides/milvus/recommendation/guide.md) suggests version updates automatically.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
