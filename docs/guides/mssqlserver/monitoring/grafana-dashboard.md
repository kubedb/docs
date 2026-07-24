---
title: Visualize MSSQLServer Metrics with Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: ms-grafana-dashboard-monitoring
    name: Grafana Dashboard
    parent: ms-monitoring
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Visualize MSSQLServer Metrics with Grafana Dashboard

KubeDB exposes MSSQLServer metrics through a sidecar exporter. Once Prometheus scrapes those metrics, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a MSSQLServer instance, and importing the Grafana dashboards.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (Step 2) uses to expose metrics to Prometheus.

- MSSQLServer requires TLS to be enabled. You need a cert-manager `Issuer` in the `demo` namespace before deploying. If cert-manager is not installed, install it first:

  ```bash
  $ helm repo add jetstack https://charts.jetstack.io
  $ helm repo update
  $ helm upgrade --install cert-manager jetstack/cert-manager \
    --namespace cert-manager --create-namespace \
    --set crds.enabled=true
  ```

  Then create a self-signed `Issuer` in the `demo` namespace:

  ```yaml
  apiVersion: cert-manager.io/v1
  kind: Issuer
  metadata:
    name: mssqlserver-ca-issuer
    namespace: demo
  spec:
    selfSigned: {}
  ```

  ```bash
  $ kubectl apply -f issuer.yaml
  issuer.cert-manager.io/mssqlserver-ca-issuer created
  ```

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mssqlserver/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Configuration

> These two steps — deploying `kube-prometheus-stack` and installing Panopticon — are shared prerequisites for all KubeDB database monitoring guides. If you have already completed them in another guide, skip to [Step 1](#step-1-deploy-mssqlserver-with-monitoring-enabled).

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

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the MSSQLServer instance.

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

## Step 1: Deploy MSSQLServer with Monitoring Enabled

Below is the MSSQLServer object with TLS and monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-grafana-demo
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
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

- `tls.issuerRef` is required for MSSQLServer; it references the cert-manager Issuer created above.
- `monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` for this instance.
- `monitor.prometheus.serviceMonitor.labels` must match the `serviceMonitorSelector` label of your Prometheus (`release: prometheus`).
- `monitor.prometheus.serviceMonitor.interval` sets the scrape interval to 10 seconds.

Create the MSSQLServer instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/monitoring/mssql-grafana-demo.yaml
mssqlserver.kubedb.com/mssql-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get mssqlserver -n demo mssql-grafana-demo
NAME                 VERSION    STATUS   AGE
mssql-grafana-demo   2022-cu12  Ready    3m
```

KubeDB creates a stats service named `{mssqlserver-name}-stats` for the exporter:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=mssql-grafana-demo"
NAME                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
mssql-grafana-demo         ClusterIP   10.96.10.1     <none>        1433/TCP    3m
mssql-grafana-demo-stats   ClusterIP   10.96.10.2     <none>        9399/TCP    3m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                       AGE
mssql-grafana-demo-stats   3m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo mssql-grafana-demo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `mssql-grafana-demo-stats`. Its state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/mssqlserver/monitoring/ms-prom-targets.png" style="padding:10px">
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
  <img alt="Grafana Login" src="/docs/images/mssqlserver/monitoring/ms-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/mssqlserver/monitoring/ms-grafana-home.png" style="padding:10px">
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

## Step 5: Import KubeDB MSSQLServer Dashboard

The KubeDB MSSQLServer dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download all three JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/mssqlserver) repository (`mssqlserver/` folder):

| File | Dashboard |
|------|-----------|
| `mssqlserver_summary_dashboard.json` | KubeDB / MSSQLServer / Summary |
| `mssqlserver_pods_dashboard.json` | KubeDB / MSSQLServer / Pod |
| `mssqlserver_databases_dashboard.json` | KubeDB / MSSQLServer / Database |

**Import steps (repeat for each of the three files):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

The import page looks like this — click **Upload dashboard JSON file** to select the file:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/mssqlserver/monitoring/ms-grafana-import.png" style="padding:10px">
</p>

After importing all three files, they will appear under `Dashboards` in the left sidebar.

| Dashboard Name | Description |
|---|---|
| KubeDB / MSSQLServer / Summary | Instance status, version, node count, resource requests/limits, CPU usage |
| KubeDB / MSSQLServer / Pod | Per-pod status, role (Primary/Secondary), uptime, server resource overview, connections |
| KubeDB / MSSQLServer / Database | Service status/uptime, AG cluster replica roles, cluster status, SQL compilations, batch requests |

## Step 6: Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable        | Applies to        | What to select                                                  |
|------------------|--------------------|-------------------------------------------------------------------|
| **Namespace**    | All dashboards     | Namespace where your MSSQLServer is deployed (e.g., `demo`)      |
| **mssqlserver**  | All dashboards     | Name of your instance (e.g., `mssql-grafana-demo`)               |
| **Pod Name**     | Pod dashboard      | A specific pod (e.g., `mssql-grafana-demo-0`)                    |
| **Job / database** | Pod dashboard   | The stats job and target database to inspect                     |

**KubeDB / MSSQLServer / Summary** — start here for an instance overview:
- **General Info** — database status, version, whether secure transport is required, deletion policy, total nodes
- **Resource Requests / Limits** — configured CPU, memory, and storage requests and limits
- **CPU Info / CPU Quota** — per-pod CPU usage over time and quota utilization

<p align="center">
  <img alt="KubeDB MSSQLServer Summary Dashboard" src="/docs/images/mssqlserver/monitoring/ms-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / MSSQLServer / Pod** — drill into a specific pod:
- **Pod Name / Status / Role / Uptime** — pod identity, running status, Availability Group role (Primary/Secondary), and uptime
- **Server Resource Overview** — server local time, total and used RAM, pagefile size and usage
- **Server Resource Overview (2)** — total page faults, batch requests/sec, page life expectancy, deadlocks, user errors/sec, kill connection errors/sec
- **Summary** — current database connections, log growth since last restart, total I/O stall wait time

<p align="center">
  <img alt="KubeDB MSSQLServer Pod Dashboard" src="/docs/images/mssqlserver/monitoring/ms-grafana-pod.png" style="padding:10px">
</p>

**KubeDB / MSSQLServer / Database** — Availability Group cluster health:
- **Service Status / Uptime** — per-pod health and how long each pod has been serving
- **AG Cluster Active Replica** — which pod is acting as the active replica
- **Cluster Status** — count of Primary vs. Secondary replicas
- **SQL Compilations/sec** — per-pod query compilation rate
- **Batch Requests per Second** — per-pod T-SQL batch throughput

<p align="center">
  <img alt="KubeDB MSSQLServer Database Dashboard" src="/docs/images/mssqlserver/monitoring/ms-grafana-database.png" style="padding:10px">
</p>


## Cleaning up

```bash
# Remove the MSSQLServer instance
kubectl delete mssqlserver -n demo mssql-grafana-demo

# Remove the TLS Issuer
kubectl delete issuer mssqlserver-ca-issuer -n demo

# Remove namespaces
kubectl delete ns demo

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your MSSQLServer instance with KubeDB using [Prometheus Operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
