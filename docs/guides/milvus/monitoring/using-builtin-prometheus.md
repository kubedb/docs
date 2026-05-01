---
title: Monitor Milvus using Builtin Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: milvus-using-builtin-prometheus-monitoring
    name: Builtin Prometheus
    parent: milvus-monitoring
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Milvus with Builtin Prometheus

This tutorial will show you how to monitor Milvus database using builtin [Prometheus](https://github.com/prometheus/prometheus) scraper.

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
  name: builtin-prom-milvus
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
    agent: prometheus.io/builtin
```

Here, `spec.monitor.agent: prometheus.io/builtin` specifies that we are going to monitor this server using builtin Prometheus scraper.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/milvus/monitoring/builtin-prom-milvus.yaml
milvus.kubedb.com/builtin-prom-milvus created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get milvus -n demo builtin-prom-milvus
NAME                  VERSION   STATUS   AGE
builtin-prom-milvus   2.4.0     Ready    2m
```

KubeDB will create a separate stats service with name `{Milvus crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=builtin-prom-milvus"
NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)      AGE
builtin-prom-milvus           ClusterIP   10.96.100.20   <none>        19530/TCP    2m
builtin-prom-milvus-stats     ClusterIP   10.96.100.21   <none>        9091/TCP     90s
```

Let's describe the stats service:

```bash
$ kubectl describe svc -n demo builtin-prom-milvus-stats
Name:              builtin-prom-milvus-stats
Namespace:         demo
Labels:            app.kubernetes.io/name=milvuses.kubedb.com
                   app.kubernetes.io/instance=builtin-prom-milvus
Annotations:       monitoring.appscode.com/agent: prometheus.io/builtin
                   prometheus.io/path: /metrics
                   prometheus.io/port: 9091
                   prometheus.io/scrape: true
```

The service contains the following annotations which are used by builtin Prometheus to discover the endpoint:

```
prometheus.io/path: /metrics
prometheus.io/port: 9091
prometheus.io/scrape: true
```

Configure your Prometheus to scrape metrics from the `monitoring` namespace service discovery:

```yaml
scrape_configs:
- job_name: kubedb-milvuses
  honor_labels: true
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - demo
  relabel_configs:
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
    action: keep
    regex: true
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
    action: replace
    target_label: __metrics_path__
    regex: (.+)
  - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
    action: replace
    target_label: __address__
    regex: ([^:]+)(?::\d+)?;(\d+)
    replacement: $1:$2
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo milvus/builtin-prom-milvus -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo milvus/builtin-prom-milvus

kubectl delete ns demo
kubectl delete ns monitoring
```
