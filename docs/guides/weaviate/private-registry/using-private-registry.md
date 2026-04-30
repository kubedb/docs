---
title: Run Weaviate using Private Registry
menu:
  docs_{{ .version }}:
    identifier: weaviate-using-private-registry
    name: Quickstart
    parent: weaviate-private-registry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Private Docker Registry

This tutorial shows how to run KubeDB managed Weaviate using private Docker images.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured for that cluster.
- Create an image pull secret for your private registry in the target namespace.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl create secret docker-registry -n demo myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
```

## Deploy Weaviate from Private Registry

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: pvt-reg-weaviate
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      imagePullSecrets:
        - name: myregistrykey
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f pvt-reg-weaviate.yaml
weaviate.kubedb.com/pvt-reg-weaviate created
```

## Verify

```bash
$ kubectl get pod -n demo -l app.kubernetes.io/instance=pvt-reg-weaviate
```

## Cleaning up

```bash
kubectl delete weaviate -n demo pvt-reg-weaviate
kubectl delete ns demo
```