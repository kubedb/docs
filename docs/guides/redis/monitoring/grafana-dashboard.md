---
title: Redis Grafana Dashboard
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

# Redis Grafana Dashboard

KubeDB exposes Redis metrics through a sidecar exporter, and its own view of each resource (status, phase, version) through Panopticon. Once Prometheus is scraping both, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a Redis instance, and importing the Grafana dashboards.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (see [Configuration](/docs/guides/redis/monitoring/using-prometheus-operator.md#configuration)) uses to expose metrics to Prometheus.

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

* Before proceeding, complete the [Configuration](/docs/guides/redis/monitoring/using-prometheus-operator.md#configuration) steps to deploy **kube-prometheus-stack** and **Panopticon**.

> Note: YAML files used in this tutorial are stored in [docs/examples/redis/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Setup

## Step 1: Deploy Redis

Below is the Redis object with monitoring configured to use Prometheus Operator. This example deploys Redis in **Cluster mode** (3 shards, 2 replicas each) — the mode needed to populate the `KubeDB / Redis / Shard` dashboard's panels; a standalone Redis instance leaves that dashboard empty.

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

- `spec.mode: Cluster` with `spec.cluster.shards`/`spec.cluster.replicas` deploys a 3-shard Redis Cluster with 2 replicas per shard (9 pods total).
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

Each shard gets its own stats service named `{redis-name}-shard{N}-stats` for the exporter:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=redis-cluster"
NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
redis-cluster                 ClusterIP   10.96.10.1     <none>        6379/TCP    5m
redis-cluster-shard0-stats    ClusterIP   10.96.10.2     <none>        56790/TCP   5m
redis-cluster-shard1-stats    ClusterIP   10.96.10.3     <none>        56790/TCP   5m
redis-cluster-shard2-stats    ClusterIP   10.96.10.4     <none>        56790/TCP   5m
```

KubeDB also creates a `ServiceMonitor` per shard in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                          AGE
redis-cluster-shard0-stats    5m
redis-cluster-shard1-stats    5m
redis-cluster-shard2-stats    5m
```

Verify one carries the correct label:

```bash
$ kubectl get servicemonitor -n demo redis-cluster-shard0-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for entries whose `service` label matches `redis-cluster-shard0-stats`, `redis-cluster-shard1-stats`, and `redis-cluster-shard2-stats`. Their state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/redis/monitoring/rd-prom-targets.png" style="padding:10px">
</p>

If a target is missing, check that the `ServiceMonitor` label (`release: prometheus`) matches the Prometheus `serviceMonitorSelector`.

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

If you're using a different Grafana instance than the one installed in the Configuration prerequisite (linked in "Before You Begin" above), add Prometheus as a data source manually:

1. Go to **Connections** → **Data sources** → **Add new data source**.
2. Select **Prometheus**.
3. Set the URL to your Prometheus service:

   ```
   http://prometheus-operated.monitoring.svc:9090
   ```

4. Click **Save & test**. You should see `Data source is working`.

## Step 5: Import Dashboard — Option A: Automatic (chart)

Rather than downloading and uploading each JSON file by hand (Option B below), KubeDB ships a chart that creates all matching dashboards for you as `GrafanaDashboard` custom resources. A separate controller, **`grafana-operator`**, watches these resources and pushes the actual dashboard JSON into your Grafana instance — both pieces are required.

**1. Install `grafana-operator`** (skip if it's already running in your cluster):

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm upgrade --install grafana-operator appscode/grafana-operator \
    --version v2026.6.12 \
    --namespace kubeops --create-namespace
```

**2. Register your Grafana instance as an `AppBinding`** (skip if you've already done this in another guide on this cluster). `grafana-operator` needs to know where to push dashboards and how to authenticate — it reads this from an `AppBinding` object, not from the chart install command itself. Since Grafana here came bundled with `kube-prometheus-stack`, reuse its existing admin credentials:

```bash
$ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80 &

$ curl -s -X POST -H "Content-Type: application/json" \
    -u admin:<grafana_password> \
    http://localhost:3000/api/auth/keys \
    -d '{"name":"kubedb-dashboards","role":"Admin"}'
# Note the returned "key"

$ kill %1

$ kubectl create secret generic grafana-admin-token -n monitoring \
    --from-literal=token='<key-from-above>'

$ cat <<EOF | kubectl apply -f -
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: grafana
  namespace: monitoring
spec:
  type: monitoring.appscode.com/grafana
  clientConfig:
    url: http://prometheus-grafana.monitoring.svc:80
  secret:
    name: grafana-admin-token
EOF
```

**3. Install the dashboards:**

The chart packages dashboard JSON for every database it supports, so install it from a copy trimmed down to just Redis's dashboards (this also keeps `grafana-operator` from creating dashboards for databases you don't run):

```bash
$ helm pull appscode/kubedb-grafana-dashboards --version v2026.7.10 --untar

$ cd kubedb-grafana-dashboards/dashboards
$ ls | grep -v '^redis$' | xargs rm -rf   # keep only dashboards/redis (includes the RedisSentinel dashboards too)
$ cd ../..

$ helm package kubedb-grafana-dashboards
Successfully packaged chart and saved it to: kubedb-grafana-dashboards-v2026.7.10.tgz

$ helm upgrade -i kubedb-grafana-dashboards-redis ./kubedb-grafana-dashboards-v2026.7.10.tgz \
    -n kubeops --create-namespace \
    --set grafana.name=grafana \
    --set grafana.namespace=monitoring
```

`grafana.name`/`grafana.namespace` point the chart's `GrafanaDashboard` resources at the `AppBinding` created in step 2 (omitting them falls back to whichever `AppBinding` in your cluster is labeled as the cluster-default Grafana, if any — explicit is safer on a shared cluster). No need to touch `featureGates` — with every other database's `dashboards/` folder removed, their gates simply match zero files and render nothing, regardless of being `true` by default.

This creates every dashboard the chart ships for Redis — `KubeDB / Redis / Summary`, `KubeDB / Redis / Pod`, `KubeDB / Redis / Shard` (plus the `RedisSentinel / Pod` and `RedisSentinel / Summary` variants) — which `grafana-operator` then pushes into your Grafana instance automatically. No manual JSON download or upload needed.

Verify they landed:

```bash
$ kubectl get grafanadashboards.openviz.dev -n kubeops | grep -i redis
NAME                                    TITLE                              SYNCED    AGE
grafana-kubedb-redis-pod                KubeDB / Redis / Pod               Current   30s
grafana-kubedb-redis-shard               KubeDB / Redis / Shard              Current   30s
grafana-kubedb-redis-summary            KubeDB / Redis / Summary           Current   30s
grafana-kubedb-redissentinel-pod        KubeDB / RedisSentinel / Pod       Current   30s
grafana-kubedb-redissentinel-summary    KubeDB / RedisSentinel / Summary   Current   30s
```

> The `grafana-` prefix on each resource name comes from the `grafana.name=grafana` value set above (the chart prepends it to the dashboard title to build the resource name) — this is expected.

`SYNCED: Current` confirms `grafana-operator` successfully pushed each dashboard into Grafana. Open Grafana — the dashboards are already there under `Dashboards`, fully wired to your Prometheus data source, ready to explore in [Step 6](#step-6-explore-the-dashboard) below.

After importing them, they will appear under `Dashboards` in the left sidebar as well:

| Dashboard Name | Description |
|---|---|
| KubeDB / Redis / Summary | Instance overview: status, version, mode, node count, resource requests/limits, CPU usage |
| KubeDB / Redis / Pod | Per-pod role, master/slaves, connected clients, memory, commands/sec, network I/O, CPU/memory |
| KubeDB / Redis / Shard | Cluster shard slot health, node/slave count, per-slave status, cluster mode |

## Step 5: Import Dashboard — Option B: Manual (JSON upload)

If you'd rather not run `grafana-operator`, or want fine-grained control over exactly which dashboards get imported, upload the same dashboard JSON files by hand instead.

The KubeDB Redis dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download the JSON files from the [opnpulse/dashboards](https://github.com/opnpulse/dashboards/tree/master/redis) repository (`redis/` folder):

| File | Dashboard |
|------|-----------|
| `redis_summary_dashboard.json` | KubeDB / Redis / Summary |
| `redis_pod_dashboard.json` | KubeDB / Redis / Pod |
| `redis_shards_dashboard.json` | KubeDB / Redis / Shard |

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

## Step 6: Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable       | Applies to              | What to select                                             |
|----------------|--------------------------|--------------------------------------------------------------|
| **datasource** | All dashboards          | Your Prometheus data source                                |
| **Namespace**  | All dashboards          | Namespace where your Redis is deployed (e.g., `demo`)      |
| **app**        | Summary dashboard       | Name of your Redis instance (e.g., `redis-cluster`)         |
| **redis**      | Pod, Shard dashboards   | Name of your Redis instance (e.g., `redis-cluster`)         |
| **Pod Name**   | Pod, Shard dashboards   | A specific pod (e.g., `redis-cluster-shard0-0`)             |
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

**KubeDB / Redis / Shard** — cluster shard health, populated because this tutorial deploys Redis in Cluster mode:
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

# Uninstall the Grafana dashboards chart, if you used Option A
helm uninstall kubedb-grafana-dashboards-redis -n kubeops

# Uninstall grafana-operator (optional — skip if other DB guides on this cluster still use it)
helm uninstall grafana-operator -n kubeops

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your Redis database with KubeDB using [Prometheus Operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis database with KubeDB using [built-in Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
