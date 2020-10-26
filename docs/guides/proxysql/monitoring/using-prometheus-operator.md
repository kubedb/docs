---
title: Monitor ProxySQL using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: prx-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: prx-monitoring-proxysql
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Monitoring ProxySQL Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native ways to deploy and configure the Prometheus server. This tutorial will show you how to use the Prometheus operator to monitor ProxySQL deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/proxysql/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use two different namespaces called,
- `monitoring` to deploy respective monitoring resources
- `demo` to deploy respective resources from KubeDB

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md).

- If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

> Note: YAML files used in this tutorial are stored in [docs/examples/proxysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/proxysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Out Required Labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` object. We are going to provide these labels in `.spec.monitor.prometheus.labels` field of ProxySQL object so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME         AGE
monitoring   prometheus   2m56s
```

> If you don't have any Prometheus server running in your cluster, deploy one following the guide specified in **Before You Begin** section.

Now, let's view the YAML of the available Prometheus server `prometheus` in `monitoring` namespace.

```yaml
$ kubectl get prometheus -n monitoring prometheus -o yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"monitoring.coreos.com/v1","kind":"Prometheus","metadata":{"annotations":{},"labels":{"prometheus":"prometheus"},"name":"prometheus","namespace":"monitoring"},"spec":{"replicas":1,"resources":{"requests":{"memory":"400Mi"}},"serviceAccountName":"prometheus","serviceMonitorSelector":{"matchLabels":{"k8s-app":"prometheus"}}}}
  creationTimestamp: "2019-11-26T11:58:19Z"
  generation: 1
  labels:
    prometheus: prometheus
  name: prometheus
  namespace: monitoring
  resourceVersion: "25919"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/prometheuses/prometheus
  uid: 394c5008-56a3-4f0a-af12-9beb873cbc3e
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

Notice the `.spec.serviceMonitorSelector` section. Here, `k8s-app: prometheus` label is used to select `ServiceMonitor` object. So, we are going to use this label in `.spec.monitor.prometheus.labels` field of ProxySQL crd.

## Deploy ProxySQL with Monitoring Enabled

ProxySQL assumes that there already exists backend servers. So we need some backend ready. Here, we will set up ProxySQL to load balance MySQL Group Replication.

### Deploy Sample MySQL Group Replication

So, let's deploy a MySQL database with Group Replication support. Below is the MySQL object that we are going to create.

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
```

Let's create the MySQL object we have shown above.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/demo-my-group.yaml
mysql.kubedb.com/my-group created
```

Now, wait for the database to go into the `Running` state.

```bash
$ kubectl get my -n demo my-group
NAME       VERSION   STATUS    AGE
my-group   5.7.25    Running   3m
```

### Deploy ProxySQL

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
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 42004
      serviceMonitor:
        labels:
          k8s-app: prometheus
        interval: 10s
```

Here,

- `.spec.monitor.agent: prometheus.io/operator` indicates that we are going to monitor this server using the Prometheus operator.
- `.spec.monitor.prometheus.port` specifies the port at which ProxySQL exporter will serve the metrics and from this port, the Prometheus server collects them.
- `.spec.monitor.prometheus.namespace: monitoring` specifies that KubeDB should create `ServiceMonitor` in `monitoring` namespace.
- `.spec.monitor.prometheus.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.
- `.spec.monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from ProxySQL exporter with 10 seconds interval.

Let's create the ProxySQL object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/coreos-prom-proxysql.yaml
proxysql.kubedb.com/coreos-prom-mysql created
```

```bash
$ kubectl get proxysql -n demo coreos-prom-proxysql
NAME                   VERSION   STATUS    AGE
coreos-prom-proxysql   2.0.4     Running   14s
```

KubeDB will create a separate stats service with the name `{ProxySQL object name}-stats` for monitoring purposes.

```bash
$ kubectl get svc -n demo --selector="proxysql.kubedb.com/name=coreos-prom-proxysql"
NAME                         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
coreos-prom-proxysql         ClusterIP   10.101.10.235   <none>        6033/TCP    73s
coreos-prom-proxysql-stats   ClusterIP   10.111.242.54   <none>        42004/TCP   68s
```

Here, `coreos-prom-proxysql-stats` service has been created for monitoring purposes.

Let's describe this stats service.

```yaml
$ kubectl describe svc -n demo coreos-prom-proxysql-stats
Name:              coreos-prom-proxysql-stats
Namespace:         demo
Labels:            kubedb.com/kind=ProxySQL
                   kubedb.com/role=stats
                   proxysql.kubedb.com/load-balance=GroupReplication
                   proxysql.kubedb.com/name=coreos-prom-proxysql
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          kubedb.com/kind=ProxySQL,proxysql.kubedb.com/load-balance=GroupReplication,proxysql.kubedb.com/name=coreos-prom-proxysql
Type:              ClusterIP
IP:                10.111.242.54
Port:              prom-http  42004/TCP
TargetPort:        prom-http/TCP
Endpoints:         10.244.2.14:42004
Session Affinity:  None
Events:            <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use this information to target its endpoints.

KubeDB will also create a `ServiceMonitor` object in `monitoring` namespace that select the endpoints of `coreos-prom-proxysql-stats` service. Verify that the `ServiceMonitor` object has been created.

```bash
$ kubectl get servicemonitor -n monitoring
NAME                               AGE
kubedb-demo-coreos-prom-proxysql   3m22s
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `.spec.monitor` section of ProxySQL object.

```yaml
$ kubectl get servicemonitor -n monitoring kubedb-demo-coreos-prom-proxysql -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2019-11-26T12:46:21Z"
  generation: 1
  labels:
    k8s-app: prometheus
    monitoring.appscode.com/service: coreos-prom-proxysql-stats.demo
  name: kubedb-demo-coreos-prom-proxysql
  namespace: monitoring
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    kind: Service
    name: coreos-prom-proxysql-stats
    uid: a37585e6-14b3-41b2-a8d2-81763845cb8d
  resourceVersion: "30358"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/servicemonitors/kubedb-demo-coreos-prom-proxysql
  uid: 09dc865e-73c0-44e1-b972-540b58efc660
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
      kubedb.com/kind: ProxySQL
      kubedb.com/role: stats
      proxysql.kubedb.com/load-balance: GroupReplication
      proxysql.kubedb.com/name: coreos-prom-proxysql
```

Notice that the `ServiceMonitor` has `k8s-app: prometheus` label that we had specified in ProxySQL object.

Also, notice that the `ServiceMonitor` has a selector that matches the labels we have seen in the `coreos-prom-proxysql-stats` service. It also, target the `prom-http` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus Pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                      READY   STATUS    RESTARTS   AGE
prometheus-prometheus-0   3/3     Running   1          56m
```

Prometheus server is listening to port `9090` of `Prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access the Prometheus dashboard.

Run the following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `prom-http` endpoint of `coreos-prom-proxysql-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/proxysql/proxysql-coreos-prom-target.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels marked by the red rectangles. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from the homepage of this Prometheus dashboard. You can also use this Prometheus server as a data source for [Grafana](https://grafana.com/) and create a beautiful dashboard with collected metrics.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run following commands

```bash
# cleanup prometheus resources
$ kubectl delete -n monitoring prometheus prometheus
$ kubectl delete -n monitoring clusterrolebinding prometheus
$ kubectl delete -n monitoring clusterrole prometheus
$ kubectl delete -n monitoring serviceaccount prometheus
$ kubectl delete -n monitoring service prometheus-operated

# cleanup prometheus operator resources
$ kubectl delete -n monitoring deployment prometheus-operator
$ kubectl delete -n dmeo serviceaccount prometheus-operator
$ kubectl delete clusterrolebinding prometheus-operator
$ kubectl delete clusterrole prometheus-operator

# cleanup proxysql and mysql resources
$ kubectl delete -n demo proxysql/coreos-prom-proxysql
$ kubectl delete -n demo my/my-group

# delete namespace
$ kubectl delete ns monitoring
$ kubectl delete ns demo
```

## Next Steps

- Monitor your ProxySQL with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/proxysql/monitoring/using-builtin-prometheus.md).
- Use custom config file to configure ProxySQL [here](/docs/guides/proxysql/configuration/using-config-file.md).
- Detail concepts of ProxySQL CRD [here](/docs/guides/proxysql/concepts/proxysql.md).
- Detail concepts of ProxySQLVersion CRD [here](/docs/guides/proxysql/concepts/catalog.md).
- Use private Docker registry to deploy ProxySQL with KubeDB [here](/docs/guides/proxysql/private-registry/using-private-registry.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
