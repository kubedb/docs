---
title: Monitor Cassandra using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: cas-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: cas-monitoring-cassandra
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Cassandra Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor Cassandra database deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one locally by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/cassandra/monitoring/overview.md).

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, you can deploy one using this helm chart [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack).
  
- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy the prometheus operator helm chart. Alternatively, you can use `--create-namespace` flag while deploying prometheus. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```



> Note: YAML files used in this tutorial are stored in [docs/examples/cassandra](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/cassandra) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.serviceMonitor.labels` field of Cassandra crd so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME                                    VERSION   DESIRED   READY   RECONCILED   AVAILABLE   AGE
monitoring   prometheus-kube-prometheus-prometheus   v3.4.2    1         1       True         True        7h43m
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
  creationTimestamp: "2025-07-24T04:20:17Z"
  finalizers:
  - monitoring.appscode.com/prometheus
  generation: 1
  labels:
    app: kube-prometheus-stack-prometheus
    app.kubernetes.io/instance: prometheus
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/part-of: kube-prometheus-stack
    app.kubernetes.io/version: 75.9.0
    chart: kube-prometheus-stack-75.9.0
    heritage: Helm
    release: prometheus
  name: prometheus-kube-prometheus-prometheus
  namespace: monitoring
  resourceVersion: "49548"
  uid: aa50a17f-9e2e-4f0e-8898-af5dd7f90c9b
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
  image: quay.io/prometheus/prometheus:v3.4.2
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
  version: v3.4.2
  walCompression: true
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2025-07-25T04:40:59Z"
    message: ""
    observedGeneration: 1
    reason: ""
    status: "True"
    type: Available
  - lastTransitionTime: "2025-07-25T04:40:59Z"
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

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` crd. So, we are going to use this label in `spec.monitor.prometheus.serviceMonitor.labels` field of Cassandra crd.

## Deploy Cassandra with Monitoring Enabled

At first, let's deploy a Cassandra database with monitoring enabled. Below is the Cassandra object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-prod
  namespace: demo
spec:
  version: 5.0.3
  topology:
    rack:
      - name: r0
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 2Gi
                    cpu: 2
                  requests:
                    memory: 1Gi
                    cpu: 1
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
  deletionPolicy: WipeOut
  monitor:
    agent: "prometheus.io/operator"
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

Let's create the cassandra object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/monitoring/cas-with-monirtoring.yaml
cassandras.kubedb.com/cassandra created
```

Now, wait for the database to go into `Running` state.

```bash
$  kubectl get cas -n demo cassandra-prod
NAME             TYPE                  VERSION   STATUS   AGE
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Ready    17h
```

KubeDB will create a separate stats service with name `{Cassandra crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=cassandra-prod"
NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                               AGE
cassandra-prod                ClusterIP   10.96.232.61   <none>        9042/TCP,7000/TCP,7199/TCP,7001/TCP   17h
cassandra-prod-rack-r0-pods   ClusterIP   None           <none>        9042/TCP,7000/TCP,7199/TCP,7001/TCP   17h
cassandra-prod-stats          ClusterIP   10.96.189.65   <none>        56790/TCP                             17h
```

Here, `cassandra-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```bash
$ kubectl describe svc -n demo cassandra-prod-stats
Name:                     cassandra-prod-stats
Namespace:                demo
Labels:                   app.kubernetes.io/component=database
                          app.kubernetes.io/instance=cassandra-prod
                          app.kubernetes.io/managed-by=kubedb.com
                          app.kubernetes.io/name=cassandras.kubedb.com
                          kubedb.com/role=stats
Annotations:              monitoring.appscode.com/agent: prometheus.io/operator
Selector:                 app.kubernetes.io/instance=cassandra-prod,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=cassandras.kubedb.com
Type:                     ClusterIP
IP Family Policy:         SingleStack
IP Families:              IPv4
IP:                       10.96.189.65
IPs:                      10.96.189.65
Port:                     metrics  56790/TCP
TargetPort:               metrics/TCP
Endpoints:                10.244.0.19:8080,10.244.0.18:8080
Session Affinity:         None
Internal Traffic Policy:  Cluster
Events:                   <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use this information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `demo` namespace that select the endpoints of `cassandra-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$  kubectl get servicemonitor -n demo
NAME                   AGE
cassandra-prod-stats   17h
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of Cassandra crd.

```bash
$ kubectl get servicemonitor -n demo cassandra-prod-stats -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2025-07-24T11:51:13Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: cassandra-prod
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: cassandras.kubedb.com
    release: prometheus
  name: cassandra-prod-stats
  namespace: demo
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    controller: true
    kind: Service
    name: cassandra-prod-stats
    uid: f162ee2b-7bf2-4c27-88ec-76ec95c1829f
  resourceVersion: "46434"
  uid: 294b03fc-acbf-4560-b940-ea1f1db2171b
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
      app.kubernetes.io/instance: cassandra-prod
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: cassandras.kubedb.com
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in Cassandra crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `cassandra-prod-stats` service. It also, target the `metrics` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app.kubernetes.io/name=prometheus
NAME                                                 READY   STATUS    RESTARTS      AGE
prometheus-prometheus-kube-prometheus-prometheus-0   2/2     Running   2 (18m ago)   24h
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-kube-prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-kube-prometheus-prometheus` service which is pointing to the prometheus pod,

```bash
$ kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `metrics` endpoint of `cassandra-stats` service as one of the targets.

Check the `endpoint` and `service` labels. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create a beautiful dashboard with collected metrics.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run following commands

```bash
kubectl delete -n demo cas/cassandra
kubectl delete ns demo
```

## Next Steps

- Learn how to use KubeDB to run a Apache Cassandra cluster [here](/docs/guides/cassandra/README.md).
- Detail concepts of [CassandraVersion object](/docs/guides/cassandra/concepts/cassandraversion.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).