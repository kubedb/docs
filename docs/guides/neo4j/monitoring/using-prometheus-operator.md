---
title: Monitor Neo4j using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: neo4j-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: neo4j-monitoring
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Neo4j Using Prometheus Operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides a simple and Kubernetes-native way to deploy and configure Prometheus server. This tutorial will show you how to use the Prometheus operator to monitor a Neo4j database deployed with KubeDB.

## Before You Begin

> Prerequisites: A running Kubernetes cluster with KubeDB installed. See the [quickstart guide](/docs/guides/neo4j/quickstart/quickstart.md) if you need to set up your environment.

- To learn how Prometheus monitoring works with KubeDB in general, please visit the [monitoring overview](/docs/guides/neo4j/monitoring/overview.md).

- Prometheus resources will be deployed in the `monitoring` namespace; the database will be in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

- A running [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance is required. If you don't have one, deploy it following [these docs](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md).

- A running Prometheus server is also required. Deploy one following [this tutorial](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

> Note: YAML files used in this tutorial are stored in the [docs/examples/neo4j](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/neo4j) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out Required Labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` CR. We will provide these labels in `spec.monitor.prometheus.labels` field of the Neo4j CR so that KubeDB creates a `ServiceMonitor` object accordingly.

Let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME                                    VERSION              DESIRED   READY   RECONCILED   AVAILABLE   AGE
monitoring   prometheus-kube-prometheus-prometheus   v3.11.3-distroless   1         1       True         True        10m
```

> If you don't have any Prometheus server running in your cluster, deploy one following the guide specified in the **Before You Begin** section.

Now, let's view the YAML of the available Prometheus server in `monitoring` namespace.

```bash
$ kubectl get prometheus -n monitoring prometheus-kube-prometheus-prometheus -o yaml
```

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus-kube-prometheus-prometheus
  namespace: monitoring
  labels:
    app: kube-prometheus-stack-prometheus
    release: prometheus
spec:
  replicas: 1
  serviceMonitorSelector:
    matchLabels:
      release: prometheus
  serviceMonitorNamespaceSelector: {}
  ...
```

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` CRs. So, we are going to use this label in `spec.monitor.prometheus.labels` field of the Neo4j CR.

## Deploy Neo4j with Monitoring Enabled

Below is the Neo4j object that we are going to create with monitoring enabled.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: coreos-prom-neo4j
  namespace: demo
spec:
  replicas: 3
  deletionPolicy: WipeOut
  version: "2025.11.2"
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here,

- `monitor.agent: prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
- `monitor.prometheus.serviceMonitor.labels` specifies that KubeDB should create a `ServiceMonitor` with these labels. We use `release: prometheus` to match the `serviceMonitorSelector` configured in the Prometheus CR above.
- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database at a 10-second interval.

Let's create the Neo4j object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/monitoring/coreos-prom-neo4j.yaml
neo4j.kubedb.com/coreos-prom-neo4j created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get neo4j -n demo coreos-prom-neo4j
NAME                VERSION      STATUS   AGE
coreos-prom-neo4j   2025.11.2    Ready    3m
```

KubeDB will create a separate stats service with the name `{Neo4j CR name}-stats` for monitoring purposes.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=coreos-prom-neo4j"
NAME                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                                                 AGE
coreos-prom-neo4j         ClusterIP   10.43.124.250   <none>        6362/TCP,7687/TCP,7474/TCP                              3m55s
coreos-prom-neo4j-0       ClusterIP   None            <none>        6362/TCP,7687/TCP,7474/TCP,7688/TCP,7000/TCP,6000/TCP   3m55s
coreos-prom-neo4j-1       ClusterIP   None            <none>        6362/TCP,7687/TCP,7474/TCP,7688/TCP,7000/TCP,6000/TCP   3m55s
coreos-prom-neo4j-2       ClusterIP   None            <none>        6362/TCP,7687/TCP,7474/TCP,7688/TCP,7000/TCP,6000/TCP   3m55s
coreos-prom-neo4j-stats   ClusterIP   10.43.214.74    <none>        2004/TCP                                                3m55s
```

Here, `coreos-prom-neo4j-stats` service has been created for monitoring purposes. It exposes metrics on port `2004`.

## Verify ServiceMonitor Creation

KubeDB will also create a `ServiceMonitor` CR in the `demo` namespace that selects the endpoints of `coreos-prom-neo4j-stats` service. Verify that the `ServiceMonitor` has been created.

```bash
$ kubectl get servicemonitor -n demo
NAME                      AGE
coreos-prom-neo4j-stats   6m8s
```

Let's verify the `ServiceMonitor` YAML.

```bash
$ kubectl get servicemonitor -n demo coreos-prom-neo4j-stats -o yaml
```

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: coreos-prom-neo4j
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: neo4js.kubedb.com
    release: prometheus
  name: coreos-prom-neo4j-stats
  namespace: demo
spec:
  endpoints:
  - honorLabels: true
    interval: 10s
    path: /metrics
    port: metrics
    relabelings:
    - action: replace
      sourceLabels:
      - __meta_kubernetes_endpoint_address_target_name
      targetLabel: pod
    scheme: http
  namespaceSelector:
    matchNames:
    - demo
  selector:
    matchLabels:
      app.kubernetes.io/component: database
      app.kubernetes.io/instance: coreos-prom-neo4j
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: neo4js.kubedb.com
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` carries the `release: prometheus` label that we specified in the Neo4j CR — this is exactly what the Prometheus server's `serviceMonitorSelector` looks for.

The `ServiceMonitor` selects the `coreos-prom-neo4j-stats` service by matching its labels and scrapes the `metrics` port (`2004`) every 10 seconds.

## Verify Monitoring Metrics

Let's find out the Prometheus pod for our Prometheus server.

```bash
$ kubectl get pod -n monitoring -l app.kubernetes.io/name=prometheus
NAME                                                     READY   STATUS    RESTARTS   AGE
prometheus-prometheus-kube-prometheus-prometheus-0       2/2     Running   0          15m
```

The Prometheus server is listening on port `9090`. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access the Prometheus dashboard.

Run the following command in a separate terminal to forward port 9090:

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-kube-prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, open [http://localhost:9090](http://localhost:9090) in your browser. Navigate to **Status → Targets** and you should see the `coreos-prom-neo4j-stats` endpoint listed as an active scrape target.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/neo4j/prometheus.png" style="padding:10px">
</p>

The `endpoint` and `service` labels confirm the target is our Neo4j database. You can now browse collected metrics from the Prometheus homepage and create graphs, or use this Prometheus server as a data source for [Grafana](https://grafana.com/) to build dashboards.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
# delete the Neo4j database
kubectl patch -n demo neo4j/coreos-prom-neo4j -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo neo4j/coreos-prom-neo4j

# delete namespaces
kubectl delete ns demo
kubectl delete ns monitoring
```

## Next Steps

- Monitor your Neo4j database with KubeDB using [built-in Prometheus](/docs/guides/neo4j/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
