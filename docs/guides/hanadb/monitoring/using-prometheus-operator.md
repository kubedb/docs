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

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring HanaDB with Prometheus Operator

This tutorial will show you how to monitor HanaDB database using [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- Install Prometheus Operator in your cluster following the steps from [here](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/coreos-operator). If you want to use an already deployed Prometheus instance, configure it to monitor all namespaces.

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/hanadb/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy the database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy HanaDB with Monitoring Enabled

At first, let's deploy a HanaDB database with monitoring enabled using the Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: coreos-prom-hanadb
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here,

- `spec.monitor.agent: prometheus.io/operator` tells KubeDB that we want to monitor using Prometheus Operator.
- `spec.monitor.prometheus.serviceMonitor.labels` specifies labels to add to the ServiceMonitor. The Prometheus CR must have matching labels in its `serviceMonitorSelector`.
- `spec.monitor.prometheus.serviceMonitor.interval` specifies the scrape interval.

Let's create the HanaDB CR:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/monitoring/coreos-prom-hanadb.yaml
hanadb.kubedb.com/coreos-prom-hanadb created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get hanadb -n demo coreos-prom-hanadb
NAME                 VERSION   STATUS   AGE
coreos-prom-hanadb   2.0       Ready    2m
```

KubeDB will create a ServiceMonitor and stats service for this HanaDB instance.

```bash
$ kubectl get servicemonitor -n demo
NAME                        AGE
coreos-prom-hanadb-stats    2m
```

Let's verify the ServiceMonitor:

```yaml
$ kubectl get servicemonitor -n demo coreos-prom-hanadb-stats -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: coreos-prom-hanadb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: hanadbs.kubedb.com
    release: prometheus
  name: coreos-prom-hanadb-stats
  namespace: demo
spec:
  endpoints:
  - honorLabels: true
    interval: 10s
    path: /metrics
    port: prom-http
  namespaceSelector:
    matchNames:
    - demo
  selector:
    matchLabels:
      app.kubernetes.io/instance: coreos-prom-hanadb
      app.kubernetes.io/name: hanadbs.kubedb.com
```

Prometheus Operator will automatically pick up this ServiceMonitor and start scraping metrics from the HanaDB stats service.

## Access Prometheus Dashboard

To verify, port-forward the Prometheus service and visit `http://localhost:9090/targets`:

```bash
$ kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090
```

You should see `demo/coreos-prom-hanadb-stats` target in an UP state.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/coreos-prom-hanadb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/coreos-prom-hanadb

kubectl delete ns demo
kubectl delete ns monitoring
```
