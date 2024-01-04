---
title: Monitor Memcached using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: mc-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: mc-monitoring-memcached
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Memcached Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor Memcached server deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/memcached/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md).

- If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

> Note: YAML files used in this tutorial are stored in [docs/examples/memcached](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/memcached) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.labels` field of Memcached crd so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
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
      {"apiVersion":"monitoring.coreos.com/v1","kind":"Prometheus","metadata":{"annotations":{},"labels":{"prometheus":"prometheus"},"name":"prometheus","namespace":"monitoring"},"spec":{"replicas":1,"resources":{"requests":{"memory":"400Mi"}},"serviceAccountName":"prometheus","serviceMonitorSelector":{"matchLabels":{"release":"prometheus"}}}}
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
      release: prometheus
```

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` crd. So, we are going to use this label in `spec.monitor.prometheus.labels` field of Memcached crd.

## Deploy Memcached with Monitoring Enabled

At first, let's deploy an Memcached server with monitoring enabled. Below is the Memcached object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Memcached
metadata:
  name: coreos-prom-memcd
  namespace: demo
spec:
  replicas: 3
  version: "1.6.22"
  terminationPolicy: WipeOut
  podTemplate:
    spec:
      resources:
        limits:
          cpu: 500m
          memory: 128Mi
        requests:
          cpu: 250m
          memory: 64Mi
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

Let's create the Memcached object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/monitoring/coreos-prom-memcd.yaml
memcached.kubedb.com/coreos-prom-memcd created
```

Now, wait for the database to go into `Running` state.

```bash
$ kubectl get mc -n demo coreos-prom-memcd
NAME                VERSION    STATUS    AGE
coreos-prom-memcd   1.6.22   Running   19s
```

KubeDB will create a separate stats service with name `{Memcached crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=coreos-prom-memcd"
NAME                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
coreos-prom-memcd         ClusterIP   10.100.207.76   <none>        11211/TCP   41s
coreos-prom-memcd-stats   ClusterIP   10.97.230.149   <none>        56790/TCP   38s
```

Here, `coreos-prom-memcd-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```yaml
$ kubectl describe svc -n demo coreos-prom-memcd-stats
Name:              coreos-prom-memcd-stats
Namespace:         demo
Labels:            app.kubernetes.io/name=memcacheds.kubedb.com
                   app.kubernetes.io/instance=coreos-prom-memcd
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/name=memcacheds.kubedb.com,app.kubernetes.io/instance=coreos-prom-memcd
Type:              ClusterIP
IP:                10.97.230.149
Port:              prom-http  56790/TCP
TargetPort:        prom-http/TCP
Endpoints:         172.17.0.7:56790,172.17.0.8:56790,172.17.0.9:56790
Session Affinity:  None
Events:            <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use these information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `monitoring` namespace that select the endpoints of `coreos-prom-memcd-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$ kubectl get servicemonitor -n monitoring
NAME                            AGE
kubedb-demo-coreos-prom-memcd   1m
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of Memcached crd.

```yaml
$ kubectl get servicemonitor -n monitoring kubedb-demo-coreos-prom-memcd -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: 2019-01-03T15:13:46Z
  generation: 1
  labels:
    release: prometheus
    monitoring.appscode.com/service: coreos-prom-memcd-stats.demo
  name: kubedb-demo-coreos-prom-memcd
  namespace: monitoring
  resourceVersion: "51236"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/servicemonitors/kubedb-demo-coreos-prom-memcd
  uid: 2aa57b5a-0f6a-11e9-b230-080027f306f3
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
      app.kubernetes.io/name: memcacheds.kubedb.com
      app.kubernetes.io/instance: coreos-prom-memcd
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in Memcached crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `coreos-prom-memcd-stats` service. It also, target the `prom-http` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                      READY   STATUS    RESTARTS   AGE
prometheus-prometheus-0   3/3     Running   1          63m
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `prom-http` endpoint of `coreos-prom-memcd-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/memcached/monitoring/mc-coreos-prom-target.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels marked by red rectangle. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboard with collected metrics.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run following commands

```bash
# cleanup database
kubectl delete -n demo mc/coreos-prom-memcd

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

## Next Steps

- Monitor your Memcached server with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/memcached/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Memcached object](/docs/guides/memcached/concepts/memcached.md).
- Use [private Docker registry](/docs/guides/memcached/private-registry/using-private-registry.md) to deploy Memcached with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
