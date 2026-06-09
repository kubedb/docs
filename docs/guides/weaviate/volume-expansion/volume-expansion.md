---
title: Expand Weaviate Volume
menu:
  docs_{{ .version }}:
    identifier: weaviate-volume-expansion-cluster
    name: Cluster
    parent: weaviate-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Expand Weaviate Volume

This guide shows how to increase Weaviate data volume size using a `WeaviateOpsRequest`.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured.
- StorageClass used by Weaviate must support volume expansion.
- Install KubeDB and Ops Manager from [setup docs](/docs/setup/README.md).
- Review [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md).

Create namespace:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Verify Expandable StorageClass

```bash
$ kubectl get storageclass
NAME       PROVISIONER               ALLOWVOLUMEEXPANSION   AGE
longhorn   driver.longhorn.io        true                   5d
```

Use a storage class where `ALLOWVOLUMEEXPANSION` is `true`.

## Deploy Weaviate

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
weaviate.kubedb.com/weaviate-sample created

$ kubectl get weaviate -n demo weaviate-sample -w
NAME              VERSION   STATUS   AGE
weaviate-sample   1.33.1    Ready    2m
```

Check current PVC size:

```bash
$ kubectl get pvc -n demo
```

## Apply VolumeExpansion OpsRequest

Use this request:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-volume-expand
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: weaviate-sample
  volumeExpansion:
    mode: Online
    weaviate: 2Gi
```

Apply the sample file:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/volume-expansion/ops-request.yaml
weaviateopsrequest.ops.kubedb.com/weaviate-volume-expand created
```

## Verify Volume Expansion

Watch request status:

```bash
$ kubectl get weaviateopsrequest -n demo weaviate-volume-expand
NAME                     TYPE              STATUS       AGE
weaviate-volume-expand   VolumeExpansion   Successful   3m
```

Inspect request events and progression:

```bash
$ kubectl describe weaviateopsrequest -n demo weaviate-volume-expand
```

Verify PVC size has been updated:

```bash
$ kubectl get pvc -n demo
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-volume-expand
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
