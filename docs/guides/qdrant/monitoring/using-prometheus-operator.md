---
title: Monitor Qdrant using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: qdrant-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: qdrant-monitoring
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Qdrant using Prometheus Operator

This guide shows how to expose Qdrant metrics using Prometheus Operator integration.

## Before You Begin

- Install KubeDB operator from [setup guide](/docs/setup/README.md).
- Ensure Prometheus Operator is installed in your cluster.

```bash
$ kubectl create ns demo
$ kubectl create ns monitoring
```

## Deploy Qdrant with Monitoring Enabled

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: coreos-prom-qdrant
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
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f coreos-prom-qdrant.yaml
qdrant.kubedb.com/coreos-prom-qdrant created
```

## Verify

```bash
$ kubectl get qdrant -n demo coreos-prom-qdrant
$ kubectl get servicemonitor -n demo
```

## Cleaning up

```bash
kubectl delete qdrant -n demo coreos-prom-qdrant
kubectl delete ns demo
kubectl delete ns monitoring
```