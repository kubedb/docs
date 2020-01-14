---
title: Monitor Percona XtraDB using Coreos Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: px-monitoring-using-coreos-prometheus-operator
    name: Coreos Prometheus Operator
    parent: px-monitoring
    weight: 15
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Monitoring PerconaXtraDB Using CoreOS Prometheus Operator

CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use CoreOS Prometheus operator to monitor PerconaXtraDB deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/concepts/database-monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy database in `demo` namespace.

  ```console
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

- We need a CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/coreos-operator/README.md).

- If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/coreos-operator/README.md#deploy-prometheus-server).

> Note: YAML files used in this tutorial are stored in [docs/examples/percona-xtradb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/percona-xtradb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` resource. We are going to provide these labels in `.spec.monitor.prometheus.labels` field of `PerconaXtraDB` object so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```console
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME         AGE
monitoring   prometheus   18m
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
  creationTimestamp: 2019-01-03T13:41:51Z
  generation: 1
  labels:
    prometheus: prometheus
  name: prometheus
  namespace: monitoring
  resourceVersion: "44402"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/prometheuses/prometheus
  uid: 5324ad98-0f5d-11e9-b230-080027f306f3
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

Notice the `.spec.serviceMonitorSelector` section. Here, `k8s-app: prometheus` label is used to select `ServiceMonitor` resource. So, we are going to use this label in `.spec.monitor.prometheus.labels` field of `PerconaXtraDB` resource.

## Deploy PerconaXtraDB with Monitoring Enabled

At first, let's deploy a sample PerconaXtraDB with monitoring enabled. Below is the `PerconaXtraDB` object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: PerconaXtraDB
metadata:
  name: px-coreos-prom
  namespace: demo
spec:
  version: "5.7-cluster"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  updateStrategy:
    type: "RollingUpdate"
  terminationPolicy: WipeOut
  monitor:
    agent: prometheus.io/coreos-operator
    prometheus:
      namespace: monitoring
      labels:
        k8s-app: prometheus
      interval: 10s
```

Here,

- `.spec.monitor.agent:  prometheus.io/coreos-operator` indicates that we are going to monitor this server using CoreOS prometheus operator.
- `.spec.monitor.prometheus.namespace: monitoring` specifies that KubeDB should create `ServiceMonitor` in `monitoring` namespace.

- `.spec.monitor.prometheus.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.

- `.spec.monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the `PerconaXtraDB` object that we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/px-coreos-prom.yaml
perconaxtradb.kubedb.com/px-coreos-prom created
```

Now, wait for the database to go into `Running` state.

```console
$ kubectl get px -n demo px-coreos-prom
NAME             VERSION       STATUS    AGE
px-coreos-prom   5.7-cluster   Running   5m4s
```

KubeDB will create a separate stats service with name `{PerconaXtraDB_obj_name}-stats` for monitoring purpose.

```console
$ kubectl get svc -n demo --selector="kubedb.com/name=px-coreos-prom"
NAME                   TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
px-coreos-prom         ClusterIP   10.107.58.214    <none>        3306/TCP    5m25s
px-coreos-prom-gvr     ClusterIP   None             <none>        3306/TCP    5m25s
px-coreos-prom-stats   ClusterIP   10.106.130.209   <none>        56790/TCP   48s
```

Here, `px-coreos-prom-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```yaml
$ kubectl describe svc -n demo px-coreos-prom-stats
Name:              px-coreos-prom-stats
Namespace:         demo
Labels:            kubedb.com/kind=PerconaXtraDB
                   kubedb.com/name=px-coreos-prom
                   kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/coreos-operator
Selector:          kubedb.com/kind=PerconaXtraDB,kubedb.com/name=px-coreos-prom
Type:              ClusterIP
IP:                10.106.130.209
Port:              prom-http  56790/TCP
TargetPort:        prom-http/TCP
Endpoints:         10.244.1.7:56790,10.244.2.11:56790,10.244.2.9:56790
Session Affinity:  None
Events:            <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use these information to target its endpoints.

KubeDB will also create a `ServiceMonitor` resource in `monitoring` namespace that select the endpoints of `px-coreos-prom-stats` service. Verify that the `ServiceMonitor` resource has been created.

```console
$ kubectl get servicemonitor -n monitoring
NAME                         AGE
kubedb-demo-px-coreos-prom   2m45s
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `.spec.monitor` section of `PerconaXtraDB` object.

```yaml
$ kubectl get servicemonitor -n monitoring kubedb-demo-px-coreos-prom -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2019-12-19T07:17:11Z"
  generation: 1
  labels:
    k8s-app: prometheus
    monitoring.appscode.com/service: px-coreos-prom-stats.demo
  name: kubedb-demo-px-coreos-prom
  namespace: monitoring
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    kind: Service
    name: px-coreos-prom-stats
    uid: fb353fac-9945-48ca-ad81-f79dcf5e8e24
  resourceVersion: "13989"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/servicemonitors/kubedb-demo-px-coreos-prom
  uid: 07fd2ae4-d294-4e7f-88d8-960ca3a63a40
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
      kubedb.com/kind: PerconaXtraDB
      kubedb.com/name: px-coreos-prom
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `k8s-app: prometheus` that we had specified in `PerconaXtraDB` object.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `px-coreos-prom-stats` service. It also, target the `prom-http` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus Pod for `prometheus` Prometheus server.

```console
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                      READY   STATUS    RESTARTS   AGE
prometheus-prometheus-0   3/3     Running   1          51m
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-0` Pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run the following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-0` Pod,

```console
$ kubectl port-forward -n monitoring prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `prom-http` endpoint of `px-coreos-prom-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/percona-xtradb/coreos-prom-targets.png" style="padding:10px">
</p>

Check the followings:

- Endpoint URLs of backend servers marked by green rectangles
- `endpoint`, `pod`, and `service` labels marked by red rectangles

They all indicate that the target is our expected database.

Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboard with collected metrics.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run following commands

```console
# cleanup database
$ kubectl delete -n demo px/px-coreos-prom

# cleanup prometheus resources
$ kubectl delete -n monitoring prometheus prometheus
$ kubectl delete clusterrolebinding prometheus
$ kubectl delete clusterrole prometheus
$ kubectl delete -n monitoring serviceaccount prometheus
$ kubectl delete -n monitoring service prometheus-operated

# cleanup prometheus operator resources
$ kubectl delete -n monitoring deployment prometheus-operator
$ kubectl delete clusterrolebinding prometheus-operator
$ kubectl delete clusterrole prometheus-operator
$ kubectl delete -n monitoring serviceaccount prometheus-operator

# delete namespace
$ kubectl delete ns monitoring
$ kubectl delete ns demo
```

## Next Steps

- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/percona-xtradb/monitoring/using-builtin-prometheus.md).
- Initialize [PerconaXtraDB with Script](/docs/guides/percona-xtradb/initialization/using-script.md).
- Use [private Docker registry](/docs/guides/percona-xtradb/private-registry/using-private-registry.md) to deploy PerconaXtraDB with KubeDB.
- How to use [custom configuration](/docs/guides/percona-xtradb/configuration/using-custom-config.md).
- How to use [custom rbac resource](/docs/guides/percona-xtradb/custom-rbac/using-custom-rbac.md) for PerconaXtraDB.
- Use Stash to [Backup PerconaXtraDB](/docs/guides/percona-xtradb/snapshot/stash.md).
- Detail concepts of [PerconaXtraDB object](/docs/concepts/databases/percona-xtradb.md).
- Detail concepts of [PerconaXtraDBVersion object](/docs/concepts/catalog/percona-xtradb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
