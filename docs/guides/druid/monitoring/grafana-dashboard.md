---
title: Visualize Druid Metrics with Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: guides-druid-grafana-dashboard
    name: Grafana Dashboard
    parent: guides-druid-monitoring
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Visualize Druid Metrics with Grafana Dashboard

KubeDB exposes Druid metrics through a JMX Exporter running as a Java agent inside each Druid container. Once Prometheus scrapes those metrics, you can visualize them in Grafana using pre-built KubeDB dashboards. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a Druid instance, and importing the Grafana dashboards.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (Step 2) uses to expose metrics to Prometheus.

- Druid requires a deep storage backend and ZooKeeper. The example below uses S3-compatible deep storage with a pre-created secret.

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/druid/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/druid/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

Find the `serviceMonitorSelector` label that Prometheus uses to pick up `ServiceMonitor` objects. You will need this label when enabling monitoring on the Druid instance.

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

## Step 3: Deploy Druid with Monitoring Enabled

Druid requires a deep storage secret. Create a secret containing the credentials for your S3-compatible storage:

```bash
$ kubectl create secret generic druid-deep-storage-config -n demo \
  --from-literal=druid.storage.type=s3 \
  --from-literal=druid.storage.bucket=<your-bucket> \
  --from-literal=druid.s3.accessKey=<your-access-key> \
  --from-literal=druid.s3.secretKey=<your-secret-key> \
  --from-literal=druid.s3.endpoint.url=<your-s3-endpoint>
```

Below is the Druid object with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-grafana-demo
  namespace: demo
spec:
  version: "28.0.1"
  deepStorage:
    type: s3
    configSecret:
      name: druid-deep-storage-config
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

Create the Druid instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/monitoring/druid-grafana-demo.yaml
druid.kubedb.com/druid-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get druid -n demo druid-grafana-demo
NAME                 VERSION   STATUS   AGE
druid-grafana-demo   28.0.1    Ready    5m
```

KubeDB creates a stats service named `{druid-name}-stats` for monitoring:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=druid-grafana-demo"
NAME                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
druid-grafana-demo         ClusterIP   10.96.10.1     <none>        8888/TCP    5m
druid-grafana-demo-stats   ClusterIP   10.96.10.2     <none>        9101/TCP    5m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                       AGE
druid-grafana-demo-stats   5m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo druid-grafana-demo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `druid-grafana-demo-stats`. Its state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/druid/monitoring/druid-prom-targets.png" style="padding:10px">
</p>

If the target is missing, check that the `ServiceMonitor` label (`release: prometheus`) matches the Prometheus `serviceMonitorSelector`.

## Step 5: Access Grafana

Port-forward the Grafana service:

```bash
$ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
Forwarding from 127.0.0.1:3000 -> 3000
Forwarding from [::1]:3000 -> 3000
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
  <img alt="Grafana Login" src="/docs/images/druid/monitoring/druid-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/druid/monitoring/druid-grafana-home.png" style="padding:10px">
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

## Step 7: Import KubeDB Druid Dashboard

The KubeDB Druid dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download all three JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/druid) repository (`druid/` folder):

| File | Dashboard |
|------|-----------|
| `druid_summary_dashboard.json` | KubeDB / Druid / Summary |
| `druid_pods_dashboard.json` | KubeDB / Druid / Pod |
| `druid_databases_dashboard.json` | KubeDB / Druid / Database |

**Import steps (repeat for each of the three files):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

The import page looks like this — click **Upload dashboard JSON file** to select the file:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/druid/monitoring/druid-grafana-import.png" style="padding:10px">
</p>

After importing all three files, they will appear under `Dashboards` in the left sidebar.

| Dashboard Name | Description |
|---|---|
| KubeDB / Druid / Summary | Segment availability, query rate, JVM heap, CPU/memory/storage |
| KubeDB / Druid / Pod | Per-component (broker/coordinator/historical/router) JVM heap, CPU/memory |
| KubeDB / Druid / Database | Datasource-level query rate, scan rate, segment count, indexing tasks |

## Step 8: Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable      | Applies to              | What to select                                             |
|---------------|-------------------------|------------------------------------------------------------|
| **namespace** | All dashboards          | Namespace where your Druid is deployed (e.g., `demo`)     |
| **app**       | All dashboards          | Name of your Druid instance (e.g., `druid-grafana-demo`)  |
| **pod**       | Pod, Database dashboards | A specific pod, or `All` for an aggregated view          |

**KubeDB / Druid / Summary** — start here for a cluster-level overview:
- **Coordinator / Overlord / Broker / Historical / MiddleManager / Router** — pod counts and readiness per node type
- **Segment Availability** — percentage of segments available for queries (aim for 100%)
- **Task Throughput** — tasks submitted and completed per second
- **Query Latency** — broker-side p50/p95/p99 query execution time
- **CPU / Memory / JVM Heap** — resource consumption across the cluster

<p align="center">
  <img alt="KubeDB Druid Summary Dashboard" src="/docs/images/druid/monitoring/druid-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / Druid / Pod** — drill into a specific Druid node:
- **JVM Heap Used** — heap usage on this pod
- **GC Time** — time spent in garbage collection
- **CPU / Memory** — per-pod resource usage
- **Query Rate** — queries received on this pod (relevant for Brokers)
- **Task Count** — pending and running tasks (relevant for MiddleManagers)

<p align="center">
  <img alt="KubeDB Druid Pod Dashboard" src="/docs/images/druid/monitoring/druid-grafana-pod.png" style="padding:10px">
</p>

**KubeDB / Druid / Database** — segment and ingestion metrics:
- **Segment Count** — total segments across all data sources
- **Segment Size** — total deep storage usage
- **Ingestion Rate** — rows ingested per second per data source
- **Failed Tasks** — number of failed ingestion tasks
- **Query Segment Ratio** — ratio of segments used in query results

<p align="center">
  <img alt="KubeDB Druid Database Dashboard" src="/docs/images/druid/monitoring/druid-grafana-database.png" style="padding:10px">
</p>

## Cleaning up

```bash
# Remove the Druid instance
kubectl delete druid -n demo druid-grafana-demo

# Remove the deep storage secret
kubectl delete secret druid-deep-storage-config -n demo

# Remove namespaces
kubectl delete ns demo

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your Druid instance with KubeDB using [Prometheus Operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
