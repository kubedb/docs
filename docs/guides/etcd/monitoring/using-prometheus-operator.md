---
title: Monitor Etcd using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: etcd-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: etcd-monitoring-guides
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Etcd Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides a simple and Kubernetes-native way to deploy and configure a Prometheus server. This tutorial shows how to use it to monitor an etcd cluster deployed with KubeDB.

> **Read this first:** unlike most KubeDB databases, etcd has **no exporter sidecar**. etcd serves Prometheus metrics natively on its own metrics listener at port `2381`. KubeDB only creates the stats `Service` and the `ServiceMonitor` that point at it — nothing is added to the pod. See the [monitoring overview](/docs/guides/etcd/monitoring/overview.md).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Etcd support is behind an alpha feature gate, so make sure the operator was installed with `--set featureGates.Etcd=true`.

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/etcd/monitoring/overview.md).

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have one, deploy it following the docs [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md).

- If you don't already have a Prometheus server running, deploy one following the tutorial [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` for the monitoring stack, and deploy the database in `demo`.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know which labels a `Prometheus` object uses to select `ServiceMonitor` objects, so we can ask KubeDB to put those labels on the `ServiceMonitor` it generates.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME         AGE
monitoring   prometheus   18m
```

```yaml
$ kubectl get prometheus -n monitoring prometheus -o yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  labels:
    prometheus: prometheus
  name: prometheus
  namespace: monitoring
spec:
  replicas: 1
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      release: prometheus
  serviceMonitorNamespaceSelector: {}
```

Notice `spec.serviceMonitorSelector`: the `release: prometheus` label is what selects a `ServiceMonitor`, so we will pass it through `spec.monitor.prometheus.serviceMonitor.labels`.

> **Watch the namespace selector.** KubeDB creates the `ServiceMonitor` in the **same namespace as the `Etcd` object** — `demo` here, not `monitoring`. Your `Prometheus` object therefore needs a `serviceMonitorNamespaceSelector` that includes `demo` (an empty selector, as above, matches all namespaces). If yours is restricted to the `monitoring` namespace, the target will silently never appear.

## Deploy Etcd with Monitoring Enabled

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: prom-etcd
  namespace: demo
spec:
  version: 3.6.4
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

Here,

- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor`.
- `spec.monitor.prometheus.serviceMonitor.labels` are the labels KubeDB puts on that `ServiceMonitor`.
- `spec.monitor.prometheus.serviceMonitor.interval` is the scrape interval written onto the generated endpoint.

Let's create it:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/monitoring/prom-etcd.yaml
etcd.kubedb.com/prom-etcd created
```

Wait for the cluster to become `Ready`:

```bash
$ kubectl get etcd -n demo prom-etcd
NAME        VERSION   STATUS   AGE
prom-etcd   3.6.4     Ready    2m
```

Note that the member pods report `1/1` containers, not `2/2` — there is no exporter sidecar:

```bash
$ kubectl get pod -n demo -l app.kubernetes.io/instance=prom-etcd
NAME          READY   STATUS    RESTARTS   AGE
prom-etcd-0   1/1     Running   0          2m
prom-etcd-1   1/1     Running   0          2m
prom-etcd-2   1/1     Running   0          2m
```

The single container exposes three ports, of which `metrics` is the one we care about here:

```bash
$ kubectl get pod -n demo prom-etcd-0 -o jsonpath='{.spec.containers[0].ports}'
[{"containerPort":2379,"name":"client","protocol":"TCP"},{"containerPort":2380,"name":"peer","protocol":"TCP"},{"containerPort":2381,"name":"metrics","protocol":"TCP"}]
```

## The stats Service

KubeDB creates a separate stats Service named `{Etcd crd name}-stats`:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=prom-etcd"
NAME              TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE
prom-etcd         ClusterIP   10.102.7.190     <none>        2379/TCP            2m
prom-etcd-pods    ClusterIP   None             <none>        2379/TCP,2380/TCP   2m
prom-etcd-stats   ClusterIP   10.102.128.153   <none>        2381/TCP           2m
```

Here, `prom-etcd` is the load-balanced client Service, `prom-etcd-pods` is the headless governing Service that gives each member its ordinal DNS name, and `prom-etcd-stats` is the monitoring Service. Let's describe it:

```bash
$ kubectl describe svc -n demo prom-etcd-stats
Name:              prom-etcd-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
                   app.kubernetes.io/instance=prom-etcd
                   app.kubernetes.io/managed-by=kubedb.com
                   app.kubernetes.io/name=etcds.kubedb.com
                   kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/instance=prom-etcd,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=etcds.kubedb.com
Type:              ClusterIP
IP:                10.102.128.153
Port:              metrics  2381/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.1.7:2381,10.244.2.9:2381,10.244.3.5:2381
Session Affinity:  None
Events:            <none>
```

