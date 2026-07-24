---
title: ZooKeeper Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: zk-monitoring-alerting
    name: Alerting
    parent: zk-monitoring-guides
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ZooKeeper Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed ZooKeeper instance using the `zookeeper-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `demo` namespace:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

* This tutorial assumes you already have a **kube-prometheus-stack** running in your cluster, with `Prometheus` configured so that both `serviceMonitorSelector` and `ruleSelector` match the label `release: prometheus`. See the [Grafana Dashboard](grafana-dashboard.md#configuration) guide for how to deploy kube-prometheus-stack if you don't have it yet.

  To verify the selectors:

  ```bash
  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.ruleSelector}'
  {"matchLabels":{"release":"prometheus"}}

  $ kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
  {"matchLabels":{"release":"prometheus"}}
  ```

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/zookeeper/monitoring/overview.md).

* For dashboards and visualisation, see [Grafana Dashboard](grafana-dashboard.md) for ZooKeeper.

> Note: YAML files used in this tutorial are stored in [docs/examples/zookeeper](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/zookeeper) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **ZooKeeper** (3.6+) ships with a built-in Prometheus `MetricsProvider` that exposes metrics natively on an admin/metrics port — KubeDB does not need to inject a separate exporter sidecar for ZooKeeper, unlike some other databases.
- **ServiceMonitor** (named `{zookeeper-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape each ZooKeeper pod's metrics endpoint every 10 seconds.
- **PrometheusRule** is created by the `zookeeper-alerts` chart and contains all ZooKeeper alert definitions grouped by concern: database health and provisioner.
- **Prometheus Operator** evaluates every rule expression on its configured evaluation interval and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

---

## Deploy ZooKeeper with Monitoring Enabled

At first, let's deploy a ZooKeeper database with monitoring enabled. Below is the ZooKeeper object we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zk-alert-demo
  namespace: demo
spec:
  version: "3.8.3"
  replicas: 3
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
```

Here,

- `spec.replicas: 3` is the documented minimum ZooKeeper ensemble size for quorum — keep it at 3 (or more, in odd numbers) for a production-like setup.
- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

Let's create the ZooKeeper resource.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/monitoring/zk-alert-demo.yaml
zookeeper.kubedb.com/zk-alert-demo created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get zookeeper -n demo zk-alert-demo
NAME            VERSION   STATUS   AGE
zk-alert-demo   3.8.3     Ready    90s
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=zk-alert-demo"
NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
zk-alert-demo                 ClusterIP   10.43.78.149    <none>        2181/TCP                     97s
zk-alert-demo-admin-server    ClusterIP   10.43.214.188   <none>        8080/TCP                     97s
zk-alert-demo-pods            ClusterIP   None            <none>        2181/TCP,2888/TCP,3888/TCP   97s
zk-alert-demo-stats           ClusterIP   10.43.250.162   <none>        7000/TCP                     97s
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n demo
NAME                    AGE
zk-alert-demo-stats     97s
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n demo zk-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install zookeeper-alerts

The `zookeeper-alerts` chart creates a `PrometheusRule` resource containing all ZooKeeper alert definitions grouped by concern: database health and provisioner.

### Why the Helm release name matters

The chart derives the PromQL `job`/instance scoping (and the `PrometheusRule` name) from the **Helm release name**, not from a values field — so the release name must match the ZooKeeper object's name (`zk-alert-demo`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### Install

```bash
$ helm upgrade -i zk-alert-demo oci://ghcr.io/appscode-charts/zookeeper-alerts \
    -n demo \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

| Flag | Value | Purpose |
|------|-------|---------|
| `zk-alert-demo` (release name) | — | Scopes every PromQL expression to this instance (`job="zk-alert-demo-stats"`) |
| `-n demo` | `demo` | Installs the `PrometheusRule` in the same namespace as the database |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n demo
NAME            AGE
zk-alert-demo   30s
```

Confirm the `release: prometheus` label is present.

```bash
$ kubectl get prometheusrule -n demo zk-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI and open the **Status → Rule health** page.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules?search=zookeeper`.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/zookeeper/monitoring/zk-alerting-prom-rules.png" style="padding:10px">
</p>

The `zookeeper.database.demo.zk-alert-demo.rules` group is visible with all rules showing **OK**, confirming that Prometheus has loaded and is evaluating the ZooKeeper alert definitions.

---

## Verify End-to-End

### 1. Check the metrics endpoint is serving

ZooKeeper's built-in metrics provider serves Prometheus metrics at `:7000/metrics` directly from the ZooKeeper JVM — there is no separate exporter sidecar container to check for this database.

```bash
$ kubectl exec -n demo zk-alert-demo-0 -c zookeeper -- \
    wget -qO- localhost:7000/metrics | grep -m1 "^# HELP"
# HELP write_commit_proc_issued write_commit_proc_issued
```

A non-empty response confirms the JVM's Prometheus servlet is up and exporting metrics.

### 2. Check the Prometheus target is UP

Open `http://localhost:9090/targets?search=zk-alert-demo`.

<p align="center">
  <img alt="Prometheus Target UP" src="/docs/images/zookeeper/monitoring/zk-alerting-prom-target.png" style="padding:10px">
</p>

All three targets under `serviceMonitor/demo/zk-alert-demo-stats/0` show **UP** — one per ensemble member (`zk-alert-demo-0`, `zk-alert-demo-1`, `zk-alert-demo-2`).

> Note: with more than 20 scrape pools on a busy cluster, the Prometheus React UI's target-search box can sometimes default to showing an unrelated pool first. If that happens, use the pool dropdown (or the `pool=` URL parameter) to explicitly pick `serviceMonitor/<namespace>/<name>-stats/0`.

### 3. Confirm the ZooKeeper alerts are inactive

Open `http://localhost:9090/alerts?search=zookeeper` to see the ZooKeeper alert groups.

<p align="center">
  <img alt="Prometheus Alerts" src="/docs/images/zookeeper/monitoring/zk-alerting-prom-alerts.png" style="padding:10px">
</p>

All 11 metric-driven rules in the `zookeeper.database` group (`ZooKeeperDown`, `ZooKeeperTooManyNodes`, `ZooKeeperTooManyConnections`, etc.) show **INACTIVE**, meaning the ensemble is healthy and no thresholds are breached.

> Note: the `DiskUsageHigh` / `DiskAlmostFull` rules in the same group compare `kubelet_volume_stats_used_bytes` against the PVC's usage on the underlying node. On a demo cluster where the node's disk is already heavily utilized by other workloads, these two rules can legitimately show **PENDING**/**FIRING** independent of ZooKeeper itself — that reflects real node-level disk pressure, not a problem with the ZooKeeper instance. Give your PVC's underlying disk enough free space if you want to keep these two alerts quiet in your own cluster.

### 4. Check AlertManager

Port-forward AlertManager to view any currently firing alerts.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`. With a healthy ZooKeeper ensemble, no `ZooKeeperDown` (or other database-group) alert for `zk-alert-demo` will be listed here.

---

## Simulating a Firing Alert

The previous section confirmed that the ZooKeeper-specific alerts are **INACTIVE** while the ensemble is healthy. This section walks through deliberately triggering the `ZooKeeperDown` critical alert so you can observe the full alert lifecycle — from firing in Prometheus through to the AlertManager dashboard — and then resolve it.

### 1. Stop the ZooKeeper process

`ZooKeeperDown` fires on the standard Prometheus `up{job="zk-alert-demo-stats"} == 0` scrape-health metric, with `for: 1m` — the target has to stay unreachable for a full minute before the alert transitions from **PENDING** to **FIRING**.

ZooKeeper ships as a single container per pod (no exporter sidecar), so killing the main JVM process brings the whole container down; the container then restarts automatically. Find the ZooKeeper JVM's PID inside the pod and kill it:

```bash
$ kubectl exec -n demo zk-alert-demo-0 -c zookeeper -- pgrep -f QuorumPeerMain
38

$ kubectl exec -n demo zk-alert-demo-0 -c zookeeper -- kill -9 38
```

Because the container typically restarts within a few seconds — faster than the 1-minute `for` window — a single kill is usually not enough to observe a **FIRING** state. Repeat the kill every few seconds until the container is caught in `CrashLoopBackOff`, which keeps it down long enough to satisfy the 1-minute condition:

```bash
$ for i in $(seq 1 20); do
    PID=$(kubectl exec -n demo zk-alert-demo-0 -c zookeeper -- pgrep -f QuorumPeerMain 2>/dev/null | head -1)
    [ -n "$PID" ] && kubectl exec -n demo zk-alert-demo-0 -c zookeeper -- kill -9 "$PID"
    sleep 3
  done

$ kubectl get pod -n demo zk-alert-demo-0
NAME              READY   STATUS   RESTARTS      AGE
zk-alert-demo-0   0/1     Error    6 (93s ago)   56m
```

Poll the Prometheus API until the rule state flips from `pending` to `firing` (roughly 60–90 seconds after the pod first goes down):

```bash
$ curl -s "http://localhost:9090/api/v1/rules?type=alert" \
    | jq -r '.data.groups[] | select(.name | contains("zookeeper.database")) | .rules[] | select(.name=="ZooKeeperDown") | .state'
firing
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts?search=zookeeper`.

<p align="center">
  <img alt="Prometheus Alerts — ZooKeeperDown Firing" src="/docs/images/zookeeper/monitoring/zk-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`ZooKeeperDown` now shows **FIRING (1)** — for the single pod (`zk-alert-demo-0`) that was taken down; the other two ensemble members remain healthy.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093/#/alerts?filter=%7Bnamespace%3D%22demo%22%7D`.

<p align="center">
  <img alt="AlertManager — ZooKeeperDown Firing" src="/docs/images/zookeeper/monitoring/zk-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows the `ZooKeeperDown` alert. The alert card displays:

- **Severity**: `critical`
- **Instance**: `zk-alert-demo-0` in the `demo` namespace
- **job**: `zk-alert-demo-stats`
- **Started**: timestamp when the alert first fired

In this demo run, AlertManager also showed a `KubeDBZooKeeperPhaseNotReady` alert (from the `provisioner` group) firing at the same time — the repeated `CrashLoopBackOff` was severe enough that the KubeDB operator itself marked the ZooKeeper object `NotReady`, which is a realistic example of two independent alert groups correctly firing together from a single underlying failure.

AlertManager routes these alerts to every receiver configured in your `alertmanagerConfig` (Slack, email, PagerDuty, webhook, etc.) based on your routing tree. If no receiver is configured, the alerts are visible here but silently dropped.

### 4. Restore ZooKeeper

Delete the pod so KubeDB recreates it cleanly instead of leaving it in `CrashLoopBackOff`.

```bash
$ kubectl delete pod -n demo zk-alert-demo-0
```

Once the pod is back to `Running` and its metrics endpoint is reachable again, Prometheus marks `ZooKeeperDown` **INACTIVE** and AlertManager sends a **resolved** notification to all receivers.

---

## Alert Reference

All alerts are scoped to the `zk-alert-demo` instance in the `demo` namespace via the PromQL label filters `job="zk-alert-demo-stats"` and `namespace="demo"`.

### Database Group

Fired based on live metrics from ZooKeeper's built-in Prometheus metrics provider.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `ZooKeeperDown` | critical | 1m | ZooKeeper instance is down — the scrape target has been unreachable for a full minute. |
| `ZooKeeperTooManyNodes` | warning | 1m | ZooKeeper ensemble has too many znodes (more than 1,000,000) — consider scaling up. |
| `ZooKeeperTooBigMemory` | warning | 1m | ZooKeeper's znode total occupied memory (`approximate_data_size`) is too big (more than 1 GB). |
| `ZooKeeperTooManyWatch` | warning | 1m | Too many watches are set (more than 10,000) on this instance. |
| `ZooKeeperTooManyConnections` | warning | 1m | More than 60 client connections are in use on this instance. |
| `ZooKeeperLeaderElection` | warning | 1m | A leader election happened in the last 5 minutes — indicates ensemble instability. |
| `ZooKeeperTooManyOpenFiles` | warning | 1m | More than 300 open file descriptors on this instance. |
| `ZooKeeperTooLongFsyncTime` | warning | 1m | fsync operations are taking too long (rate over 1m exceeds 100). |
| `ZooKeeperTooLongSnapshotTime` | warning | 1m | Snapshotting is taking too long (rate over 5m exceeds 100). |
| `ZooKeeperTooHighAvgLatency` | warning | 1m | Average request latency (`avg_latency`) exceeds 100ms. |
| `ZooKeeperJvmMemoryFilingUp` | warning | 1m | JVM heap usage exceeds 80% of the max heap size. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage for this instance's data directory exceeds 80%. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage for this instance's data directory exceeds 95%. |

### Provisioner Group

Monitors the KubeDB operator's view of the ZooKeeper resource phase.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBZooKeeperPhaseNotReady` | critical | 1m | KubeDB marked the ZooKeeper resource `NotReady` — operator cannot reach the ensemble. |
| `KubeDBZooKeeperPhaseCritical` | warning | 15m | The instance is in a degraded/critical phase. |

### OpsManager Group

The chart's `values.yaml` declares an `opsManager` group, intended to track `ZooKeeperOpsRequest` lifecycle during upgrades, scaling, and reconfiguration — the same pattern used by other `*-alerts` charts:

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `opsRequestOnProgress` | info | instant | An ops request is currently in progress. |
| `opsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes — likely stuck. |
| `opsRequestFailed` | critical | instant | An ops request failed — check the `OpsRequest` object for the error. |

> Verified against chart `v2026.7.14`: unlike the `database` and `provisioner` groups, this `opsManager` group is **not** currently rendered into the `PrometheusRule` template (`kubectl get prometheusrule -n demo zk-alert-demo -o jsonpath='{.spec.groups[*].name}'` only returns `zookeeper.database...` and `zookeeper.provisioner...`). The values above document the intended configuration surface; confirm with `helm template` against the chart version you install whether opsManager rules are actually emitted before relying on them for `ZooKeeperOpsRequest` alerting:
>
> ```bash
> $ helm template zk-alert-demo oci://ghcr.io/appscode-charts/zookeeper-alerts \
>     --version=v2026.7.14 \
>     --set form.alert.labels.release=prometheus \
>     | grep "alert: "
> ```

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
          zookeeperTooManyConnections:
            enabled: true
            duration: "5m"
            val: 100        # fire at 100 connections instead of the default 60
            severity: warning
          diskUsageHigh:
            enabled: true
            val: 90          # raise the disk-usage-high threshold to 90%
            duration: "5m"
            severity: warning
      provisioner:
        enabled: "none"      # disable all provisioner alerts
```

```bash
$ helm upgrade zk-alert-demo oci://ghcr.io/appscode-charts/zookeeper-alerts \
    -n demo \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the zookeeper-alerts release
$ helm uninstall zk-alert-demo -n demo

# Remove the ZooKeeper instance
$ kubectl delete zookeeper -n demo zk-alert-demo

# Delete namespace
$ kubectl delete ns demo
```

## Next Steps

- Monitor your ZooKeeper database with KubeDB using [builtin Prometheus](/docs/guides/zookeeper/monitoring/using-builtin-prometheus.md).
- Monitor your ZooKeeper database with KubeDB using [Prometheus operator](/docs/guides/zookeeper/monitoring/using-prometheus-operator.md).
- Visualise ZooKeeper metrics with [Grafana Dashboard](grafana-dashboard.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
