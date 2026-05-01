---
title: Upgrade Weaviate Version
menu:
  docs_{{ .version }}:
    identifier: weaviate-version-upgrading
    name: Version Upgrading
    parent: weaviate-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Upgrade Weaviate Version

This guide demonstrates how to update Weaviate from one supported version to another using `WeaviateOpsRequest`.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured.
- Install KubeDB and Ops Manager from [setup docs](/docs/setup/README.md).
- Review [WeaviateVersion](/docs/guides/weaviate/concepts/catalog.md) and [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md).

Create a namespace for the tutorial:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Check Available Versions

List available versions:

```bash
$ kubectl get weaviateversions
```

Choose a target version that is present and not deprecated.

## Deploy Weaviate

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
weaviate.kubedb.com/weaviate-sample created

$ kubectl get weaviate -n demo weaviate-sample -w
NAME              VERSION   STATUS   AGE
weaviate-sample   1.33.1    Ready    2m
```

## Apply UpdateVersion OpsRequest

Use this request:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: weaviate-sample
  updateVersion:
    targetVersion: 1.34.0
```

Apply from the example file:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/update-version/ops-request.yaml
weaviateopsrequest.ops.kubedb.com/weaviate-update-version created
```

## Verify Version Update

Watch request status:

```bash
$ kubectl get weaviateopsrequest -n demo weaviate-update-version
NAME                       TYPE            STATUS       AGE
weaviate-update-version    UpdateVersion   Successful   4m
```

Confirm resulting database version:

```bash
$ kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.version}{"\n"}'
1.34.0
```

Check pod image and health:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=weaviate-sample
$ kubectl get weaviate -n demo weaviate-sample
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-update-version
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```