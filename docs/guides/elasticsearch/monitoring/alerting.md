---
title: Elasticsearch Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: es-monitoring-alerting
    name: Alerting
    parent: es-monitoring-elasticsearch
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Elasticsearch instance using the `elasticsearch-alerts` Helm chart, and how to visualise live metrics using the `kubedb-grafana-dashboards` chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in a dedicated namespace, so the alerting resources created in this tutorial stay isolated from other workloads:

  ```bash
  $ kubectl create ns alert-elasticsearch
  namespace/alert-elasticsearch created
  ```

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`. See the [Grafana Dashboard](grafana-dashboard.md#configuration) guide for how to deploy kube-prometheus-stack if you don't have it yet.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/elasticsearch/monitoring/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

The diagram below shows the full alerting architecture — from Elasticsearch metric export through to alert delivery and Grafana visualisation.

<p align="center">
  <img alt="Elasticsearch Alerting Architecture" src="/docs/images/elasticsearch/monitoring/es-alerting-overview.svg">
</p>

- **KubeDB** deploys Elasticsearch with a built-in [elasticsearch_exporter](https://github.com/prometheus-community/elasticsearch_exporter) sidecar that exposes metrics on port `56790`.
- **ServiceMonitor** (named `{elasticsearch-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `elasticsearch-alerts` chart and contains all Elasticsearch alert definitions grouped by concern: database health, provisioner, ops-manager, Stash backup/restore, and KubeStash backup/restore.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** visualises metrics through pre-built dashboards provisioned by the `kubedb-grafana-dashboards` chart.

Unlike some KubeDB databases, Elasticsearch's exporter does not publish a single boolean "is the database up" gauge. Instead, the chart watches the health signals a real Elasticsearch cluster actually exposes — JVM heap usage, filesystem usage on the data path, cluster health color (`green`/`yellow`/`red`), node/data-node counts, and shard state — and fires alerts when any of those cross a threshold.

---

## Deploy Elasticsearch with Monitoring Enabled

At first, let's deploy an Elasticsearch database with monitoring enabled. This tutorial uses a topology cluster (dedicated master, data, and ingest nodes) rather than a single-node instance, since that's representative of a real deployment and is what the rest of this guide's screenshots are taken from. Below is the Elasticsearch object we are going to create.

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-alert
  namespace: alert-elasticsearch
spec:
  version: xpack-9.2.3
  deletionPolicy: WipeOut
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
      replicas: 2
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
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here,

- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.
- `spec.topology.*.storage.storageClassName: "local-path"` — use whichever storage class is available/default in your cluster (`kubectl get storageclass`). Note that `local-path` is a `hostPath`-backed class with no capacity quota — the PVC's `1Gi` request is only used for scheduling, and the volume is really backed by however much space is free on the node's own disk. That's fine throughout this tutorial, including the [firing-alert simulation](#simulating-a-firing-alert) later, since that simulation scales the data-node count rather than filling disk.

Let's create the Elasticsearch resource.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/monitoring/es-alert.yaml
elasticsearch.kubedb.com/es-alert created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get elasticsearch -n alert-elasticsearch es-alert
NAME       VERSION       STATUS   AGE
es-alert   xpack-9.2.3   Ready    37m
```

KubeDB brings up 2 master, 2 data, and 2 ingest pods for this topology — 6 nodes total:

```bash
$ kubectl get pods -n alert-elasticsearch
NAME                READY   STATUS    RESTARTS   AGE
es-alert-data-0     2/2     Running   0          37m
es-alert-data-1     2/2     Running   0          37m
es-alert-ingest-0   2/2     Running   0          37m
es-alert-ingest-1   2/2     Running   0          37m
es-alert-master-0   2/2     Running   0          37m
es-alert-master-1   2/2     Running   0          37m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-elasticsearch --selector="app.kubernetes.io/instance=es-alert"
NAME              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
es-alert          ClusterIP   10.43.126.120   <none>        9200/TCP    37m
es-alert-master   ClusterIP   None            <none>        9300/TCP    37m
es-alert-pods     ClusterIP   None            <none>        9200/TCP    37m
es-alert-stats    ClusterIP   10.43.49.18     <none>        56790/TCP   37m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-elasticsearch
NAME             AGE
es-alert-stats   115s
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-elasticsearch es-alert-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install elasticsearch-alerts

The `elasticsearch-alerts` chart creates a `PrometheusRule` resource containing all Elasticsearch alert definitions grouped by concern: database health, provisioner, ops-manager, Stash, and KubeStash.

### Why the Helm release name matters

The chart derives the PromQL `job`/instance scoping (and the `PrometheusRule` name) from the **Helm release name**, not from a values field — so the release name must match the Elasticsearch object's name (`es-alert`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### A note on chart defaults

The chart's default `database` group rules assume a specific minimum topology: `elasticsearchHealthyNodes` and `elasticsearchHealthyDataNodes` both default to `val: 3` — i.e. "fire if fewer than 3 nodes / fewer than 3 data nodes are healthy." **These defaults do not match every topology, including this tutorial's own `es-alert` (2 master + 2 data + 2 ingest = 6 nodes total, 2 data nodes)** — confirmed live: installing with the chart's plain defaults left `ElasticsearchHealthyDataNodes` firing permanently (`2 < 3` is always true for this topology), even though the cluster was fully healthy. Always override both `val`s to match your actual node counts at install time — don't assume the defaults fit just because you're running a multi-node cluster instead of a single node.

One rule pair needs overriding regardless of topology:

- `diskUsageHigh` / `diskAlmostFull` compute PVC usage as `kubelet_volume_stats_used_bytes / (kubelet_volume_stats_used_bytes + kube_pod_spec_volumes_persistentvolumeclaims_info)`. The `..._info` series is a constant label metric (always `1`), not a byte count, so this expression evaluates to ~100% regardless of actual usage — a chart-level expression defect, confirmed live (real PVC usage was 69.6% while both alerts read ~100% and fired). We disable both and rely instead on `elasticsearchDiskOutOfSpace` / `elasticsearchDiskSpaceLow`, which are computed from the exporter's own accurate `elasticsearch_filesystem_data_available_bytes` / `elasticsearch_filesystem_data_size_bytes` metrics.

### Install

```bash
$ helm repo add appscode oci://ghcr.io/appscode-charts
$ helm repo update
$ helm search repo appscode/elasticsearch-alerts --version=v2026.7.14
NAME                         	CHART VERSION	APP VERSION	DESCRIPTION                                     
appscode/elasticsearch-alerts	v2026.7.14   	v0.7.0     	A Helm chart for Elasticsearch Alert by AppsCode

$ helm upgrade -i es-alert appscode/elasticsearch-alerts -n alert-elasticsearch --create-namespace --version=v2026.7.14 \
  --set form.alert.labels.release=prometheus \
  --set form.alert.groups.database.rules.diskUsageHigh.enabled=false \
  --set form.alert.groups.database.rules.diskAlmostFull.enabled=false \
  --set form.alert.groups.database.rules.elasticsearchHealthyNodes.val=6 \
  --set form.alert.groups.database.rules.elasticsearchHealthyDataNodes.val=2
```

| Flag | Value | Purpose |
|------|-------|---------|
| `es-alert` (release name) | — | Scopes every PromQL expression to this instance (`job="es-alert-stats"`). **This must exactly match the Elasticsearch object's name** — see [above](#why-the-helm-release-name-matters). A mismatched release name is the most common cause of alerts silently never firing (and Grafana/Prometheus showing nothing for a healthy instance): the chart's rules end up scoped to a `job` label that no target ever carries. |
| `-n alert-elasticsearch` | `alert-elasticsearch` | Installs the `PrometheusRule` in the same namespace as the database |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |
| `...diskUsageHigh.enabled` / `...diskAlmostFull.enabled` | `false` | Works around the PVC-usage expression defect described above |
| `...elasticsearchHealthyNodes.val` | `6` | Matches this tutorial's real total node count (2 master + 2 data + 2 ingest) |
| `...elasticsearchHealthyDataNodes.val` | `2` | Matches this tutorial's real data-node count |

> Whatever topology you actually deploy, set both `val`s to your real node counts — total nodes for `elasticsearchHealthyNodes`, data nodes for `elasticsearchHealthyDataNodes`. For a single-node instance that means `val: 1` for both, and you should also disable `elasticsearchUnassignedShards` (a single-node cluster can never assign a replica shard, so this rule fires permanently).

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-elasticsearch
NAME       AGE
es-alert   22s
```

Confirm the `release: prometheus` label is present.

```bash
$ kubectl get prometheusrule -n alert-elasticsearch es-alert \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI and open the **Status → Rule health** page.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules?search=elasticsearch`.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/elasticsearch/monitoring/es-alerting-prom-rules.png" style="padding:10px">
</p>

The `elasticsearch.database.alert-elasticsearch.es-alert.rules` group is visible with all rules showing **OK**, confirming that Prometheus has loaded and is evaluating the Elasticsearch alert definitions every 30 seconds.

---

## Step 2 — Install kubedb-grafana-dashboards

The `kubedb-grafana-dashboards` chart creates `GrafanaDashboard` CRDs containing pre-built Elasticsearch dashboard JSON. A separate controller, `grafana-operator`, watches these CRDs and pushes the dashboards into Grafana over its HTTP API — both pieces are required.

### Install grafana-operator

If your cluster doesn't already have it (check with `kubectl get crd grafanadashboards.openviz.dev`), install the operator that reconciles `GrafanaDashboard`/`GrafanaDatasource` objects into a real Grafana instance:

```bash
$ helm upgrade -i grafana-operator appscode/grafana-operator \
    -n kubeops --create-namespace \
    --version=v2026.6.12 \
    --wait
```

### Mark your Grafana instance as the cluster default

The chart looks up Grafana connection details from an `AppBinding` annotated as the cluster's default Grafana. If you deployed Grafana via `kube-prometheus-stack` (as in this tutorial), that `AppBinding` doesn't exist yet and must be created once per cluster:

```bash
# Create a Grafana API key (adjust the endpoint/payload shape for your Grafana version)
$ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80 &
$ GRAFANA_PW=$(kubectl get secret -n monitoring prometheus-grafana -o jsonpath='{.data.admin-password}' | base64 -d)
$ curl -s -X POST -H "Content-Type: application/json" -u admin:$GRAFANA_PW \
    http://localhost:3000/api/auth/keys \
    -d '{"name":"kubedb-dashboards","role":"Admin"}'
# Note the returned "key"
$ kill %1
```

```yaml
# grafana-appbinding.yaml
apiVersion: v1
kind: Secret
metadata:
  name: grafana-admin-token
  namespace: kubeops
type: Opaque
stringData:
  token: "<api-key-from-above>"
---
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: grafana
  namespace: kubeops
  annotations:
    monitoring.appscode.com/is-default-grafana: "true"   # must be an ANNOTATION, not a label
spec:
  type: monitoring.appscode.com/grafana
  clientConfig:
    url: "http://prometheus-grafana.monitoring.svc:80"
  secret:
    name: grafana-admin-token
```

```bash
$ kubectl apply -f grafana-appbinding.yaml
```

> **Why an AppBinding at all?** `GrafanaDashboard` objects don't carry connection details themselves — `grafana-operator` looks up the one `AppBinding` across the cluster marked with the `monitoring.appscode.com/is-default-grafana: "true"` **annotation** and uses its `clientConfig.url` + referenced `secret` (must contain a `token` key) to talk to Grafana. Skip this step only if your cluster already provisions Grafana through an Appscode-managed chart that creates this `AppBinding` automatically.

### Install the dashboards

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update appscode

$ helm template kubedb-grafana-dashboards appscode/kubedb-grafana-dashboards \
    -n kubeops \
    --version=v2026.7.10 \
    --set featureGates.Elasticsearch=true \
    --set grafana.url="http://prometheus-grafana.monitoring.svc:80" \
    --set grafana.apikey="<api-key-from-above>" \
  | kubectl apply -n kubeops -f -
```

> **Note:** The `kubedb-grafana-dashboards` chart bundles many large Grafana dashboard JSON files. Even with a single `featureGate` enabled, the rendered manifests can exceed Kubernetes' hard 1 MB Secret limit that Helm uses to store release state. To work around this, render the chart locally with `helm template` and apply the output directly with `kubectl apply`, which bypasses Helm's Secret storage entirely. Because this doesn't create a Helm release object, `helm uninstall` will not work for cleanup — use `kubectl delete` directly (see [Cleaning up](#cleaning-up)). Also note that `featureGates.<DB>` defaults to `true` for almost every database in this chart (only `Aerospike` defaults `false`), so one `helm template | kubectl apply` installs dashboards for many databases at once, not just Elasticsearch — this is expected.

### Verify dashboards are created

```bash
$ kubectl get grafanadashboards -n kubeops | grep elasticsearch
NAME                            TITLE                            STATUS    AGE
kubedb-elasticsearch-database   KubeDB / Elasticsearch / Database   Current   2m
kubedb-elasticsearch-pod        KubeDB / Elasticsearch / Pod         Current   2m
kubedb-elasticsearch-summary    KubeDB / Elasticsearch / Summary     Current   2m
```

`Current` means `grafana-operator` successfully pushed the dashboard into Grafana. If a dashboard stays `Failed` with a message like `no default Grafana appbinding found`, revisit the AppBinding step above.

---

## Verify End-to-End

### 1. Check the exporter is running

The `exporter` sidecar inside the Elasticsearch pod serves metrics at `:56790/metrics`. The `elasticsearch_cluster_health_status` series confirms the exporter can reach Elasticsearch and report cluster health.

```bash
$ kubectl exec -n alert-elasticsearch es-alert-data-0 -c exporter -- \
    wget -qO- localhost:56790/metrics | grep elasticsearch_cluster_health_status
elasticsearch_cluster_health_status{cluster="es-alert",color="green"} 1
elasticsearch_cluster_health_status{cluster="es-alert",color="red"} 0
elasticsearch_cluster_health_status{cluster="es-alert",color="yellow"} 0
```

With master, data, and ingest nodes all up, the cluster can fully assign both primary and replica shards, so it reports `green`. (A single-node cluster would instead report `yellow` — it can never assign replica shards without a second node to place them on — which is expected and not an outage.)

### 2. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-elasticsearch%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus Target UP" src="/docs/images/elasticsearch/monitoring/es-alerting-prom-target.png" style="padding:10px">
</p>

All 6 series report `up == 1` — one entry per master/data/ingest pod, confirming metrics are being scraped from every node in the `alert-elasticsearch` namespace.

### 3. Confirm all Elasticsearch alerts are inactive

Open `http://localhost:9090/alerts?search=elasticsearch` to see the Elasticsearch alert groups.

<p align="center">
  <img alt="Prometheus Alerts — All Inactive" src="/docs/images/elasticsearch/monitoring/es-alerting-prom-alerts.png" style="padding:10px">
</p>

All 6 rules in the `elasticsearch.database` group show **INACTIVE (6)**, meaning the database is healthy and no thresholds are breached.

### 4. Check AlertManager

Port-forward AlertManager to view any currently firing alerts.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`. With a healthy Elasticsearch instance, no alerts for `es-alert` will be listed here.

<p align="center">
  <img alt="AlertManager — No Active Alerts" src="/docs/images/elasticsearch/monitoring/es-alerting-alertmanager.png" style="padding:10px">
</p>

### 5. Explore Grafana dashboards

Port-forward Grafana and log in.

```bash
$ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
```

Open `http://localhost:3000` (username: `admin`). Search for **elasticsearch** in the Dashboards section.

<p align="center">
  <img alt="Grafana — Elasticsearch Dashboard List" src="/docs/images/elasticsearch/monitoring/es-alerting-grafana-dashboards.png" style="padding:10px">
</p>

Three pre-built dashboards are available. The `Namespace` and `app` drop-downs at the top of each dashboard let you switch between instances.

**KubeDB / Elasticsearch / Summary** — cluster-wide health: database status, version, node count, CPU/memory/storage requests vs. usage.

<p align="center">
  <img alt="Grafana — KubeDB Elasticsearch Summary" src="/docs/images/elasticsearch/monitoring/es-alerting-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / Elasticsearch / Pod** — per-node detail: node status color, open file count, connected/active data nodes, memory and heap usage, GC time, documents indexed.

<p align="center">
  <img alt="Grafana — KubeDB Elasticsearch Pod" src="/docs/images/elasticsearch/monitoring/es-alerting-grafana-pod.png" style="padding:10px">
</p>

**KubeDB / Elasticsearch / Database** — cluster status, shard counts, documents indexed, index size, indexing/query rate, and ingest-node system metrics.

<p align="center">
  <img alt="Grafana — KubeDB Elasticsearch Database" src="/docs/images/elasticsearch/monitoring/es-alerting-grafana-database.png" style="padding:10px">
</p>

---

## Simulating a Firing Alert

The previous section confirmed that all alerts are **INACTIVE** while the database is healthy. This section walks through deliberately triggering `ElasticsearchHealthyDataNodes` — along with two alerts it drags along with it, see the note below — so you can observe the full alert lifecycle and then resolve it.

Elasticsearch doesn't have a single "process down" style alert the way some other databases do — its exporter reports live cluster metrics rather than a boolean liveness gauge. Killing the `elasticsearch` process inside a pod doesn't work either: the container restarts in under 2 seconds (faster than the cluster's fault-detection window), so the other nodes never actually perceive the node as gone. Instead, we shrink the `data` role from 2 nodes to 1 — a real, sustained, cleanly-reversible change that reliably crosses this tutorial's `elasticsearchHealthyDataNodes.val: 2` threshold set in [Step 1](#install).

### 1. Scale down the data nodes

```bash
$ kubectl patch elasticsearch -n alert-elasticsearch es-alert \
    --type=merge -p '{"spec":{"topology":{"data":{"replicas":1}}}}'
elasticsearch.kubedb.com/es-alert patched
```

KubeDB terminates one data pod to bring the topology down to the new desired count:

```bash
$ kubectl get pods -n alert-elasticsearch -l app.kubernetes.io/instance=es-alert
NAME                READY   STATUS    RESTARTS   AGE
es-alert-data-0     2/2     Running   0          37m
es-alert-ingest-0   2/2     Running   0          37m
es-alert-ingest-1   2/2     Running   0          37m
es-alert-master-0   2/2     Running   0          37m
es-alert-master-1   2/2     Running   0          37m
```

Wait 30–60 seconds for the next Prometheus scrape cycle (configured at 10 s) and rule-evaluation cycle (30 s) to register the smaller data-node count.

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts?search=elasticsearch`.

<p align="center">
  <img alt="Prometheus Alerts — ElasticsearchHealthyDataNodes Firing" src="/docs/images/elasticsearch/monitoring/es-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

Dropping to 5 total nodes (1 data + 2 master + 2 ingest) crosses **three** thresholds at once, confirmed live — all `for: instant`, so all three move directly from **INACTIVE** to **FIRING** within one evaluation cycle, while the rest of the `elasticsearch.database` group stays **INACTIVE**:

- `ElasticsearchHealthyDataNodes` — data-node count (1) is below `val: 2`.
- `ElasticsearchHealthyNodes` — total node count (5) is below `val: 6`.
- `ElasticsearchUnassignedShards` — with only one data node left, replica shards have nowhere to be placed.

Each fires once per surviving node's exporter (5 series each here — one per remaining pod), since every node independently reports its own view of cluster-wide state; that's 15 alert instances in total, not 15 separate incidents.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093/#/alerts?filter={namespace="alert-elasticsearch"}`.

<p align="center">
  <img alt="AlertManager — ElasticsearchHealthyDataNodes Firing" src="/docs/images/elasticsearch/monitoring/es-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows all three alerts grouped by namespace (15 alerts total). Each alert card displays:

- **Severity**: `critical`
- **app** / **job**: `es-alert` / `es-alert-stats`
- **pod**: the surviving node reporting the condition (e.g. `es-alert-data-0`, `es-alert-master-1`, ...)
- **Started**: timestamp when the alert first fired

AlertManager routes these alerts to every receiver configured in your `alertmanagerConfig` (Slack, email, PagerDuty, webhook, etc.) based on your routing tree. If no receiver is configured, the alerts are visible here but silently dropped.

### 4. Restore the data nodes

Scale the data role back to 2 to resolve the alert.

```bash
$ kubectl patch elasticsearch -n alert-elasticsearch es-alert \
    --type=merge -p '{"spec":{"topology":{"data":{"replicas":2}}}}'
elasticsearch.kubedb.com/es-alert patched
```

Wait for the pod to rejoin and for the next scrape cycle to register the recovered count.

```bash
$ kubectl get elasticsearch -n alert-elasticsearch es-alert
NAME       VERSION       STATUS   AGE
es-alert   xpack-9.2.3   Ready    41m
```

Once the Elasticsearch resource returns to `Ready` and `elasticsearch_cluster_health_number_of_data_nodes` reports `2` again, Prometheus marks all three alerts **INACTIVE** and AlertManager sends **resolved** notifications to all receivers.

---

## Alert Reference

All alerts are scoped to the `es-alert` instance in the `alert-elasticsearch` namespace via the PromQL label filters `job="es-alert-stats"` and `namespace="alert-elasticsearch"`.

### Database Group

Fired based on live metrics from the Elasticsearch exporter.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `ElasticsearchHeapUsageTooHigh` | critical | 2m | The JVM heap usage is over 90%. |
| `ElasticsearchHeapUsageWarning` | warning | 2m | The JVM heap usage is over 80%. |
| `ElasticsearchDiskOutOfSpace` | critical | instant | The disk usage is over 90%. |
| `ElasticsearchDiskSpaceLow` | warning | 2m | The disk usage is over 80%. |
| `ElasticsearchClusterRed` | critical | instant | Elastic Cluster Red status — one or more primary shards are not allocated. |
| `ElasticsearchClusterYellow` | warning | instant | Elastic Cluster Yellow status — one or more replica shards are not allocated. |
| `ElasticsearchHealthyNodes` | critical | instant | Fewer than the configured minimum number of nodes are healthy in the cluster (default `val: 3`; this tutorial overrides it to `6` — see [Step 1](#install)). |
| `ElasticsearchHealthyDataNodes` | critical | instant | Fewer than the configured minimum number of data nodes are healthy in the cluster (default `val: 3`; this tutorial overrides it to `2`). |
| `ElasticsearchRelocatingShards` | info | instant | Elasticsearch is relocating shards. |
| `ElasticsearchInitializingShards` | info | instant | Elasticsearch is initializing shards. |
| `ElasticsearchUnassignedShards` | critical | instant | Elasticsearch has unassigned shards. |
| `ElasticsearchPendingTasks` | warning | 15m | Elasticsearch has pending tasks — the cluster is working slowly. |
| `ElasticsearchNoNewDocuments10m` | info | instant | No new documents were indexed in the last 10 minutes (disabled by default). |
| `DiskUsageHigh` | warning | 1m | **Disabled by the install command above** — broken denominator always reads ~100% usage regardless of real usage (confirmed: real usage 69.6% while this alert read ~100%). |
| `DiskAlmostFull` | critical | 1m | **Disabled by the install command above** — same broken-denominator bug as `DiskUsageHigh`. |

### Provisioner Group

Monitors the KubeDB operator's view of the Elasticsearch resource phase.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBElasticsearchPhaseNotReady` | critical | 1m | KubeDB marked the Elasticsearch resource `NotReady` — operator cannot reach the database. |
| `KubeDBElasticsearchPhaseCritical` | warning | 15m | The instance is in a degraded/critical phase. |

### OpsManager Group

Tracks `ElasticsearchOpsRequest` lifecycle during upgrades, scaling, and reconfiguration.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBElasticsearchOpsRequestOnProgress` | info | instant | An ops request is currently in progress. |
| `KubeDBElasticsearchOpsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes — likely stuck. |
| `KubeDBElasticsearchOpsRequestFailed` | critical | instant | An ops request failed — check the `ElasticsearchOpsRequest` object for the error. |

### Stash Group

Tracks backup/restore health for Elasticsearch instances backed up with [Stash](https://stash.run/).

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `ElasticsearchStashBackupSessionFailed` | critical | instant | The most recent Stash backup session failed. |
| `ElasticsearchStashRestoreSessionFailed` | critical | instant | The most recent Stash restore session failed. |
| `ElasticsearchStashNoBackupSessionForTooLong` | warning | instant | No successful backup session in the last 18000s (5 hours). |
| `ElasticsearchStashRepositoryCorrupted` | critical | 5m | The Stash backup repository failed its integrity check. |
| `ElasticsearchStashRepositoryStorageRunningLow` | warning | 5m | The Stash repository has grown beyond 10 GB. |
| `ElasticsearchStashBackupSessionPeriodTooLong` | warning | instant | A backup session took longer than 1800s (30 minutes) to complete. |
| `ElasticsearchStashRestoreSessionPeriodTooLong` | warning | instant | A restore session took longer than 1800s (30 minutes) to complete. |

### KubeStash Group

Tracks backup/restore health for Elasticsearch instances backed up with [KubeStash](https://kubestash.com/).

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `ElasticsearchKubeStashBackupSessionFailed` | critical | instant | The most recent KubeStash backup session failed. |
| `ElasticsearchKubeStashRestoreSessionFailed` | critical | instant | The most recent KubeStash restore session failed. |
| `ElasticsearchKubeStashNoBackupSessionForTooLong` | warning | instant | No successful backup session in the last 18000s (5 hours). |
| `ElasticsearchKubeStashRepositoryCorrupted` | critical | 5m | The KubeStash repository failed its integrity check. |
| `ElasticsearchKubeStashRepositoryStorageRunningLow` | warning | 5m | The KubeStash repository has grown beyond 10 GB. |
| `ElasticsearchKubeStashBackupSessionPeriodTooLong` | warning | instant | A backup session took longer than 1800s (30 minutes) to complete. |
| `ElasticsearchKubeStashRestoreSessionPeriodTooLong` | warning | instant | A restore session took longer than 1800s (30 minutes) to complete. |

> Stash and KubeStash alerts are only relevant if you've configured backups for this Elasticsearch instance. This tutorial doesn't set up backups — the tables above are included so you know what's available in the chart if you do.

---

## Customising Alerts

To override thresholds or disable specific alert groups, create a custom values file and upgrade the chart.

```yaml
# custom-alerts.yaml
form:
  alert:
    labels:
      release: prometheus
    groups:
      database:
        enabled: warning
        rules:
          elasticsearchHeapUsageWarning:
            enabled: true
            duration: "5m"
            val: 70        # fire at 70% heap usage instead of the default 80%
            severity: warning
      opsManager:
        enabled: "none"    # disable all ops-manager alerts
```

```bash
$ helm upgrade es-alert oci://ghcr.io/appscode-charts/elasticsearch-alerts \
    -n alert-elasticsearch \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the Grafana dashboards (installed via helm template | kubectl apply, not helm install)
$ helm template kubedb-grafana-dashboards appscode/kubedb-grafana-dashboards \
    -n kubeops \
    --version=v2026.7.10 \
    --set featureGates.Elasticsearch=true \
    --set grafana.url="http://prometheus-grafana.monitoring.svc:80" \
    --set grafana.apikey="<api-key>" \
  | kubectl delete -n kubeops -f - --ignore-not-found

# Remove the elasticsearch-alerts release
$ helm uninstall es-alert -n alert-elasticsearch

# Remove the Elasticsearch instance
$ kubectl delete elasticsearch -n alert-elasticsearch es-alert

# Delete namespace
$ kubectl delete ns alert-elasticsearch

# Optional: only if nothing else in the cluster depends on them
$ kubectl delete appbinding -n kubeops grafana
$ kubectl delete secret -n kubeops grafana-admin-token
$ helm uninstall grafana-operator -n kubeops
```

## Next Steps

- Monitor your Elasticsearch database with KubeDB using [builtin Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Visualise Elasticsearch metrics with [Grafana Dashboard](grafana-dashboard.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
