---
title: Monitor Milvus using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: milvus-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: milvus-monitoring
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Milvus using Prometheus Operator

[Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator) provides a simple, Kubernetes-native way to deploy and configure Prometheus. This tutorial will show you how to monitor a KubeDB-managed Milvus database using the Prometheus Operator.

## Before You Begin

- You need a running Kubernetes cluster and a Prometheus Operator installation. Note the labels its `Prometheus` object uses to select `ServiceMonitor`s (here, `release: prometheus`).

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [Monitoring Overview](/docs/guides/milvus/monitoring/overview.md)

- An object-storage secret named `my-release-minio` must exist in the `demo` namespace.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/monitoring/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/monitoring/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Enable Monitoring in the Milvus Manifest

Monitoring is enabled through `spec.monitor`. The base [standalone](/docs/guides/milvus/quickstart/standalone.md) and [distributed](/docs/guides/milvus/quickstart/distributed.md) manifests already include it:

```yaml
spec:
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

- `agent: prometheus.io/operator` selects Prometheus-Operator integration.
- `serviceMonitor.labels` are applied to the generated `ServiceMonitor` so the Prometheus Operator picks it up (`release: prometheus` must match your Prometheus `serviceMonitorSelector`).
- `serviceMonitor.interval` is the scrape interval.

Deploy the database and wait until it is `Ready`.

## Stats Service

When monitoring is enabled, KubeDB creates a dedicated **stats service** named `<db>-stats` exposing the metrics port `9091`:

```bash
kubectl get svc -n demo -l app.kubernetes.io/instance=milvus-standalone
```
NAME                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
milvus-standalone         ClusterIP   10.43.144.154   <none>        19530/TCP   91s
milvus-standalone-stats   ClusterIP   10.43.12.191    <none>        9091/TCP    91s

## ServiceMonitor

KubeDB also creates a `ServiceMonitor` named `<db>-stats` that selects the stats service:

```bash
kubectl get servicemonitor -n demo -l app.kubernetes.io/instance=milvus-standalone
```
NAME                      AGE
milvus-standalone-stats   90s

```bash
kubectl get servicemonitor milvus-standalone-stats -n demo -o yaml
```
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: milvus-standalone
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: milvuses.kubedb.com
    release: prometheus
  name: milvus-standalone-stats
  namespace: demo
spec:
  endpoints:
  - honorLabels: true
    interval: 10s
    path: /metrics
    port: metrics
    relabelings:
    - action: replace
      sourceLabels:
      - __meta_kubernetes_endpoint_address_target_name
      targetLabel: pod
    scheme: http
  namespaceSelector:
    matchNames:
    - demo
  selector:
    matchLabels:
      app.kubernetes.io/component: database
      app.kubernetes.io/instance: milvus-standalone
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: milvuses.kubedb.com
      kubedb.com/role: stats

Key points:

- The `release: prometheus` label (from `serviceMonitor.labels`) is what lets the Prometheus Operator discover this `ServiceMonitor`.
- The scrape `interval` is `10s`, as configured.
- The endpoint scrapes the `metrics` port at `/metrics`.
- The selector matches the stats service via the `kubedb.com/role: stats` label.

Once the Prometheus Operator reconciles this `ServiceMonitor`, Milvus metrics begin appearing in Prometheus.

## Distributed Milvus

Monitoring works identically for a distributed Milvus. A single stats service and `ServiceMonitor` named `milvus-cluster-stats` are created, and metrics are scraped from the distributed components (each role's pods expose port `9091`).

```bash
kubectl get svc -n demo -l app.kubernetes.io/instance=milvus-cluster
```
NAME                           TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)     AGE
milvus-cluster                 ClusterIP   10.43.221.1   <none>        19530/TCP   3m
milvus-cluster-datanode        ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-mixcoord        ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-querynode       ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-stats           ClusterIP   10.43.95.57   <none>        9091/TCP    3m
milvus-cluster-streamingnode   ClusterIP   None          <none>        9091/TCP    3m

```bash
kubectl get servicemonitor milvus-cluster-stats -n demo -o yaml
```
...
spec:
  endpoints:
  - honorLabels: true
    interval: 10s
    path: /metrics
    port: metrics
    scheme: http
  namespaceSelector:
    matchNames:
    - demo
  selector:
    matchLabels:
      app.kubernetes.io/instance: milvus-cluster
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: milvuses.kubedb.com
      kubedb.com/role: stats

## Cleaning up

```bash
kubectl delete milvus.kubedb.com -n demo milvus-standalone
```

```bash
kubectl delete ns demo
```

## Next Steps

- Secure your Milvus database with [TLS/SSL](/docs/guides/milvus/tls/configure/index.md).
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
