---
title: Visualize Kafka Metrics with Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: kf-grafana-dashboard-monitoring
    name: Grafana Dashboard
    parent: kf-monitoring-kafka
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Visualize Kafka Metrics with Grafana Dashboard

KubeDB exposes Kafka metrics through a JMX Exporter running as a Java agent inside each Kafka container. Once Prometheus scrapes those metrics, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a Kafka instance, and importing the Grafana dashboards.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (Step 2) uses to expose metrics to Prometheus.

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/kafka/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Step 1: Deploy kube-prometheus-stack

`kube-prometheus-stack` installs Prometheus, Prometheus Operator, Alertmanager, and Grafana together. This is the recommended way to get the full monitoring stack on Kubernetes.

Add the prometheus-community Helm repo and install:

```bash
$ helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
$ helm repo update

$ helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --set grafana.image.tag=7.5.5
```

Wait for all pods to be ready:

```bash
$ kubectl get pods -n monitoring
NAME                                                   READY   STATUS    RESTARTS   AGE
alertmanager-prometheus-kube-prometheus-alertmanager-0 2/2     Running   0          2m
prometheus-grafana-xxxx                                3/3     Running   0          2m
prometheus-kube-prometheus-operator-xxxx               1/1     Running   0          2m
prometheus-kube-prometheus-prometheus-0                2/2     Running   0          2m
prometheus-kube-state-metrics-xxxx                     1/1     Running   0          2m
```

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the Kafka instance.

```bash
$ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
{"matchLabels":{"release":"prometheus"}}
```

The label is `release: prometheus`.

## Step 2: Install Panopticon

Panopticon is the Appscode operator that reads `MetricsConfiguration` objects created by `kubedb-metrics` and exposes them to Prometheus. It must be installed before enabling `kubedb-metrics`.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm upgrade --install panopticon appscode/panopticon \
  --version v2026.4.30 \
  --namespace kubeops --create-namespace \
  --set monitoring.enabled=true \
  --set monitoring.agent=prometheus.io/operator \
  --set monitoring.serviceMonitor.labels.release=prometheus \
  --set-file license=/path/to/kubedb-license.txt \
  --wait --timeout 5m0s
```

Verify panopticon is running:

```bash
$ kubectl get pods -n kubeops
NAME                          READY   STATUS    RESTARTS   AGE
panopticon-xxxx               1/1     Running   0          1m
```

## Step 3: Deploy Kafka with Monitoring Enabled

Below is the Kafka object with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-grafana-demo
  namespace: demo
spec:
  version: "3.9.0"
  replicas: 1
  deletionPolicy: WipeOut
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here,

- `monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` for this instance.
- `monitor.prometheus.serviceMonitor.labels` must match the `serviceMonitorSelector` label of your Prometheus (`release: prometheus`).
- `monitor.prometheus.serviceMonitor.interval` sets the scrape interval to 10 seconds.

Create the Kafka instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/monitoring/kafka-grafana-demo.yaml
kafka.kubedb.com/kafka-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get kafka -n demo kafka-grafana-demo
NAME                 VERSION   STATUS   AGE
kafka-grafana-demo   3.9.0     Ready    3m
```

KubeDB creates a stats service named `{kafka-name}-stats` for monitoring:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=kafka-grafana-demo"
NAME                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
kafka-grafana-demo         ClusterIP   10.96.10.1     <none>        9092/TCP    3m
kafka-grafana-demo-stats   ClusterIP   10.96.10.2     <none>        9101/TCP    3m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                       AGE
kafka-grafana-demo-stats   3m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo kafka-grafana-demo-stats -o jsonpath='{.metadata.labels}'
{"release":"prometheus", ...}
```

## Step 4: Verify Prometheus is Scraping

Port-forward the Prometheus pod:

```bash
$ kubectl port-forward -n monitoring \
  prometheus-prometheus-kube-prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `kafka-grafana-demo-stats`. Its state should be **UP**.

If the target is missing, check that the `ServiceMonitor` label (`release: prometheus`) matches the Prometheus `serviceMonitorSelector`.

## Step 5: Access Grafana

Port-forward the Grafana service:

```bash
$ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
Forwarding from 127.0.0.1:3000 -> 80
```

