---
title: Monitor MySQL using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: my-using-coreos-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: my-monitoring-mysql
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Monitoring MySQL Using CoreOS Prometheus Operator

CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use CoreOS Prometheus operator to monitor MySQL database deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/concepts/database-monitoring/overview.md).

- To keep database resources isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

- We need a CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/coreos-operator/README.md).

- If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/coreos-operator/README.md#deploy-prometheus-server).

> Note: YAML files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.labels` field of MySQL crd so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE   NAME         VERSION   REPLICAS   AGE
default     prometheus             1          2m19s
```

> If you don't have any Prometheus server running in your cluster, deploy one following the guide specified in **Before You Begin** section.

Now, let's view the YAML of the available Prometheus server `prometheus` in `default` namespace.

```yaml
$ kubectl get prometheus -n default prometheus -o yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"monitoring.coreos.com/v1","kind":"Prometheus","metadata":{"annotations":{},"labels":{"prometheus":"prometheus"},"name":"prometheus","namespace":"default"},"spec":{"replicas":1,"resources":{"requests":{"memory":"400Mi"}},"serviceAccountName":"prometheus","serviceMonitorNamespaceSelector":{"matchLabels":{"prometheus":"prometheus"}},"serviceMonitorSelector":{"matchLabels":{"k8s-app":"prometheus"}}}}
  creationTimestamp: "2020-08-25T04:02:07Z"
  generation: 1
  labels:
    prometheus: prometheus
  ...
    manager: kubectl
    operation: Update
    time: "2020-08-25T04:02:07Z"
  name: prometheus
  namespace: default
  resourceVersion: "2087"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/default/prometheuses/prometheus
  uid: 972a50cb-b751-418b-b2bc-e0ecc9232730
spec:
  replicas: 1
  resources:
    requests:
      memory: 400Mi
  serviceAccountName: prometheus
  serviceMonitorNamespaceSelector:
    matchLabels:
      prometheus: prometheus
  serviceMonitorSelector:
    matchLabels:
      k8s-app: prometheus
```

- `spec.serviceMonitorSelector` field specifies which ServiceMonitors should be included. The Above label `k8s-app: prometheus` is used to select `ServiceMonitors` by its selector. So, we are going to use this label in `spec.monitor.prometheus.labels` field of MySQL crd.
- `spec.serviceMonitorNamespaceSelector` field specifies that the `ServiceMonitors` can be selected outside the Prometheus namespace by Prometheus using namespace selector. The Above label `prometheus: prometheus` is used to select the namespace where the `ServiceMonitor` is created.

### Add Label to database namespace

KubeDB creates a `ServiceMonitor` in database namespace `demo`. We need to add label to `demo` namespace. Prometheus will select this namespace by using its `spec.serviceMonitorNamespaceSelector` field.

Let's add label `prometheus: prometheus` to `demo` namespace,

```bash
$ kubectl patch namespace demo -p '{"metadata":{"labels": {"prometheus":"prometheus"}}}'
namespace/demo patched
```

## Deploy MySQL with Monitoring Enabled

At first, let's deploy an MySQL database with monitoring enabled. Below is the MySQL object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: coreos-prom-mysql
  namespace: demo
spec:
  version: "8.0.21"
  terminationPolicy: WipeOut
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/coreos-operator
    prometheus:
      labels:
        k8s-app: prometheus
      interval: 10s
```

Here,

- `monitor.agent:  prometheus.io/coreos-operator` indicates that we are going to monitor this server using CoreOS prometheus operator.

- `monitor.prometheus.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.

- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the MySQL object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/monitoring/coreos-prom-mysql.yaml
mysql.kubedb.com/coreos-prom-mysql created
```

Now, wait for the database to go into `Running` state.

```bash
$ watch -n 3 kubectl get mysql -n demo coreos-prom-mysql
Every 3.0s: kubectl get mysql -n demo coreos-prom-mysql         suaas-appscode: Tue Aug 25 11:53:34 2020

