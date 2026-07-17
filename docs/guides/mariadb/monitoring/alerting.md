---
title: MariaDB Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-monitoring-alerting
    name: Alerting
    parent: guides-mariadb-monitoring
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MariaDB Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed MariaDB instance using the `mariadb-alerts` Helm chart, and how to visualise live metrics using the `kubedb-grafana-dashboards` chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-mariadb` namespace:

  ```bash
  $ kubectl create ns alert-mariadb
  namespace/alert-mariadb created
  ```

* Before proceeding, complete the [Configuration](grafana-dashboard.md#configuration) steps to deploy **kube-prometheus-stack** and **Panopticon**.

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/mariadb/monitoring/overview/index.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/mariadb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mariadb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

The diagram below shows the full alerting architecture — from MariaDB metric export through to alert delivery and Grafana visualisation.

<p align="center">
  <img alt="MariaDB Alerting Architecture" src="/docs/images/mariadb/monitoring/mariadb-alerting-overview.svg">
</p>

- **KubeDB** deploys MariaDB with a `mysqld_exporter`-compatible sidecar (container `exporter`) that exposes metrics used by both MySQL and MariaDB alert charts (`mysql_*` metric names).
- **ServiceMonitor** (named `{mariadb-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `mariadb-alerts` chart and contains MariaDB alert definitions grouped by concern: database health, Galera cluster, provisioner, ops-manager, Stash backup/restore, KubeStash backup/restore, and schema manager.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** visualises metrics through pre-built dashboards provisioned by the `kubedb-grafana-dashboards` chart.

---

## Deploy MariaDB with Monitoring Enabled

Below is the MariaDB object we are going to create — a 3-node Galera cluster (Primary-Primary multi-master replication) with monitoring enabled. This tutorial uses a real Galera cluster rather than a standalone instance since that's representative of a real deployment and is what the rest of this guide's screenshots are taken from — the `cluster` group's `GaleraReplicationLatencyTooLong` alert only produces real data on a Galera topology; a standalone instance simply leaves it permanently INACTIVE with no series at all.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-alert-demo
  namespace: alert-mariadb
spec:
  version: "12.1.2"
  deletionPolicy: WipeOut
  replicas: 3
  topology:
    mode: GaleraCluster
  wsrepSSTMethod: rsync
  storageType: Durable
  storage:
    storageClassName: "longhorn"
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

- `spec.topology.mode: GaleraCluster` tells KubeDB to bootstrap a multi-master Galera cluster instead of a standalone/async-replica instance.
- `spec.wsrepSSTMethod: rsync` selects the State Snapshot Transfer method Galera uses to bring a rejoining node's dataset back in sync with the cluster.
- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/monitoring/mariadb-alert-demo.yaml
mariadb.kubedb.com/mariadb-alert-demo created
```

Wait for the database to go into `Ready` state.

```bash
$ kubectl get mariadb -n alert-mariadb mariadb-alert-demo
NAME                 VERSION   STATUS   AGE
mariadb-alert-demo   12.1.2    Ready    21h
```

KubeDB brings up 3 Galera pods, each running as a Primary:

```bash
$ kubectl get pods -n alert-mariadb
NAME                   READY   STATUS    RESTARTS   AGE
mariadb-alert-demo-0   3/3     Running   0          21h
mariadb-alert-demo-1   3/3     Running   0          21h
mariadb-alert-demo-2   3/3     Running   0          21h
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-mariadb --selector="app.kubernetes.io/instance=mariadb-alert-demo"
NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
mariadb-alert-demo         ClusterIP   10.43.111.157   <none>        3306/TCP    21h
mariadb-alert-demo-pods    ClusterIP   None            <none>        3306/TCP    21h
mariadb-alert-demo-stats   ClusterIP   10.43.30.195    <none>        56790/TCP   21h
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-mariadb
NAME                       AGE
mariadb-alert-demo-stats   21h
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-mariadb mariadb-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install mariadb-alerts

The `mariadb-alerts` chart creates a `PrometheusRule` resource containing all MariaDB alert definitions.

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression (via `job="{release-name}-stats"` / `app="{release-name}"`) from the **Helm release name** — so the release name must match the MariaDB object's name (`mariadb-alert-demo`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### A note on chart defaults

`mariadb-alerts` has one chart-level bug worth knowing about before you install: `diskUsageHigh` / `diskAlmostFull` compute PVC usage as `kubelet_volume_stats_used_bytes / (kubelet_volume_stats_used_bytes + kube_pod_spec_volumes_persistentvolumeclaims_info)`. The `..._info` series is a constant label metric (always `1`), not a byte count, so this expression evaluates to ~100% regardless of actual usage — a chart-level expression defect (the same one documented for several other `*-alerts` charts in this project). Confirmed on this instance: `df -h /var/lib/mysql` showed real usage at **31%**, but both alerts were firing permanently because the broken expression read them as ~100%. Unlike Elasticsearch's `elasticsearch-alerts` chart, `mariadb-alerts` has **no accurate alternative** disk-space rule to fall back on — so the fix here is simply to disable both.

### Install

```bash
$ helm upgrade -i mariadb-alert-demo oci://ghcr.io/appscode-charts/mariadb-alerts \
    -n alert-mariadb \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus \
    --set form.alert.groups.database.rules.diskUsageHigh.enabled=false \
    --set form.alert.groups.database.rules.diskAlmostFull.enabled=false
```

| Flag | Value | Purpose |
|------|-------|---------|
| `mariadb-alert-demo` (release name) | — | Scopes every PromQL expression to this instance. **This must exactly match the MariaDB object's name** — see [above](#why-the-helm-release-name-matters). |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |
| `...diskUsageHigh.enabled` / `...diskAlmostFull.enabled` | `false` | Works around the PVC-usage expression defect described above |

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-mariadb
NAME                 AGE
mariadb-alert-demo   30s

$ kubectl get prometheusrule -n alert-mariadb mariadb-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules?search=mariadb` and locate the `mariadb.database`, `mariadb.cluster`, `mariadb.provisioner`, `mariadb.opsManager`, `mariadb.stash`, `mariadb.kubeStash`, and `mariadb.schemaManager` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/mariadb/monitoring/mariadb-alerting-prom-rules.png" style="padding:10px">
</p>

All groups show **OK**, confirming that Prometheus has loaded and is evaluating the MariaDB alert definitions every 30 seconds. Unlike several other `*-alerts` charts in this project, `mariadb-alerts` v2026.7.14 renders every group declared in its `values.yaml` — no missing-group gap found here. Note `mariadb.database` now has 11 rules, not 13 — `diskUsageHigh`/`diskAlmostFull` are gone entirely rather than merely disabled-and-hidden, since a disabled rule isn't rendered into the `PrometheusRule` at all.

---

## Step 2 — Install kubedb-grafana-dashboards

The `kubedb-grafana-dashboards` chart creates `GrafanaDashboard` CRDs containing pre-built MariaDB dashboard JSON. A separate controller, `grafana-operator`, watches these CRDs and pushes the dashboards into Grafana over its HTTP API — both pieces are required. If you've already set these up for another database on this cluster (see the [Elasticsearch alerting guide](/docs/guides/elasticsearch/monitoring/alerting.md) for the full walkthrough), skip straight to [Install the dashboards](#install-the-dashboards) below.

### Install grafana-operator

If your cluster doesn't already have it (check with `kubectl get crd grafanadashboards.openviz.dev`):

```bash
$ helm upgrade -i grafana-operator appscode/grafana-operator \
    -n kubeops --create-namespace \
    --version=v2026.6.12 \
    --wait
```

### Mark your Grafana instance as the cluster default

Skip this if you already have a Grafana `AppBinding` annotated as the cluster default (one is shared across every database). Otherwise:

```bash
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

### Install the dashboards

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update appscode

$ helm template kubedb-grafana-dashboards appscode/kubedb-grafana-dashboards \
    -n kubeops \
    --version=v2026.7.10 \
    --set featureGates.MariaDB=true \
    --set grafana.url="http://prometheus-grafana.monitoring.svc:80" \
    --set grafana.apikey="<api-key-from-above>" \
  | kubectl apply -n kubeops -f -
```

> **Note:** `featureGates.<DB>` defaults to `true` for almost every database in this chart, so one `helm template | kubectl apply` installs dashboards for many databases at once, not just MariaDB — this is expected. See the render-vs-Secret-size caveat in the [Elasticsearch alerting guide](/docs/guides/elasticsearch/monitoring/alerting.md#install-the-dashboards) for why `helm template | kubectl apply` is used instead of `helm install`.

### Verify dashboards are created

```bash
$ kubectl get grafanadashboards -n kubeops | grep mariadb
NAME                            TITLE                          STATUS    AGE
kubedb-mariadb-database         KubeDB / MariaDB / Database     Current   2m
kubedb-mariadb-galera-cluster   KubeDB / MariaDB / Galera-Cluster   Current   2m
kubedb-mariadb-pod              KubeDB / MariaDB / Pod          Current   2m
kubedb-mariadb-summary          KubeDB / MariaDB / Summary      Current   2m
```

Four dashboards this time, not three — MariaDB's chart ships a dedicated **Galera-Cluster** dashboard alongside the usual Summary/Pod/Database triplet.

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-mariadb%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — all 3 mariadb-alert-demo pods UP" src="/docs/images/mariadb/monitoring/mariadb-alerting-prom-target.png" style="padding:10px">
</p>

All 3 pods (`mariadb-alert-demo-0/1/2`) should report `up == 1` via the `mariadb-alert-demo-stats` service/job.

### 2. Confirm the MariaDB alerts are inactive

Open `http://localhost:9090/alerts?search=mariadb`.

<p align="center">
  <img alt="Prometheus Alerts — MariaDB groups inactive" src="/docs/images/mariadb/monitoring/mariadb-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules show **INACTIVE**, including `GaleraReplicationLatencyTooLong` (the `cluster` group) — on a real Galera topology this rule has live data (unlike a standalone instance, where it would have none at all), it's just currently below threshold.

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/mariadb/monitoring/mariadb-alerting-alertmanager.png" style="padding:10px">
</p>

No alerts should be firing for the `alert-mariadb` namespace.

### 4. Explore Grafana dashboards

Port-forward Grafana and log in.

```bash
$ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
```

Open `http://localhost:3000` (username: `admin`). Search for **mariadb** in the Dashboards section.

<p align="center">
  <img alt="Grafana — MariaDB Dashboard List" src="/docs/images/mariadb/monitoring/mariadb-alerting-grafana-dashboards.png" style="padding:10px">
</p>

Four pre-built dashboards are available. The `Namespace` and `MariaDB` drop-downs at the top of each dashboard let you switch between instances.

**KubeDB / MariaDB / Summary** — database status, version, node count, CPU/memory/storage requests vs. usage.

<p align="center">
  <img alt="Grafana — KubeDB MariaDB Summary" src="/docs/images/mariadb/monitoring/mariadb-alerting-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / MariaDB / Galera-Cluster** — cluster name, per-node ONLINE/Primary status, and Galera replication latency (average, standard deviation, sample size) per node.

<p align="center">
  <img alt="Grafana — KubeDB MariaDB Galera-Cluster" src="/docs/images/mariadb/monitoring/mariadb-alerting-grafana-galera.png" style="padding:10px">
</p>

**KubeDB / MariaDB / Database** — per-pod service status/uptime, cluster size, primary status, QPS, connections, disk I/O, and top command counters.

<p align="center">
  <img alt="Grafana — KubeDB MariaDB Database" src="/docs/images/mariadb/monitoring/mariadb-alerting-grafana-database.png" style="padding:10px">
</p>

**KubeDB / MariaDB / Pod** — per-pod CPU/memory/file descriptors, connections, thread activity, temporary objects, slow queries, table locks, and network traffic.

<p align="center">
  <img alt="Grafana — KubeDB MariaDB Pod" src="/docs/images/mariadb/monitoring/mariadb-alerting-grafana-pod.png" style="padding:10px">
</p>

---

## Simulating a Firing Alert

This section deliberately triggers `MariaDBInstanceDown` (instant, `for: 0m`) by crashing the main `mariadb` process, and observes the alert through Prometheus and AlertManager.

Unlike Elasticsearch, killing the main process here works well: MariaDB's container `PID 1` is `tini` supervising a wrapper script, not `mariadbd` itself, so killing `mariadbd` doesn't take the container down — it just leaves `mysql_up` at `0` until the script notices and restarts the daemon. A single `kill -9` self-heals in roughly 20–30 seconds (too fast to reliably catch, since it beats one evaluation cycle), so hold it down with a short kill-loop instead.

### 1. Crash the MariaDB process

```bash
$ kubectl exec -n alert-mariadb mariadb-alert-demo-0 -c mariadb -- sh -c '
    end=$(( $(date +%s) + 45 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -x mariadbd | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

Run this in the background (or a separate terminal) — it holds `mariadbd` down for 45 seconds, comfortably past one Prometheus scrape (10s) and evaluation (30s) cycle.

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts?search=mariadb`.

<p align="center">
  <img alt="Prometheus Alerts — MariaDBInstanceDown Firing" src="/docs/images/mariadb/monitoring/mariadb-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`MariaDBInstanceDown` (`mysql_up == 0`) transitions straight to **FIRING** since it has no `for` delay, while the rest of the `mariadb.database` group stays **INACTIVE**.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093/#/alerts?filter={namespace="alert-mariadb"}`.

<p align="center">
  <img alt="AlertManager — MariaDBInstanceDown Firing" src="/docs/images/mariadb/monitoring/mariadb-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows the `MariaDBInstanceDown` alert. The alert card displays:

- **Severity**: `critical`
- **pod**: `mariadb-alert-demo-0`
- **job**: `mariadb-alert-demo-stats`
- **Started**: timestamp when the alert first fired

### 4. Restore MariaDB

Let the loop from step 1 finish (or stop it early). `run.sh` inside the container restarts `mariadbd` on its own — no pod restart needed.

```bash
$ kubectl get mariadb -n alert-mariadb mariadb-alert-demo -w
NAME                 VERSION   STATUS     AGE
mariadb-alert-demo   12.1.2    Critical   21h
mariadb-alert-demo   12.1.2    Ready      21h
```

Recovery took about 10–15 seconds after the kill-loop ended in testing — `mariadbd` restarts, performs a quick Galera State Snapshot Transfer (SST via `rsync`) to catch back up with the other two nodes, and `mysql_up` returns to `1`. The **KubeDB / MariaDB / Galera-Cluster** dashboard's replication-latency panel is a good place to watch this recovery happen in real time. Once `mysql_up` is back to `1`, Prometheus marks the alert **INACTIVE** and AlertManager sends a **resolved** notification. If MariaDB does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-mariadb mariadb-alert-demo-0`.

---

## Alert Reference

All alerts are scoped to the `mariadb-alert-demo` instance in the `alert-mariadb` namespace via the PromQL label filters `job="mariadb-alert-demo-stats"` / `namespace="alert-mariadb"` (database/cluster groups), or `app="mariadb-alert-demo"` / `namespace="alert-mariadb"` (provisioner/opsManager/stash/kubeStash/schemaManager groups).

### Database Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MariaDBInstanceDown` | critical | instant | `mysql_up == 0` on this instance. |
| `MariaDBServiceDown` | critical | instant | No replica behind the service is answering. |
| `MariaDBTooManyConnections` | warning | 2m | Connection count is high relative to `max_connections`. |
| `MariaDBHighThreadsRunning` | warning | 2m | Too many threads actively running. |
| `MariaDBSlowQueries` | warning | 2m | Slow-query count is increasing. |
| `MariaDBInnoDBLogWaits` | warning | instant | InnoDB log waits are occurring — I/O may be a bottleneck. |
| `MariaDBRestarted` | warning | instant | Uptime indicates a recent restart. |
| `MariaDBHighQPS` | critical | instant | Query rate is unusually high. |
| `MariaDBHighIncomingBytes` | critical | instant | Inbound network traffic is unusually high. |
| `MariaDBHighOutgoingBytes` | critical | instant | Outbound network traffic is unusually high. |
| `MariaDBTooManyOpenFiles` | warning | 2m | Open file count is high relative to the limit. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80%. **Disabled in this tutorial** — see [above](#a-note-on-chart-defaults): the expression is a chart-level defect that always reads ~100% regardless of real usage, with no accurate alternative in this chart. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95%. **Disabled in this tutorial** — same defect as `DiskUsageHigh` above. |

### Cluster Group

Only produces data when `spec.topology` (Galera) is configured — this tutorial's instance is a Galera cluster, so this group has live data.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `GaleraReplicationLatencyTooLong` | warning | 5m | Galera replication latency is high. |

### Provisioner Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMariaDBPhaseNotReady` | critical | 1m | KubeDB marked the MariaDB resource `NotReady`. |
| `KubeDBMariaDBPhaseCritical` | warning | 15m | MariaDB is degraded but not fully unavailable. |

### OpsManager Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMariaDBOpsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes. |
| `KubeDBMariaDBOpsRequestFailed` | critical | instant | An ops request failed. |

### Stash / KubeStash Groups

Only meaningful once Stash or KubeStash backup/restore is configured.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MariaDBStashBackupSessionFailed` / `MariaDBKubeStashBackupSessionFailed` | critical | instant | The most recent backup session failed. |
| `MariaDBStashRestoreSessionFailed` / `MariaDBKubeStashRestoreSessionFailed` | critical | instant | The most recent restore session failed. |
| `MariaDBStashNoBackupSessionForTooLong` / `MariaDBKubeStashNoBackupSessionForTooLong` | warning | instant | No recent successful backup. |
| `MariaDBStashRepositoryCorrupted` / `MariaDBKubeStashRepositoryCorrupted` | critical | 5m | Backup repository integrity check failed. |
| `MariaDBStashRepositoryStorageRunningLow` / `MariaDBKubeStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage usage is high. |
| `MariaDBStashBackupSessionPeriodTooLong` / `MariaDBKubeStashBackupSessionPeriodTooLong` | warning | instant | A backup session is taking unusually long. |
| `MariaDBStashRestoreSessionPeriodTooLong` / `MariaDBKubeStashRestoreSessionPeriodTooLong` | warning | instant | A restore session is taking unusually long. |

### SchemaManager Group

Only meaningful when using `MariaDBDatabase` schema-manager objects.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMariaDBSchemaPendingForTooLong` | warning | 30m | A `MariaDBDatabase` object stuck `Pending`. |
| `KubeDBMariaDBSchemaInProgressForTooLong` | warning | 30m | A `MariaDBDatabase` object stuck `InProgress`. |
| `KubeDBMariaDBSchemaTerminatingForTooLong` | warning | 30m | A `MariaDBDatabase` object stuck `Terminating`. |
| `KubeDBMariaDBSchemaFailed` | warning | instant | A `MariaDBDatabase` object failed. |
| `KubeDBMariaDBSchemaExpired` | warning | instant | A `MariaDBDatabase` object expired. |

---

## Customising Alerts

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
          mariadbTooManyConnections:
            enabled: true
            duration: "5m"
            severity: warning
      cluster:
        enabled: "none"    # disable if you don't run Galera
```

```bash
$ helm upgrade mariadb-alert-demo oci://ghcr.io/appscode-charts/mariadb-alerts \
    -n alert-mariadb \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

```bash
# Remove the Grafana dashboards (installed via helm template | kubectl apply, not helm install)
$ helm template kubedb-grafana-dashboards appscode/kubedb-grafana-dashboards \
    -n kubeops \
    --version=v2026.7.10 \
    --set featureGates.MariaDB=true \
    --set grafana.url="http://prometheus-grafana.monitoring.svc:80" \
    --set grafana.apikey="<api-key>" \
  | kubectl delete -n kubeops -f - --ignore-not-found

# Remove the mariadb-alerts release
$ helm uninstall mariadb-alert-demo -n alert-mariadb

# Remove the MariaDB instance
$ kubectl delete mariadb -n alert-mariadb mariadb-alert-demo

# Delete namespace
$ kubectl delete ns alert-mariadb

# Optional: only if nothing else in the cluster depends on them
$ kubectl delete appbinding -n kubeops grafana
$ kubectl delete secret -n kubeops grafana-admin-token
$ helm uninstall grafana-operator -n kubeops
```

## Next Steps

- Monitor your MariaDB database with KubeDB using [built-in Prometheus](/docs/guides/mariadb/monitoring/builtin-prometheus/index.md).
- Monitor your MariaDB database with KubeDB using [Prometheus operator](/docs/guides/mariadb/monitoring/prometheus-operator/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
