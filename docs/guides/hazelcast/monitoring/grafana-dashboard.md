---
title: Visualize Hazelcast Metrics with Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: guides-hz-grafana-dashboard
    name: Grafana Dashboard
    parent: hz-monitoring-hazelcast
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Visualize Hazelcast Metrics with Grafana Dashboard

KubeDB exposes Hazelcast metrics through a sidecar exporter. Once Prometheus scrapes those metrics, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a Hazelcast instance, and importing the Grafana dashboards.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (Step 2) uses to expose metrics to Prometheus.

- Hazelcast requires a valid license secret. Create the license secret before deploying:

  ```bash
  $ kubectl create secret generic hz-license-key -n demo \
    --from-literal=license=<your-hazelcast-license-key>
  ```

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/hazelcast/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hazelcast/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the Hazelcast instance.

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

## Step 3: Deploy Hazelcast with Monitoring Enabled

Below is the Hazelcast object with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: hz-grafana-demo
  namespace: demo
spec:
  version: "5.5.2"
  replicas: 3
  licenseSecret:
    name: hz-license-key
  deletionPolicy: WipeOut
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
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

Create the Hazelcast instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/monitoring/hz-grafana-demo.yaml
hazelcast.kubedb.com/hz-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get hazelcast -n demo hz-grafana-demo
NAME              VERSION   STATUS   AGE
hz-grafana-demo   5.5.2     Ready    3m
```

KubeDB creates a stats service named `{hazelcast-name}-stats` for the exporter:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=hz-grafana-demo"
NAME                   TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)     AGE
hz-grafana-demo        ClusterIP   10.96.10.1    <none>        5701/TCP    3m
hz-grafana-demo-stats  ClusterIP   10.96.10.2    <none>        56790/TCP   3m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                   AGE
hz-grafana-demo-stats  3m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo hz-grafana-demo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `hz-grafana-demo-stats`. Its state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/hazelcast/monitoring/hz-prom-targets.png" style="padding:10px">
</p>

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

| Field    | Value                       |
|----------|-----------------------------|
| Username | `admin`                     |
| Password | output of the command above |

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

## Step 7: Import KubeDB Hazelcast Dashboards

The KubeDB Hazelcast dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download all three JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/hazelcast) repository (`hazelcast/` folder):

| File | Dashboard |
|------|-----------|
| `hazelcast_summary_dashboard.json` | KubeDB / Hazelcast / Summary |
| `hazelcast_pods_dashboard.json` | KubeDB / Hazelcast / Pod |
| `hazelcast_databases_dashboard.json` | KubeDB / Hazelcast / Database |

**Import steps (repeat for each of the three files):**

1. In Grafana, click **Dashboards** in the left sidebar.
2. Select **Import** from the menu.
3. Click **Upload dashboard JSON file** and select one of the downloaded `.json` files.
4. In the **Prometheus** dropdown that appears, select your Prometheus data source.
5. Click **Import**.

After importing all three files, they will appear under **Dashboards** in the left sidebar.

## Step 8: Explore the Dashboards

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable      | Applies to     | What to select                                               |
|---------------|----------------|--------------------------------------------------------------|
| **namespace** | All dashboards | Namespace where your Hazelcast is deployed (e.g., `demo`)   |
| **app**       | All dashboards | Name of your instance (e.g., `hz-grafana-demo`)             |
| **pod**       | Pod dashboard  | A specific pod, or `All` for an aggregated view             |

### KubeDB / Hazelcast / Summary

Cluster-level overview showing general health, resource usage, storage, and network metrics.

**General Info section:**
- **Database Status** — current health (`Ready`, `NotReady`, etc.)
- **Database Up-time** — how long the cluster has been running
- **Version** — Hazelcast version (e.g., `5.5.2`)
- **Total Nodes** — number of active member pods
- **CPU / Memory Request & Limit** — configured resource bounds per pod
- **Deletion Policy** — KubeDB deletion policy (e.g., `WipeOut`)

**CPU Info section:**
- **CPU Usage** — per-pod CPU usage over time
- **CPU Quota** — table of CPU requests, limits, and usage percentages per pod

<p align="center">
  <img alt="KubeDB Hazelcast Summary Dashboard - General Info and CPU" src="/docs/images/hazelcast/monitoring/hz-grafana-summary.png" style="padding:10px">
</p>

**Memory Info section:**
- **Memory Usage** — RSS memory per pod over time
- **Memory Quota** — table of memory requests, limits, and usage percentages per pod

**Storage Info section:**
- **Disk Usage** — per-pod disk write growth over time
- **Disk R/W Info** — read/write throughput per pod

<p align="center">
  <img alt="KubeDB Hazelcast Summary Dashboard - Memory and Storage" src="/docs/images/hazelcast/monitoring/hz-grafana-summary-2.png" style="padding:10px">
</p>

**Network Info section:**
- **Persistent Volume Usage History** — PV capacity vs used over time
- **Receive Bandwidth** — inbound network traffic per pod
- **Transmit Bandwidth** — outbound network traffic per pod

<p align="center">
  <img alt="KubeDB Hazelcast Summary Dashboard - PV and Network" src="/docs/images/hazelcast/monitoring/hz-grafana-summary-3.png" style="padding:10px">
</p>

### KubeDB / Hazelcast / Pod

Per-member drill-down. Use the **pod** dropdown to select a specific `hz-grafana-demo-N` pod.

**Hazelcast Memory Consumption section:**
- **Partition Counts** — number of partitions owned by this member
- **System Load** — CPU system load average over time
- **Heap Used / Max** — JVM heap in use vs. configured maximum
- **Physical Usage / Max** — OS physical memory used vs. total
- **Non-Heap Used / Max** — JVM non-heap (metaspace, code cache) used vs. max
- **Free Physical Memory** — OS free memory over time
- **Memory Usage** — owned, backup, and heap-cost entry memory per map

<p align="center">
  <img alt="KubeDB Hazelcast Pod Dashboard - Memory Consumption" src="/docs/images/hazelcast/monitoring/hz-grafana-pod.png" style="padding:10px">
</p>

**Hazelcast Operate Rate PerMinute section:**
- **Put Rate Per Minute** — map put operations per minute on this pod
- **Get Rate Per Minute** — map get operations per minute on this pod
- **Average Get Times** — mean latency for get operations
- **Remove Rate Per Minute** — map remove operations per minute on this pod

<p align="center">
  <img alt="KubeDB Hazelcast Pod Dashboard - Operate Rate" src="/docs/images/hazelcast/monitoring/hz-grafana-pod-2.png" style="padding:10px">
</p>

### KubeDB / Hazelcast / Database

Cluster-wide map and memory metrics across all members.

**Hazelcast Server section:**
- **Active Members** — total cluster member count
- **Cluster Version** — Hazelcast cluster protocol version
- **Client Connections** — number of connected Hazelcast clients
- **Partition Counts** — gauge per pod showing owned partitions

**Hazelcast Memory Usage section:**
- **System Load** — system load over time across members
- **Heap Usage** — JVM heap usage gauge per member pod
- **Owned Entry Memory Cost** — heap memory consumed by owned map entries
- **Backup Entry Memory Cost** — heap memory consumed by backup entries

<p align="center">
  <img alt="KubeDB Hazelcast Database Dashboard - Server Stats and Memory" src="/docs/images/hazelcast/monitoring/hz-grafana-database.png" style="padding:10px">
</p>

**Hazelcast Operation Latency section:**
- **Heap Cost** — total heap cost per map across all members
- **GetLatency and GetCount** — get operation latency and throughput
- **PutLatency and PutCount** — put operation latency and throughput

<p align="center">
  <img alt="KubeDB Hazelcast Database Dashboard - Operation Latency" src="/docs/images/hazelcast/monitoring/hz-grafana-database-2.png" style="padding:10px">
</p>

## Cleaning up

```bash
# Remove the Hazelcast instance
kubectl delete hazelcast -n demo hz-grafana-demo

# Remove the license secret
kubectl delete secret hz-license-key -n demo

# Remove namespaces
kubectl delete ns demo

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your Hazelcast database with KubeDB using [built-in Prometheus](/docs/guides/hazelcast/monitoring/prometheus-builtin.md).
- Monitor your Hazelcast database with KubeDB using [Prometheus Operator](/docs/guides/hazelcast/monitoring/prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
