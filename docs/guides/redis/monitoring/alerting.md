---
title: Redis Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: rd-monitoring-alerting
    name: Alerting
    parent: rd-monitoring-redis
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Redis Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed Redis instance using the `redis-alerts` Helm chart.

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

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/redis/monitoring/overview.md).

* For dashboards and visualisation, see [Grafana Dashboard](grafana-dashboard.md) for Redis.

> Note: YAML files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys Redis with a built-in `redis_exporter` sidecar that exposes metrics on port `56790`.
- **ServiceMonitor** (named `{redis-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `redis-alerts` chart and contains all Redis alert definitions grouped by concern: database health, provisioner, ops-manager, and backups (Stash and KubeStash).
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

---

## Deploy Redis with Monitoring Enabled

At first, let's deploy a Redis database with monitoring enabled. Below is the Redis object we are going to create.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: rd-alert-demo
  namespace: alert-redis
spec:
  version: "6.0.20"
  deletionPolicy: WipeOut
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
        interval: 10s
        labels:
          release: prometheus
```

Here,

- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

Let's create the namespace and the Redis resource.

```bash
$ kubectl create ns alert-redis
namespace/alert-redis created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/monitoring/rd-alert-demo.yaml
redis.kubedb.com/rd-alert-demo created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get redis -n alert-redis rd-alert-demo
NAME            VERSION   STATUS   AGE
rd-alert-demo   6.0.20    Ready    62m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-redis --selector="app.kubernetes.io/instance=rd-alert-demo"
NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)              AGE
rd-alert-demo         ClusterIP   10.43.236.232   <none>        6379/TCP             62m
rd-alert-demo-pods    ClusterIP   None            <none>        6379/TCP,16379/TCP   62m
rd-alert-demo-stats   ClusterIP   10.43.191.236   <none>        56790/TCP            62m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-redis
NAME                  AGE
rd-alert-demo-stats   62m
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-redis rd-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install redis-alerts

The `redis-alerts` chart creates a `PrometheusRule` resource containing all Redis alert definitions grouped by concern: database health, provisioner, ops-manager, and backups (Stash / KubeStash).

### Why the Helm release name matters

The chart derives the PromQL `job`/instance scoping (and the `PrometheusRule` name) from the **Helm release name**, not from a values field — so the release name must match the Redis object's name (`rd-alert-demo`) for the rules to be correctly scoped to this instance.

The chart's default label is `release: kube-prometheus-stack`, so we must also override it at install time to match the Prometheus `ruleSelector`.

### Install

```bash
$ helm upgrade -i rd-alert-demo oci://ghcr.io/appscode-charts/redis-alerts \
    -n alert-redis \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

| Flag | Value | Purpose |
|------|-------|---------|
| `rd-alert-demo` (release name) | — | Scopes every PromQL expression to this instance (`job="rd-alert-demo-stats"`) |
| `-n alert-redis` | `alert-redis` | Installs the `PrometheusRule` in the same namespace as the database |
| `form.alert.labels.release` | `prometheus` | Matches the Prometheus `ruleSelector` so the rules are loaded |

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-redis
NAME            AGE
rd-alert-demo   60m
```

Confirm the `release: prometheus` label is present.

```bash
$ kubectl get prometheusrule -n alert-redis rd-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

Port-forward the Prometheus UI and open the **Status → Rule health** page.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules?search=rd-alert-demo`.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/redis/monitoring/rd-alerting-prom-rules.png" style="padding:10px">
</p>

The `redis.database.alert-redis.rd-alert-demo.rules` group (and the accompanying `redis.kubeStash.alert-redis.rd-alert-demo.rules` group) is visible with all rules showing **OK**, confirming that Prometheus has loaded and is evaluating the Redis alert definitions every 30 seconds.

---

## Verify End-to-End

### 1. Check the exporter is running

The `exporter` sidecar inside the Redis pod serves metrics at `:56790/metrics`. A value of `redis_up 1` confirms the exporter can reach Redis.

```bash
$ kubectl exec -n alert-redis rd-alert-demo-0 -c exporter -- \
    wget -qO- localhost:56790/metrics | grep redis_up
redis_up 1
```

### 2. Check the Prometheus target is UP

Open `http://localhost:9090/targets?search=rd-alert-demo`. Prometheus discovers more than 20 scrape pools on this cluster, so it will ask you to pick one from the dropdown — select `serviceMonitor/alert-redis/rd-alert-demo-stats/0`.

<p align="center">
  <img alt="Prometheus Target UP" src="/docs/images/redis/monitoring/rd-alerting-prom-target.png" style="padding:10px">
</p>

The target `serviceMonitor/alert-redis/rd-alert-demo-stats/0` shows **UP**, confirming metrics are being scraped from `rd-alert-demo-0` in the `alert-redis` namespace.

### 3. Confirm all Redis alerts are inactive

Open `http://localhost:9090/alerts?search=rd-alert-demo` to see the Redis alert groups.

<p align="center">
  <img alt="Prometheus Alerts — All Inactive" src="/docs/images/redis/monitoring/rd-alerting-prom-alerts.png" style="padding:10px">
</p>

All 7 rules in the `redis.database` group show **INACTIVE (7)**, meaning the database is healthy and no thresholds are breached.

### 4. Check AlertManager

Port-forward AlertManager to view any currently firing alerts.

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`. With a healthy Redis instance, no alerts for `rd-alert-demo` will be listed here.

---

## Simulating a Firing Alert

The previous section confirmed that all alerts are **INACTIVE** while the database is healthy. This section walks through deliberately triggering the `RedisDown` critical alert so you can observe the full alert lifecycle — from firing in Prometheus through to the AlertManager dashboard — and then resolve it.

Unlike some other database exporters, the Redis exporter runs as a **separate sidecar container** that keeps running fine even if the main `redis` container crashes and gets restarted by Kubernetes. Because `RedisDown` requires the outage to persist for `for: 2m` (not instant), and Kubernetes tends to restart a crashed container within a few seconds, a single `kill` is usually not enough to keep Redis down long enough to breach the 2-minute window. In practice, repeatedly stopping the Redis process for a couple of minutes (so the container keeps crash-looping) reliably keeps `redis_up` at `0` for long enough for the alert to fire.

### 1. Stop the Redis process repeatedly

Shut down the `redis` process inside the pod. This crashes the main container so the `exporter` sidecar can no longer reach it and reports `redis_up 0` on the next scrape, while Kubernetes restarts the crashed container in the background.

```bash
$ kubectl exec -n alert-redis rd-alert-demo-0 -c redis -- redis-cli shutdown nosave
```

Because Kubernetes restarts the container quickly, repeat this command every few seconds for about two minutes to keep Redis down continuously long enough for the `for: 2m` window on `RedisDown` to be satisfied:

```bash
$ while true; do
    kubectl exec -n alert-redis rd-alert-demo-0 -c redis -- redis-cli shutdown nosave >/dev/null 2>&1
    sleep 1
  done
```

Let this loop run for about two minutes, then move on to the next step (leave the loop running while you check).

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts?search=rd-alert-demo`.

<p align="center">
  <img alt="Prometheus Alerts — RedisDown Firing" src="/docs/images/redis/monitoring/rd-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`RedisDown` moves from **INACTIVE** to **FIRING** once the `redis_up == 0` condition has held continuously for the configured `for: 2m` duration. `RedisMissingMaster` fires alongside it, since with no master node visible the cluster is also missing a master.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093/#/alerts?filter=%7Bnamespace%3D%22alert-redis%22%7D`.

<p align="center">
  <img alt="AlertManager — RedisDown Firing" src="/docs/images/redis/monitoring/rd-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows the `RedisDown` alert (alongside `KubeDBRedisPhaseNotReady`, since the KubeDB operator also observes the pod is not ready). The alert card displays:

- **Severity**: `critical`
- **pod**: `rd-alert-demo-0` in the `alert-redis` namespace
- **job**: `rd-alert-demo-stats`
- **Started**: timestamp when the alert first fired

AlertManager routes this alert to every receiver configured in your `alertmanagerConfig` (Slack, email, PagerDuty, webhook, etc.) based on your routing tree. If no receiver is configured, the alert is visible here but silently dropped.

### 4. Restore Redis

Stop the loop from step 1, then delete the pod so KubeDB recreates it cleanly.

```bash
$ kubectl delete pod -n alert-redis rd-alert-demo-0
```

Once `redis_up` returns to `1` continuously, Prometheus marks the alert **INACTIVE** again and AlertManager sends a **resolved** notification to all receivers.

---

## Alert Reference

All alerts are scoped to the `rd-alert-demo` instance in the `alert-redis` namespace via the PromQL label filters `job="rd-alert-demo-stats"` and `namespace="alert-redis"`.

### Database Group

Fired based on live metrics from the Redis exporter.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `RedisDown` | critical | 2m | Exporter cannot reach Redis — instance is down or crashed. |
| `RedisMissingMaster` | critical | 2m | Redis cluster has fewer nodes marked as master than expected. |
| `RedisTooManyMasters` | critical | 2m | Redis cluster has more nodes marked as master than expected — possible split-brain. |
| `RedisDisconnectedSlaves` | warning | 2m | Redis is not replicating for all slaves — review the replication status. |
| `RedisTooManyConnections` | warning | 2m | More than 80% of `maxclients` are in use — client pool nearing exhaustion. |
| `DiskUsageHigh` | warning | 5m | Persistent volume usage is between 80% and 95% — plan for capacity. |
| `DiskAlmostFull` | critical | 5m | Persistent volume usage has exceeded 95% — the volume is nearly full. |

### Provisioner Group

Monitors the KubeDB operator's view of the Redis resource phase.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBRedisPhaseNotReady` | critical | 1m | KubeDB marked the Redis resource `NotReady` — operator cannot reach the database. |
| `KubeDBRedisPhaseCritical` | warning | 15m | The instance is in a degraded/critical phase. |

### OpsManager Group

Tracks `RedisOpsRequest` lifecycle during upgrades, scaling, and reconfiguration.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBRedisOpsRequestOnProgress` | info | instant | A `RedisOpsRequest` is currently in progress. |
| `KubeDBRedisOpsRequestStatusProgressingToLong` | critical | 30m | A `RedisOpsRequest` has been in progress for 30+ minutes — likely stuck. |
| `KubeDBRedisOpsRequestFailed` | critical | instant | A `RedisOpsRequest` failed — check the `RedisOpsRequest` object for the error. |

### Stash Group

Tracks Stash-driven backup/restore health for this instance. (You do not need to configure backups to see these rules; they are included in the `PrometheusRule` regardless.)

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `RedisStashBackupSessionFailed` | critical | instant | A Stash backup session failed. |
| `RedisStashRestoreSessionFailed` | critical | instant | A Stash restore session failed. |
| `RedisStashNoBackupSessionForTooLong` | warning | instant | No successful backup session for more than 18000s (5 hours). |
| `RedisStashRepositoryCorrupted` | critical | 5m | The Stash backup repository integrity check failed — repository is corrupted. |
| `RedisStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage size has exceeded 10GB. |
| `RedisStashBackupSessionPeriodTooLong` | warning | instant | A backup session took more than 1800s (30 minutes) to complete. |
| `RedisStashRestoreSessionPeriodTooLong` | warning | instant | A restore session took more than 1800s (30 minutes) to complete. |

### KubeStash Group

Tracks KubeStash-driven backup/restore health for this instance. Same semantics as the Stash group above, sourced from KubeStash metrics instead.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `RedisKubeStashBackupSessionFailed` | critical | instant | A KubeStash backup session failed. |
| `RedisKubeStashRestoreSessionFailed` | critical | instant | A KubeStash restore session failed. |
| `RedisKubeStashNoBackupSessionForTooLong` | warning | instant | No successful backup session for more than 18000s (5 hours). |
| `RedisKubeStashRepositoryCorrupted` | critical | 5m | The KubeStash backup repository integrity check failed — repository is corrupted. |
| `RedisKubeStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage size has exceeded 10GB. |
| `RedisKubeStashBackupSessionPeriodTooLong` | warning | instant | A backup session took more than 1800s (30 minutes) to complete. |
| `RedisKubeStashRestoreSessionPeriodTooLong` | warning | instant | A restore session took more than 1800s (30 minutes) to complete. |

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
          redisTooManyConnections:
            enabled: true
            duration: "5m"
            val: 90        # fire at 90% of maxclients instead of the default 80%
            severity: warning
      opsManager:
        enabled: "none"    # disable all ops-manager alerts
```

```bash
$ helm upgrade rd-alert-demo oci://ghcr.io/appscode-charts/redis-alerts \
    -n alert-redis \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the redis-alerts release
$ helm uninstall rd-alert-demo -n alert-redis

# Remove the Redis instance
$ kubectl delete redis -n alert-redis rd-alert-demo

# Delete namespace
$ kubectl delete ns alert-redis
```

## Next Steps

- Monitor your Redis database with KubeDB using [builtin Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Monitor your Redis database with KubeDB using [Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Visualise Redis metrics with [Grafana Dashboard](grafana-dashboard.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
