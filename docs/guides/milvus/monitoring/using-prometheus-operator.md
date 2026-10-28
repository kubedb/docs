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

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO and creates the object-storage secrets (`milvus-storage-config` for `milvus-standalone`, `milvus-cluster-storage-config` for `milvus-cluster`).

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/monitoring/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/monitoring/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Enable Monitoring in the Milvus Manifest

Monitoring is enabled through `spec.monitor`. The base [standalone](/docs/guides/milvus/quickstart/standalone.md) and [distributed](/docs/guides/milvus/quickstart/distributed.md) quickstarts intentionally omit it so the first deployment only needs MinIO and etcd. Add the following block to your `Milvus` manifest:

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

The `monitoring/yamls/standalone.yaml` and `monitoring/yamls/distributed.yaml` files are the quickstart manifests with this block added. Deploy the standalone database and wait until it is `Ready`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/monitoring/yamls/standalone.yaml
milvus.kubedb.com/milvus-standalone created

$ kubectl get milvuses.kubedb.com -n demo milvus-standalone -w
```

> If you deleted a Milvus of the same name in an earlier guide, delete its object-storage prefix first, as shown in [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#delete-a-milvus-instances-object-storage-data). Otherwise the new database can hang in `Provisioning`.

## Stats Service

When monitoring is enabled, the primary Milvus service still exposes gRPC, metrics, and REST, and KubeDB also creates a dedicated **stats service** named `<db>-stats` for scraping:

```bash
$ kubectl get svc -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                       AGE
milvus-standalone         ClusterIP   10.43.144.154   <none>        19530/TCP,9091/TCP,8080/TCP   91s
milvus-standalone-stats   ClusterIP   10.43.12.191    <none>        9091/TCP                      91s
```

The `ServiceMonitor` targets the dedicated `-stats` service, not the primary service.

## ServiceMonitor

KubeDB also creates a `ServiceMonitor` named `<db>-stats` that selects the stats service:

```bash
$ kubectl get servicemonitor -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                      AGE
milvus-standalone-stats   90s
```

```bash
$ kubectl get servicemonitor milvus-standalone-stats -n demo -o yaml
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
```

Key points:

- The `release: prometheus` label (from `serviceMonitor.labels`) is what lets the Prometheus Operator discover this `ServiceMonitor`.
- The scrape `interval` is `10s`, as configured.
- The endpoint scrapes the `metrics` port at `/metrics`.
- The selector matches the stats service via the `kubedb.com/role: stats` label.

Once the Prometheus Operator reconciles this `ServiceMonitor`, Milvus metrics begin appearing in Prometheus.

## Verify in Prometheus

Port-forward the Prometheus service created by your Prometheus Operator installation (with kube-prometheus-stack release `prometheus` in namespace `monitoring`, it is `prometheus-kube-prometheus-prometheus`; adjust the name and namespace to your installation):

```bash
$ kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Then open `http://localhost:9090/targets` and check that the `serviceMonitor/demo/milvus-standalone-stats/0` target is `UP`, or query the `up` metric for the namespace:

```bash
$ curl -s 'http://localhost:9090/api/v1/query' --data-urlencode 'query=up{namespace="demo"}' | jq '.data.result[] | {job: .metric.job, pod: .metric.pod, value: .value[1]}'
```

Each scraped Milvus pod reports `"value": "1"`. Milvus metrics are prefixed with `milvus_`, for example `milvus_num_node`.

## Distributed Milvus

Monitoring works identically for a distributed Milvus. A single stats service and `ServiceMonitor` named `milvus-cluster-stats` are created, and metrics are scraped from the distributed components (each role's pods expose port `9091`).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/monitoring/yamls/distributed.yaml
milvus.kubedb.com/milvus-cluster created

$ kubectl get svc -n demo -l app.kubernetes.io/instance=milvus-cluster
NAME                           TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                       AGE
milvus-cluster                 ClusterIP   10.43.221.1   <none>        19530/TCP,9091/TCP,8080/TCP   3m
milvus-cluster-datanode        ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-mixcoord        ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-querynode       ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-stats           ClusterIP   10.43.95.57   <none>        9091/TCP    3m
milvus-cluster-streamingnode   ClusterIP   None          <none>        9091/TCP    3m

$ kubectl get servicemonitor milvus-cluster-stats -n demo -o yaml
...
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
      app.kubernetes.io/instance: milvus-cluster
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: milvuses.kubedb.com
      kubedb.com/role: stats
```

In Prometheus, the `milvus-cluster-stats` job has one target per role pod (mixcoord, datanode, querynode, streamingnode and proxy), each with `up` = `1`.

## Cleaning up

```bash
$ kubectl patch -n demo milvus.kubedb.com milvus-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl patch -n demo milvus.kubedb.com milvus-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone milvus-cluster
$ kubectl delete ns demo
```

## Next Steps

- Secure your Milvus database with [TLS/SSL](/docs/guides/milvus/tls/guide.md).
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
