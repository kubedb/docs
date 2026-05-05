---
title: Monitor HanaDB using Built-in Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: hanadb-using-builtin-prometheus-monitoring
    name: Built-in Prometheus
    parent: hanadb-monitoring
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring HanaDB with Built-in Prometheus

This tutorial shows how to monitor a HanaDB instance using the built-in [Prometheus](https://github.com/prometheus/prometheus) scraper.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl` to communicate with it. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- If you are not familiar with how to configure Prometheus to scrape metrics from various Kubernetes resources, please read the tutorial from [here](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/hanadb/monitoring/overview.md).

- This tutorial deploys Prometheus resources in the `monitoring` namespace and the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy HanaDB with Monitoring Enabled

Deploy a HanaDB instance with monitoring enabled. The manifest is shown below.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: builtin-prom-hanadb
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/builtin
    prometheus:
      exporter:
        port: 9668
```

Here, `spec.monitor.agent: prometheus.io/builtin` tells KubeDB to use Prometheus annotation-based discovery. `spec.monitor.prometheus.exporter.port` specifies the exporter port. If omitted, KubeDB defaults it to `9668`.

Create the HanaDB object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/monitoring/builtin-prom-hanadb.yaml
hanadb.kubedb.com/builtin-prom-hanadb created
```

Wait for the database to reach the `Ready` state.

```bash
$ kubectl get hanadb -n demo builtin-prom-hanadb
NAME                  VERSION   STATUS   AGE
builtin-prom-hanadb   2.0.82    Ready    2m
```

KubeDB creates a separate stats service named `{hanadb-name}-stats` for metrics scraping.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=builtin-prom-hanadb"
NAME                         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
builtin-prom-hanadb          ClusterIP   10.96.100.10    <none>        39017/TCP   2m
builtin-prom-hanadb-stats    ClusterIP   10.96.100.11    <none>        9668/TCP    90s
```

The `builtin-prom-hanadb-stats` service exposes the exporter endpoint. Describe the service:

```bash
$ kubectl describe svc -n demo builtin-prom-hanadb-stats
Name:              builtin-prom-hanadb-stats
Namespace:         demo
Labels:            app.kubernetes.io/name=hanadbs.kubedb.com
                   app.kubernetes.io/instance=builtin-prom-hanadb
Annotations:       monitoring.appscode.com/agent: prometheus.io/builtin
                   prometheus.io/path: /metrics
                   prometheus.io/port: 9668
                   prometheus.io/scheme: http
                   prometheus.io/scrape: true
Selector:          app.kubernetes.io/name=hanadbs.kubedb.com,app.kubernetes.io/instance=builtin-prom-hanadb
Type:              ClusterIP
Port:              metrics  9668/TCP
```

The service contains the following annotations, which are used by Prometheus to discover the endpoint:

```
prometheus.io/path: /metrics
prometheus.io/port: 9668
prometheus.io/scheme: http
prometheus.io/scrape: true
```

## Configure Prometheus

Configure Prometheus to scrape metrics from this service. Add the following `scrape_config` to your Prometheus configuration:

```yaml
scrape_configs:
- job_name: kubedb-hanadbs
  honor_labels: true
  kubernetes_sd_configs:
  - role: endpoints
  relabel_configs:
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
    separator: ;
    regex: true
    target_label: __tmp_prometheus_service_scrape
    replacement: $1
    action: keep
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
    separator: ;
    regex: (https?)
    target_label: __scheme__
    replacement: $1
    action: replace
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
    separator: ;
    regex: (.+)
    target_label: __metrics_path__
    replacement: $1
    action: replace
  - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
    separator: ;
    regex: ([^:]+)(?::\d+)?;(\d+)
    target_label: __address__
    replacement: $1:$2
    action: replace
```

Now Prometheus will discover the HanaDB stats service and scrape metrics automatically.

## Access Prometheus Dashboard

To access the Prometheus dashboard, port-forward the Prometheus service and visit `http://localhost:9090` in your browser.

```bash
$ kubectl port-forward -n monitoring svc/prometheus 9090:9090
```

You should see the HanaDB metrics in the Prometheus dashboard under the `kubedb-hanadbs` job.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/builtin-prom-hanadb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/builtin-prom-hanadb

kubectl delete ns demo
kubectl delete ns monitoring
```