NAME                VERSION   STATUS    AGE
coreos-prom-mysql   8.0.21    Running   2m53s
```

KubeDB will create a separate stats service with name `{MySQL crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="kubedb.com/name=coreos-prom-mysql"
NAME                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
coreos-prom-mysql         ClusterIP   10.103.228.135   <none>        3306/TCP    3m36s
coreos-prom-mysql-gvr     ClusterIP   None             <none>        3306/TCP    3m36s
coreos-prom-mysql-stats   ClusterIP   10.106.236.14    <none>        56790/TCP   50s
```

Here, `coreos-prom-mysql-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```yaml
$ kubectl describe svc -n demo coreos-prom-mysql-stats
Name:              coreos-prom-mysql-stats
Namespace:         demo
Labels:            kubedb.com/kind=MySQL
                   kubedb.com/name=coreos-prom-mysql
                   kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/coreos-operator
Selector:          kubedb.com/kind=MySQL,kubedb.com/name=coreos-prom-mysql
Type:              ClusterIP
IP:                10.106.236.14
Port:              prom-http  56790/TCP
TargetPort:        prom-http/TCP
Endpoints:         10.244.2.6:56790
Session Affinity:  None
Events:            <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use these information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `demo` namespace that select the endpoints of `coreos-prom-mysql-stats` service. Verify that the `ServiceMonitor` crd has been created.

```console
$ kubectl get servicemonitor -n demo
NAME                            AGE
kubedb-demo-coreos-prom-mysql   3m16s
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of MySQL crd.

```yaml
$ kubectl get servicemonitor -n demo kubedb-demo-coreos-prom-mysql -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2020-08-25T05:53:27Z"
  generation: 1
  labels:
    k8s-app: prometheus
    operation: Update
    time: "2020-08-25T05:53:27Z"
  ...
  name: kubedb-demo-coreos-prom-mysql
  namespace: demo
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    controller: true
    kind: Service
    name: coreos-prom-mysql-stats
    uid: cf4ce3ec-a78e-4828-9fee-941c77eb965e
  resourceVersion: "28659"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/demo/servicemonitors/kubedb-demo-coreos-prom-mysql
  uid: 9cec794a-dfee-49dc-a809-6c9d6faac1df
spec:
  endpoints:
  - bearerTokenSecret:
      key: ""
    honorLabels: true
    interval: 10s
    path: /metrics
    port: prom-http
  namespaceSelector:
    matchNames:
    - demo
  selector:
    matchLabels:
      kubedb.com/kind: MySQL
      kubedb.com/name: coreos-prom-mysql
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `k8s-app: prometheus` that we had specified in MySQL crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `coreos-prom-mysql-stats` service. It also, target the `prom-http` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```console
$ kubectl get pod -n default -l=app=prometheus
NAME                      READY   STATUS    RESTARTS   AGE
prometheus-prometheus-0   3/3     Running   1          121m
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-0` pod,

```console
$ kubectl port-forward -n default prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `prom-http` endpoint of `coreos-prom-mysql-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/mysql/monitoring/mysql-coreos-prom-target.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels marked by red rectangle. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboard with collected metrics.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run following commands

```console
# cleanup database
kubectl delete -n demo my/coreos-prom-mysql

# cleanup Prometheus resources if exist
kubectl delete -f https://raw.githubusercontent.com/appscode/third-party-tools/master/monitoring/prometheus/coreos-operator/artifacts/prometheus.yaml
kubectl delete -f https://raw.githubusercontent.com/appscode/third-party-tools/master/monitoring/prometheus/coreos-operator/artifacts/prometheus-rbac.yaml

# cleanup Prometheus operator resources if exist
kubectl delete -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.41/bundle.yaml

# delete namespace
kubectl delete ns demo
```

## Next Steps

- Monitor your MySQL database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/using-builtin-prometheus.md).
- Detail concepts of [MySQL object](/docs/concepts/databases/mysql.md).
- Detail concepts of [MySQLVersion object](/docs/concepts/catalog/mysql.md).
- Initialize [MySQL with Script](/docs/guides/mysql/initialization/using-script.md).
- Use [private Docker registry](/docs/guides/mysql/private-registry/using-private-registry.md) to deploy MySQL with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
