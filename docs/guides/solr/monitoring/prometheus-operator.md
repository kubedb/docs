---
title: Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: sl-promtheus-operator-monitoring-solr
    name: Prometheus Operator
    parent: sl-monitoring-solr
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Solr Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor Elasticsearch database deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/elasticsearch/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy database in `demo` namespace.

```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md).

- If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.labels` field of Elasticsearch crd so that KubeDB creates `ServiceMonitor` object accordingly.

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
  creationTimestamp: "2024-10-17T08:16:24Z"
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
  resourceVersion: "632331"
  uid: 9cbf73fe-d9b5-456d-becf-770e07e298af
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
    - lastTransitionTime: "2024-10-29T06:44:10Z"
      message: ""
      observedGeneration: 1
      reason: ""
      status: "True"
      type: Available
    - lastTransitionTime: "2024-10-29T06:44:10Z"
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

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` crd. So, we are going to use this label in `spec.monitor.prometheus.labels` field of Solr crd.

## Deploy Solr with Monitoring Enabled

At first, let's deploy an Solr database with monitoring enabled. Below is the Solr object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: operator-prom-sl
  namespace: demo
spec:
  version: 9.4.1
  replicas: 2
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  solrModules:
  - s3-repository
  - gcs-repository
  - prometheus-exporter
  zookeeperRef:
    name: zoo
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard

```

Here,

- `monitor.agent:  prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
- `monitor.prometheus.namespace: monitoring` specifies that KubeDB should create `ServiceMonitor` in `monitoring` namespace.

- `monitor.prometheus.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.

- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the Elasticsearch object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/monitoring/solr-operator.yaml
solr.kubedb.com/operator-prom-sl created
```

Now, wait for the database to go into `Running` state.

```bash
$ kubectl get sl -n demo
NAME            TYPE                  VERSION   STATUS   AGE
operator-prom-sl   kubedb.com/v1alpha2   9.6.1     Ready    104m
```

KubeDB will create a separate stats service with name `{Solr crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo -l 'app.kubernetes.io/instance=operator-prom-sl'
NAME                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
operator-prom-sl         ClusterIP   10.96.76.207   <none>        8983/TCP   122m
operator-prom-sl-pods    ClusterIP   None           <none>        8983/TCP   122m
operator-prom-sl-stats   ClusterIP   10.96.192.50   <none>        9854/TCP   122m
```

Here, `operator-prom-sl-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```bash
$ kubectl describe svc -n demo operator-prom-sl-stats
Name:              operator-prom-sl-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
  app.kubernetes.io/instance=operator-prom-sl
  app.kubernetes.io/managed-by=kubedb.com
  app.kubernetes.io/name=solrs.kubedb.com
  kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/instance=operator-prom-sl,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=solrs.kubedb.com
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                10.96.192.50
IPs:               10.96.192.50
Port:              metrics  9854/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.0.37:9854,10.244.0.39:9854
Session Affinity:  None
Events:            <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use these information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `monitoring` namespace that select the endpoints of `coreos-prom-es-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$ kubectl get servicemonitor -n demo 
NAME                  AGE
operator-prom-sl-stats   125m
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of Elasticsearch crd.

```bash
$ kubectl get servicemonitor -n demo operator-prom-sl-stats -oyaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2024-10-29T06:44:10Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: operator-prom-sl
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: solrs.kubedb.com
    release: prometheus
  name: operator-prom-sl-stats
  namespace: demo
  ownerReferences:
    - apiVersion: v1
      blockOwnerDeletion: true
      controller: true
      kind: Service
      name: operator-prom-sl-stats
      uid: 6728f31e-8840-45ff-b6bd-f9c5839c74a6
  resourceVersion: "632313"
  uid: 74194ab9-d148-4d84-b973-764606f5f7b6
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
      app.kubernetes.io/instance: operator-prom-sl
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: solrs.kubedb.com
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in Solr crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `operator-prom-sl-stats` service. It also, target the `prom-http` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=release=prometheus
NAME                                                   READY   STATUS    RESTARTS         AGE
prometheus-kube-prometheus-operator-6c8698f59d-cljvq   1/1     Running   12 (4h11m ago)   12d
prometheus-kube-state-metrics-5548456c74-ksh5n         1/1     Running   13 (4h10m ago)   12d
prometheus-prometheus-node-exporter-n5ht8              1/1     Running   9 (4h11m ago)    12d
```

Prometheus server is listening to port `9090` of `prometheus-kube-prometheus-operator-6c8698f59d-cljvq` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-kube-prometheus-operator-6c8698f59d-cljvq 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `prom-http` endpoint of `operator-prom-sl-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/solr/monitoring/solr-operator-prom-target.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels marked by red rectangle. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboard with collected metrics.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run following commands

```bash
# cleanup database
kubectl delete -n demo sl/operator-prom-sl

# cleanup prometheus resources
kubectl delete -n monitoring prometheus prometheus
kubectl delete -n monitoring clusterrolebinding prometheus
kubectl delete -n monitoring clusterrole prometheus
kubectl delete -n monitoring serviceaccount prometheus
kubectl delete -n monitoring service prometheus-operated

# cleanup prometheus operator resources
kubectl delete -n monitoring deployment prometheus-operator
kubectl delete -n dmeo serviceaccount prometheus-operator
kubectl delete clusterrolebinding prometheus-operator
kubectl delete clusterrole prometheus-operator

# delete namespace
kubectl delete ns monitoring
kubectl delete ns demo
```
