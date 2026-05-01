---
title: Monitor Oracle using Builtin Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: oracle-using-builtin-prometheus-monitoring
    name: Builtin Prometheus
    parent: oracle-monitoring
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Oracle with Builtin Prometheus

This tutorial will show you how to monitor `Oracle` database using builtin [Prometheus](https://github.com/prometheus/prometheus) scraper.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).

- If you are not familiar with how to configure Prometheus to scrape metrics from various Kubernetes resources, please read the tutorial from [here](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/oracle/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy the database in `demo` namespace.

```bash
$ kubectl create ns monitoring
namespace/monitoring created

$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/oracle/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Oracle with Monitoring Enabled

At first, let's deploy a `Oracle` database with monitoring enabled. Below is the `Oracle` object that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: builtin-prom-oracle
  namespace: demo
spec:
  version: "21.3.0"
  replicas: 3
  storage:
    storageClassName: "standard"
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

- `spec.monitor.agent: prometheus.io/builtin` specifies that we are going to monitor this server using builtin Prometheus scraper.

Let's create the `Oracle` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/monitoring/builtin-prom-oracle.yaml
oracle.kubedb.com/builtin-prom-oracle created
```

Now, wait for the database to go into `Ready` state:

```bash
$ kubectl get oracle -n demo builtin-prom-oracle
NAME                   VERSION   STATUS   AGE
builtin-prom-oracle    1.17.0    Ready    1m
```

KubeDB will create a separate stats service with name `{Oracle cr name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=builtin-prom-oracle"
NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
builtin-prom-oracle           ClusterIP   10.102.7.190    <none>        1521/TCP    87s
builtin-prom-oracle-stats     ClusterIP   10.102.128.153  <none>        1521/TCP    56s
```

Here, `builtin-prom-oracle-stats` service has been created for monitoring purpose. Let's describe the service:

```bash
$ kubectl describe svc -n demo builtin-prom-oracle-stats
Name:              builtin-prom-oracle-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
                   app.kubernetes.io/instance=builtin-prom-oracle
                   app.kubernetes.io/managed-by=kubedb.com
                   app.kubernetes.io/name=oracles.kubedb.com
Annotations:       monitoring.appscode.com/agent: prometheus.io/builtin
                   prometheus.io/path: /metrics
                   prometheus.io/port: 1521
                   prometheus.io/scrape: true
Selector:          app.kubernetes.io/instance=builtin-prom-oracle,app.kubernetes.io/name=oracles.kubedb.com
Type:              ClusterIP
Port:              metrics  1521/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.1.5:1521,10.244.1.6:1521,10.244.1.7:1521
```

You can see that the service contains the following annotations:

```
prometheus.io/scrape: true
prometheus.io/path: /metrics
prometheus.io/port: 1521
```

The Prometheus server will discover this service endpoint using these annotations and will scrape metrics from all endpoints.

## Configure Prometheus to Scrape

To get the monitoring of this `Oracle` database, you need to configure a Prometheus server. Below is the necessary configuration:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
- job_name: 'kubedb-databases'
  kubernetes_sd_configs:
  - role: endpoints
  relabel_configs:
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
    regex: true
    action: keep
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
    regex: (.+)
    target_label: __metrics_path__
    action: replace
```

## Verify Monitoring

Once Prometheus is configured and running, you can check the monitoring is working by navigating to the Prometheus dashboard (by default at `localhost:9090`). You should see the `builtin-prom-oracle-stats` endpoint in the list of scrape targets.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracle -n demo builtin-prom-oracle
kubectl delete ns demo
kubectl delete ns monitoring
```