---
title: Monitor Milvus using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: milvus-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: milvus-monitoring
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Milvus using Prometheus Operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus, Alertmanager and related monitoring components. This tutorial will show you how to use the Prometheus operator to monitor Milvus database deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/milvus/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy the database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/milvus](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/milvus) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.serviceMonitor.labels` field of Milvus crd so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME                                    VERSION   DESIRED   READY   RECONCILED   AVAILABLE   AGE
monitoring   prometheus-kube-prometheus-prometheus   v3.11.3   1         1       True         True        25h
```

> If you don't have any Prometheus server running in your cluster, deploy one following the guide specified in **Before You Begin** section.

Now, let's view the YAML of the available Prometheus server `prometheus-kube-prometheus-prometheus` in `monitoring` namespace.

```bash
$ kubectl get prometheus -n monitoring prometheus-kube-prometheus-prometheus -oyaml
```
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  annotations:
    meta.helm.sh/release-name: prometheus
    meta.helm.sh/release-namespace: monitoring
  creationTimestamp: "2026-05-04T04:56:27Z"
  generation: 1
  labels:
    app: kube-prometheus-stack-prometheus
    app.kubernetes.io/instance: prometheus
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/part-of: kube-prometheus-stack
    app.kubernetes.io/version: 84.5.0
    chart: kube-prometheus-stack-84.5.0
    heritage: Helm
    release: prometheus
  name: prometheus-kube-prometheus-prometheus
  namespace: monitoring
  resourceVersion: "404630"
  uid: 6cf79306-5f52-46e7-9342-b7447016823e
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
  enableOTLPReceiver: false
  evaluationInterval: 30s
  externalUrl: http://prometheus-kube-prometheus-prometheus.monitoring:9090
  hostNetwork: false
  image: quay.io/prometheus/prometheus:v3.11.3
  imagePullPolicy: IfNotPresent
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
  version: v3.11.3
  walCompression: true
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2026-05-04T04:57:09Z"
    message: ""
    observedGeneration: 1
    reason: ""
    status: "True"
    type: Available
  - lastTransitionTime: "2026-05-04T04:56:37Z"
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
Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` CR. So, we are going to use this label in `spec.monitor.prometheus.labels` field of Milvus CR.


## Deploy Milvus with Monitoring Enabled

At first, let's deploy a Milvus database with monitoring enabled.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: coreos-prom-milvus
  namespace: demo
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: "my-release-minio"
  topology:
    mode: Distributed
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Here,
- `spec.monitor.agent: prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.
- `spec.monitor.prometheus.serviceMonitor.interval` indicates that the Prometheus should scrape metrics from this database with 10 seconds interval.

Let's create the Milvus object that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/milvus/monitoring/coreos-prom-milvus.yaml
milvus.kubedb.com/coreos-prom-milvus created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get milvus -n demo coreos-prom-milvus
NAME                 VERSION   STATUS   AGE
coreos-prom-milvus   2.6.11     Ready    2m
```
KubeDB will create a separate stats service with name `{Milvus crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n kubedb --selector="app.kubernetes.io/instance=coreos-prom-milvus"
NAME                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
coreos-prom-milvus-stats      ClusterIP   10.43.113.92    <none>        9091/TCP    50m
```
Here, `coreos-prom-milvus-stats` service has been created for monitoring purpose.
Let's describe this stats service.

```bash
$ kubectl describe svc -n demo coreos-prom-milvus-stats
```
```yaml
Name:              coreos-prom-milvus-stats
Namespace:         kubedb
Labels:            app.kubernetes.io/component=database
  app.kubernetes.io/instance=coreos-prom-milvus-stats
  app.kubernetes.io/managed-by=kubedb.com
  app.kubernetes.io/name=milvuses.kubedb.com
  kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/instance=coreos-prom-milvus-stats,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=milvuses.kubedb.com
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                10.43.113.92
IPs:               10.43.113.92
Port:              metrics  9091/TCP
TargetPort:        metrics/TCP
Endpoints:         10.42.0.102:9091,10.42.0.107:9091,10.42.0.111:9091 + 2 more...
Session Affinity:  None
Events:            <none>
```
Notice the `Labels` and `Port` fields. `ServiceMonitor` will use this information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `demo` namespace that select the endpoints of `coreos-prom-milvus-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$ kubectl get servicemonitor -n demo
NAME                      AGE
coreos-prom-milvus-stats   2m
```


Let's verify the `ServiceMonitor` has the right label to be discovered by Prometheus:

```bash
$ kubectl get servicemonitor -n demo coreos-prom-milvus-stats -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    release: prometheus
  name: coreos-prom-milvus-stats
  namespace: demo
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
      app.kubernetes.io/instance: coreos-prom-milvus
```
Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in Milvus crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `coreos-prom-milvus-stats` service. It also, target the `metrics` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app.kubernetes.io/name=prometheus
NAME                                                 READY   STATUS    RESTARTS   AGE
prometheus-prometheus-kube-prometheus-prometheus-0   2/2     Running   0          27h
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-kube-prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-kube-prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-kube-prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `metrics` endpoint of `coreos-prom-milvus-stats` service as one of the targets.

Check the `endpoint` and `service` labels marked by the red rectangles. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create a beautiful dashboard with collected metrics.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo milvus/coreos-prom-milvus -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo milvus/coreos-prom-milvus

kubectl delete ns demo
kubectl delete ns monitoring
```
