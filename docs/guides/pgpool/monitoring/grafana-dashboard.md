---
title: Visualize Pgpool Metrics with Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: pp-grafana-dashboard-monitoring
    name: Grafana Dashboard
    parent: pp-monitoring-pgpool
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Visualize Pgpool Metrics with Grafana Dashboard

KubeDB exposes Pgpool metrics through a sidecar exporter. Once Prometheus scrapes those metrics, you can visualize them in Grafana using a pre-built KubeDB dashboard. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a Pgpool instance, and importing the Grafana dashboard.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (Step 2) uses to expose metrics to Prometheus.

- Pgpool sits in front of a PostgreSQL server. Prepare a KubeDB Postgres instance (for example `ha-postgres` in the `demo` namespace) following the [streaming replication guide](/docs/guides/postgres/clustering/streaming_replication.md), or use any externally managed Postgres.

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgpool/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgpool/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Configuration

> These two steps — deploying `kube-prometheus-stack` and installing Panopticon — are shared prerequisites for all KubeDB database monitoring guides. If you have already completed them in another guide, skip to [Step 1](#step-1-deploy-pgpool-with-monitoring-enabled).

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

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the Pgpool instance.

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

## Step 1: Deploy Pgpool with Monitoring Enabled

Below is the Pgpool object pointing at the `ha-postgres` backend, with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pp-grafana-demo
  namespace: demo
spec:
  version: "4.5.0"
  postgresRef:
    name: ha-postgres
    namespace: demo
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

- `postgresRef` points at the backend Postgres instance Pgpool load-balances connections across.
- `monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` for this instance.
- `monitor.prometheus.serviceMonitor.labels` must match the `serviceMonitorSelector` label of your Prometheus (`release: prometheus`).
- `monitor.prometheus.serviceMonitor.interval` sets the scrape interval to 10 seconds.

Create the Pgpool instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/monitoring/coreos-prom-pp.yaml
pgpool.kubedb.com/pp-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get pp -n demo pp-grafana-demo
NAME             TYPE                 VERSION   STATUS   AGE
pp-grafana-demo  kubedb.com/v1alpha2  4.5.0     Ready    65s
```

KubeDB creates a stats service named `{pgpool-name}-stats` for the exporter:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=pp-grafana-demo"
NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
pp-grafana-demo       ClusterIP   10.96.201.180   <none>        9999/TCP,9595/TCP   2m
pp-grafana-demo-pods  ClusterIP   None            <none>        9999/TCP            2m
pp-grafana-demo-stats ClusterIP   10.96.73.22     <none>        9719/TCP            2m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                  AGE
pp-grafana-demo-stats 2m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo pp-grafana-demo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `pp-grafana-demo-stats`. Its state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/pgpool/monitoring/pp-prom-targets.png" style="padding:10px">
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
  <img alt="Grafana Login" src="/docs/images/pgpool/monitoring/pp-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/pgpool/monitoring/pp-grafana-home.png" style="padding:10px">
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

## Step 5: Import KubeDB Pgpool Dashboard

The KubeDB Pgpool dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download all three JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/pgpool) repository (`pgpool/` folder):

| File | Dashboard |
|------|-----------|
| `pgpool_summary_dashboard.json` | KubeDB / Pgpool / Summary |
| `pgpool_pods_dashboard.json` | KubeDB / Pgpool / Pod |
| `pgpool_databases_dashboard.json` | KubeDB / Pgpool / Database |

**Import steps (repeat for each of the three files):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

The import page looks like this — click **Upload dashboard JSON file** to select the file:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/pgpool/monitoring/pp-grafana-import.png" style="padding:10px">
</p>

After importing all three files, they will appear under `Dashboards` in the left sidebar.

| Dashboard Name | Description |
|---|---|
| KubeDB / Pgpool / Summary | Node health, client/server connections, query throughput, replication delay, CPU/memory/storage |
| KubeDB / Pgpool / Pod | Per-pod connections, query rate, cache hit ratio, CPU/memory |
| KubeDB / Pgpool / Database | Per-backend connections, replication status, statement throughput |

## Step 6: Explore the Dashboard

After opening a dashboard, you will see dropdown filters at the top. These control which data is shown across all panels — change them to focus on a specific instance without editing any queries.

| Variable      | Applies to              | What to select                                            |
|---------------|-------------------------|-----------------------------------------------------------|
| **namespace** | All dashboards          | Namespace where your Pgpool is deployed (e.g., `demo`)   |
| **app**       | All dashboards          | Name of your Pgpool instance (e.g., `pp-grafana-demo`)   |
| **pod**       | Pod, Database dashboards | A specific pod, or `All` to see aggregated view          |

Once you set these, all panels update automatically. Below is what each dashboard shows:

**KubeDB / Pgpool / Summary** — start here for a pool-wide overview:
- **Node Health** — Pgpool node status and process liveness
- **Client / Server Connections** — total client and backend Postgres connections
- **Queries per Second** — statement throughput routed through Pgpool
- **Replication Delay** — lag between primary and replica backends (for load-balanced setups)
- **Cache Hit Ratio** — in-memory query cache effectiveness
- **CPU / Memory / Storage** — resource consumption vs. requests and limits

<p align="center">
  <img alt="KubeDB Pgpool Summary Dashboard" src="/docs/images/pgpool/monitoring/pp-grafana-summary.png" style="padding:10px">
</p>


**KubeDB / Pgpool / Pod** — drill into a specific pod:
- **Client / Server Connections** — connections held on this pod
- **Queries per Second** — per-pod statement throughput
- **Cache Hit Ratio** — per-pod query cache effectiveness
- **Backend Pool Status** — active/idle backend connections per pod
- **CPU / Memory** — per-pod resource usage

<p align="center">
  <img alt="KubeDB Pgpool Pod Dashboard" src="/docs/images/pgpool/monitoring/pp-grafana-pod.png" style="padding:10px">
</p>

**KubeDB / Pgpool / Database** — per-backend and per-database metrics:
- **Backend Connection Status** — up/down/recovering status per backend Postgres node
- **Connections per Backend** — active and idle connections per backend
- **Replication Status** — primary/standby role and replication health per backend
- **Statement Throughput** — SELECT/INSERT/UPDATE/DELETE rate per database
- **Backend Select / Load Balance Ratio** — query distribution across backends

<p align="center">
  <img alt="KubeDB Pgpool Database Dashboard" src="/docs/images/pgpool/monitoring/pp-grafana-database.png" style="padding:10px">
</p>

## Cleaning up

```bash
# Remove the Pgpool instance
kubectl delete pp -n demo pp-grafana-demo

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

- Monitor your Pgpool database with KubeDB using [built-in Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Monitor your Pgpool database with KubeDB using [Prometheus Operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
