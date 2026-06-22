---
title: Monitor HanaDB using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: hanadb-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: hanadb-monitoring
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# Monitoring HanaDB with Prometheus Operator

This tutorial shows how to monitor a HanaDB instance using [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator).

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl` to communicate with it. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster by following the [setup guide](/docs/setup/README.md).

- Install Prometheus Operator in your cluster by following the [Prometheus Operator setup guide](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/coreos-operator). If you want to use an already deployed Prometheus instance, configure it to monitor all namespaces.

- To learn how Prometheus monitoring works with KubeDB in general, read the [HanaDB monitoring overview](/docs/guides/hanadb/monitoring/overview.md).

- This tutorial deploys Prometheus resources in the `monitoring` namespace and the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy HanaDB with Monitoring Enabled

Deploy a HanaDB instance with monitoring enabled using Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-prometheus-operator
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
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9668
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here,

- `spec.monitor.agent: prometheus.io/operator` tells KubeDB that we want to monitor using Prometheus Operator.
- `spec.monitor.prometheus.exporter.port` specifies the exporter port. If omitted, KubeDB defaults it to `9668`.
- `spec.monitor.prometheus.serviceMonitor.labels` specifies labels to add to the ServiceMonitor. The Prometheus CR must have matching labels in its `serviceMonitorSelector`.
- `spec.monitor.prometheus.serviceMonitor.interval` specifies the scrape interval.

Create the HanaDB object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/monitoring/coreos-prom-hanadb.yaml
hanadb.kubedb.com/hanadb-prometheus-operator created
```

Wait for the database to reach the `Ready` state.

```bash
$ kubectl get hanadb -n demo hanadb-prometheus-operator
NAME                         VERSION   STATUS   AGE
hanadb-prometheus-operator   2.0.82    Ready    2m
```

KubeDB will create a ServiceMonitor and stats service for this HanaDB instance.

```bash
$ kubectl get servicemonitor -n demo
NAME                              AGE
hanadb-prometheus-operator-stats    2m
```

Verify the `ServiceMonitor`:

```yaml
$ kubectl get servicemonitor -n demo hanadb-prometheus-operator-stats -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: hanadb-prometheus-operator
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: hanadbs.kubedb.com
    release: prometheus
  name: hanadb-prometheus-operator-stats
  namespace: demo
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
      app.kubernetes.io/instance: hanadb-prometheus-operator
      app.kubernetes.io/name: hanadbs.kubedb.com
```

Prometheus Operator will automatically pick up this ServiceMonitor and start scraping metrics from the HanaDB stats service.

## Access Prometheus Dashboard

To verify, port-forward the Prometheus service and visit `http://localhost:9090/targets`:

```bash
$ kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090
```

You should see `demo/hanadb-prometheus-operator-stats` target in an UP state.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/hanadb-prometheus-operator -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/hanadb-prometheus-operator

kubectl delete ns demo
kubectl delete ns monitoring
```
