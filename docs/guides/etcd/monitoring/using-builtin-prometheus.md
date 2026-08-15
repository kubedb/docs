---
title: Monitor Etcd using Builtin Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: etcd-using-builtin-prometheus-monitoring
    name: Builtin Prometheus
    parent: etcd-monitoring-guides
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Etcd with builtin Prometheus

This tutorial shows how to monitor a KubeDB-managed etcd cluster with a plain [Prometheus](https://github.com/prometheus/prometheus) server — no Prometheus operator, no CRDs, just annotation-based service discovery.

> **Read this first:** unlike most KubeDB databases, etcd has **no exporter sidecar**. etcd serves Prometheus metrics natively on its own metrics listener at port `2381`. When you set `spec.monitor`, KubeDB adds nothing to the pod; it only creates and annotates a stats `Service` so Prometheus can find what etcd is already serving. See the [monitoring overview](/docs/guides/etcd/monitoring/overview.md).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Etcd support is behind an alpha feature gate, so make sure the operator was installed with `--set featureGates.Etcd=true`.

- If you are not familiar with how to configure Prometheus to scrape metrics from Kubernetes resources, please read the tutorial [here](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/etcd/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` for the monitoring stack, and deploy the database in `demo`.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Etcd with Monitoring Enabled

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: builtin-prom-etcd
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
    agent: prometheus.io/builtin
  deletionPolicy: WipeOut
```

Here,

- `spec.monitor.agent: prometheus.io/builtin` tells KubeDB to annotate the stats Service for annotation-based discovery.

Create it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/monitoring/builtin-prom-etcd.yaml
etcd.kubedb.com/builtin-prom-etcd created
```

Wait for the cluster to become `Ready`:

```bash
$ kubectl get etcd -n demo builtin-prom-etcd
NAME                VERSION   STATUS   AGE
builtin-prom-etcd   3.6.4     Ready    2m
```

Note the member pods report `1/1`, not `2/2` — there is no exporter container:

```bash
$ kubectl get pod -n demo -l app.kubernetes.io/instance=builtin-prom-etcd
NAME                  READY   STATUS    RESTARTS   AGE
builtin-prom-etcd-0   1/1     Running   0          2m
builtin-prom-etcd-1   1/1     Running   0          2m
builtin-prom-etcd-2   1/1     Running   0          2m
```

## The stats Service

KubeDB creates a separate stats Service named `{Etcd crd name}-stats`:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=builtin-prom-etcd"
NAME                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE
builtin-prom-etcd         ClusterIP   10.102.7.190     <none>        2379/TCP            2m
builtin-prom-etcd-pods    ClusterIP   None             <none>        2379/TCP,2380/TCP   2m
builtin-prom-etcd-stats   ClusterIP   10.102.128.153   <none>        56790/TCP           2m
```

Let's describe the stats Service:

```bash
$ kubectl describe svc -n demo builtin-prom-etcd-stats
Name:              builtin-prom-etcd-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
                   app.kubernetes.io/instance=builtin-prom-etcd
                   app.kubernetes.io/managed-by=kubedb.com
                   app.kubernetes.io/name=etcds.kubedb.com
                   kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/builtin
                   prometheus.io/path: /metrics
                   prometheus.io/port: 56790
                   prometheus.io/scheme: http
                   prometheus.io/scrape: true
Selector:          app.kubernetes.io/instance=builtin-prom-etcd,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=etcds.kubedb.com
Type:              ClusterIP
IP:                10.102.128.153
Port:              metrics  56790/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.1.7:2381,10.244.2.9:2381,10.244.3.5:2381
Session Affinity:  None
Events:            <none>
```

The annotations that matter for discovery are:

```bash
prometheus.io/scrape: true
prometheus.io/scheme: http
prometheus.io/path: /metrics
prometheus.io/port: 56790
```

The Prometheus server will discover the Service endpoints using these and scrape metrics **straight from etcd** — the `TargetPort: metrics` resolves to port `2381` on each member pod. Note there is one endpoint per member: each etcd member reports its own view of the cluster, so scraping all of them is the point.

`prometheus.io/scheme` is `http` even if you have TLS enabled on the cluster, because the metrics listener is a separate socket from the mutually authenticated client API. See the [monitoring overview](/docs/guides/etcd/monitoring/overview.md) for why.

## Configure Prometheus Server

Now we have to configure a Prometheus scraping job to scrape metrics using this Service. We are going to configure a scraping job similar to this [kubernetes-service-endpoints](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin#kubernetes-service-endpoints) job that scrapes metrics from the endpoints of a service.

```yaml
- job_name: 'kubedb-databases'
  honor_labels: true
  scheme: http
  kubernetes_sd_configs:
  - role: endpoints
  # by default Prometheus server selects all Kubernetes services as possible targets.
  # relabel_config is used to filter only the desired endpoints
  relabel_configs:
  # keep only those services that have the "prometheus.io/scrape" and "prometheus.io/port" annotations
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape, __meta_kubernetes_service_annotation_prometheus_io_port]
    separator: ;
    regex: true;(.*)
    action: keep
  # KubeDB stats services use the "http" scheme, so drop any service that uses "https".
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
    action: drop
    regex: https
  # only keep the stats services created by KubeDB for monitoring purposes, which have a "-stats" suffix
  - source_labels: [__meta_kubernetes_service_name]
    separator: ;
    regex: (.*-stats)
    action: keep
  # services created by KubeDB have the "app.kubernetes.io/name" label. keep only those services.
  - source_labels: [__meta_kubernetes_service_label_app_kubernetes_io_name]
    separator: ;
    regex: (.*)
    action: keep
  # read the metric path from the "prometheus.io/path: <path>" annotation
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
    action: replace
    target_label: __metrics_path__
    regex: (.+)
  # read the port from the "prometheus.io/port: <port>" annotation and update the scraping address accordingly
  - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
    action: replace
    target_label: __address__
    regex: ([^:]+)(?::\d+)?;(\d+)
    replacement: $1:$2
  # add the service namespace as a label on the scraped metrics
  - source_labels: [__meta_kubernetes_namespace]
    separator: ;
    regex: (.*)
    target_label: namespace
    replacement: $1
    action: replace
  # add the service name as a label on the scraped metrics
  - source_labels: [__meta_kubernetes_service_name]
    separator: ;
    regex: (.*)
    target_label: service
    replacement: $1
    action: replace
  # add the pod name as a label, so you can tell which etcd member a sample came from
  - source_labels: [__meta_kubernetes_pod_name]
    separator: ;
    regex: (.*)
    target_label: pod
    replacement: $1
    action: replace
  # add the stats service's labels to the scraped metrics
  - action: labelmap
    regex: __meta_kubernetes_service_label_(.+)
```

The `pod` relabeling rule at the end is worth keeping for etcd specifically: every member is scraped separately and reports its own view of the cluster, so without a `pod` label you cannot tell which member reported `etcd_server_has_leader 0`.

### Configure Existing Prometheus Server

If you already have a Prometheus server running, add the scraping job above to the `ConfigMap` used to configure it, then restart it for the updated configuration to take effect.

> If you don't use a persistent volume for Prometheus storage, you will lose your previously scraped data on restart.

### Deploy New Prometheus Server

If you don't have a Prometheus server yet, deploy one in the `monitoring` namespace to collect metrics using this stats Service.

**Create ConfigMap:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  labels:
    app: prometheus-demo
  namespace: monitoring
data:
  prometheus.yml: |-
    global:
      scrape_interval: 5s
      evaluation_interval: 5s
    scrape_configs:
    - job_name: 'kubedb-databases'
      honor_labels: true
      scheme: http
      kubernetes_sd_configs:
      - role: endpoints
      relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape, __meta_kubernetes_service_annotation_prometheus_io_port]
        separator: ;
        regex: true;(.*)
        action: keep
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: drop
        regex: https
      - source_labels: [__meta_kubernetes_service_name]
        separator: ;
        regex: (.*-stats)
        action: keep
      - source_labels: [__meta_kubernetes_service_label_app_kubernetes_io_name]
        separator: ;
        regex: (.*)
        action: keep
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      - source_labels: [__meta_kubernetes_namespace]
        separator: ;
        regex: (.*)
        target_label: namespace
        replacement: $1
        action: replace
      - source_labels: [__meta_kubernetes_service_name]
        separator: ;
        regex: (.*)
        target_label: service
        replacement: $1
        action: replace
      - source_labels: [__meta_kubernetes_pod_name]
        separator: ;
        regex: (.*)
        target_label: pod
        replacement: $1
        action: replace
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/monitoring/prom-config.yaml
configmap/prometheus-config created
```

**Create RBAC:**

If you are using an RBAC-enabled cluster, give Prometheus the necessary permissions:

```bash
$ kubectl apply -f https://github.com/appscode/third-party-tools/raw/master/monitoring/prometheus/builtin/artifacts/rbac.yaml
clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created
```

> YAML for the RBAC resources created above can be found [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/artifacts/rbac.yaml).

**Deploy Prometheus:**

```bash
$ kubectl apply -f https://github.com/appscode/third-party-tools/raw/master/monitoring/prometheus/builtin/artifacts/deployment.yaml
deployment.apps/prometheus created
```

## Verify Monitoring Metrics

Prometheus listens on port `9090`. Use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to reach its dashboard.

```bash
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                          READY   STATUS    RESTARTS   AGE
prometheus-8568c86d86-95zhn   1/1     Running   0          77s
```

```bash
$ kubectl port-forward -n monitoring prometheus-8568c86d86-95zhn 9090
Forwarding from 127.0.0.1:9090 -> 9090
```

Open [http://localhost:9090/targets](http://localhost:9090/targets). Under the `kubedb-databases` job you should see **three** targets — one per etcd member — each with an address ending in `:2381`, and each carrying the `service="builtin-prom-etcd-stats"`, `namespace="demo"` and `pod="builtin-prom-etcd-N"` labels.

Now try a few queries on the graph page. These are upstream etcd's own metric names, since etcd is doing the exporting:

```promql
# 1 for every member that currently sees a leader. Any 0 here is a problem.
etcd_server_has_leader

# Current backend database size, per member.
etcd_mvcc_db_total_size_in_bytes

# How close the backend is to its configured --quota-backend-bytes.
etcd_mvcc_db_total_size_in_bytes / etcd_server_quota_backend_bytes

# Leader elections over the last hour. Should be flat in a healthy cluster.
increase(etcd_server_leader_changes_seen_total[1h])
```

You can also use this Prometheus server as a data source for [Grafana](https://grafana.com/). KubeDB does not ship a Grafana dashboard for etcd, but since these are upstream etcd metric names, community etcd dashboards work as-is.

## Cleaning up

```bash
$ kubectl delete etcd -n demo builtin-prom-etcd

$ kubectl delete -n monitoring deployment.apps/prometheus
$ kubectl delete -n monitoring configmap/prometheus-config

$ kubectl delete clusterrole.rbac.authorization.k8s.io/prometheus
$ kubectl delete -n monitoring serviceaccount/prometheus
$ kubectl delete clusterrolebinding.rbac.authorization.k8s.io/prometheus

$ kubectl delete ns demo
$ kubectl delete ns monitoring
```

## Next Steps

- Monitor your etcd cluster with KubeDB using the [`out-of-the-box` Prometheus operator](/docs/guides/etcd/monitoring/using-prometheus-operator.md).
- Run your etcd cluster with [TLS/SSL encryption](/docs/guides/etcd/tls/configure-ssl.md).
- Tune the backend quota and compaction policy with [custom configuration](/docs/guides/etcd/custom-configuration/using-config.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
