---
title: Monitor PgBouncer using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: pb-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: pb-monitoring-pgbouncer
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Monitoring PgBouncer using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor PgBouncer deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/pgbouncer/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created
  ```

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md).

- If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

> Note: YAML files used in this tutorial are stored in [docs/examples/pgbouncer](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.labels` field of PgBouncer crd so that KubeDB creates `ServiceMonitor` object accordingly.

As a prerequisite, we need to have Prometheus operator running, and a prometheus server created to monitor PgBouncer exporter. In this tutorial we are going to use a prometheus server named `promethus` in `monitoring` namespace. You can use the following to install `Prometheus operator`.

```bash
$ kubectl apply -f https://raw.githubusercontent.com/appscode/third-party-tools/master/monitoring/prometheus/coreos-operator/artifacts/operator.yaml
```

Now, get a prometheus server up and running.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/monitoring/coreos-prom-server.yaml

clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created
prometheus.monitoring.coreos.com/prometheus created
```

Now, let's find out the available Prometheus server in our cluster.

```bash

$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME                                    AGE
default      tufted-rodent-prometheus-o-prometheus   3h42m
monitoring   prometheus                              18m
```

Now, let's view the YAML of the available Prometheus server `prometheus` in `monitoring` namespace.

```yaml
$ kubectl get prometheus -n monitoring prometheus -o yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"monitoring.coreos.com/v1","kind":"Prometheus","metadata":{"annotations":{},"labels":{"prometheus":"prometheus"},"name":"prometheus","namespace":"monitoring"},"spec":{"replicas":1,"resources":{"requests":{"memory":"400Mi"}},"serviceAccountName":"prometheus","serviceMonitorSelector":{"matchLabels":{"k8s-app":"prometheus"}}}}
  creationTimestamp: "2019-09-19T09:32:12Z"
  generation: 1
  labels:
    prometheus: prometheus
  name: prometheus
  namespace: monitoring
  resourceVersion: "38348"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/prometheuses/prometheus
  uid: f9285974-3349-40e8-815a-8f50c3a8a4f5
spec:
  replicas: 1
  resources:
    requests:
      memory: 400Mi
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      k8s-app: prometheus
```

Notice the `spec.serviceMonitorSelector` section. Here, `k8s-app: prometheus` label is used to select `ServiceMonitor` crd. So, we are going to use this label in `spec.monitor.prometheus.labels` field of PgBouncer crd.

## Deploy PgBouncer with Monitoring Enabled

We will need a PgBouncer with monitoring enabled. This PgBouncer needs to be connected to PostgreSQL database(s). You can get a PgBouncer setup with active connection(s) to PostgreSQL by following the [quickstart](/docs/guides/pgbouncer/quickstart/quickstart.md) guide. PgBouncer object in that guide didn't come with monitoring. So we are going to enable monitoring in it. Below is the PgBouncer object that contains Prometheus operator based monitoring:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PgBouncer
metadata:
  name: pgbouncer-server
  namespace: demo
spec:
  version: "1.11.0"
  replicas: 1
  databases:
  - alias: "postgres"
    databaseName: "postgres"
    databaseRef:
      name: "quick-postgres"
  connectionPool:
    maxClientConnections: 20
    reservePoolSize: 5
    adminUsers:
    - admin
    - admin1
  userListSecretRef:
    name: db-user-pass
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          k8s-app: prometheus
        interval: 10s
```

Here,

- `monitor.agent:  prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
- `monitor.prometheus.namespace: monitoring` specifies that KubeDB should create `ServiceMonitor` in `monitoring` namespace.

- `monitor.prometheus.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.

- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the PgBouncer object that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/monitoring/coreos-prom-pgbouncer.yaml
pgbouncer.kubedb.com/pgbouncer-server configured
```

Now, wait for the database to go into `Running` state.

```bash
$ kubectl get pb -n demo pgbouncer-server
NAME               VERSION   STATUS    AGE
pgbouncer-server   1.11.0    Running   10s
```

KubeDB will create a separate stats service with name `{PgBouncer crd name}-stats` for monitoring purpose.

```bash
$  kubectl get svc -n demo --selector="kubedb.com/name=pgbouncer-server"
NAME                     TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
pgbouncer-server         ClusterIP   10.104.83.201    <none>        5432/TCP    52s
pgbouncer-server-stats   ClusterIP   10.101.214.117   <none>        56790/TCP   50s
```

Here, `pgbouncer-server-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```yaml
$ kubectl describe svc -n demo pgbouncer-server-stats
Name:              pgbouncer-server-stats
Namespace:         demo
Labels:            kubedb.com/kind=PgBouncer
                   kubedb.com/name=pgbouncer-server
                   kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent:prometheus.io/operator
Selector:          kubedb.com/kind=PgBouncer,kubedb.com/name=pgbouncer-server
Type:              ClusterIP
IP:                10.101.214.117
Port:              prom-http  56790/TCP
TargetPort:        prom-http/TCP
Endpoints:         172.17.0.7:56790
Session Affinity:  None
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use these information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `monitoring` namespace that select the endpoints of `pgbouncer-server-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$ kubectl get servicemonitor -n monitoring
NAME                           AGE
kubedb-demo-pgbouncer-server   3m4s
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of PgBouncer crd.

```yaml
$ kubectl get servicemonitor -n monitoring kubedb-demo-pgbouncer-server -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2019-09-19T10:03:24Z"
  generation: 1
  labels:
    k8s-app: prometheus
    monitoring.appscode.com/service: pgbouncer-server-stats.demo
  name: kubedb-demo-pgbouncer-server
  namespace: monitoring
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    kind: Service
    name: pgbouncer-server-stats
    uid: 749bc2ed-e14c-4a9e-9688-9d319af2b902
  resourceVersion: "41639"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/servicemonitors/kubedb-demo-pgbouncer-server
  uid: 4a68d942-a003-4b47-a8cb-f20e526e9748
spec:
  endpoints:
  - honorLabels: true
    interval: 5s
    path: /metrics
    port: prom-http
  namespaceSelector:
    matchNames:
    - demo
  selector:
    matchLabels:
      kubedb.com/kind: PgBouncer
      kubedb.com/name: pgbouncer-server
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `k8s-app: prometheus` that we had specified in PgBouncer crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `pgbouncer-server-stats` service. It also, target the `prom-http` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                      READY   STATUS    RESTARTS   AGE
prometheus-prometheus-0   3/3     Running   1          35m
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. You should see `prom-http` endpoint of `pgbouncer-server-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/pgbouncer/monitoring/pb-coreos-prom-target.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels which verify that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboard with collected metrics.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run the following commands

```bash
# cleanup prometheus resources
kubectl delete -n monitoring prometheus prometheus
kubectl delete -n monitoring clusterrolebinding prometheus
kubectl delete -n monitoring clusterrole prometheus
kubectl delete -n monitoring serviceaccount prometheus
kubectl delete -n monitoring service prometheus-operated

# delete namespace
kubectl delete ns monitoring
```

## Next Steps

- Monitor your PgBouncer with KubeDB using [built-in Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
