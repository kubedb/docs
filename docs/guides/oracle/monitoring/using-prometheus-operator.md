---
title: Monitor Oracle using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-monitoring-prometheus-operator
    name: Prometheus Operator
    parent: guides-oracle-monitoring
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Oracle Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to monitor a KubeDB managed Oracle database using Prometheus operator.

KubeDB collects Oracle metrics using the free, public **Oracle AI Database Metrics Exporter**, which gathers standard Oracle database metrics (and supports custom metrics collection). The metrics can then be visualized in Grafana through flexible dashboards, enabling users to monitor database health and performance.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/monitoring/overview.md).

- You have to have a Prometheus operator installed in your cluster. A quick way is the `prometheus-community/kube-prometheus-stack` Helm chart. The Prometheus instance must be configured to discover `ServiceMonitor`s in the database namespace (for example with `serviceMonitorSelectorNilUsesHelmValues=false`).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy the Prometheus operator. We are going to deploy the database in the `demo` namespace.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl create ns monitoring
namespace/monitoring created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`.

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` CR. We are going to provide these labels in `spec.monitor.prometheus.serviceMonitor.labels` field of the Oracle CR so that KubeDB creates a `ServiceMonitor` object that the Prometheus server will pick up.

Let's find out the available Prometheus servers and the labels they use to select ServiceMonitors,

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME                                    VERSION              DESIRED   READY   RECONCILED   AVAILABLE   AGE
monitoring   prometheus-operator-kube-p-prometheus   v3.12.0-distroless   1         1       True         True        18m
```

Inspect the Prometheus CR to see its `serviceMonitorSelector`,

```bash
$ kubectl get prometheus -n monitoring prometheus-operator-kube-p-prometheus -o jsonpath='{.spec.serviceMonitorSelector}'
{}
```

In this tutorial the Prometheus server has an empty `serviceMonitorSelector` (`{}`), which means it discovers **all** `ServiceMonitor`s across the namespaces it watches. If, instead, your Prometheus selects ServiceMonitors by a specific label (e.g. `release: prometheus`), inspect `spec.serviceMonitorSelector.matchLabels` and use that label in the Oracle CR below. We set `release: prometheus` on our ServiceMonitor as an example.

## Deploy Oracle with Monitoring Enabled

Below is the YAML of the `Oracle` CR with monitoring enabled through the Prometheus operator agent,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: standalone-monitoring
  namespace: demo
spec:
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9161
        resources:
          limits:
            memory: 512Mi
          requests:
            cpu: 200m
            memory: 256Mi
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
```

Here,

- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to use the Prometheus operator to monitor the database.
- `spec.monitor.prometheus.exporter.port: 9161` is the port the Oracle metrics exporter listens on.
- `spec.monitor.prometheus.serviceMonitor.labels` are the labels added to the `ServiceMonitor` so that the Prometheus server selects it.
- `spec.monitor.prometheus.serviceMonitor.interval` is the scrape interval.

Let's create the `Oracle` CR,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/monitoring/standalone-monitoring.yaml
oracle.kubedb.com/standalone-monitoring created
```

Wait until the database is `Ready` and the pod prints the `DATABASE IS READY TO USE!!!` banner.

## Verify Monitoring Metrics

KubeDB creates a stats `Service` (named `<db-name>-stats`) for the exporter and a `ServiceMonitor` so the Prometheus operator can scrape it.

Let's check the stats service and the ServiceMonitor,

```bash
$ kubectl get service -n demo -l app.kubernetes.io/instance=standalone-monitoring
NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
standalone-monitoring         ClusterIP   10.43.50.21     <none>        1521/TCP   12m
standalone-monitoring-pods    ClusterIP   None            <none>        1521/TCP   12m
standalone-monitoring-stats   ClusterIP   10.43.224.116   <none>        9161/TCP   12m

$ kubectl get servicemonitor -n demo
NAME                          AGE
standalone-monitoring-stats   12m
```

The `standalone-monitoring-stats` service exposes the exporter on port `9161`, and the `standalone-monitoring-stats` ServiceMonitor tells Prometheus how to scrape it. Let's look at the ServiceMonitor spec,

```yaml
$ kubectl get servicemonitor -n demo standalone-monitoring-stats -o yaml
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
      app.kubernetes.io/component: database
      app.kubernetes.io/instance: standalone-monitoring
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: oracles.kubedb.com
      kubedb.com/role: stats
```

Let's verify the exporter is actually serving Oracle metrics by scraping the `/metrics` endpoint of the stats service from inside the cluster,

```bash
$ kubectl run mon-curl -n demo --image=curlimages/curl:8.10.1 --restart=Never --command -- sleep 90
$ kubectl exec -n demo mon-curl -- curl -s http://standalone-monitoring-stats.demo.svc:9161/metrics | grep '^oracledb' | head

# HELP oracledb_activity_execute_count Generic counter metric from gv$sysstat view in Oracle.
oracledb_activity_execute_count{database="default",inst_id="1"} 31038
# HELP oracledb_activity_parse_count_total Generic counter metric from gv$sysstat view in Oracle.
oracledb_activity_parse_count_total{database="default",inst_id="1"} 11874
oracledb_activity_user_commits{database="default",inst_id="1"} 35
oracledb_activity_user_rollbacks{database="default",inst_id="1"} 2
oracledb_ag_cluster_size_cluster_size{database="default"} 1
oracledb_batch_requests_batch_requests_total{database="default"} 31032
oracledb_buffer_cache_hit_ratio_cache_hit_ratio{database="default"} 0.9372
```

These `oracledb_*` metrics are produced by the Oracle AI Database Metrics Exporter that KubeDB runs alongside the database.

Once the ServiceMonitor is discovered, the Oracle database will appear as a target in your Prometheus server (Status → Targets) and the metrics can be visualized in Grafana.

## Monitoring a DataGuard cluster

Monitoring is enabled the same way for a DataGuard cluster — set `mode: DataGuard`, `replicas: 3`, and the same `spec.monitor` block. KubeDB runs a metrics exporter alongside each database pod and creates the stats service and ServiceMonitor:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-dg-sample
  namespace: demo
spec:
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: DataGuard
  storageType: Durable
  replicas: 3
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9161
        resources:
          limits:
            memory: 512Mi
          requests:
            cpu: 200m
            memory: 256Mi
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo oracle/standalone-monitoring -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete oracle -n demo standalone-monitoring
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Configure [TLS/SSL encryption](/docs/guides/oracle/tls/overview/index.md) for your Oracle database.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
