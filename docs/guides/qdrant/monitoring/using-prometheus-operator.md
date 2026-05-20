---
title: Monitor Qdrant using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: qdrant-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: qdrant-monitoring
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Qdrant Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor Qdrant deployed with KubeDB.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/qdrant/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, you can deploy one using this helm chart [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack).

> Note: YAML files used in this tutorial are stored in [docs/examples/qdrant/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/qdrant/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by `Prometheus` Operator. We are going to provide these labels in `spec.monitor.prometheus.serviceMonitor.labels` field of Qdrant CR so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME                                    VERSION   DESIRED   READY   RECONCILED   AVAILABLE   AGE
monitoring   prometheus-kube-prometheus-prometheus   v2.54.1   1         1       True         True        16d
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
  creationTimestamp: "2024-10-14T10:14:36Z"
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
  resourceVersion: "1004097"
  uid: b7879d3e-e4bb-4425-8d78-f917561d95f7
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
    - lastTransitionTime: "2024-10-31T07:38:36Z"
      message: ""
      observedGeneration: 1
      reason: ""
      status: "True"
      type: Available
    - lastTransitionTime: "2024-10-31T07:38:36Z"
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

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` CR. So, we are going to use this label in `spec.monitor.prometheus.serviceMonitor.labels` field of Qdrant CR.

## Deploy Qdrant with Monitoring Enabled

Now, let's deploy a Qdrant cluster with monitoring enabled. Below is the Qdrant object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-monitoring
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
  deletionPolicy: WipeOut
```

Here,

- `monitor.agent: prometheus.io/operator` indicates that we are going to monitor this Qdrant cluster using Prometheus operator.

- `monitor.prometheus.serviceMonitor.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.

- `monitor.prometheus.serviceMonitor.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the Qdrant object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/monitoring/qdrant-monitoring.yaml
qdrant.kubedb.com/qdrant-monitoring created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get qdrant -n demo qdrant-monitoring
NAME                VERSION   STATUS   AGE
qdrant-monitoring   1.17.0    Ready    1m
```

KubeDB will create a separate stats service with name `{qdrant cr name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=qdrant-monitoring"
NAME                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
qdrant-monitoring           ClusterIP   10.96.225.130   <none>        6333/TCP    1m
qdrant-monitoring-stats     ClusterIP   10.96.147.93    <none>        6333/TCP    1m
```

Here, `qdrant-monitoring-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```bash
$ kubectl describe svc -n demo qdrant-monitoring-stats
```
```yaml
Name:              qdrant-monitoring-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
  app.kubernetes.io/instance=qdrant-monitoring
  app.kubernetes.io/managed-by=kubedb.com
  app.kubernetes.io/name=qdrants.kubedb.com
  kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/instance=qdrant-monitoring,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=qdrants.kubedb.com
Type:              ClusterIP
Port:              metrics  6333/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.0.47:6333,10.244.0.48:6333,10.244.0.49:6333
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use these information to target its endpoints.

KubeDB will also create a `ServiceMonitor` CR in `demo` namespace that select the endpoints of `qdrant-monitoring-stats` service. Verify that the `ServiceMonitor` CR has been created.

```bash
$ kubectl get servicemonitor -n demo
NAME                        AGE
qdrant-monitoring-stats     1m
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of Qdrant CR.

```bash
$ kubectl get servicemonitor -n demo qdrant-monitoring-stats -o yaml
```

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2024-10-31T07:38:36Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: qdrant-monitoring
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: qdrants.kubedb.com
    release: prometheus
  name: qdrant-monitoring-stats
  namespace: demo
  ownerReferences:
    - apiVersion: v1
      blockOwnerDeletion: true
      controller: true
      kind: Service
      name: qdrant-monitoring-stats
      uid: 99193679-301b-41fd-aae5-a732b3070d19
  resourceVersion: "1004080"
  uid: 87635ad4-dfb2-4544-89af-e48b40783205
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
      app.kubernetes.io/instance: qdrant-monitoring
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: qdrants.kubedb.com
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in Qdrant CR.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `qdrant-monitoring-stats` service. It also, target the `metrics` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus-kube-prometheus-prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app.kubernetes.io/name=prometheus
NAME                                                 READY   STATUS    RESTARTS         AGE
prometheus-prometheus-kube-prometheus-prometheus-0   2/2     Running   1                16d
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-kube-prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-kube-prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `metrics` endpoint of `qdrant-monitoring-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/qdrant/monitoring/qdrant-monitoring-targets.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboards with collected metrics.

## Grafana Dashboards

There are pre-built Grafana dashboards to monitor Qdrant databases managed by KubeDB:

- KubeDB / Qdrant / Summary: Shows overall summary of Qdrant instance.
- KubeDB / Qdrant / Pod: Shows individual pod-level information.

To use these dashboards, download them from [qdrant-dashboards](https://github.com/ops-center/grafana-dashboards/tree/master/qdrant) and import them into your Grafana instance.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run following commands

```bash
kubectl delete -n demo qdrant/qdrant-monitoring
kubectl delete ns demo

helm uninstall prometheus -n monitoring
kubectl delete ns monitoring
```

## Next Steps
- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
