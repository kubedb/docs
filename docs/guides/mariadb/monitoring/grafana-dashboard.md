---
title: Visualize MariaDB Metrics with Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-grafana-dashboard
    name: Grafana Dashboard
    parent: guides-mariadb-monitoring
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Visualize MariaDB Metrics with Grafana Dashboard

KubeDB exposes MariaDB metrics through a sidecar exporter. Once Prometheus scrapes those metrics, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a MariaDB instance, and importing the Grafana dashboards.

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

> Note: YAML files used in this tutorial are stored in [docs/examples/mariadb/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mariadb/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the MariaDB instance.

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

## Step 3: Deploy MariaDB with Monitoring Enabled

Below is the MariaDB object with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-grafana-demo
  namespace: demo
spec:
  version: "11.5.2"
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

Create the MariaDB instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/monitoring/coreos-prom-mariadb.yaml
mariadb.kubedb.com/mariadb-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get mariadb -n demo mariadb-grafana-demo
NAME                   VERSION   STATUS   AGE
mariadb-grafana-demo   11.5.2    Ready    2m
```

KubeDB creates a stats service named `{mariadb-name}-stats` for the exporter:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=mariadb-grafana-demo"
NAME                         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
mariadb-grafana-demo         ClusterIP   10.96.10.1     <none>        3306/TCP    2m
mariadb-grafana-demo-stats   ClusterIP   10.96.10.2     <none>        9104/TCP    2m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                         AGE
mariadb-grafana-demo-stats   2m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo mariadb-grafana-demo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `mariadb-grafana-demo-stats`. Its state should be **UP**.

If the target is missing, check that the `ServiceMonitor` label (`release: prometheus`) matches the Prometheus `serviceMonitorSelector`.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/mariadb/monitoring/mariadb-prom-targets.png" style="padding:10px">
</p>

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

<p align="center">
  <img alt="Grafana Login" src="/docs/images/mariadb/monitoring/mariadb-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/mariadb/monitoring/mariadb-grafana-home.png" style="padding:10px">
</p>

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

## Step 7: Import KubeDB MariaDB Dashboard

The KubeDB MariaDB dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Four dashboards are available. Download all JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/mariadb) repository (`mariadb/` folder):

| File | Dashboard |
|------|-----------|
| `mariadb_summary_dashboard.json` | KubeDB / MariaDB / Summary |
| `mariadb_pods_dashboard.json` | KubeDB / MariaDB / Pod |
| `mariadb_databases_dashboard.json` | KubeDB / MariaDB / Database |
| `mariadb_galera_cluster_dashboard.json` | KubeDB / MariaDB / Galera Cluster |

> The Galera Cluster dashboard is only relevant for MariaDB Galera cluster deployments (`spec.topology.mode: GaleraCluster`).

**Import steps (repeat for each file):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

The import page looks like this — click **Upload dashboard JSON file** to select the file:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/mariadb/monitoring/mariadb-grafana-import.png" style="padding:10px">
</p>

After importing all files, they will appear under `Dashboards` in the left sidebar.

| Dashboard Name                    | Description                                                                                      |
|-----------------------------------|--------------------------------------------------------------------------------------------------|
| KubeDB / MariaDB / Summary        | Instance overview: connections, QPS, slow queries, aborted connections, CPU/memory/storage       |
| KubeDB / MariaDB / Pod            | Per-pod connections, threads, CPU/memory usage, table lock contention                            |
| KubeDB / MariaDB / Database       | InnoDB buffer pool, row operations, query cache, temporary tables, handler statistics            |
| KubeDB / MariaDB / Galera Cluster | Cluster size, node state, wsrep_ready, flow control, replication bytes, commit/cert failure rate |

## Step 8: Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable      | Applies to              | What to select                                                |
|---------------|-------------------------|---------------------------------------------------------------|
| **namespace** | All dashboards          | Namespace where your MariaDB is deployed (e.g., `demo`)      |
| **app**       | All dashboards          | Name of your MariaDB instance (e.g., `mariadb-grafana-demo`) |
| **pod**       | Pod, Database dashboards | A specific pod, or `All` for an aggregated view              |

**KubeDB / MariaDB / Summary** — start here for an instance overview:
- **Version / Uptime** — MariaDB version and how long the server has been running
- **Connections** — max connections configured vs. current active connections
- **Questions per Second** — total query throughput
- **Slow Queries** — queries exceeding the `long_query_time` threshold
- **Aborted Clients / Connections** — network or authentication failures
- **CPU / Memory / Storage / Network** — resource consumption vs. requests and limits

<p align="center">
  <img alt="KubeDB MariaDB Summary Dashboard" src="/docs/images/mariadb/monitoring/mariadb-grafana-summary.png" style="padding:10px">
</p>
<p align="center">
  <img alt="KubeDB MariaDB Summary Dashboard - continued" src="/docs/images/mariadb/monitoring/mariadb-grafana-summary-2.png" style="padding:10px">
</p>

**KubeDB / MariaDB / Pod** — drill into a specific pod:
- **Connections / Questions / Slow Queries** — per-pod query metrics
- **Thread Count** — running and connected threads
- **CPU / Memory** — per-pod resource usage
- **Table Locks Waited / Immediate** — lock contention indicator

<p align="center">
  <img alt="KubeDB MariaDB Pod Dashboard" src="/docs/images/mariadb/monitoring/mariadb-grafana-pod.png" style="padding:10px">
</p>
<p align="center">
  <img alt="KubeDB MariaDB Pod Dashboard - continued" src="/docs/images/mariadb/monitoring/mariadb-grafana-pod-2.png" style="padding:10px">
</p>

**KubeDB / MariaDB / Database** — storage engine and query cache metrics:
- **InnoDB Buffer Pool** — reads and writes through the buffer pool
- **InnoDB Row Operations** — read, insert, update, delete rows per second
- **Query Cache** — hits vs. inserts (if query cache is enabled)
- **Temporary Tables** — tables created on disk vs. in memory
- **Handler Statistics** — low-level read operations per index

<p align="center">
  <img alt="KubeDB MariaDB Database Dashboard" src="/docs/images/mariadb/monitoring/mariadb-grafana-database.png" style="padding:10px">
</p>
<p align="center">
  <img alt="KubeDB MariaDB Database Dashboard - continued" src="/docs/images/mariadb/monitoring/mariadb-grafana-database-2.png" style="padding:10px">
</p>

**KubeDB / MariaDB / Galera Cluster** — Galera-specific metrics:
- **Cluster Size** — number of nodes in the cluster
- **Local State** — wsrep state per node (Synced, Donor, Joiner, etc.)
- **wsrep_ready** — whether each node is ready to accept queries
- **Flow Control Paused** — percentage of time replication was paused due to flow control
- **Replication Bytes** — bytes sent and received via Galera replication per node
- **Local Commits / Cert Failures** — commit throughput and certification conflict rate

<p align="center">
  <img alt="KubeDB MariaDB Galera Cluster Dashboard" src="/docs/images/mariadb/monitoring/mariadb-grafana-galera.png" style="padding:10px">
</p>
<p align="center">
  <img alt="KubeDB MariaDB Galera Cluster Dashboard - continued" src="/docs/images/mariadb/monitoring/mariadb-grafana-galera-2.png" style="padding:10px">
</p>

## Cleaning up

```bash
# Remove the MariaDB instance
kubectl delete mariadb -n demo mariadb-grafana-demo

# Remove namespaces
kubectl delete ns demo

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your MariaDB database with KubeDB using [built-in Prometheus](/docs/guides/mariadb/monitoring/builtin-prometheus/).
- Monitor your MariaDB database with KubeDB using [Prometheus Operator](/docs/guides/mariadb/monitoring/prometheus-operator/).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
