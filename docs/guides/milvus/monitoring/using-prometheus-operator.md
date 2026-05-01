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
      name: my-release-minio
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
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
- `spec.monitor.agent: prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.
- `spec.monitor.prometheus.serviceMonitor.interval` indicates that the Prometheus should scrape metrics from this database with 10 seconds interval.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/milvus/monitoring/coreos-prom-milvus.yaml
milvus.kubedb.com/coreos-prom-milvus created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get milvus -n demo coreos-prom-milvus
NAME                 VERSION   STATUS   AGE
coreos-prom-milvus   2.4.0     Ready    2m
```

KubeDB will create a `ServiceMonitor` CRD in the same namespace as the Milvus database.

```bash
$ kubectl get servicemonitor -n demo
NAME                 AGE
coreos-prom-milvus   2m
```

Let's verify the `ServiceMonitor` has the right label to be discovered by Prometheus:

```bash
$ kubectl get servicemonitor -n demo coreos-prom-milvus -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    release: prometheus
  name: coreos-prom-milvus
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

Now, if we go to the Prometheus dashboard, we should see the target being scraped.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo milvus/coreos-prom-milvus -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo milvus/coreos-prom-milvus

kubectl delete ns demo
kubectl delete ns monitoring
```
