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

KubeDB exposes Redis metrics through a sidecar exporter, and its own view of each resource (status, phase, version) through Panopticon. Once Prometheus is scraping both, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial covers only the Grafana-specific part of that setup — deploying Grafana, connecting it to Prometheus, and importing the dashboards.

## Before You Begin

- Complete the [Monitoring Redis Using Prometheus Operator](/docs/guides/redis/monitoring/using-prometheus-operator.md) tutorial first. It covers the shared prerequisites this page builds on: a running Prometheus (Operator) instance, a Redis instance deployed with `spec.monitor.agent: prometheus.io/operator`, and **Panopticon** installed — required for the dashboards' database status/version/phase panels, without which they show "No data".

  This tutorial assumes you completed that guide as written, so:
  - Prometheus is running in the `monitoring` namespace, reachable in-cluster at `prometheus-operated.monitoring.svc:9090`.
  - The Redis instance `coreos-prom-redis` is running in the `demo` namespace, with a `ServiceMonitor` already scraping it.
  - Panopticon is running in the `kubeops` namespace.

> Note: YAML files used in this tutorial are stored in [docs/examples/redis/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Grafana

The Prometheus operator setup used in the prerequisite tutorial doesn't bundle Grafana, so install it separately with Helm.

```bash
$ helm repo add grafana https://grafana.github.io/helm-charts
$ helm repo update

$ helm upgrade --install grafana grafana/grafana -n monitoring --set persistence.enabled=false
```

Wait for the Grafana pod to be ready:

```bash
$ kubectl get pods -n monitoring -l=app.kubernetes.io/name=grafana
NAME                       READY   STATUS    RESTARTS   AGE
grafana-6b9c8f9c7d-x2n4p   1/1     Running   0          45s
```

Port-forward the Grafana service:

```bash
$ kubectl port-forward -n monitoring svc/grafana 3000:80
Forwarding from 127.0.0.1:3000 -> 3000
```

Open [http://localhost:3000](http://localhost:3000). The username is `admin`. Retrieve the auto-generated password from the secret:

```bash
$ kubectl get secret -n monitoring grafana -o jsonpath='{.data.admin-password}' | base64 -d
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

## Configure Prometheus as a Data Source

1. Go to **Connections** → **Data sources** → **Add new data source**.
2. Select **Prometheus**.
3. Set the URL to the `Prometheus` operator's service (`prometheus-operated`, created automatically by the operator for the `prometheus` CR from the prerequisite tutorial):

   ```
   http://prometheus-operated.monitoring.svc:9090
   ```

4. Click **Save & test**. You should see `Data source is working`.

## Import Dashboard — Option A: Automatic (chart)

Rather than downloading and uploading each JSON file by hand (Option B below), KubeDB ships a chart that creates all matching dashboards for you as `GrafanaDashboard` custom resources. A separate controller, **`grafana-operator`**, watches these resources and pushes the actual dashboard JSON into your Grafana instance — both pieces are required.

**1. Install `grafana-operator`** (skip if it's already running in your cluster):

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm upgrade --install grafana-operator appscode/grafana-operator \
    --version v2026.6.12 \
    --namespace kubeops --create-namespace
```

**2. Register your Grafana instance as an `AppBinding`.** `grafana-operator` needs to know where to push dashboards and how to authenticate — it reads this from an `AppBinding` object, not from the chart install command itself. Create an API key for the Grafana deployed above, store it in a Secret, and reference both in an `AppBinding`:

```bash
$ kubectl port-forward -n monitoring svc/grafana 3000:80 &

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
    url: http://grafana.monitoring.svc:80
  secret:
    name: grafana-admin-token
EOF
```

**3. Install the dashboards:**

```bash
$ helm upgrade -i kubedb-grafana-dashboards appscode/kubedb-grafana-dashboards \
    -n kubeops --create-namespace --version=v2026.8.14-rc.0 \
    --set featureGates.Redis=true \
    --set grafana.name=grafana \
    --set grafana.namespace=monitoring
```

`featureGates.Redis` already defaults to `true` — set explicitly above for clarity. `grafana.name`/`grafana.namespace` point the chart's `GrafanaDashboard` resources at the `AppBinding` created in step 2 (omitting them falls back to whichever `AppBinding` in your cluster is labeled as the cluster-default Grafana, if any — explicit is safer on a shared cluster).

This single command creates every dashboard this chart ships for Redis — `KubeDB / Redis / Summary`, `KubeDB / Redis / Pod`, `KubeDB / Redis / Shard` (plus the `RedisSentinel` variants) — which `grafana-operator` then pushes into your Grafana instance automatically. No manual JSON download or upload needed.

Verify they landed:

```bash
$ kubectl get grafanadashboards.openviz.dev -n kubeops | grep -i redis
NAME                            TITLE                      SYNCED    AGE
grafana-kubedb-redis-pod        KubeDB / Redis / Pod       Current   30s
grafana-kubedb-redis-shard      KubeDB / Redis / Shard     Current   30s
grafana-kubedb-redis-summary    KubeDB / Redis / Summary   Current   30s
```

> The `grafana-` prefix on each resource name comes from the `grafana.name=grafana` value set above (the chart prepends it to the dashboard title to build the resource name) — this is expected.

`SYNCED: Current` confirms `grafana-operator` successfully pushed each dashboard into Grafana. Open Grafana — the three dashboards are already there under `Dashboards`, fully wired to your Prometheus data source, ready to explore in [Explore the Dashboard](#explore-the-dashboard) below.

## Import Dashboard — Option B: Manual (JSON upload)

If you'd rather not run `grafana-operator`, or want fine-grained control over exactly which dashboards get imported, upload the same dashboard JSON files by hand instead.

The KubeDB Redis dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download the JSON files from the [opnpulse/dashboards](https://github.com/opnpulse/dashboards/tree/master/redis) repository (`redis/` folder):

| File | Dashboard |
|------|-----------|
| `redis_summary_dashboard.json` | KubeDB / Redis / Summary |
| `redis_pod_dashboard.json` | KubeDB / Redis / Pod |
| `redis_shards_dashboard.json` | KubeDB / Redis / Shard |

> The Shard dashboard is relevant for Redis Cluster mode (`spec.mode: Cluster`); its panels stay empty for a standalone (non-cluster) Redis instance like the one deployed in the prerequisite tutorial.

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

## Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable       | Applies to              | What to select                                             |
|----------------|--------------------------|--------------------------------------------------------------|
| **datasource** | All dashboards          | Your Prometheus data source                                |
| **Namespace**  | All dashboards          | Namespace where your Redis is deployed (e.g., `demo`)      |
| **app**        | Summary dashboard       | Name of your Redis instance (e.g., `coreos-prom-redis`)    |
| **redis**      | Pod, Shard dashboards   | Name of your Redis instance (e.g., `coreos-prom-redis`)    |
| **Pod Name**   | Pod, Shard dashboards   | A specific pod (e.g., `coreos-prom-redis-0`)               |
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

This page only cleans up the Grafana-specific resources it created. For Redis, Prometheus, and Panopticon, see the [Cleaning up](/docs/guides/redis/monitoring/using-prometheus-operator.md#cleaning-up) section of the prerequisite tutorial.

```bash
# If you used Option A (automatic dashboard import)
$ helm uninstall kubedb-grafana-dashboards -n kubeops
$ helm uninstall grafana-operator -n kubeops
$ kubectl delete secret -n monitoring grafana-admin-token
$ kubectl delete appbinding -n monitoring grafana

$ helm uninstall grafana -n monitoring
```

## Next Steps

- Monitor your Redis database with KubeDB using [Prometheus Operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis database with KubeDB using [built-in Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
