---
title: Monitor Oracle using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: oracle-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: oracle-monitoring
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Oracle Using Prometheus Operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides a simple and Kubernetes-native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor `Oracle` database deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/oracle/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy the database in `demo` namespace.

```bash
$ kubectl create ns monitoring
namespace/monitoring created

$ kubectl create ns demo
namespace/demo created
```

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md).

- If you don't already have a Prometheus server running, deploy one following the tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

> **Note:** YAML files used in this tutorial are stored in [docs/examples/oracle/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` CR. We are going to provide these labels in `spec.monitor.prometheus.serviceMonitor.labels` field of the `Oracle` CR so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster:

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME         AGE
monitoring   prometheus   18m
```

Now, let's view the YAML of the available Prometheus server `prometheus` in `monitoring` namespace:

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
  resources:
    requests:
      memory: 400Mi
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      release: prometheus
```

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` CR. So, we are going to use this label in `spec.monitor.prometheus.serviceMonitor.labels` field of the `Oracle` CR.

## Deploy Oracle with Monitoring Enabled

At first, let's deploy a `Oracle` database with monitoring enabled. Below is the `Oracle` object that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: coreos-prom-oracle
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    storageClassName: "standard"
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

- `spec.monitor.agent: prometheus.io/operator` specifies that we are going to monitor this server using Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.
- `spec.monitor.prometheus.serviceMonitor.interval` specifies how frequently Prometheus should scrape this database.

Let's create the `Oracle` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/monitoring/coreos-prom-oracle.yaml
oracle.kubedb.com/coreos-prom-oracle created
```

Now, wait for the database to go into `Ready` state:

```bash
$ kubectl get oracle -n demo coreos-prom-oracle
NAME                  VERSION   STATUS   AGE
coreos-prom-oracle    1.17.0    Ready    1m
```

KubeDB will create a `ServiceMonitor` object for this `Oracle` database:

```bash
$ kubectl get servicemonitor -n demo
NAME                       AGE
coreos-prom-oracle         65s
```

Let's verify the `ServiceMonitor` has the labels we specified:

```bash
$ kubectl get servicemonitor -n demo coreos-prom-oracle -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: coreos-prom-oracle
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: oracles.kubedb.com
    release: prometheus
  name: coreos-prom-oracle
  namespace: demo
spec:
  endpoints:
  - honorLabels: true
    interval: 10s
    path: /metrics
    port: metrics
  namespaceSelector:
    matchNames:
    - demo
  selector:
    matchLabels:
      app.kubernetes.io/instance: coreos-prom-oracle
      app.kubernetes.io/name: oracles.kubedb.com
```

Notice that the `ServiceMonitor` has the label `release: prometheus`, which will be picked up by the Prometheus server.

## Verify Monitoring

Once everything is set up, you can visit the Prometheus dashboard. The `coreos-prom-oracle` service monitor will be discovered and its metrics will be scraped. You should be able to query Oracle metrics in Prometheus.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracle -n demo coreos-prom-oracle
kubectl delete ns demo
kubectl delete ns monitoring
```