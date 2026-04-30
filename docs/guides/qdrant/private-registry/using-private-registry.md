---
title: Run Qdrant using Private Registry
menu:
  docs_{{ .version }}:
    identifier: qdrant-using-private-registry
    name: Quickstart
    parent: qdrant-private-registry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Private Docker Registry

This tutorial shows how to run KubeDB managed Qdrant using private Docker images.

## Before You Begin

- Prepare a Kubernetes cluster and `kubectl`.

```bash
kubectl create ns demo
```

## Create ImagePullSecret

```bash
kubectl create secret docker-registry -n demo myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
```

## Create QdrantVersion CRD

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: QdrantVersion
metadata:
  name: "1.17.0"
spec:
  db:
    image: PRIVATE_REGISTRY/qdrant:1.17.0
  version: "1.17.0"
```

## Deploy Qdrant from Private Registry

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: pvt-reg-qdrant
  namespace: demo
spec:
  version: 1.17.0
  mode: Distributed
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  podTemplate:
    spec:
      imagePullSecrets:
        - name: myregistrykey
  deletionPolicy: WipeOut
```

## Cleaning up

```bash
kubectl delete qdrant -n demo pvt-reg-qdrant
kubectl delete ns demo
```