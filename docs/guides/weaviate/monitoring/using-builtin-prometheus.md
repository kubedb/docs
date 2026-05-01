---
title: Monitor Weaviate using Builtin Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: weaviate-using-builtin-prometheus-monitoring
    name: Builtin Prometheus
    parent: weaviate-monitoring
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Weaviate with Builtin Prometheus

This tutorial will show you how to monitor `Weaviate` database using builtin [Prometheus](https://github.com/prometheus/prometheus) scraper.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).

- If you are not familiar with how to configure Prometheus to scrape metrics from various Kubernetes resources, please read the tutorial from [here](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/weaviate/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy the database in `demo` namespace.

```bash
$ kubectl create ns monitoring
namespace/monitoring created

$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Weaviate with Monitoring Enabled

At first, let's deploy a `Weaviate` database with monitoring enabled. Below is the `Weaviate` object that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: builtin-prom-weaviate
  namespace: demo
spec:
  version: "1.26.4"
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

Let's create the `Weaviate` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/monitoring/builtin-prom-weaviate.yaml
weaviate.kubedb.com/builtin-prom-weaviate created
```

Now, wait for the database to go into `Ready` state:

```bash
$ kubectl get weaviate -n demo builtin-prom-weaviate
NAME                     VERSION   STATUS   AGE
builtin-prom-weaviate    1.26.4    Ready    1m
```

KubeDB will create a separate stats service with name `{Weaviate cr name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=builtin-prom-weaviate"
NAME                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
builtin-prom-weaviate           ClusterIP   10.102.7.190    <none>        8080/TCP    87s
builtin-prom-weaviate-stats     ClusterIP   10.102.128.153  <none>        8080/TCP    56s
```

Here, `builtin-prom-weaviate-stats` service has been created for monitoring purpose. Let's describe the service:

```bash
$ kubectl describe svc -n demo builtin-prom-weaviate-stats
Name:              builtin-prom-weaviate-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
                   app.kubernetes.io/instance=builtin-prom-weaviate
                   app.kubernetes.io/managed-by=kubedb.com
                   app.kubernetes.io/name=weaviates.kubedb.com
Annotations:       monitoring.appscode.com/agent: prometheus.io/builtin
                   prometheus.io/path: /metrics
                   prometheus.io/port: 8080
                   prometheus.io/scrape: true
Selector:          app.kubernetes.io/instance=builtin-prom-weaviate,app.kubernetes.io/name=weaviates.kubedb.com
Type:              ClusterIP
Port:              metrics  8080/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.1.5:8080,10.244.1.6:8080,10.244.1.7:8080
```

You can see that the service contains the following annotations:

```
prometheus.io/scrape: true
prometheus.io/path: /metrics
prometheus.io/port: 8080
```

The Prometheus server will discover this service endpoint using these annotations and will scrape metrics from all endpoints.

## Configure Prometheus to Scrape

To get the monitoring of this `Weaviate` database, you need to configure a Prometheus server. Below is the necessary configuration:

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

Once Prometheus is configured and running, you can check the monitoring is working by navigating to the Prometheus dashboard (by default at `localhost:9090`). You should see the `builtin-prom-weaviate-stats` endpoint in the list of scrape targets.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo builtin-prom-weaviate
kubectl delete ns demo
kubectl delete ns monitoring
```
