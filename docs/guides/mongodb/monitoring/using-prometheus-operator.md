---
title: Monitor MongoDB using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: mg-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: mg-monitoring-mongodb
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring MongoDB Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor MongoDB database deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/mongodb/monitoring/overview.md).

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, you can deploy one using this helm chart [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack).
  
- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy the prometheus operator helm chart. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```



> Note: YAML files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.labels` field of MongoDB crd so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME                                    VERSION   REPLICAS   AGE
monitoring   prometheus-kube-prometheus-prometheus   v2.24.0   1          6h48m
```

> If you don't have any Prometheus server running in your cluster, deploy one following the guide specified in **Before You Begin** section.

Now, let's view the YAML of the available Prometheus server `prometheus` in `monitoring` namespace.

```yaml
$ kubectl get prometheus -n monitoring prometheus-kube-prometheus-prometheus -o yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  annotations:
    meta.helm.sh/release-name: prometheus
    meta.helm.sh/release-namespace: monitoring
  creationTimestamp: "2021-03-09T10:47:17Z"
  generation: 1
  labels:
    app: kube-prometheus-stack-prometheus
    app.kubernetes.io/managed-by: Helm
    chart: kube-prometheus-stack-13.13.0
    heritage: Helm
    release: prometheus
  managedFields:
    - apiVersion: monitoring.coreos.com/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            .: {}
            f:meta.helm.sh/release-name: {}
            f:meta.helm.sh/release-namespace: {}
          f:labels:
            .: {}
            f:app: {}
            f:app.kubernetes.io/managed-by: {}
            f:chart: {}
            f:heritage: {}
            f:release: {}
        f:spec:
          .: {}
          f:alerting:
            .: {}
            f:alertmanagers: {}
          f:enableAdminAPI: {}
          f:externalUrl: {}
          f:image: {}
          f:listenLocal: {}
          f:logFormat: {}
          f:logLevel: {}
          f:paused: {}
          f:podMonitorNamespaceSelector: {}
          f:podMonitorSelector:
            .: {}
            f:matchLabels:
              .: {}
              f:release: {}
          f:portName: {}
          f:probeNamespaceSelector: {}
          f:probeSelector:
            .: {}
            f:matchLabels:
              .: {}
              f:release: {}
          f:replicas: {}
          f:retention: {}
          f:routePrefix: {}
          f:ruleNamespaceSelector: {}
          f:ruleSelector:
            .: {}
            f:matchLabels:
              .: {}
              f:app: {}
              f:release: {}
          f:securityContext:
            .: {}
            f:fsGroup: {}
            f:runAsGroup: {}
            f:runAsNonRoot: {}
            f:runAsUser: {}
          f:serviceAccountName: {}
          f:serviceMonitorNamespaceSelector: {}
          f:serviceMonitorSelector:
            .: {}
            f:matchLabels:
              .: {}
              f:release: {}
          f:shards: {}
          f:version: {}
      manager: Go-http-client
      operation: Update
      time: "2021-03-09T10:47:17Z"
  name: prometheus-kube-prometheus-prometheus
  namespace: monitoring
  resourceVersion: "100084"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/prometheuses/prometheus-kube-prometheus-prometheus
  uid: 4b7a8c5b-09c4-4858-8232-13cbb71c766b
spec:
  alerting:
    alertmanagers:
      - apiVersion: v2
        name: prometheus-kube-prometheus-alertmanager
        namespace: monitoring
        pathPrefix: /
        port: web
  enableAdminAPI: false
  externalUrl: http://prometheus-kube-prometheus-prometheus.monitoring:9090
  image: quay.io/prometheus/prometheus:v2.24.0
  listenLocal: false
  logFormat: logfmt
  logLevel: info
  paused: false
  podMonitorNamespaceSelector: {}
  podMonitorSelector:
    matchLabels:
      release: prometheus
  portName: web
  probeNamespaceSelector: {}
  probeSelector:
    matchLabels:
      release: prometheus
  replicas: 1
  retention: "10d"
  routePrefix: /
  ruleNamespaceSelector: {}
  ruleSelector:
    matchLabels:
      app: kube-prometheus-stack
      release: prometheus
  securityContext:
    fsGroup: 2000
    runAsGroup: 2000
    runAsNonRoot: true
    runAsUser: 1000
  serviceAccountName: prometheus-kube-prometheus-prometheus
  serviceMonitorNamespaceSelector: {}
  serviceMonitorSelector:
    matchLabels:
      release: prometheus
  shards: 1
  version: v2.24.0
```

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` crd. So, we are going to use this label in `spec.monitor.prometheus.labels` field of MongoDB crd.

## Deploy MongoDB with Monitoring Enabled

