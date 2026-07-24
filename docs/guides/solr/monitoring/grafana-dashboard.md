---
title: Visualize Solr Metrics with Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: guides-sl-grafana-dashboard
    name: Grafana Dashboard
    parent: sl-monitoring-solr
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Visualize Solr Metrics with Grafana Dashboard

KubeDB exposes Solr metrics through the built-in Prometheus exporter module. Once Prometheus scrapes those metrics, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a Solr instance, and importing the Grafana dashboards.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (Step 2) uses to expose metrics to Prometheus.

- Solr requires a ZooKeeper cluster for coordination. Deploy a ZooKeeper instance in the `demo` namespace before creating the Solr instance. See the [ZooKeeper quickstart](/docs/guides/zookeeper/quickstart/quickstart.md) for details.

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/solr/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/solr/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Configuration

> These two steps — deploying `kube-prometheus-stack` and installing Panopticon — are shared prerequisites for all KubeDB database monitoring guides. If you have already completed them in another guide, skip to [Step 1](#step-1-deploy-solr-with-monitoring-enabled).

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

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the Solr instance.

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

## Step 1: Deploy Solr with Monitoring Enabled

Below is the Solr object with monitoring configured to use Prometheus Operator. The `prometheus-exporter` module must be listed in `solrModules` to enable the metrics endpoint.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-grafana-demo
  namespace: demo
spec:
  version: "9.8.0"
  replicas: 1
  solrModules:
    - prometheus-exporter
  zookeeperRef:
    name: zk-grafana-demo
    namespace: demo
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
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

- `solrModules: [prometheus-exporter]` enables Solr's built-in Prometheus metrics module.
- `monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` for this instance.
- `monitor.prometheus.serviceMonitor.labels` must match the `serviceMonitorSelector` label of your Prometheus (`release: prometheus`).

Create the Solr instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/monitoring/solr-grafana-demo.yaml
solr.kubedb.com/solr-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get solr -n demo solr-grafana-demo
NAME                VERSION   STATUS   AGE
solr-grafana-demo   9.8.0     Ready    3m
```

KubeDB creates a stats service named `{solr-name}-stats` for the exporter:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=solr-grafana-demo"
NAME                     TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
solr-grafana-demo        ClusterIP   10.96.10.1    <none>        8983/TCP   3m
solr-grafana-demo-stats  ClusterIP   10.96.10.2    <none>        9854/TCP   3m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                     AGE
solr-grafana-demo-stats  3m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo solr-grafana-demo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `solr-grafana-demo-stats`. Its state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/solr/monitoring/sl-prom-targets.png" style="padding:10px">
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
  <img alt="Grafana Login" src="/docs/images/kafka/monitoring/kf-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/kafka/monitoring/kf-grafana-home.png" style="padding:10px">
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

## Step 5: Import KubeDB Solr Dashboards

The KubeDB Solr dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download all three JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/solr) repository (`solr/` folder):

| File | Dashboard |
|------|-----------|
| `solr_summary_dashboard.json` | KubeDB / Solr / Summary |
| `solr_pods_dashboard.json` | KubeDB / Solr / Pod |
| `solr_databases_dashboard.json` | KubeDB / Solr / Database |

**Import steps (repeat for each of the three files):**

1. In Grafana, click **Dashboards** in the left sidebar.
2. Select **Import** from the menu.
3. Click **Upload dashboard JSON file** and select one of the downloaded `.json` files.
4. In the **Prometheus** dropdown that appears, select your Prometheus data source.
5. Click **Import**.

The import page looks like this:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/kafka/monitoring/kf-grafana-import.png" style="padding:10px">
</p>

After importing all three files, they will appear under **Dashboards** in the left sidebar.

## Step 6: Explore the Dashboards

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable      | Applies to     | What to select                                           |
|---------------|----------------|----------------------------------------------------------|
| **namespace** | All dashboards | Namespace where your Solr is deployed (e.g., `demo`)    |
| **app**       | All dashboards | Name of your instance (e.g., `solr-grafana-demo`)       |
| **pod**       | Pod dashboard  | A specific pod, or `All` for an aggregated view         |

**KubeDB / Solr / Summary** — cluster-level overview:
- **Database Status** — current health of the Solr cluster
- **Version** — Solr version running
- **Queries per Second** — total query throughput across the cluster
- **Update Rate** — documents added and deleted per second
- **Cache Hit Rates** — query result cache and filter cache effectiveness
- **CPU / Memory** — resource usage over time

<p align="center">
  <img alt="KubeDB Solr Summary Dashboard" src="/docs/images/solr/monitoring/sl-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / Solr / Pod** — per-node drill-down:
- **Uptime** — how long this Solr node has been running
- **JVM Heap** — heap used vs. max on this node
- **GC Activity** — garbage collection pause frequency and duration
- **Requests on Pod** — queries and updates handled by this node
- **CPU / Memory** — per-pod resource usage

<p align="center">
  <img alt="KubeDB Solr Pod Dashboard" src="/docs/images/solr/monitoring/sl-grafana-pod.png" style="padding:10px">
</p>

**KubeDB / Solr / Database** — collection-level metrics:
- **Document Count** — total documents per collection
- **Index Size** — disk space used by each collection's index
- **Shard Distribution** — documents and replicas across shards
- **Merge Activity** — index segment merges in progress

<p align="center">
  <img alt="KubeDB Solr Database Dashboard" src="/docs/images/solr/monitoring/sl-grafana-database.png" style="padding:10px">
</p>

## Cleaning up

```bash
# Remove the Solr instance
kubectl delete solr -n demo solr-grafana-demo

# Remove namespaces
kubectl delete ns demo

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your Solr cluster with KubeDB using [built-in Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md).
- Monitor your Solr cluster with KubeDB using [Prometheus Operator](/docs/guides/solr/monitoring/prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
