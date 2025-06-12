---
title: Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: hz-promtheus-operator-monitoring-hazelcast
    name: Prometheus Operator
    parent: hz-monitoring-hazelcast
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Hazelcast Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor Hazelcast database deployed with KubeDB.

<p align="center">
  <img alt="Database Monitoring Flow"  src="/docs/images/solr/monitoring/database-monitoring-overview.png">
</p>

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/hazelcast/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy database in `demo` namespace.

```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md).

- If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

> Note: YAML files used in this tutorial are stored in [docs/examples/hazelcast](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hazelcast) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.labels` field of Hazelcast crd so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus -A
NAMESPACE    NAME                                    VERSION   DESIRED   READY   RECONCILED   AVAILABLE   AGE
monitoring   prometheus-kube-prometheus-prometheus   v2.54.1   1         1       True         True        11d
```

> If you don't have any Prometheus server running in your cluster, deploy one following the guide specified in **Before You Begin** section.

Now, let's view the YAML of the available Prometheus server `prometheus` in `monitoring` namespace.

```bash
$ kubectl get prometheus -n monitoring prometheus-kube-prometheus-prometheus -oyaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  annotations:
    meta.helm.sh/release-name: prometheus
    meta.helm.sh/release-namespace: monitoring
  creationTimestamp: "2025-05-30T05:23:20Z"
  generation: 1
  labels:
    app: kube-prometheus-stack-prometheus
    app.kubernetes.io/instance: prometheus
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/part-of: kube-prometheus-stack
    app.kubernetes.io/version: 72.6.4
    chart: kube-prometheus-stack-72.6.4
    heritage: Helm
    release: prometheus
  name: prometheus-kube-prometheus-prometheus
  namespace: monitoring
  resourceVersion: "1277397"
  uid: f38d5c58-fd41-43ff-91a7-8f137c4658e0
spec:
  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - podAffinityTerm:
          labelSelector:
            matchExpressions:
            - key: app.kubernetes.io/name
              operator: In
              values:
              - prometheus
            - key: app.kubernetes.io/instance
              operator: In
              values:
              - prometheus-kube-prometheus-prometheus
          topologyKey: kubernetes.io/hostname
        weight: 100
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
  image: quay.io/prometheus/prometheus:v3.4.0
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
  version: v3.4.0
  walCompression: true
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2025-06-12T05:01:24Z"
    message: ""
    observedGeneration: 1
    reason: ""
    status: "True"
    type: Available
  - lastTransitionTime: "2025-06-12T05:01:24Z"
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

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` crd. So, we are going to use this label in `spec.monitor.prometheus.labels` field of Hazelcast crd.

## Deploy Hazelcast with Monitoring Enabled

At first, let's deploy a Hazelcast database with monitoring enabled. Below is the Hazelcast object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: operator-prom-hz
  namespace: demo
spec:
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: Halt
  licenseSecret:
    name: hz-license-key
  replicas: 3
  version: 5.5.2
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
    storageClassName: standard
```

Here,