Two things to notice:

- The Service port is `2381` (the default `spec.monitor.prometheus.exporter.port` for etcd) and the **target port is `metrics`**, which also resolves to `2381` on the pod — etcd's own metrics listener. There is no intermediate exporter process; Prometheus talks to etcd directly.
- The endpoint list has one address **per member**. Each etcd member reports its own view of the cluster, so `etcd_server_has_leader` is a per-member metric and the whole point is to scrape all three.

## The generated ServiceMonitor

```bash
$ kubectl get servicemonitor -n demo
NAME              AGE
prom-etcd-stats   2m
```

> The `ServiceMonitor` lives in `demo`, the database's namespace — not in `monitoring`.

```yaml
$ kubectl get servicemonitor -n demo prom-etcd-stats -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: prom-etcd
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: etcds.kubedb.com
    release: prometheus
  name: prom-etcd-stats
  namespace: demo
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    controller: true
    kind: Service
    name: prom-etcd-stats
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
      app.kubernetes.io/instance: prom-etcd
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: etcds.kubedb.com
      kubedb.com/role: stats
```

Here,

- `labels` includes `release: prometheus`, exactly what we asked for in the `Etcd` CRD — that is what makes the `Prometheus` object adopt this `ServiceMonitor`.
- `selector.matchLabels` matches the labels on the `prom-etcd-stats` Service, including `kubedb.com/role: stats`.
- `endpoints[0].port: metrics` refers to the **Service port name**, not a number.
- `scheme: http` — the metrics listener is plain HTTP even when `spec.tls` is enabled on the cluster, because it is a separate socket from the mutually authenticated client API. There is no `tlsConfig` on the endpoint.
- The `relabelings` entry copies the pod name onto a `pod` label, so you can tell which member a sample came from.
- The `ServiceMonitor` is owned by the stats Service, so deleting the `Etcd` object garbage-collects it.

## Verify Monitoring Metrics

Find the Prometheus pod:

```bash
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                      READY   STATUS    RESTARTS   AGE
prometheus-prometheus-0   3/3     Running   1          63m
```

Forward its port:

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
```

Open [http://localhost:9090/targets](http://localhost:9090/targets). You should see three targets under the `serviceMonitor/demo/prom-etcd-stats/0` job — one per member — all `UP`, each with an address ending in `:2381`.

Now try a few queries on the graph page:

```promql
# 1 for every member that currently sees a leader. Any 0 here is a problem.
etcd_server_has_leader

# How close the backend is to its quota. Alert well before this reaches 1.
etcd_mvcc_db_total_size_in_bytes / etcd_server_quota_backend_bytes

# Elections over the last hour. Should be flat in a healthy cluster.
increase(etcd_server_leader_changes_seen_total[1h])

# 99th percentile WAL fsync latency, per member.
histogram_quantile(0.99, sum by (le, pod) (rate(etcd_disk_wal_fsync_duration_seconds_bucket[5m])))
```

Because the scrape hits each member individually, the `pod` label added by the relabeling rule lets you break any of these down per member — which is how you spot a single slow disk before it starts causing elections.

You can also use this Prometheus server as a data source for [Grafana](https://grafana.com/). KubeDB does not ship a Grafana dashboard for etcd, but since these are upstream etcd metric names, the community etcd dashboards work as-is.

## Cleaning up

```bash
# cleanup database
$ kubectl delete etcd -n demo prom-etcd

# cleanup prometheus resources
$ kubectl delete -n monitoring prometheus prometheus
$ kubectl delete -n monitoring clusterrolebinding prometheus
$ kubectl delete -n monitoring clusterrole prometheus
$ kubectl delete -n monitoring serviceaccount prometheus
$ kubectl delete -n monitoring service prometheus-operated

# cleanup prometheus operator resources
$ kubectl delete -n monitoring deployment prometheus-operator
$ kubectl delete -n monitoring serviceaccount prometheus-operator
$ kubectl delete clusterrolebinding prometheus-operator
$ kubectl delete clusterrole prometheus-operator

# delete namespaces
$ kubectl delete ns monitoring
$ kubectl delete ns demo
```

## Next Steps

- Monitor your etcd cluster with KubeDB using [builtin Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md).
- Run your etcd cluster with [TLS/SSL encryption](/docs/guides/etcd/tls/configure-ssl.md).
- Tune the backend quota and compaction policy with [custom configuration](/docs/guides/etcd/custom-configuration/using-config.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
