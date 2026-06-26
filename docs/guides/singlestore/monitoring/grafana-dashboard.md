---
title: Visualize Singlestore Metrics with Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-grafana-dashboard
    name: Grafana Dashboard
    parent: guides-sdb-monitoring
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Visualize Singlestore Metrics with Grafana Dashboard

KubeDB exposes Singlestore metrics through a sidecar exporter. Once Prometheus scrapes those metrics, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a Singlestore instance, and importing the Grafana dashboards.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (Step 2) uses to expose metrics to Prometheus.

- Singlestore requires a valid license secret. Create the license secret before deploying:

  ```bash
  $ kubectl create secret generic sdb-license -n demo \
    --from-literal=username=<your-singlestore-email> \
    --from-literal=password=<your-license-key>
  ```

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/singlestore/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/singlestore/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the Singlestore instance.

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

## Step 3: Deploy Singlestore with Monitoring Enabled

Below is the Singlestore object with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-grafana-demo
  namespace: demo
spec:
  version: "8.7.10"
  licenseSecret:
    name: sdb-license
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

Create the Singlestore instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/singlestore/monitoring/sdb-grafana-demo.yaml
singlestore.kubedb.com/sdb-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get singlestore -n demo sdb-grafana-demo
NAME               VERSION   STATUS   AGE
sdb-grafana-demo   8.7.10    Ready    3m
```

KubeDB creates a stats service named `{singlestore-name}-stats` for the exporter:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=sdb-grafana-demo"
NAME                     TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
sdb-grafana-demo         ClusterIP   10.96.10.1     <none>        3306/TCP    3m
sdb-grafana-demo-stats   ClusterIP   10.96.10.2     <none>        9104/TCP    3m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                     AGE
sdb-grafana-demo-stats   3m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo sdb-grafana-demo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `sdb-grafana-demo-stats`. Its state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/singlestore/monitoring/sdb-prom-targets.png" style="padding:10px">
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

<p align="center">
  <img alt="Grafana Login" src="/docs/images/singlestore/monitoring/sdb-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/singlestore/monitoring/sdb-grafana-home.png" style="padding:10px">
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

## Step 7: Import KubeDB Singlestore Dashboard

The KubeDB Singlestore dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download all three JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/singlestore) repository (`singlestore/` folder):

| File | Dashboard |
|------|-----------|
| `singlestore_summary_dashboard.json` | KubeDB / Singlestore / Summary |
| `singlestore_pods_dashboard.json` | KubeDB / Singlestore / Pod |
| `singlestore_databases_dashboard.json` | KubeDB / Singlestore / Database |

**Import steps (repeat for each of the three files):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

The import page looks like this — click **Upload dashboard JSON file** to select the file:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/singlestore/monitoring/sdb-grafana-import.png" style="padding:10px">
</p>

After importing all three files, they will appear under `Dashboards` in the left sidebar.

| Dashboard Name | Description |
|---|---|
| KubeDB / Singlestore / Summary | Cluster status, connections, queries/sec, memory usage, CPU/storage |
| KubeDB / Singlestore / Pod | Per-pod connections, query rate, CPU/memory usage |
| KubeDB / Singlestore / Database | Database-level DML rates, row count, table size, index usage |

## Step 8: Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable      | Applies to              | What to select                                                |
|---------------|-------------------------|---------------------------------------------------------------|
| **namespace** | All dashboards          | Namespace where your Singlestore is deployed (e.g., `demo`)  |
| **app**       | All dashboards          | Name of your instance (e.g., `sdb-grafana-demo`)             |
| **pod**       | Pod, Database dashboards | A specific pod, or `All` for an aggregated view             |

**KubeDB / Singlestore / Summary** — start here for an instance overview:
- **Node Count** — number of aggregator and leaf nodes
- **Connections** — active client connections
- **Queries per Second** — total query throughput
- **Memory Usage** — used vs. total memory across nodes
- **CPU / Network** — resource consumption vs. requests and limits

<p align="center">
  <img alt="KubeDB Singlestore Summary Dashboard" src="/docs/images/singlestore/monitoring/sdb-grafana-summary.png" style="padding:10px">
</p>
<p align="center">
  <img alt="KubeDB Singlestore Summary Dashboard - continued" src="/docs/images/singlestore/monitoring/sdb-grafana-summary-2.png" style="padding:10px">
</p>

**KubeDB / Singlestore / Pod** — drill into a specific node:
- **Connections** — connections on this specific pod
- **Query Throughput** — queries executed on this pod per second
- **Rows Read / Written** — per-pod row-level throughput
- **CPU / Memory** — per-pod resource usage

<p align="center">
  <img alt="KubeDB Singlestore Pod Dashboard" src="/docs/images/singlestore/monitoring/sdb-grafana-pod.png" style="padding:10px">
</p>
<p align="center">
  <img alt="KubeDB Singlestore Pod Dashboard - continued" src="/docs/images/singlestore/monitoring/sdb-grafana-pod-2.png" style="padding:10px">
</p>

**KubeDB / Singlestore / Database** — database-level storage metrics:
- **Table Count** — number of tables per database
- **Row Count** — estimated rows per table
- **Index Size** — disk space used by indexes
- **Data Size** — disk space used by table data
- **Replication Lag** — replica lag relative to the aggregator (for distributed configurations)

<p align="center">
  <img alt="KubeDB Singlestore Database Dashboard" src="/docs/images/singlestore/monitoring/sdb-grafana-database.png" style="padding:10px">
</p>
<p align="center">
  <img alt="KubeDB Singlestore Database Dashboard - continued" src="/docs/images/singlestore/monitoring/sdb-grafana-database-2.png" style="padding:10px">
</p>

## Cleaning up

```bash
# Remove the Singlestore instance
kubectl delete singlestore -n demo sdb-grafana-demo

# Remove the license secret
kubectl delete secret sdb-license -n demo

# Remove namespaces
kubectl delete ns demo

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your Singlestore instance with KubeDB using [Prometheus Operator](/docs/guides/singlestore/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
