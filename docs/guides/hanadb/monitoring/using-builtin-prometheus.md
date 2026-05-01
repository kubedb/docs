---
title: Monitor HanaDB using Builtin Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: hanadb-using-builtin-prometheus-monitoring
    name: Builtin Prometheus
    parent: hanadb-monitoring
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring HanaDB with Builtin Prometheus

This tutorial will show you how to monitor HanaDB database using builtin [Prometheus](https://github.com/prometheus/prometheus) scraper.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- If you are not familiar with how to configure Prometheus to scrape metrics from various Kubernetes resources, please read the tutorial from [here](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/hanadb/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy the database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy HanaDB with Monitoring Enabled

At first, let's deploy a HanaDB database with monitoring enabled. Below is the HanaDB object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: builtin-prom-hanadb
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/builtin
```

Here, `spec.monitor.agent: prometheus.io/builtin` specifies that we are going to monitor this server using builtin Prometheus scraper.

Let's create the HanaDB CR we have shown above.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/monitoring/builtin-prom-hanadb.yaml
hanadb.kubedb.com/builtin-prom-hanadb created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get hanadb -n demo builtin-prom-hanadb
NAME                  VERSION   STATUS   AGE
builtin-prom-hanadb   2.0       Ready    2m
```

KubeDB will create a separate stats service with name `{HanaDB crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=builtin-prom-hanadb"
NAME                         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
builtin-prom-hanadb          ClusterIP   10.96.100.10    <none>        39017/TCP   2m
builtin-prom-hanadb-stats    ClusterIP   10.96.100.11    <none>        56790/TCP   90s
```

Here, `builtin-prom-hanadb-stats` service has been created for monitoring purpose. Let's describe the service.

```bash
$ kubectl describe svc -n demo builtin-prom-hanadb-stats
Name:              builtin-prom-hanadb-stats
Namespace:         demo
Labels:            app.kubernetes.io/name=hanadbs.kubedb.com
                   app.kubernetes.io/instance=builtin-prom-hanadb
Annotations:       monitoring.appscode.com/agent: prometheus.io/builtin
                   prometheus.io/path: /metrics
                   prometheus.io/port: 56790
                   prometheus.io/scrape: true
Selector:          app.kubernetes.io/name=hanadbs.kubedb.com,app.kubernetes.io/instance=builtin-prom-hanadb
Type:              ClusterIP
Port:              prom-http  56790/TCP
```

You can see that the service contains the following annotations, which are used by builtin Prometheus to discover the endpoint:

```
prometheus.io/path: /metrics
prometheus.io/port: 56790
prometheus.io/scrape: true
```

## Configure Prometheus

Now we need to configure Prometheus to scrape metrics from this service. Add the following `scrape_config` to your Prometheus configuration:

```yaml
scrape_configs:
- job_name: kubedb-hanadbs
  honor_labels: true
  kubernetes_sd_configs:
  - role: endpoints
  relabel_configs:
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
    separator: ;
    regex: true
    target_label: __tmp_prometheus_service_scrape
    replacement: $1
    action: keep
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
    separator: ;
    regex: (https?)
    target_label: __scheme__
    replacement: $1
    action: replace
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
    separator: ;
    regex: (.+)
    target_label: __metrics_path__
    replacement: $1
    action: replace
  - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
    separator: ;
    regex: ([^:]+)(?::\d+)?;(\d+)
    target_label: __address__
    replacement: $1:$2
    action: replace
```

Now Prometheus will discover the HanaDB stats service and scrape metrics automatically.

## Access Prometheus Dashboard

To access the Prometheus dashboard, port-forward the Prometheus service and visit `http://localhost:9090` in your browser.

```bash
$ kubectl port-forward -n monitoring svc/prometheus 9090:9090
```

You should see the HanaDB metrics in the Prometheus dashboard under the `kubedb-hanadbs` job.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/builtin-prom-hanadb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/builtin-prom-hanadb

kubectl delete ns demo
kubectl delete ns monitoring
```
