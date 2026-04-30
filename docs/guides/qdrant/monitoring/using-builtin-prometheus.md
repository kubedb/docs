---
title: Monitor Qdrant using Builtin Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: qdrant-using-builtin-prometheus-monitoring
    name: Builtin Prometheus
    parent: qdrant-monitoring
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Qdrant with Builtin Prometheus

This guide shows how to enable builtin Prometheus scraping for Qdrant.

## Before You Begin

- Install KubeDB operator from [setup guide](/docs/setup/README.md).
- Use separate namespaces for database and monitoring resources.

```bash
$ kubectl create ns demo
$ kubectl create ns monitoring
```

## Deploy Qdrant with Monitoring Enabled

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: builtin-prom-qdrant
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
    agent: prometheus.io/builtin
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f builtin-prom-qdrant.yaml
qdrant.kubedb.com/builtin-prom-qdrant created
```

## Verify

```bash
$ kubectl get qdrant -n demo builtin-prom-qdrant
$ kubectl get svc -n demo --selector=app.kubernetes.io/instance=builtin-prom-qdrant
```

The operator creates a stats service with scrape annotations for builtin Prometheus discovery.

## Cleaning up

```bash
kubectl delete qdrant -n demo builtin-prom-qdrant
kubectl delete ns demo
kubectl delete ns monitoring
```