Open [http://localhost:3000](http://localhost:3000). The username is `admin`. Retrieve the auto-generated password from the secret:

```bash
$ kubectl get secret -n monitoring prometheus-grafana \
  -o jsonpath='{.data.admin-password}' | base64 -d
```

## Step 6: Configure Prometheus as a Data Source

If you installed Grafana via `kube-prometheus-stack`, Prometheus is already configured as the default data source — skip to Step 7.

For a standalone Grafana installation:

1. Go to **Connections** → **Data sources** → **Add new data source**.
2. Select **Prometheus**.
3. Set the URL to your Prometheus service:

   ```
   http://prometheus-operated.monitoring.svc:9090
   ```

4. Click **Save & test**. You should see `Data source is working`.

## Step 7: Import KubeDB Kafka Dashboard

The KubeDB Kafka dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Six dashboards are available. Download the JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/kafka) repository (`kafka/` folder):

| File | Dashboard |
|------|-----------|
| `kafka_summary_dashboard.json` | KubeDB / Kafka / Summary |
| `kafka_pods_dashboard.json` | KubeDB / Kafka / Pod |
| `kafka_databases_dashboard.json` | KubeDB / Kafka / Database |
| `kafka_connectcluster_summary_dashboard.json` | KubeDB / Kafka / ConnectCluster Summary |
| `kafka_connectcluster_pods_dashboard.json` | KubeDB / Kafka / ConnectCluster Pod |
| `kafka_connectcluster_connect_dashboard.json` | KubeDB / Kafka / ConnectCluster Connect |

> The ConnectCluster dashboards (last three) are only relevant if you deploy a `ConnectCluster` resource alongside your Kafka cluster. The core three dashboards (Summary, Pod, Database) cover the base Kafka cluster.

**Import steps (repeat for each file you need):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

## Step 8: Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable      | Applies to              | What to select                                           |
|---------------|-------------------------|----------------------------------------------------------|
| **namespace** | All dashboards          | Namespace where your Kafka is deployed (e.g., `demo`)   |
| **app**       | All dashboards          | Name of your Kafka instance (e.g., `kafka-grafana-demo`) |
| **pod**       | Pod dashboards          | A specific pod, or `All` for an aggregated view         |
| **topic**     | Database dashboard      | A specific Kafka topic, or `All`                        |

**KubeDB / Kafka / Summary** — start here for a broker-level overview:
- **Broker Count** — number of active brokers
- **Under-Replicated Partitions** — partitions with fewer in-sync replicas than configured (non-zero means degraded replication)
- **Offline Partitions** — partitions with no leader (non-zero means data unavailability)
- **Active Controller** — exactly one broker should be controller at all times
- **Message Rate** — messages in/out per second
- **Bytes Rate** — bytes in/out per second across all topics
- **CPU / Memory / JVM Heap** — resource consumption per broker

**KubeDB / Kafka / Pod** — drill into a specific broker:
- **JVM Heap Used** — heap usage on this broker
- **GC Time** — time spent in garbage collection
- **Network Request Rate** — fetch and produce requests per second
- **Log Flush Rate** — frequency of log segment flushes to disk
- **CPU / Memory** — per-pod resource usage over time

**KubeDB / Kafka / Database** — topic-level metrics:
- **Produce / Fetch Rate** — per-topic throughput
- **Partition Count** — partitions per topic
- **Leader Election Rate** — frequency of leader elections (spikes indicate instability)
- **Consumer Group Lag** — messages pending consumption per group

**KubeDB / Kafka / ConnectCluster Summary** — connector fleet health:
- **Task Count** — total running tasks across all connectors
- **Failed Tasks** — tasks in failed state
- **Connector Count** — number of deployed connectors
- **Worker Rebalancing** — whether a worker rebalance is in progress

**KubeDB / Kafka / ConnectCluster Pod** — per-worker metrics:
- **CPU / Memory** — resource usage per worker pod
- **Task Throughput** — records processed per second

**KubeDB / Kafka / ConnectCluster Connect** — per-connector metrics:
- **Offset Lag** — how far behind the connector is from the source
- **Record Throughput** — records processed per second
- **Error Rate** — records skipped due to errors

## Cleaning up

```bash
# Remove the Kafka instance
kubectl delete kafka -n demo kafka-grafana-demo

# Remove namespaces
kubectl delete ns demo

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your Kafka instance with KubeDB using [built-in Prometheus](/docs/guides/kafka/monitoring/using-builtin-prometheus.md).
- Monitor your Kafka instance with KubeDB using [Prometheus Operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