At first, let's deploy an MongoDB database with monitoring enabled. Below is the MongoDB object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: coreos-prom-mgo
  namespace: demo
spec:
  version: "4.2.3"
  terminationPolicy: WipeOut
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
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
- `monitor.prometheus.namespace: monitoring` specifies that KubeDB should create `ServiceMonitor` in `monitoring` namespace.
- `monitor.prometheus.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.
- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the MongoDB object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/monitoring/coreos-prom-mgo.yaml
mongodb.kubedb.com/coreos-prom-mgo created
```

Now, wait for the database to go into `Running` state.

```bash
$ kubectl get mg -n demo coreos-prom-mgo
NAME              VERSION   STATUS    AGE
coreos-prom-mgo   4.2.3     Ready     34s
```

KubeDB will create a separate stats service with name `{MongoDB crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=coreos-prom-mgo"
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
coreos-prom-mgo         ClusterIP   10.96.150.171   <none>        27017/TCP   84s
coreos-prom-mgo-pods    ClusterIP   None            <none>        27017/TCP   84s
coreos-prom-mgo-stats   ClusterIP   10.96.218.41    <none>        56790/TCP   64s
```

Here, `coreos-prom-mgo-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```yaml
$ kubectl describe svc -n demo coreos-prom-mgo-stats
Name:              coreos-prom-mgo-stats
Namespace:         demo
Labels:            app.kubernetes.io/instance=coreos-prom-mgo
  app.kubernetes.io/managed-by=kubedb.com
  app.kubernetes.io/name=mongodbs.kubedb.com
  kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/instance=coreos-prom-mgo,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mongodbs.kubedb.com
Type:              ClusterIP
IP Families:       <none>
IP:                10.96.218.41
IPs:               <none>
Port:              metrics  56790/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.0.110:56790
Session Affinity:  None
Events:            <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use this information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `demo` namespace that select the endpoints of `coreos-prom-mgo-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$ kubectl get servicemonitor -n demo
NAME                    AGE
coreos-prom-mgo-stats   2m40s
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of MongoDB crd.

```yaml
$ kubectl get servicemonitor -n demo coreos-prom-mgo-stats -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2021-03-09T17:40:16Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: coreos-prom-mgo
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mongodbs.kubedb.com
    release: prometheus
  managedFields:
    - apiVersion: monitoring.coreos.com/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:labels:
            .: {}
            f:app.kubernetes.io/component: {}
            f:app.kubernetes.io/instance: {}
            f:app.kubernetes.io/managed-by: {}
            f:app.kubernetes.io/name: {}
            f:release: {}
          f:ownerReferences: {}
        f:spec:
          .: {}
          f:endpoints: {}
          f:namespaceSelector:
            .: {}
            f:matchNames: {}
          f:selector:
            .: {}
            f:matchLabels:
              .: {}
              f:app.kubernetes.io/instance: {}
              f:app.kubernetes.io/managed-by: {}
              f:app.kubernetes.io/name: {}
              f:kubedb.com/role: {}
      manager: mg-operator
      operation: Update
      time: "2021-03-09T17:40:16Z"
  name: coreos-prom-mgo-stats
  namespace: demo
  ownerReferences:
    - apiVersion: v1
      blockOwnerDeletion: true
      controller: true
      kind: Service
      name: coreos-prom-mgo-stats
      uid: 906358eb-90dc-4a06-b9d3-89f557ad6ef4
  resourceVersion: "184540"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/demo/servicemonitors/coreos-prom-mgo-stats
  uid: b0df2b5e-b6dd-4e8b-bf48-9da14f099d83
spec:
  endpoints:
    - bearerTokenSecret:
        key: ""
      honorLabels: true
      interval: 10s
      path: /metrics
      port: metrics
  namespaceSelector:
    matchNames:
      - demo
  selector:
    matchLabels:
      app.kubernetes.io/instance: coreos-prom-mgo
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: mongodbs.kubedb.com
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in MongoDB crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `coreos-prom-mgo-stats` service. It also, target the `metrics` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                                                 READY   STATUS    RESTARTS   AGE
prometheus-prometheus-kube-prometheus-prometheus-0   2/2     Running   1          6h58m
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-kube-prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-kube-prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-kube-prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `metrics` endpoint of `coreos-prom-mgo-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/mongodb/monitoring/mg-coreos-prom-target.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels marked by the red rectangles. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create a beautiful dashboard with collected metrics.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run following commands

```bash
kubectl delete -n demo mg/coreos-prom-mgo
kubectl delete ns demo
```

## Next Steps

- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- [Backup and Restore](/docs/guides/mongodb/backup/stash.md) process of MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
