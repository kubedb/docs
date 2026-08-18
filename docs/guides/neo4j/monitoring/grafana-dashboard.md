---
title: Neo4j Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: guides-neo4j-grafana-dashboard
    name: Grafana Dashboard
    parent: neo4j-monitoring
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Grafana Dashboard

KubeDB exposes Neo4j metrics by enabling Neo4j's built-in metrics support. Once Prometheus scrapes those metrics, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a Neo4j instance, and importing the Grafana dashboards.

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

> Note: YAML files used in this tutorial are stored in [docs/examples/neo4j/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/neo4j/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Configuration

> These two steps — deploying `kube-prometheus-stack` and installing Panopticon — are shared prerequisites for all KubeDB database monitoring guides. If you have already completed them in another guide, skip to [Step 1](#step-1-deploy-neo4j-with-monitoring-enabled).

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

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the Neo4j instance.

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

## Step 1: Deploy Neo4j

Below is the Neo4j object with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-grafana-demo
  namespace: demo
spec:
  version: "2025.11.2"
  replicas: 3
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

Create the Neo4j instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/monitoring/neo4j-grafana-demo.yaml
neo4j.kubedb.com/neo4j-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get neo4j -n demo neo4j-grafana-demo
NAME                 VERSION     STATUS   AGE
neo4j-grafana-demo   2025.11.2   Ready    3m
```

KubeDB creates a stats service named `{neo4j-name}-stats` for the exporter:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=neo4j-grafana-demo"
NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)            AGE
neo4j-grafana-demo         ClusterIP   10.96.10.1      <none>        7687/TCP,7474/TCP  3m
neo4j-grafana-demo-stats   ClusterIP   10.96.10.2      <none>        2004/TCP           3m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                       AGE
neo4j-grafana-demo-stats   3m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo neo4j-grafana-demo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `neo4j-grafana-demo-stats`. Its state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/neo4j/monitoring/neo4j-prom-targets.png" style="padding:10px">
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

```bash
$ helm upgrade -i kubedb-grafana-dashboards appscode/kubedb-grafana-dashboards \
    -n kubeops --create-namespace --version=v2026.8.14-rc.0 \
    --set featureGates.Neo4j=true \
    --set grafana.name=grafana \
    --set grafana.namespace=monitoring
```

`featureGates.Neo4j` already defaults to `true` — set explicitly above for clarity. `grafana.name`/`grafana.namespace` point the chart's `GrafanaDashboard` resources at the `AppBinding` created in step 2 (omitting them falls back to whichever `AppBinding` in your cluster is labeled as the cluster-default Grafana, if any — explicit is safer on a shared cluster).

This single command creates every dashboard this chart ships for Neo4j — `KubeDB / Neo4j / Summary`, `KubeDB / Neo4j / Pod`, `KubeDB / Neo4j / Database` — which `grafana-operator` then pushes into your Grafana instance automatically. No manual JSON download or upload needed.

Verify they landed:

```bash
$ kubectl get grafanadashboards.openviz.dev -n kubeops | grep -i neo4j
NAME                             TITLE                        SYNCED    AGE
grafana-kubedb-neo4j-summary     KubeDB / Neo4j / Summary     Current   30s
grafana-kubedb-neo4j-pod         KubeDB / Neo4j / Pod         Current   30s
grafana-kubedb-neo4j-database    KubeDB / Neo4j / Database    Current   30s
```

`SYNCED: Current` confirms `grafana-operator` successfully pushed each dashboard into Grafana. Open Grafana — the dashboards are already there under `Dashboards`, fully wired to your Prometheus data source, ready to explore in Step 6 below.

## Step 5: Import Dashboard — Option B: Manual (JSON upload)

If you'd rather not run `grafana-operator`, or want fine-grained control over exactly which dashboards get imported, upload the same dashboard JSON files by hand instead.

The KubeDB Neo4j dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download all three JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/neo4j) repository (`neo4j/` folder):

| File | Dashboard |
|------|-----------|
| `neo4j-summary.json` | KubeDB / Neo4j / Summary |
| `neo4j-pod.json` | KubeDB / Neo4j / Pod |
| `neo4j-database.json` | KubeDB / Neo4j / Database |

**Import steps (repeat for each of the three files):**

1. In Grafana, click **Dashboards** in the left sidebar.
2. Select **Import** from the menu.
3. Click **Upload dashboard JSON file** and select one of the downloaded `.json` files.
4. In the **Prometheus** dropdown that appears, select your Prometheus data source.
5. Click **Import**.

After importing all three files, they will appear under **Dashboards** in the left sidebar.

## Step 6: Explore the Dashboards

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable      | Applies to              | What to select                                             |
|---------------|--------------------------|-------------------------------------------------------------|
| **namespace** | All dashboards           | Namespace where your Neo4j is deployed (e.g., `demo`)      |
| **Neo4j**     | All dashboards           | Name of your instance (e.g., `neo4j-grafana-demo`)         |
| **pod**       | Pod dashboard             | A specific pod (e.g., `neo4j-grafana-demo-0`)              |
| **database**  | Pod, Database dashboards | Target database name (e.g., `neo4j`, `system`)             |
| **leader**    | Pod, Database dashboards | The pod currently holding database leadership                |

**KubeDB / Neo4j / Summary** — instance-level overview:

- **General Info** — database status, version, whether secure transport is required, deletion policy, total nodes
- **Resource Requests / Limits** — configured CPU, memory, and storage requests and limits
- **CPU Info / CPU Quota** — per-pod CPU usage over time and quota utilization

<p align="center">
  <img alt="KubeDB Neo4j Summary Dashboard" src="/docs/images/neo4j/monitoring/neo4j-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / Neo4j / Pod** — per-pod drill-down:

- **Neo4j** — status, uptime, and active Bolt connections for the selected pod
- **CPU And Memory Usage Stats** — total memory used and average CPU usage over time
- **Transactions** — last committed write transaction ID, committed transaction speed, property creation speed, rolled-back transaction speed, committed transaction rate

<p align="center">
  <img alt="KubeDB Neo4j Pod Dashboard" src="/docs/images/neo4j/monitoring/neo4j-grafana-pod.png" style="padding:10px">
</p>

**KubeDB / Neo4j / Database** — graph database metrics:

- **Neo4j** — uptime, node count, and relationship count for the selected database
- **CPU And Memory Usage Stats** — total memory used and average CPU usage over time
- **Transactions** — last committed write transaction ID, committed transaction rate, committed write transaction rate

<p align="center">
  <img alt="KubeDB Neo4j Database Dashboard" src="/docs/images/neo4j/monitoring/neo4j-grafana-database.png" style="padding:10px">
</p>

## Cleaning up

```bash
# Remove the Neo4j instance
kubectl delete neo4j -n demo neo4j-grafana-demo

# Remove namespaces
kubectl delete ns demo

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your Neo4j database with KubeDB using [built-in Prometheus](/docs/guides/neo4j/monitoring/using-builtin-prometheus.md).
- Monitor your Neo4j database with KubeDB using [Prometheus Operator](/docs/guides/neo4j/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
