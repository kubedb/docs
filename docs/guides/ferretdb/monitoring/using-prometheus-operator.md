---
title: Monitor FerretDB using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: fr-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: fr-monitoring-ferretdb
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring FerretDB Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor FerretDB database deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/ferretdb/monitoring/overview.md).

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, you can deploy one using this helm chart [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy the prometheus operator helm chart. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/ferretdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/ferretdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.serviceMonitor.labels` field of FerretDB crd so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME                                    VERSION   DESIRED   READY   RECONCILED   AVAILABLE   AGE
monitoring   prometheus-kube-prometheus-prometheus   v2.54.1   1         1       True         True        13m
```

> If you don't have any Prometheus server running in your cluster, deploy one following the guide specified in **Before You Begin** section.

Now, let's view the YAML of the available Prometheus server `prometheus` in `monitoring` namespace.
```bash
$ kubectl get prometheus -n monitoring prometheus-kube-prometheus-prometheus -o yaml
```
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  annotations:
    meta.helm.sh/release-name: prometheus
    meta.helm.sh/release-namespace: monitoring
  creationTimestamp: "2024-10-14T17:17:25Z"
  generation: 1
  labels:
    app: kube-prometheus-stack-prometheus
    app.kubernetes.io/instance: prometheus
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/part-of: kube-prometheus-stack
    app.kubernetes.io/version: 65.2.0
    chart: kube-prometheus-stack-65.2.0
    heritage: Helm
    release: prometheus
  name: prometheus-kube-prometheus-prometheus
  namespace: monitoring
  resourceVersion: "58118"
  uid: b1bf237b-2fdc-459c-b92f-e087a1119f33
spec:
  alerting:
    alertmanagers:
      - apiVersion: v2
        name: prometheus-kube-prometheus-alertmanager
        namespace: monitoring
        pathPrefix: /
        port: http-web
  automountServiceAccountToken: true
  enableAdminAPI: false
  evaluationInterval: 30s
  externalUrl: http://prometheus-kube-prometheus-prometheus.monitoring:9090
  hostNetwork: false
  image: quay.io/prometheus/prometheus:v2.54.1
  listenLocal: false
  logFormat: logfmt
  logLevel: info
  paused: false
  podMonitorNamespaceSelector: {}
  podMonitorSelector:
    matchLabels:
      release: prometheus
  portName: http-web
  probeNamespaceSelector: {}
  probeSelector:
    matchLabels:
      release: prometheus
  replicas: 1
  retention: 10d
  routePrefix: /
  ruleNamespaceSelector: {}
  ruleSelector:
    matchLabels:
      release: prometheus
  scrapeConfigNamespaceSelector: {}
  scrapeConfigSelector:
    matchLabels:
      release: prometheus
  scrapeInterval: 30s
  securityContext:
    fsGroup: 2000
    runAsGroup: 2000
    runAsNonRoot: true
    runAsUser: 1000
    seccompProfile:
      type: RuntimeDefault
  serviceAccountName: prometheus-kube-prometheus-prometheus
  serviceMonitorNamespaceSelector: {}
  serviceMonitorSelector:
    matchLabels:
      release: prometheus
  shards: 1
  tsdb:
    outOfOrderTimeWindow: 0s
  version: v2.54.1
  walCompression: true
status:
  availableReplicas: 1
  conditions:
    - lastTransitionTime: "2024-10-14T17:27:17Z"
      message: ""
      observedGeneration: 1
      reason: ""
      status: "True"
      type: Available
    - lastTransitionTime: "2024-10-14T17:27:17Z"
      message: ""
      observedGeneration: 1
      reason: ""
      status: "True"
      type: Reconciled
  paused: false
  replicas: 1
  selector: app.kubernetes.io/instance=prometheus-kube-prometheus-prometheus,app.kubernetes.io/managed-by=prometheus-operator,app.kubernetes.io/name=prometheus,operator.prometheus.io/name=prometheus-kube-prometheus-prometheus,prometheus=prometheus-kube-prometheus-prometheus
  shardStatuses:
    - availableReplicas: 1
      replicas: 1
      shardID: "0"
      unavailableReplicas: 0
      updatedReplicas: 1
  shards: 1
  unavailableReplicas: 0
  updatedReplicas: 1
```

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` crd. So, we are going to use this label in `spec.monitor.prometheus.serviceMonitor.labels` field of FerretDB crd.

## Deploy FerretDB with Monitoring Enabled

At first, let's deploy an FerretDB database with monitoring enabled. Below is the FerretDB object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: coreos-prom-fr
  namespace: demo
spec:
  version: "2.0.0"
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
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
- `monitor.prometheus.serviceMonitor.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.
- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the FerretDB object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/monitoring/coreos-prom-fr.yaml
ferretdb.kubedb.com/coreos-prom-fr created
```

Now, wait for the database to go into `Running` state.

```bash
$ kubectl get fr -n demo coreos-prom-fr
NAME             NAMESPACE   VERSION   STATUS   AGE
coreos-prom-fr   demo        2.0.0     Ready    111s
```

KubeDB will create a separate stats service with name `{FerretDB crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=coreos-prom-fr"
NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
coreos-prom-fr         ClusterIP   10.96.234.253   <none>        27017/TCP   2m16s
coreos-prom-fr-stats   ClusterIP   10.96.27.143    <none>        56790/TCP   2m16s
```

Here, `coreos-prom-fr-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```bash
$ kubectl describe svc -n demo coreos-prom-fr-stats
```
```yaml
Name:              coreos-prom-fr-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
  app.kubernetes.io/instance=coreos-prom-fr
  app.kubernetes.io/managed-by=kubedb.com
  app.kubernetes.io/name=ferretdbs.kubedb.com
  kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/instance=coreos-prom-fr,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=ferretdbs.kubedb.com
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                10.96.27.143
IPs:               10.96.27.143
Port:              metrics  56790/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.0.59:8080,10.244.0.60:8080
Session Affinity:  None
Events:            <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use this information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `demo` namespace that select the endpoints of `coreos-prom-fr-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$ kubectl get servicemonitor -n demo
NAME                              AGE
coreos-prom-fr-pg-backend-stats   3m33s
coreos-prom-fr-stats              2m24s
```

> If backend Postgres is managed by KubeDB, KubeDB operator will also enable monitoring to backend Postgres. That's why `coreos-prom-fr-pg-backend-stats` `ServiceMonitor` also created. 
To look at the more details of KubeDB managed Postgres monitoring, you can look at [this documentation](/docs/guides/postgres/monitoring/using-prometheus-operator.md). 

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of FerretDB crd.

```bash
$ kubectl get servicemonitor -n demo coreos-prom-fr-stats -o yaml
```
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2024-10-14T17:35:38Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: coreos-prom-fr
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: ferretdbs.kubedb.com
    release: prometheus
  name: coreos-prom-fr-stats
  namespace: demo
  ownerReferences:
    - apiVersion: v1
      blockOwnerDeletion: true
      controller: true
      kind: Service
      name: coreos-prom-fr-stats
      uid: d0811d68-6e31-4357-b35a-8a7793ab4918
  resourceVersion: "59094"
  uid: cfa29869-8000-44fe-bc9b-e7e78b08da36
spec:
  endpoints:
    - honorLabels: true
      interval: 10s
      path: /debug/metrics
      port: metrics
  namespaceSelector:
    matchNames:
      - demo
  selector:
    matchLabels:
      app.kubernetes.io/component: database
      app.kubernetes.io/instance: coreos-prom-fr
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: ferretdbs.kubedb.com
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in FerretDB crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `coreos-prom-fr-stats` service. It also, target the `metrics` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app.kubernetes.io/name=prometheus
NAME                                                 READY   STATUS    RESTARTS   AGE
prometheus-prometheus-kube-prometheus-prometheus-0   2/2     Running   0          27m
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-kube-prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-kube-prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-kube-prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `metrics` endpoint of `coreos-prom-fr-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/ferretdb/fr-coreos-prom-target.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels marked by the red rectangles. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create a beautiful dashboard with collected metrics.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run following commands

```bash
kubectl delete -n demo fr/coreos-prom-fr
kubectl delete ns demo
```

## Next Steps

- Monitor your FerretDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/ferretdb/monitoring/using-builtin-prometheus.md).
- Detail concepts of [FerretDB object](/docs/guides/ferretdb/concepts/ferretdb.md).
- Detail concepts of [FerretDBVersion object](/docs/guides/ferretdb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
