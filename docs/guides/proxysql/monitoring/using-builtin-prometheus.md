---
title: Monitor ProxySQL using Builtin Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: monitor-proxysql-using-builtin-prometheus
    name: Builtin Prometheus Discovery
    parent: proxysql-monitoring
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Monitoring ProxySQL with builtin Prometheus

This tutorial will show you how to monitor ProxySQL using builtin [Prometheus](https://github.com/prometheus/prometheus) scraper.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- If you are not familiar with how to configure Prometheus to scrape metrics from various Kubernetes resources, please read the tutorial from [here](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/concepts/database-monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use two different namespaces called,
  - `monitoring` to deploy respective monitoring resources
  - `demo` to deploy respective resources from KubeDB

  ```console
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/proxysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/proxysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Sample MySQL Group Replication

At first, let's deploy a MySQL database with Group Replication support. Below is the MySQL object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  version: "5.7.25"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      name: "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"
      baseServerID: 100
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
  updateStrategy:
    type: RollingUpdate
```

Let's create the MySQL object we have shown above.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/demo-my-group.yaml
mysql.kubedb.com/my-group created
```

Now, wait for the database to go into the `Running` state.

```console
$ kubectl get my -n demo my-group
NAME       VERSION   STATUS    AGE
my-group   5.7.25    Running   3m
```

## Deploy ProxySQL with Monitoring Enabled

Now we are going to create a sample ProxySQL object to load balance the previously created MySQL group. Keep note that monitoring is enabled in this sample ProxySQL object. See below:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: builtin-prom-proxysql
  namespace: demo
spec:
  version: "2.0.4"
  replicas: 1
  mode: GroupReplication
  backend:
    ref:
      apiGroup: "kubedb.com"
      kind: MySQL
      name: my-group
    replicas: 3
  updateStrategy:
    type: RollingUpdate
  monitor:
    agent: prometheus.io/builtin
    prometheus:
      port: 42004
```

- `.spec.monitor.agent: prometheus.io/builtin` specifies that we are going to monitor this server using builtin Prometheus scraper.

Let's create the ProxySQL object we have shown above.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/builtin-prom-proxysql.yaml
proxysql.kubedb.com/builtin-prom-proxysql created
```

Now, wait for the ProxySQL object to go into the `Running` state.

```console
$ kubectl get proxysql -n demo builtin-prom-proxysql
NAME                    VERSION   STATUS    AGE
builtin-prom-proxysql   2.0.4     Running   3m
```

KubeDB will create a separate stats service with the name `{ProxySQL object name}-stats` for monitoring purposes.

```console
$ kubectl get svc -n demo --selector="proxysql.kubedb.com/name=builtin-prom-proxysql"
NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
builtin-prom-proxysql         ClusterIP   10.101.12.24    <none>        6033/TCP    23m
builtin-prom-proxysql-stats   ClusterIP   10.97.112.192   <none>        42004/TCP   23m
```

Here, `builtin-prom-proxysql-stats` service has been created for monitoring purposes. Let's describe the service.

```console
$ kubectl describe svc -n demo builtin-prom-proxysql-stats
Name:              builtin-prom-proxysql-stats
Namespace:         demo
Labels:            kubedb.com/kind=ProxySQL
                   kubedb.com/role=stats
                   proxysql.kubedb.com/load-balance=GroupReplication
                   proxysql.kubedb.com/name=builtin-prom-proxysql
Annotations:       monitoring.appscode.com/agent: prometheus.io/builtin
                   prometheus.io/path: /metrics
                   prometheus.io/port: 42004
                   prometheus.io/scrape: true
Selector:          kubedb.com/kind=ProxySQL,proxysql.kubedb.com/load-balance=GroupReplication,proxysql.kubedb.com/name=builtin-prom-proxysql
Type:              ClusterIP
IP:                10.97.112.192
Port:              prom-http  42004/TCP
TargetPort:        prom-http/TCP
Endpoints:         10.244.1.6:42004
Session Affinity:  None
Events:            <none>
```

You can see that the service contains the following annotations.

```console
prometheus.io/path: /metrics
prometheus.io/port: 42004
prometheus.io/scrape: true
```

The Prometheus server will discover the service endpoint using these specifications and will scrape metrics from the exporter.

## Configure Prometheus Server

Now, we have to configure a Prometheus scraping job to scrape the metrics using this service. We are going to configure scraping jobs similar to this [kubernetes-service-endpoints](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin#kubernetes-service-endpoints) job that scrapes metrics from endpoints of a service.

Let's configure a Prometheus scraping job to collect metrics from this service.

```yaml
- job_name: 'kubedb-databases'
  honor_labels: true
  scheme: http
  kubernetes_sd_configs:
  - role: endpoints
  # by default Prometheus server select all kubernetes services as possible target.
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
  # service created by KubeDB will have "kubedb.com/kind" and "kubedb.com/name" annotations. keep only those services that have these annotations.
  - source_labels: [__meta_kubernetes_service_label_kubedb_com_kind]
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

If you already have a Prometheus server running, you have to add the above scraping job in the `ConfigMap` used to configure the Prometheus server. Then, you have to restart it for the updated configuration to take effect.

> If you don't use a persistent volume for Prometheus storage, you will lose your previously scraped data on restart.

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
      # by default Prometheus server select all kubernetes services as possible target.
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
      # service created by KubeDB will have "kubedb.com/kind" and "kubedb.com/name" annotations. keep only those services that have these annotations.
      - source_labels: [__meta_kubernetes_service_label_kubedb_com_kind]
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

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/monitoring/builtin-prometheus/prom-config.yaml
configmap/prometheus-config created
```

**Create RBAC:**

If you are using an RBAC enabled cluster, you have to give necessary RBAC permissions for Prometheus. Let's create necessary RBAC stuff for Prometheus,

```console
$ kubectl apply -f https://github.com/appscode/third-party-tools/raw/master/monitoring/prometheus/builtin/artifacts/rbac.yaml
clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created
```

>YAML for the RBAC resources created above can be found [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/artifacts/rbac.yaml).

**Deploy Prometheus:**

Now, we are ready to deploy the Prometheus server. We are going to use the following [deployment](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/artifacts/deployment.yaml) to deploy the Prometheus server.

Let's deploy the Prometheus server.

```console
$ kubectl apply -f https://github.com/appscode/third-party-tools/raw/master/monitoring/prometheus/builtin/artifacts/deployment.yaml
deployment.apps/prometheus created
```

### Verify Monitoring Metrics

Prometheus server is listening to port `9090`. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

At first, let's check if the Prometheus pod is in `Running` state.

```console
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                          READY   STATUS    RESTARTS   AGE
prometheus-789c9695fc-v8gjg   1/1     Running   0          27s
```

Now, run the following command on a separate terminal to forward 9090 port of `prometheus-789c9695fc-v8gjg` pod,

```console
$ kubectl port-forward -n monitoring prometheus-8568c86d86-95zhn 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see the endpoint of `builtin-prom-proxysql-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" height="100%" src="/docs/images/proxysql/proxysql-builtin-prom-target.png" style="padding:10px">
</p>

Check the labels marked with the red rectangles. These labels confirm that the metrics are coming from `ProxySQL` database `builtin-prom-proxysql` through stats service `builtin-prom-proxysql-stats`.

Now, you can view the collected metrics and create a graph from the homepage of this Prometheus dashboard. You can also use this Prometheus server as a data source for [Grafana](https://grafana.com/) and create a beautiful dashboard with collected metrics.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run following commands

```console
$ kubectl delete -n monitoring deployment.apps/prometheus

$ kubectl delete -n monitoring clusterrole.rbac.authorization.k8s.io/prometheus
$ kubectl delete -n monitoring serviceaccount/prometheus
$ kubectl delete -n monitoring clusterrolebinding.rbac.authorization.k8s.io/prometheus

$ kubectl delete -n demo proxysql/builtin-prom-proxysql
$ kubectl delete -n demo my/my-group

$ kubectl delete ns demo
$ kubectl delete ns monitoring
```

## Next Steps

- Monitor your ProxySQL database with KubeDB using [`out-of-the-box` CoreOS Prometheus Operator](/docs/guides/proxysql/monitoring/using-coreos-prometheus-operator.md).
- Use private Docker registry to deploy ProxySQL with KubeDB [here](/docs/guides/proxysql/private-registry/using-private-registry.md).
- Use custom config file to configure ProxySQL [here](/docs/guides/proxysql/configuration/using-custom-config.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
