---
title: Monitor Druid using Builtin Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: guides-druid-monitoring-builtin-monitoring
    name: Builtin Prometheus
    parent: guides-druid-monitoring
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Druid with builtin Prometheus

This tutorial will show you how to monitor Druid cluster using builtin [Prometheus](https://github.com/prometheus/prometheus) scraper.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- If you are not familiar with how to configure Prometheus to scrape metrics from various Kubernetes resources, please read the tutorial from [here](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/druid/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/druid](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/druid) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Druid with Monitoring Enabled

At first, let's deploy a Druid cluster with monitoring enabled. Below is the Druid object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-with-monitoring
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  monitor:
    agent: prometheus.io/builtin
    prometheus:
      exporter:
        port: 56790
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

Here,

- `spec.monitor.agent: prometheus.io/builtin` specifies that we are going to monitor this server using builtin Prometheus scraper.
- `spec.monitor.prometheus.exporter.port: 56790` specifies the port where the exporter is running.

Let's create the Druid crd we have shown above.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/monitoring/yamls/druid-builtin-monitoring.yaml
druid.kubedb.com/druid-with-monitoring created
```

Now, wait for the cluster to go into `Ready` state.

```bash
NAME                    TYPE                  VERSION    STATUS           AGE
druid-with-monitoring   kubedb.com/v1alpha2   28.0.1     Ready          31s
```

KubeDB will create a separate stats service with name `{Druid crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=druid-with-monitoring"
NAME                                  TYPE          CLUSTER-IP      EXTERNAL-IP   PORT(S)                                                  AGE
druid-with-monitoring-brokers         ClusterIP     10.96.28.252    <none>        8082/TCP                                                2m13s
druid-with-monitoring-coordinators    ClusterIP     10.96.52.186    <none>        8081/TCP                                                2m13s
druid-with-monitoring-pods            ClusterIP     None            <none>        8081/TCP,8090/TCP,8083/TCP,8091/TCP,8082/TCP,8888/TCP   2m13s
druid-with-monitoring-routers         ClusterIP     10.96.134.202   <none>        8888/TCP                                                2m13s
druid-with-monitoring-stats           ClusterIP     10.96.222.96    <none>        56790/TCP                                               2m13s
```

Here, `druid-with-monitoring-stats` service has been created for monitoring purpose. Let's describe the service.

```bash
$ kubectl describe svc -n druid-demo builtin-prom-stats
Name:              druid-with-monitoring-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
                   app.kubernetes.io/instance=druid-with-monitoring
                   app.kubernetes.io/managed-by=kubedb.com
                   app.kubernetes.io/name=druids.kubedb.com
                   kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/builtin
                   prometheus.io/path: /metrics
                   prometheus.io/port: 56790
                   prometheus.io/scrape: true
Selector:          app.kubernetes.io/instance=druid-with-monitoring,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                10.96.222.96
IPs:               10.96.222.96
Port:              metrics  56790/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.0.31:56790,10.244.0.33:56790
Session Affinity:  None
Events:            <none>
```

You can see that the service contains following annotations.

```bash
prometheus.io/path: /metrics
prometheus.io/port: 56790
prometheus.io/scrape: true
```

The Prometheus server will discover the service endpoint using these specifications and will scrape metrics from the exporter.

## Configure Prometheus Server

Now, we have to configure a Prometheus scraping job to scrape the metrics using this service. We are going to configure scraping job similar to this [kubernetes-service-endpoints](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin#kubernetes-service-endpoints) job that scrapes metrics from endpoints of a service.

Let's configure a Prometheus scraping job to collect metrics from this service.

```yaml
- job_name: 'kubedb-databases'
  honor_labels: true
  scheme: http
  kubernetes_sd_configs:
  - role: endpoints
  # by default Prometheus server select all Kubernetes services as possible target.
  # relabel_config is used to filter only desired endpoints
  relabel_configs:
  # keep only those services that has "prometheus.io/scrape","prometheus.io/path" and "prometheus.io/port" anootations
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape, __meta_kubernetes_service_annotation_prometheus_io_port]
    separator: ;
    regex: true;(.*)
    action: keep
  # currently KubeDB supported databases uses only "http" scheme to export metrics. so, drop any service that uses "https" scheme.
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
    action: drop
    regex: https
  # only keep the stats services created by KubeDB for monitoring purpose which has "-stats" suffix
  - source_labels: [__meta_kubernetes_service_name]
    separator: ;
    regex: (.*-stats)
    action: keep
  # service created by KubeDB will have "app.kubernetes.io/name" and "app.kubernetes.io/instance" annotations. keep only those services that have these annotations.
  - source_labels: [__meta_kubernetes_service_label_app_kubernetes_io_name]
    separator: ;
    regex: (.*)
    action: keep
  # read the metric path from "prometheus.io/path: <path>" annotation
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
    action: replace
    target_label: __metrics_path__
    regex: (.+)
  # read the port from "prometheus.io/port: <port>" annotation and update scraping address accordingly
  - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
    action: replace
    target_label: __address__
    regex: ([^:]+)(?::\d+)?;(\d+)
    replacement: $1:$2
  # add service namespace as label to the scraped metrics
  - source_labels: [__meta_kubernetes_namespace]
    separator: ;
    regex: (.*)
    target_label: namespace
    replacement: $1
    action: replace
  # add service name as a label to the scraped metrics
  - source_labels: [__meta_kubernetes_service_name]
    separator: ;
    regex: (.*)
    target_label: service
    replacement: $1
    action: replace
  # add stats service's labels to the scraped metrics
  - action: labelmap
    regex: __meta_kubernetes_service_label_(.+)
```

### Configure Existing Prometheus Server

If you already have a Prometheus server running, you have to add above scraping job in the `ConfigMap` used to configure the Prometheus server. Then, you have to restart it for the updated configuration to take effect.

>If you don't use a persistent volume for Prometheus storage, you will lose your previously scraped data on restart.

### Deploy New Prometheus Server

If you don't have any existing Prometheus server running, you have to deploy one. In this section, we are going to deploy a Prometheus server in `monitoring` namespace to collect metrics using this stats service.

**Create ConfigMap:**

At first, create a ConfigMap with the scraping configuration. Bellow, the YAML of ConfigMap that we are going to create in this tutorial.

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
      # by default Prometheus server select all Kubernetes services as possible target.
      # relabel_config is used to filter only desired endpoints
      relabel_configs:
      # keep only those services that has "prometheus.io/scrape","prometheus.io/path" and "prometheus.io/port" anootations
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape, __meta_kubernetes_service_annotation_prometheus_io_port]
        separator: ;
        regex: true;(.*)
        action: keep
      # currently KubeDB supported databases uses only "http" scheme to export metrics. so, drop any service that uses "https" scheme.
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: drop
        regex: https
      # only keep the stats services created by KubeDB for monitoring purpose which has "-stats" suffix
      - source_labels: [__meta_kubernetes_service_name]
        separator: ;
        regex: (.*-stats)
        action: keep
      # service created by KubeDB will have "app.kubernetes.io/name" and "app.kubernetes.io/instance" annotations. keep only those services that have these annotations.
      - source_labels: [__meta_kubernetes_service_label_app_kubernetes_io_name]
        separator: ;
        regex: (.*)
        action: keep
      # read the metric path from "prometheus.io/path: <path>" annotation
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      # read the port from "prometheus.io/port: <port>" annotation and update scraping address accordingly
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      # add service namespace as label to the scraped metrics
      - source_labels: [__meta_kubernetes_namespace]
        separator: ;
        regex: (.*)
        target_label: namespace
        replacement: $1
        action: replace
      # add service name as a label to the scraped metrics
      - source_labels: [__meta_kubernetes_service_name]
        separator: ;
        regex: (.*)
        target_label: service
        replacement: $1
        action: replace
      # add stats service's labels to the scraped metrics
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
```

Let's create above `ConfigMap`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/monitoring/builtin-prometheus/prom-config.yaml
configmap/prometheus-config created
```

**Create RBAC:**

If you are using an RBAC enabled cluster, you have to give necessary RBAC permissions for Prometheus. Let's create necessary RBAC stuffs for Prometheus,

```bash
$ kubectl apply -f https://github.com/appscode/third-party-tools/raw/master/monitoring/prometheus/builtin/artifacts/rbac.yaml
clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created
```

>YAML for the RBAC resources created above can be found [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/artifacts/rbac.yaml).

**Deploy Prometheus:**

Now, we are ready to deploy Prometheus server. We are going to use following [deployment](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/artifacts/deployment.yaml) to deploy Prometheus server.

Let's deploy the Prometheus server.

```bash
$ kubectl apply -f https://github.com/appscode/third-party-tools/raw/master/monitoring/prometheus/builtin/artifacts/deployment.yaml
deployment.apps/prometheus created
```

### Verify Monitoring Metrics

Prometheus server is listening to port `9090`. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

At first, let's check if the Prometheus pod is in `Running` state.

```bash
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                          READY   STATUS    RESTARTS   AGE
prometheus-7bd56c6865-8dlpv   1/1     Running   0          28s
```

Now, run following command on a separate terminal to forward 9090 port of `prometheus-7bd56c6865-8dlpv` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-7bd56c6865-8dlpv 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see the endpoint of `druid-with-monitoring-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" height="100%" src="/docs/guides/druid/monitoring/images/druid-prometheus.png" style="padding:10px">
</p>

Check the labels marked with red rectangle. These labels confirm that the metrics are coming from `Druid` cluster `druid-with-monitoring` through stats service `druid-with-monitoring-stats`.

Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboard with collected metrics.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run following commands

```bash
kubectl delete -n demo druid/druid-with-monitoring

kubectl delete -n monitoring deployment.apps/prometheus

kubectl delete -n monitoring clusterrole.rbac.authorization.k8s.io/prometheus
kubectl delete -n monitoring serviceaccount/prometheus
kubectl delete -n monitoring clusterrolebinding.rbac.authorization.k8s.io/prometheus

kubectl delete ns demo
kubectl delete ns monitoring
```

## Next Steps

- Learn how to configure [Druid Topology](/docs/guides/druid/clustering/overview/index.md).
- Monitor your Druid database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
