---
title: Monitor SingleStore using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-monitoring-prometheus-operator
    name: Prometheus Operator
    parent: guides-sdb-monitoring
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring SingleStore Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor SingleStore database deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/singlestore/monitoring/overview/index.md).

- To keep database resources isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md).

- If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

> Note: YAML files used in this tutorial are stored in [docs/guides/singlestore/monitoring/prometheus-operator/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/singlestore/monitoring/prometheus-operator/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.labels` field of SingleStore crd so that KubeDB creates `ServiceMonitor` object accordingly.

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
      {"apiVersion":"monitoring.coreos.com/v1","kind":"Prometheus","metadata":{"annotations":{},"labels":{"prometheus":"prometheus"},"name":"prometheus","namespace":"default"},"spec":{"replicas":1,"resources":{"requests":{"memory":"400Mi"}},"serviceAccountName":"prometheus","serviceMonitorNamespaceSelector":{"matchLabels":{"prometheus":"prometheus"}},"serviceMonitorSelector":{"matchLabels":{"release":"prometheus"}}}}
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
      release: prometheus
```

- `spec.serviceMonitorSelector` field specifies which ServiceMonitors should be included. The Above label `release: prometheus` is used to select `ServiceMonitors` by its selector. So, we are going to use this label in `spec.monitor.prometheus.labels` field of SingleStore crd.
- `spec.serviceMonitorNamespaceSelector` field specifies that the `ServiceMonitors` can be selected outside the Prometheus namespace by Prometheus using namespace selector. The Above label `prometheus: prometheus` is used to select the namespace where the `ServiceMonitor` is created.

### Add Label to database namespace

KubeDB creates a `ServiceMonitor` in database namespace `demo`. We need to add label to `demo` namespace. Prometheus will select this namespace by using its `spec.serviceMonitorNamespaceSelector` field.

Let's add label `prometheus: prometheus` to `demo` namespace,

```bash
$ kubectl patch namespace demo -p '{"metadata":{"labels": {"prometheus":"prometheus"}}}'
namespace/demo patched
```

## Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

## Deploy SingleStore with Monitoring Enabled

At first, let's deploy an SingleStore database with monitoring enabled. Below is the SingleStore object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: prom-operator-sdb
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 2
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                      
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here,

- `monitor.agent:  prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.

- `monitor.prometheus.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.

- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the SingleStore object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/monitoring/prometheus-operator/yamls/prom-operator-singlestore.yaml
singlestore.kubedb.com/prom-operator-sdb created
```

Now, wait for the database to go into `Running` state.

```bash
$ watch -n 3 kubectl get singlestore -n demo prom-operator-sdb

NAME                TYPE                  VERSION   STATUS   AGE
prom-operator-sdb   kubedb.com/v1alpha2   8.7.10    Ready    10m

```

KubeDB will create a separate stats service with name `{SingleStore crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=prom-operator-sdb"
NAME                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE
prom-operator-sdb         ClusterIP   10.128.249.124   <none>        3306/TCP,8081/TCP   12m
prom-operator-sdb-pods    ClusterIP   None             <none>        3306/TCP            12m
prom-operator-sdb-stats   ClusterIP   10.128.25.236    <none>        9104/TCP            12m

```

Here, `prom-operator-sdb-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```yaml
$ kubectl describe svc -n demo prom-operator-sdb-stats
Name:              prom-operator-sdb-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
                   app.kubernetes.io/instance=prom-operator-sdb
                   app.kubernetes.io/managed-by=kubedb.com
                   app.kubernetes.io/name=singlestores.kubedb.com
                   kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/instance=prom-operator-sdb,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=singlestores.kubedb.com
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                10.128.25.236
IPs:               10.128.25.236
Port:              metrics  9104/TCP
TargetPort:        metrics/TCP
Endpoints:         10.2.1.140:9104,10.2.1.141:9104
Session Affinity:  None
Events:            <none>

```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use these information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `demo` namespace that select the endpoints of `prom-operator-sdb-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$ kubectl get servicemonitor -n demo
NAME                      AGE
prom-operator-sdb-stats   32m

```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of SingleStore crd.

```yaml
$ kubectl get servicemonitor -n demo prom-operator-sdb-stats -oyaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2024-10-01T05:37:40Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: prom-operator-sdb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: singlestores.kubedb.com
    release: prometheus
  name: prom-operator-sdb-stats
  namespace: demo
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    controller: true
    kind: Service
    name: prom-operator-sdb-stats
    uid: 33802913-be0f-49ea-ac81-cf0136ed9fbc
  resourceVersion: "98648"
  uid: f26855f0-5f0e-45a6-8bf2-531d2a370377
spec:
  endpoints:
  - honorLabels: true
    interval: 10s
    path: /metrics
    port: metrics
  namespaceSelector:
    matchNames:
    - demo
  selector:
    matchLabels:
      app.kubernetes.io/component: database
      app.kubernetes.io/instance: prom-operator-sdb
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: singlestores.kubedb.com
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in SingleStore crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `prom-operator-sdb-stats` service. It also, target the `prom-http` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n default -l=app=prometheus
NAME                      READY   STATUS    RESTARTS   AGE
prometheus-prometheus-0   3/3     Running   1          121m
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n default prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `prom-http` endpoint of `prom-operator-sdb-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/guides/singlestore/monitoring/prometheus-operator/images/prom-operator-sdb-target.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels marked by red rectangle. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboard with collected metrics.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run following commands

```bash
# cleanup database
kubectl delete -n demo sdb/prom-operator-sdb

# cleanup Prometheus resources if exist
kubectl delete -f https://raw.githubusercontent.com/appscode/third-party-tools/master/monitoring/prometheus/coreos-operator/artifacts/prometheus.yaml
kubectl delete -f https://raw.githubusercontent.com/appscode/third-party-tools/master/monitoring/prometheus/coreos-operator/artifacts/prometheus-rbac.yaml

# cleanup Prometheus operator resources if exist
kubectl delete -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.41/bundle.yaml

# delete namespace
kubectl delete ns demo
```

## Next Steps

- Monitor your SingleStore database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/singlestore/monitoring/builtin-prometheus/index.md).
- Detail concepts of [SingleStore object](/docs/guides/singlestore/concepts/singlestore.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
