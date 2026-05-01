---
title: Monitor Neo4j using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: neo4j-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: neo4j-monitoring
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Neo4j using Prometheus Operator

This tutorial will show you how to use the Prometheus operator to monitor Neo4j database deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster.
- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).
- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/neo4j/monitoring/overview.md).

```bash
$ kubectl create ns monitoring
$ kubectl create ns demo
```

## Deploy Neo4j with Monitoring Enabled

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: coreos-prom-neo4j
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/monitoring/coreos-prom-neo4j.yaml
neo4j.kubedb.com/coreos-prom-neo4j created
```

```bash
$ kubectl get servicemonitor -n demo coreos-prom-neo4j
NAME                AGE
coreos-prom-neo4j   2m
```

## Cleaning up

```bash
kubectl patch -n demo neo4j/coreos-prom-neo4j -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo neo4j/coreos-prom-neo4j
kubectl delete ns demo
kubectl delete ns monitoring
```
