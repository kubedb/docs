---
title: Run Weaviate with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: weaviate-using-config-file
    name: Config File
    parent: weaviate-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB uses `spec.configuration.secretName` to provide custom Weaviate configuration.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured for that cluster.
- Install KubeDB operator from [setup guide](/docs/setup/README.md).

```bash
$ kubectl create ns demo
namespace/demo created
```

## Create Configuration Secret

Create a custom configuration file and then create a Secret:

```yaml
authentication:
  anonymous_access:
    enabled: false
query_defaults:
  limit: 25
```

```bash
$ kubectl create secret generic -n demo weaviate-config \
  --from-file=weaviate.yaml=./weaviate.yaml
secret/weaviate-config created
```

## Deploy Weaviate with Configuration Secret

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: custom-weaviate
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  configuration:
    secretName: weaviate-config
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f custom-weaviate.yaml
weaviate.kubedb.com/custom-weaviate created
```

## Verify

```bash
$ kubectl get weaviate -n demo custom-weaviate
NAME              VERSION   STATUS   AGE
custom-weaviate   1.33.1    Ready    2m
```

## Cleaning up

```bash
kubectl delete weaviate -n demo custom-weaviate
kubectl delete secret -n demo weaviate-config
kubectl delete ns demo
```