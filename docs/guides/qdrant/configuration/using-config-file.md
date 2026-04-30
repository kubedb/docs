---
title: Run Qdrant with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: qdrant-using-config-file
    name: Config File
    parent: qdrant-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB uses `spec.configuration.secretName` to provide a custom Qdrant configuration.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured for that cluster.
- Install KubeDB operator following [the setup guide](/docs/setup/README.md).

```bash
$ kubectl create ns demo
namespace/demo created
```

## Create Configuration Secret

Create a `production.yaml` file with your desired runtime settings:

```yaml
log_level: INFO
service:
  max_request_size_mb: 32
```

Create a Secret from this file:

```bash
$ kubectl create secret generic -n demo qdrant-config \
  --from-file=production.yaml=./production.yaml
secret/qdrant-config created
```

## Deploy Qdrant with Custom Configuration

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: custom-qdrant
  namespace: demo
spec:
  version: 1.17.0
  mode: Distributed
  replicas: 3
  configuration:
    secretName: qdrant-config
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f qdrant-configuration.yaml
qdrant.kubedb.com/custom-qdrant created
```

## Verify

```bash
$ kubectl get qdrant -n demo custom-qdrant
NAME           VERSION   STATUS   AGE
custom-qdrant  1.17.0    Ready    2m
```

## Cleaning up

```bash
kubectl delete qdrant -n demo custom-qdrant
kubectl delete secret -n demo qdrant-config
kubectl delete ns demo
```