- `monitor.agent:  prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
- `monitor.prometheus.namespace: monitoring` specifies that KubeDB should create `ServiceMonitor` in `monitoring` namespace.
- `monitor.prometheus.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.
- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the Hazelcast object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/monitoring/hazelcast-operator.yaml
hazelcast.kubedb.com/operator-prom-hz created
```

Now, wait for the database to go into `Running` state.

```bash
$ kubectl get hz -n demo
NAME               TYPE                  VERSION   STATUS   AGE
operator-prom-hz   kubedb.com/v1alpha2   5.5.2     Ready    55m
```

KubeDB will create a separate stats service with name `{Hazelcast crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo -l 'app.kubernetes.io/instance=operator-prom-hz'
NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
operator-prom-hz         ClusterIP   10.43.177.245   <none>        5701/TCP    56m
operator-prom-hz-pods    ClusterIP   None            <none>        5701/TCP    56m
operator-prom-hz-stats   ClusterIP   10.43.64.206    <none>        56790/TCP   56m
```

Here, `operator-prom-hz-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```bash
$ kubectl describe svc -n demo operator-prom-hz-stats 
Name:                     operator-prom-hz-stats
Namespace:                demo
Labels:                   app.kubernetes.io/component=database
                          app.kubernetes.io/instance=operator-prom-hz
                          app.kubernetes.io/managed-by=kubedb.com
                          app.kubernetes.io/name=hazelcasts.kubedb.com
                          kubedb.com/role=stats
Annotations:              monitoring.appscode.com/agent: prometheus.io/operator
Selector:                 app.kubernetes.io/instance=operator-prom-hz,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=hazelcasts.kubedb.com
Type:                     ClusterIP
IP Family Policy:         SingleStack
IP Families:              IPv4
IP:                       10.43.64.206
IPs:                      10.43.64.206
Port:                     metrics  56790/TCP
TargetPort:               metrics/TCP
Endpoints:                10.42.0.58:56790,10.42.0.59:56790,10.42.0.60:56790
Session Affinity:         None
Internal Traffic Policy:  Cluster
Events:                   <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use these information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `monitoring` namespace that select the endpoints of `operator-prom-hz-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$ kubectl get servicemonitor -n demo 
NAME                     AGE
operator-prom-hz-stats   125m
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of Hazelcast crd.

```bash
$ kubectl get servicemonitor -n demo operator-prom-hz-stats -oyaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2025-06-12T05:01:23Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: operator-prom-hz
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: hazelcasts.kubedb.com
    release: prometheus
  name: operator-prom-hz-stats
  namespace: demo
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    controller: true
    kind: Service
    name: operator-prom-hz-stats
    uid: 22bac407-71ee-4fb5-a9d0-310840a05869
  resourceVersion: "1277360"
  uid: 221d5403-e133-4138-8e9b-a1472f93f069
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
      app.kubernetes.io/instance: operator-prom-hz
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: hazelcasts.kubedb.com
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in Hazelcast crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `operator-prom-hz-stats` service. It also, target the `prom-http` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring
NAME                                                     READY   STATUS    RESTARTS   AGE
alertmanager-prometheus-kube-prometheus-alertmanager-0   2/2     Running   0          13d
prometheus-grafana-6956bd7864-lggrr                      3/3     Running   0          13d
prometheus-kube-prometheus-operator-777c5cb7b4-dcmc5     1/1     Running   0          13d
prometheus-kube-state-metrics-f8fc86d54-x8qjk            1/1     Running   0          13d
prometheus-prometheus-kube-prometheus-prometheus-0       2/2     Running   0          13d
prometheus-prometheus-node-exporter-szndm                1/1     Running   0          13d

```

Prometheus server is listening to port `9090` of `prometheus-prometheus-kube-prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-kube-prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-kube-prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
Handling connection for 9090

```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `prom-http` endpoint of `operator-prom-hz-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/hazelcast/monitoring/hazelcast-operator-prom-target.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels marked by red rectangle. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboard with collected metrics.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run following commands

```bash
# cleanup database
kubectl delete sl -n demo operator-prom-sl

# cleanup prometheus resources
kubectl delete -n monitoring prometheus prometheus
kubectl delete -n monitoring clusterrolebinding prometheus
kubectl delete -n monitoring clusterrole prometheus
kubectl delete -n monitoring serviceaccount prometheus
kubectl delete -n monitoring service prometheus-operated

# cleanup prometheus operator resources
kubectl delete -n monitoring deployment prometheus-kube-prometheus-operator
kubectl delete -n dmeo serviceaccount prometheus-operator
kubectl delete clusterrolebinding prometheus-operator
kubectl delete clusterrole prometheus-operator

# delete namespace
kubectl delete ns monitoring
kubectl delete ns demo
```
