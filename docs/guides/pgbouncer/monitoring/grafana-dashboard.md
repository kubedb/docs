---
title: Visualize PgBouncer Metrics with Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: pb-grafana-dashboard-monitoring
    name: Grafana Dashboard
    parent: pb-monitoring-pgbouncer
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Visualize PgBouncer Metrics with Grafana Dashboard

KubeDB exposes PgBouncer metrics through a sidecar exporter. Once Prometheus scrapes those metrics, you can visualize them in Grafana using a pre-built KubeDB dashboard. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a PgBouncer instance, and importing the Grafana dashboard.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (Step 2) uses to expose metrics to Prometheus.

- PgBouncer sits in front of a PostgreSQL server. Prepare a KubeDB Postgres instance (for example `ha-postgres` in the `demo` namespace) following the [streaming replication guide](/docs/guides/postgres/clustering/streaming_replication.md), or use any externally managed Postgres.

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgbouncer/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Configuration

> These two steps — deploying `kube-prometheus-stack` and installing Panopticon — are shared prerequisites for all KubeDB database monitoring guides. If you have already completed them in another guide, skip to [Step 1](#step-1-deploy-pgbouncer-with-monitoring-enabled).

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

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the PgBouncer instance.

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

## Step 1: Deploy PgBouncer with Monitoring Enabled

Below is the PgBouncer object pointing at the `ha-postgres` backend, with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb-grafana-demo
  namespace: demo
spec:
  replicas: 1
  version: "1.24.0"
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "ha-postgres"
      namespace: demo
  connectionPool:
    poolMode: session
    port: 5432
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

- `database.databaseRef` points at the backend Postgres instance PgBouncer pools connections for.
- `monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` for this instance.
- `monitor.prometheus.serviceMonitor.labels` must match the `serviceMonitorSelector` label of your Prometheus (`release: prometheus`).
- `monitor.prometheus.serviceMonitor.interval` sets the scrape interval to 10 seconds.

Create the PgBouncer instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/monitoring/coreos-prom-pb.yaml
pgbouncer.kubedb.com/pb-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get pb -n demo pb-grafana-demo
NAME             TYPE           VERSION   STATUS   AGE
pb-grafana-demo  kubedb.com/v1  1.24.0    Ready    65s
```

KubeDB creates a stats service named `{pgbouncer-name}-stats` for the exporter:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=pb-grafana-demo"
NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
pb-grafana-demo       ClusterIP   10.96.201.180   <none>        5432/TCP            2m
pb-grafana-demo-pods  ClusterIP   None            <none>        5432/TCP            2m
pb-grafana-demo-stats ClusterIP   10.96.73.22     <none>        9719/TCP            2m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                  AGE
pb-grafana-demo-stats 2m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo pb-grafana-demo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `pb-grafana-demo-stats`. Its state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/pgbouncer/monitoring/pb-prom-targets.png" style="padding:10px">
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
  <img alt="Grafana Login" src="/docs/images/pgbouncer/monitoring/pb-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/pgbouncer/monitoring/pb-grafana-home.png" style="padding:10px">
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

## Step 5: Import KubeDB PgBouncer Dashboard

The KubeDB PgBouncer dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download all three JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/pgbouncer) repository (`pgbouncer/` folder):

| File | Dashboard |
|------|-----------|
| `pgbouncer_summary_dashboard.json` | KubeDB / PgBouncer / Summary |
| `pgbouncer_pods_dashboard.json` | KubeDB / PgBouncer / Pod |
| `pgbouncer_databases_dashboard.json` | KubeDB / PgBouncer / Database |

**Import steps (repeat for each of the three files):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

The import page looks like this — click **Upload dashboard JSON file** to select the file:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/pgbouncer/monitoring/pb-grafana-import.png" style="padding:10px">
</p>

After importing all three files, they will appear under `Dashboards` in the left sidebar.

| Dashboard Name | Description |
|---|---|
| KubeDB / PgBouncer / Summary | Total/client/server connections, active pools, query throughput, CPU/memory/storage |
| KubeDB / PgBouncer / Pod | Per-pod connections, pools, waiting clients, query rate, CPU/memory |
| KubeDB / PgBouncer / Database | Per-database pool state, active/idle/waiting clients, max connections, query rate |

## Step 6: Explore the Dashboard

After opening a dashboard, you will see dropdown filters at the top. These control which data is shown across all panels — change them to focus on a specific instance without editing any queries.

| Variable      | Applies to              | What to select                                              |
|---------------|-------------------------|-------------------------------------------------------------|
| **namespace** | All dashboards          | Namespace where your PgBouncer is deployed (e.g., `demo`)  |
| **app**       | All dashboards          | Name of your PgBouncer instance (e.g., `pb-grafana-demo`)  |
| **pod**       | Pod, Database dashboards | A specific pod, or `All` to see aggregated view            |

Once you set these, all panels update automatically. Below is what each dashboard shows:

**KubeDB / PgBouncer / Summary** — start here for a pool-wide overview:
- **Total / Client / Server Connections** — connections across all pools
- **Active Pools** — number of connection pools to the backend
- **Queries per Second** — query throughput routed through PgBouncer
- **Waiting Clients** — clients queued waiting for a server connection (sustained non-zero indicates pool exhaustion)
- **Avg Query Time** — average time queries spend waiting + executing
- **CPU / Memory / Storage** — resource consumption vs. requests and limits

<p align="center">
  <img alt="KubeDB PgBouncer Summary Dashboard" src="/docs/images/pgbouncer/monitoring/pb-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / PgBouncer / Pod** — drill into a specific pod:
- **Client / Server Connections** — connections held on this pod
- **Active Pools** — pools managed by this pod
- **Queries per Second** — per-pod query throughput
- **Waiting Clients** — queued clients on this pod
- **CPU / Memory** — per-pod resource usage

<p align="center">
  <img alt="KubeDB PgBouncer Pod Dashboard" src="/docs/images/pgbouncer/monitoring/pb-grafana-pod.png" style="padding:10px">
</p>


**KubeDB / PgBouncer / Database** — per-database pool metrics:
- **Pool State** — cl_active, cl_waiting, sv_active, sv_idle per database
- **Active / Idle Server Connections** — backend connection usage per database
- **Max Client Connections** — configured limit per database
- **Query Rate** — queries per second per database
- **Total Xact / Query Count** — transaction and query counters per database

<p align="center">
  <img alt="KubeDB PgBouncer Database Dashboard" src="/docs/images/pgbouncer/monitoring/pb-grafana-database.png" style="padding:10px">
</p>

## Cleaning up

```bash
# Remove the PgBouncer instance
kubectl delete pb -n demo pb-grafana-demo

# Remove the backend Postgres instance
kubectl delete pg -n demo ha-postgres

# Remove namespaces
kubectl delete ns demo

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your PgBouncer database with KubeDB using [built-in Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Monitor your PgBouncer database with KubeDB using [Prometheus Operator](/docs/guides/pgbouncer/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
