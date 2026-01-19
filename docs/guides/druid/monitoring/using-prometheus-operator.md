---
title: Monitor Druid using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: guides-druid-monitoring-operator-monitoring
    name: Prometheus Operator
    parent: guides-druid-monitoring
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Druid Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor Druid database deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one locally by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/druid/monitoring/overview.md).

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, you can deploy one using this helm chart [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack).
  
- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy the prometheus operator helm chart. Alternatively, you can use `--create-namespace` flag while deploying prometheus. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```



> Note: YAML files used in this tutorial are stored in [docs/examples/druid](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/druid) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.serviceMonitor.labels` field of Druid crd so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME                                    VERSION   DESIRED   READY   RECONCILED   AVAILABLE   AGE
monitoring   prometheus-kube-prometheus-prometheus   v2.42.0   1         1       True         True        2d23h
```

> If you don't have any Prometheus server running in your cluster, deploy one following the guide specified in **Before You Begin** section.

Now, let's view the YAML of the available Prometheus server `prometheus` in `monitoring` namespace.

```bash
$ kubectl get prometheus -n monitoring prometheus-kube-prometheus-prometheus -o yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  annotations:
    meta.helm.sh/release-name: prometheus
    meta.helm.sh/release-namespace: monitoring
  creationTimestamp: "2023-03-27T07:56:04Z"
  generation: 1
  labels:
    app: kube-prometheus-stack-prometheus
    app.kubernetes.io/instance: prometheus
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/part-of: kube-prometheus-stack
    app.kubernetes.io/version: 45.7.1
    chart: kube-prometheus-stack-45.7.1
    heritage: Helm
    release: prometheus
  name: prometheus-kube-prometheus-prometheus
  namespace: monitoring
  resourceVersion: "638797"
  uid: 0d1e7b8a-44ae-4794-ab45-95a5d7ae7f91
spec:
  alerting:
    alertmanagers:
    - apiVersion: v2
      name: prometheus-kube-prometheus-alertmanager
      namespace: monitoring
      pathPrefix: /
      port: http-web
  enableAdminAPI: false
  evaluationInterval: 30s
  externalUrl: http://prometheus-kube-prometheus-prometheus.monitoring:9090
  hostNetwork: false
  image: quay.io/prometheus/prometheus:v2.42.0
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
  scrapeInterval: 30s
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
  version: v2.42.0
  walCompression: true
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2023-03-27T07:56:23Z"
    observedGeneration: 1
    status: "True"
    type: Available
  - lastTransitionTime: "2023-03-30T03:39:18Z"
    observedGeneration: 1
    status: "True"
    type: Reconciled
  paused: false
  replicas: 1
  shardStatuses:
  - availableReplicas: 1
    replicas: 1
    shardID: "0"
    unavailableReplicas: 0
    updatedReplicas: 1
  unavailableReplicas: 0
  updatedReplicas: 1
```

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` crd. So, we are going to use this label in `spec.monitor.prometheus.serviceMonitor.labels` field of Druid crd.

## Deploy Druid with Monitoring Enabled

At first, let's deploy a Druid database with monitoring enabled. Below is the Druid object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-with-monitoring
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configuration:
      secretName: deep-storage-config
  topology:
    routers:
      replicas: 1
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

Here,

- `monitor.agent:  prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
- `monitor.prometheus.serviceMonitor.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.
- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the druid object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/monitoring/yamls/druid-with-monirtoring.yaml
druids.kubedb.com/druid-with-monitoring created
```

Now, wait for the database to go into `Running` state.

```bash
$ kubectl get dr -n demo druid
NAME                    TYPE                  VERSION   STATUS   AGE
druid-with-monitoring   kubedb.com/v1alpha2   28.0.1    Ready    2m24s
```

KubeDB will create a separate stats service with name `{Druid crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=druid-with-monitoring"
NAME                                  TYPE          CLUSTER-IP      EXTERNAL-IP   PORT(S)                                                  AGE
druid-with-monitoring-brokers         ClusterIP     10.96.28.252    <none>        8082/TCP                                                2m13s
druid-with-monitoring-coordinators    ClusterIP     10.96.52.186    <none>        8081/TCP                                                2m13s
druid-with-monitoring-pods            ClusterIP     None            <none>        8081/TCP,8090/TCP,8083/TCP,8091/TCP,8082/TCP,8888/TCP   2m13s
druid-with-monitoring-routers         ClusterIP     10.96.134.202   <none>        8888/TCP                                                2m13s
druid-with-monitoring-stats           ClusterIP     10.96.222.96    <none>        56790/TCP                                               2m13s
```

Here, `druid-with-monitoring-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```bash
$ kubectl describe svc -n demo druid-with-monitoring-stats
Name:              druid-with-monitoring-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
                   app.kubernetes.io/instance=druid-with-monitoring
                   app.kubernetes.io/managed-by=kubedb.com
                   app.kubernetes.io/name=druids.kubedb.com
                   kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/instance=druid-with-monitoring,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                10.96.29.174
IPs:               10.96.29.174
Port:              metrics  9104/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.0.68:9104,10.244.0.71:9104,10.244.0.72:9104 + 2 more...
Session Affinity:  None
Events:            <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use this information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `demo` namespace that select the endpoints of `druid-with-monitoring-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$ kubectl get servicemonitor -n demo
NAME                          AGE
druid-with-monitoring-stats   4m49s
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of Druid crd.

```bash
$ kubectl get servicemonitor -n demo druid-with-monitoring-stats -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2024-11-01T10:25:14Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: druid-with-monitoring
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: druids.kubedb.com
    release: prometheus
  name: druid-with-monitoring-stats
  namespace: demo
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    controller: true
    kind: Service
    name: druid-with-monitoring-stats
    uid: b3ae48f3-476e-4cec-95f6-f8e28538b605
  resourceVersion: "597152"
  uid: ff385538-eba5-48a3-91c1-1a4b15f3018a
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
      app.kubernetes.io/instance: druid-with-monitoring
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: druids.kubedb.com
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in Druid crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `druid-with-monitoring-stats` service. It also, target the `metrics` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app.kubernetes.io/name=prometheus
NAME                                                 READY   STATUS    RESTARTS        AGE
prometheus-prometheus-kube-prometheus-prometheus-0   2/2     Running   8 (4h27m ago)   3d
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-kube-prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-kube-prometheus-prometheus` service which is pointing to the prometheus pod,

```bash
$ kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `metrics` endpoint of `druid-with-monitoring-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/guides/druid/monitoring/images/druid-prometheus.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create a beautiful dashboard with collected metrics.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run following commands

```bash
kubectl delete -n demo dr/druid-with-monitoring
kubectl delete ns demo
```

## Next Steps

- Learn how to use KubeDB to run Apache Druid cluster [here](/docs/guides/druid/README.md).
- Deploy [dedicated  cluster](/docs/guides/druid/clustering/overview/index.md) for Apache Druid
[//]: # (- Deploy [combined cluster]&#40;/docs/guides/druid/clustering/combined-cluster/index.md&#41; for Apache Druid)
- Detail concepts of [DruidVersion object](/docs/guides/druid/concepts/druidversion.md).
[//]: # (- Learn to use KubeDB managed Druid objects using [CLIs]&#40;/docs/guides/druid/cli/cli.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).