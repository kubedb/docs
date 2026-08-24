---
title: Elasticsearch Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: es-grafana-dashboard-monitoring
    name: Grafana Dashboard
    parent: es-monitoring-elasticsearch
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Grafana Dashboard

KubeDB exposes Elasticsearch metrics through a sidecar exporter, and its own view of each resource (status, phase, version) through Panopticon. Once Prometheus is scraping both, you can visualize them in Grafana using a pre-built KubeDB dashboard. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on an Elasticsearch instance, and importing the Grafana dashboard.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (see [Configuration](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md#configuration)) uses to expose metrics to Prometheus.

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `grafana-es` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns grafana-es
  namespace/grafana-es created
  ```

* Before proceeding, complete the [Configuration](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md#configuration) steps to deploy **kube-prometheus-stack** and **Panopticon**.

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Setup

## Step 1: Deploy Elasticsearch

Below is the Elasticsearch object with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-grafana-topo
  namespace: grafana-es
spec:
  version: "xpack-9.2.3"
  topology:
    master:
      replicas: 2
      storage:
        storageClassName: "local-path"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 3
      storage:
        storageClassName: "local-path"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 2
      storage:
        storageClassName: "local-path"
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

- `monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` for this instance.
- `monitor.prometheus.serviceMonitor.labels` must match the `serviceMonitorSelector` label of your Prometheus (`release: prometheus`).
- `monitor.prometheus.serviceMonitor.interval` sets the scrape interval to 10 seconds.

Create the Elasticsearch instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/monitoring/coreos-prom-es.yaml
elasticsearch.kubedb.com/es-grafana-topo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get elasticsearch -n grafana-es es-grafana-topo
NAME              VERSION       STATUS   AGE
es-grafana-topo   xpack-9.2.3   Ready    83m
```

KubeDB creates a stats service named `{elasticsearch-name}-stats` for the exporter:

```bash
$ kubectl get svc -n grafana-es --selector="app.kubernetes.io/instance=es-grafana-topo"
NAME                     TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
es-grafana-topo          ClusterIP   10.43.201.71   <none>        9200/TCP    84m
es-grafana-topo-master   ClusterIP   None           <none>        9300/TCP    84m
es-grafana-topo-pods     ClusterIP   None           <none>        9200/TCP    84m
es-grafana-topo-stats    ClusterIP   10.43.18.99    <none>        56790/TCP   84m
```

KubeDB also creates a `ServiceMonitor` in the `grafana-es` namespace:

```bash
$ kubectl get servicemonitor -n grafana-es
NAME                    AGE
es-grafana-topo-stats   3m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n grafana-es es-grafana-topo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `es-grafana-topo-stats`. Its state should be **UP**.

If the target is missing, check that the `ServiceMonitor` label (`release: prometheus`) matches the Prometheus `serviceMonitorSelector`.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/elasticsearch/monitoring/es-prom-targets.png" style="padding:10px">
</p>

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
  <img alt="Grafana Login" src="/docs/images/elasticsearch/monitoring/es-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/elasticsearch/monitoring/es-grafana-home.png" style="padding:10px">
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
    --set featureGates.Elasticsearch=true \
    --set grafana.name=grafana \
    --set grafana.namespace=monitoring
```

`featureGates.Elasticsearch` already defaults to `true` — set explicitly above for clarity. `grafana.name`/`grafana.namespace` point the chart's `GrafanaDashboard` resources at the `AppBinding` created in step 2 (omitting them falls back to whichever `AppBinding` in your cluster is labeled as the cluster-default Grafana, if any — explicit is safer on a shared cluster).

This single command creates every dashboard this chart ships for Elasticsearch — `KubeDB / Elasticsearch / Summary`, `KubeDB / Elasticsearch / Pod`, `KubeDB / Elasticsearch / Database` — which `grafana-operator` then pushes into your Grafana instance automatically. No manual JSON download or upload needed.

Verify they landed:

```bash
$ kubectl get grafanadashboards.openviz.dev -n kubeops | grep -i elasticsearch
NAME                                     TITLE                                SYNCED    AGE
grafana-kubedb-elasticsearch-summary     KubeDB / Elasticsearch / Summary     Current   30s
grafana-kubedb-elasticsearch-pod         KubeDB / Elasticsearch / Pod         Current   30s
grafana-kubedb-elasticsearch-database    KubeDB / Elasticsearch / Database    Current   30s
```

`SYNCED: Current` confirms `grafana-operator` successfully pushed each dashboard into Grafana. Open Grafana — the dashboard are already there under `Dashboards`, fully wired to your Prometheus data source, ready to explore in Step 6 below.

## Step 5: Import Dashboard — Option B: Manual (JSON upload)

If you'd rather not run `grafana-operator`, or want fine-grained control over exactly which dashboards get imported, upload the same dashboard JSON files by hand instead.

The KubeDB Elasticsearch dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download all three JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/elasticsearch) repository (`elasticsearch/` folder):

| File | Dashboard |
|------|-----------|
| `elasticsearch_summary_dashboard.json` | KubeDB / Elasticsearch / Summary |
| `elasticsearch_pods_dashboard.json` | KubeDB / Elasticsearch / Pod |
| `elasticsearch_databases_dashboard.json` | KubeDB / Elasticsearch / Database |

**Import steps (repeat for each of the three files):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

The import page looks like this — click **Upload dashboard JSON file** to select the file:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/elasticsearch/monitoring/es-grafana-import.png" style="padding:10px">
</p>

After importing all three files, they will appear under `Dashboards` in the left sidebar.

| Dashboard Name                        | Description                                                                              |
|---------------------------------------|------------------------------------------------------------------------------------------|
| KubeDB / Elasticsearch / Summary      | Cluster health, shard status, JVM heap usage, CPU/memory/storage, network                |
| KubeDB / Elasticsearch / Pod          | Per-node JVM heap, GC time, thread pool queues and rejections, CPU/memory usage          |
| KubeDB / Elasticsearch / Database     | Index-level indexing rate, search rate, search latency, field data cache, segment count  |

## Step 6: Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable      | Applies to              | What to select                                                 |
|---------------|-------------------------|----------------------------------------------------------------|
| **namespace** | All dashboards          | Namespace where your Elasticsearch is deployed (e.g., `grafana-es`) |
| **app**       | All dashboards          | Name of your Elasticsearch instance (e.g., `es-grafana-topo`) |
| **pod**       | Pod, Database dashboards | A specific pod, or `All` for an aggregated view              |
| **index**     | Database dashboard only | A specific index, or `All`                                    |

**KubeDB / Elasticsearch / Summary** — start here for a cluster health overview:
- **Cluster Health** — green/yellow/red status, active shards, relocating shards, unassigned shards
- **Node Count** — total nodes in the cluster
- **JVM Heap Used %** — aggregate heap usage across all nodes
- **CPU / Memory / Storage** — resource consumption vs. requests and limits
- **Network** — receive and transmit bandwidth

<p align="center">
  <img alt="KubeDB Elasticsearch Summary Dashboard" src="/docs/images/elasticsearch/monitoring/es-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / Elasticsearch / Pod** — drill into a specific node:
- **JVM Heap** — used vs. max heap per node
- **GC Collection Time** — time spent in young/old generation GC
- **Thread Pool** — queue size and rejected count per thread pool (search, index, bulk)
- **CPU / Memory** — per-pod resource usage over time

<p align="center">
  <img alt="KubeDB Elasticsearch Pod Dashboard" src="/docs/images/elasticsearch/monitoring/es-grafana-pod.png" style="padding:10px">
</p>

**KubeDB / Elasticsearch / Database** — index-level metrics:
- **Indexing Rate** — documents indexed per second
- **Search Rate** — queries executed per second
- **Search Latency** — p50/p95/p99 query latency
- **Field Data Cache Evictions** — high eviction rates indicate memory pressure
- **Segment Count** — number of Lucene segments per index

<p align="center">
  <img alt="KubeDB Elasticsearch Database Dashboard" src="/docs/images/elasticsearch/monitoring/es-grafana-database.png" style="padding:10px">
</p>
## Cleaning up

```bash
# Remove the Elasticsearch instance
kubectl delete elasticsearch -n grafana-es es-grafana-topo

# Remove namespaces
kubectl delete ns grafana-es

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your Elasticsearch database with KubeDB using [built-in Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [Prometheus Operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
