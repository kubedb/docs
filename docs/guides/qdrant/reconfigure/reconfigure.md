---
title: Reconfigure Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-reconfigure-cluster
    name: Cluster
    parent: qdrant-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Qdrant

This guide shows how to apply configuration changes to a running Qdrant database using `QdrantOpsRequest`.

## Before You Begin

- Install KubeDB Community and Enterprise operators from [setup guide](/docs/setup/README.md).
- Deploy a Qdrant database and ensure it is in `Ready` state.

```bash
$ kubectl create ns demo
```

## Apply Reconfigure OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: qdrant-sample
  configuration:
    configSecret:
      name: qdrant-config-updated
```

```bash
$ kubectl apply -f qdrant-reconfigure.yaml
qdrantopsrequest.ops.kubedb.com/qdrant-reconfigure created
```

## Verify

```bash
$ kubectl get qdrantopsrequest -n demo qdrant-reconfigure
$ kubectl describe qdrantopsrequest -n demo qdrant-reconfigure
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-reconfigure
kubectl delete ns demo
```