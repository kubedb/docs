---
title: Visualize Redis Metrics with Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: rd-grafana-dashboard-monitoring
    name: Grafana Dashboard
    parent: rd-monitoring-redis
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Visualize Redis Metrics with Grafana Dashboard

KubeDB exposes Redis metrics through a sidecar exporter. Once Prometheus scrapes those metrics, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a Redis instance, and importing the Grafana dashboards.

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

> Note: YAML files used in this tutorial are stored in [docs/examples/redis/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Configuration

> These two steps — deploying `kube-prometheus-stack` and installing Panopticon — are shared prerequisites for all KubeDB database monitoring guides. If you have already completed them in another guide, skip to [Step 1](#step-1-deploy-redis-with-monitoring-enabled).

### Step 1: Deploy kube-prometheus-stack

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

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the Redis instance.

```bash
$ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
{"matchLabels":{"release":"prometheus"}}
```

The label is `release: prometheus`.

### Step 2: Install Panopticon

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

## Setup

## Step 1: Deploy Redis with Monitoring Enabled

Below is the Redis object with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-cluster
  namespace: demo
spec:
  version: 8.2.2
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
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

- `monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` for this instance.
- `monitor.prometheus.serviceMonitor.labels` must match the `serviceMonitorSelector` label of your Prometheus (`release: prometheus`).
- `monitor.prometheus.serviceMonitor.interval` sets the scrape interval to 10 seconds.

Create the Redis instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/monitoring/redis-cluster.yaml
redis.kubedb.com/redis-cluster created
```

Wait for it to be `Ready`:

```bash
$ kubectl get redis -n demo redis-cluster
NAME            VERSION   STATUS   AGE
redis-cluster   8.2.2     Ready    5m
```

KubeDB creates a stats service named `{redis-name}-stats` for the exporter:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=redis-cluster"
NAME                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
redis-cluster         ClusterIP   10.96.10.1     <none>        6379/TCP    5m
redis-cluster-stats   ClusterIP   10.96.10.2     <none>        56790/TCP   5m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                  AGE
redis-cluster-stats   5m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo redis-cluster-stats -o jsonpath='{.metadata.labels}'
{"release":"prometheus", ...}
```

## Step 2: Verify Prometheus is Scraping

Port-forward the Prometheus pod:

```bash
$ kubectl port-forward -n monitoring \
  prometheus-prometheus-kube-prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `redis-cluster-stats`. Its state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/redis/monitoring/rd-prom-targets.png" style="padding:10px">
</p>

If the target is missing, check that the `ServiceMonitor` label (`release: prometheus`) matches the Prometheus `serviceMonitorSelector`.

## Step 3: Access Grafana

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

<p align="center">
  <img alt="Grafana Login" src="/docs/images/redis/monitoring/rd-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/redis/monitoring/rd-grafana-home.png" style="padding:10px">
</p>

## Step 4: Configure Prometheus as a Data Source

If you installed Grafana via `kube-prometheus-stack`, Prometheus is already configured as the default data source — skip to Step 5.

For a standalone Grafana installation:

1. Go to **Connections** → **Data sources** → **Add new data source**.
2. Select **Prometheus**.
3. Set the URL to your Prometheus service:

   ```
   http://prometheus-operated.monitoring.svc:9090
   ```

4. Click **Save & test**. You should see `Data source is working`.

## Step 5: Import KubeDB Redis Dashboard

The KubeDB Redis dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download the JSON files from the [opnpulse/dashboards](https://github.com/opnpulse/dashboards/tree/master/redis) repository (`redis/` folder):

| File | Dashboard |
|------|-----------|
| `redis_summary_dashboard.json` | KubeDB / Redis / Summary |
| `redis_pod_dashboard.json` | KubeDB / Redis / Pod |
| `redis_shards_dashboard.json` | KubeDB / Redis / Shard |

> The Shard dashboard is relevant for Redis Cluster mode (`spec.mode: Cluster`); its panels stay empty for a standalone (non-cluster) Redis instance.

**Import steps (repeat for each file you need):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

The import page looks like this — click **Upload dashboard JSON file** to select the file:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/redis/monitoring/rd-grafana-import.png" style="padding:10px">
</p>

After importing the files you need, they will appear under `Dashboards` in the left sidebar.

| Dashboard Name | Description |
|---|---|
| KubeDB / Redis / Summary | Instance overview: status, version, mode, node count, resource requests/limits, CPU usage |
| KubeDB / Redis / Pod | Per-pod role, master/slaves, connected clients, memory, commands/sec, network I/O, CPU/memory |
| KubeDB / Redis / Shard | Cluster shard slot health, node/slave count, per-slave status, cluster mode |

## Step 6: Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable       | Applies to              | What to select                                             |
|----------------|--------------------------|--------------------------------------------------------------|
| **datasource** | All dashboards          | Your Prometheus data source                                |
| **Namespace**  | All dashboards          | Namespace where your Redis is deployed (e.g., `demo`)      |
| **app**        | Summary dashboard       | Name of your Redis instance (e.g., `redis-cluster`)        |
| **redis**      | Pod, Shard dashboards   | Name of your Redis instance (e.g., `redis-cluster`)        |
| **Pod Name**   | Pod, Shard dashboards   | A specific pod (e.g., `redis-cluster-shard0-0`)            |
| **Filters**    | Shard dashboard         | Additional label filters for the selected shard             |

**KubeDB / Redis / Summary** — start here for an instance overview:
- **General Info** — database status, version, max clients, Redis mode, deletion policy, total nodes
- **Resource Requests / Limits** — configured CPU, memory, and storage requests and limits
- **CPU Info / CPU Quota** — per-pod CPU usage over time and quota utilization

<p align="center">
  <img alt="KubeDB Redis Summary Dashboard" src="/docs/images/redis/monitoring/rd-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / Redis / Pod** — drill into a specific pod:
- **General Counters And File Descriptor Stats** — status, role (master/slave), my master, my slaves, connected clients, Go routines
- **Uptime / Memory Usage / Commands Executed / Hits-Misses** — pod uptime, memory usage, command execution rate, cache hit/miss rate
- **Network I/O / Command Calls / Connected Clients** — network throughput, per-command call breakdown, connected client count over time
- **CPU And Memory Usage Stats** — total memory usage, average CPU usage, average memory usage

<p align="center">
  <img alt="KubeDB Redis Pod Dashboard" src="/docs/images/redis/monitoring/rd-grafana-pod.png" style="padding:10px">
</p>

**KubeDB / Redis / Shard** — cluster shard health for Cluster mode:
- **Cluster Shard Slots / Cluster Shard Slots Failed** — hash slot coverage and any failed slots
- **Cluster Nodes / Cluster Masters** — total nodes and master count in the cluster
- **Connected Slaves / My Slaves** — number of connected slaves and their IP, port, and online status
- **Mode** — confirms the instance is running in `cluster` mode

<p align="center">
  <img alt="KubeDB Redis Shard Dashboard" src="/docs/images/redis/monitoring/rd-grafana-shard.png" style="padding:10px">
</p>

## Cleaning up

```bash
# Remove the Redis instance
kubectl delete redis -n demo redis-cluster

# Remove namespaces
kubectl delete ns demo

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your Redis database with KubeDB using [built-in Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Monitor your Redis database with KubeDB using [Prometheus Operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
