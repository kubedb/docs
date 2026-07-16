---
title: MongoDB Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: mg-monitoring-alerting
    name: Alerting
    parent: mg-monitoring-mongodb
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed MongoDB instance using the `mongodb-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-mongodb` namespace:

  ```bash
  $ kubectl create ns alert-mongodb
  namespace/alert-mongodb created
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

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/mongodb/monitoring/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys MongoDB with a `mongodb_exporter` sidecar (container `exporter`) that exposes metrics (`mongodb_*`).
- **ServiceMonitor** (named `{mongodb-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape the exporter every 10 seconds.
- **PrometheusRule** is created by the `mongodb-alerts` chart and contains MongoDB alert definitions grouped by concern: database health (which also embeds the KubeDB-operator-sourced `MongoDBDown`/`MongoDBPhaseCritical` pair), provisioner, ops-manager, Stash backup/restore, KubeStash backup/restore, and schema manager.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** dashboards for MongoDB are covered separately — see [Grafana Dashboard](grafana-dashboard.md) rather than duplicated here.

---

## Deploy MongoDB with Monitoring Enabled

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongodb-alert-demo
  namespace: alert-mongodb
spec:
  version: "6.0.14"
  replicas: 1
  storageType: Durable
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/monitoring/mongodb-alert-demo.yaml
mongodb.kubedb.com/mongodb-alert-demo created
```

Wait for the database to go into `Ready` state.

```bash
$ kubectl get mongodb -n alert-mongodb mongodb-alert-demo
NAME                 VERSION   STATUS   AGE
mongodb-alert-demo   6.0.14    Ready    3m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-mongodb --selector="app.kubernetes.io/instance=mongodb-alert-demo"
NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
mongodb-alert-demo          ClusterIP   10.43.10.20    <none>        27017/TCP   3m
mongodb-alert-demo-pods     ClusterIP   None           <none>        27017/TCP   3m
mongodb-alert-demo-stats    ClusterIP   10.43.10.21    <none>        56790/TCP   3m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-mongodb
NAME                     AGE
mongodb-alert-demo-stats 3m

$ kubectl get servicemonitor -n alert-mongodb mongodb-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install mongodb-alerts

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression (via `job="{release-name}-stats"` / `app="{release-name}"`) from the **Helm release name** — so the release name must match the MongoDB object's name (`mongodb-alert-demo`).

### Install

```bash
$ helm upgrade -i mongodb-alert-demo oci://ghcr.io/appscode-charts/mongodb-alerts \
    -n alert-mongodb \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-mongodb
NAME                 AGE
mongodb-alert-demo   30s

$ kubectl get prometheusrule -n alert-mongodb mongodb-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `mongodb.database`, `mongodb.provisioner`, `mongodb.opsManager`, `mongodb.stash`, `mongodb.kubeStash`, and `mongodb.schemaManager` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/mongodb/monitoring/mongodb-alerting-prom-rules.png" style="padding:10px">
</p>

All groups should show **OK**. Every group declared in `values.yaml` is rendered for `mongodb-alerts` v2026.7.14 — no missing-group gap found here.

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-mongodb%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — mongodb-alert-demo-0 UP" src="/docs/images/mongodb/monitoring/mongodb-alerting-prom-target.png" style="padding:10px">
</p>

### 2. Confirm the MongoDB alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — MongoDB groups inactive" src="/docs/images/mongodb/monitoring/mongodb-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules should show **INACTIVE**, including `MongoDBDown` and `MongoDBPhaseCritical` — note these two are placed inside the `database` group even though, like the `provisioner` group's alerts, they key off `kubedb_com_mongodb_status_phase` (the KubeDB operator's own view), not a MongoDB-native metric. `MongoDBDown` fires much faster (`for: 30s`) than the provisioner group's `KubeDBMongoDBPhaseNotReady` (`for: 1m`).

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/mongodb/monitoring/mongodb-alerting-alertmanager.png" style="padding:10px">
</p>

### 4. Grafana dashboard

See [Grafana Dashboard](grafana-dashboard.md) for how to provision and explore the MongoDB dashboards (via the `kubedb-grafana-dashboards` chart, `--set featureGates.MongoDB=true`).

---

## Simulating a Firing Alert

This section deliberately triggers `MongoDBDown` (the fastest of the down-detection alerts, `for: 30s`) by crashing the main `mongod` process.

### 1. Crash the MongoDB process

```bash
$ kubectl exec -n alert-mongodb mongodb-alert-demo-0 -c mongodb -- sh -c '
    end=$(( $(date +%s) + 60 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -x mongod | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — MongoDBDown Firing" src="/docs/images/mongodb/monitoring/mongodb-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`MongoDBDown` (`kubedb_com_mongodb_status_phase{phase!="Ready"} == 1`, `for: 30s`) should transition to **FIRING** once the KubeDB operator observes the resource leaving `Ready`.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — MongoDBDown Firing" src="/docs/images/mongodb/monitoring/mongodb-alerting-alertmanager-firing.png" style="padding:10px">
</p>

### 4. Restore MongoDB

Stop the loop from step 1.

```bash
$ kubectl get mongodb -n alert-mongodb mongodb-alert-demo -w
NAME                 VERSION   STATUS   AGE
mongodb-alert-demo   6.0.14    Ready    24m
```

If MongoDB does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-mongodb mongodb-alert-demo-0`.

---

## Alert Reference

All alerts are scoped to the `mongodb-alert-demo` instance in the `alert-mongodb` namespace via the PromQL label filters `job="mongodb-alert-demo-stats"` / `namespace="alert-mongodb"` (most of the database group), or `app="mongodb-alert-demo"` / `namespace="alert-mongodb"` (provisioner/opsManager/stash/kubeStash/schemaManager groups, plus the two operator-phase alerts embedded in the database group).

### Database Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MongodbVirtualMemoryUsage` | warning | 1m | Virtual memory usage is high. |
| `MongodbReplicationLag` | critical | instant | Replica set member is lagging behind the primary. |
| `MongodbNumberCursorsOpen` | warning | 2m | Too many open cursors. |
| `MongodbCursorsTimeouts` | warning | 2m | Cursor timeout rate is increasing. |
| `MongodbTooManyConnections` | warning | 2m | Connection growth rate is high. |
| `MongoDBPhaseCritical` | warning | 10m | KubeDB operator view: resource `Critical` (embedded here, duplicates provisioner group's own version at a different `for`). |
| `MongoDBDown` | critical | 30s | KubeDB operator view: resource not `Ready`. Fastest down-signal available for MongoDB. |
| `MongoDBHighLatency` | warning | 10m | Operation latency is elevated. |
| `MongoDBHighTicketUtilization` | warning | 10m | WiredTiger concurrency tickets are close to exhausted. |
| `MongoDBRecurrentCursorTimeout` | warning | 30m | Cursor timeouts recurring over a longer window. |
| `MongoDBRecurrentMemoryPageFaults` | warning | 30m | Page faults recurring over a longer window. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80%. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95%. |

### Provisioner Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMongoDBPhaseNotReady` | critical | 1m | KubeDB marked the MongoDB resource `NotReady`. |
| `KubeDBMongoDBPhaseCritical` | warning | 15m | MongoDB is degraded but not fully unavailable. |

### OpsManager Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMongoDBOpsRequestStatusProgressingToLong` | critical | 30m | An ops request has been running for 30+ minutes. |
| `KubeDBMongoDBOpsRequestFailed` | critical | instant | An ops request failed. |

### Stash / KubeStash Groups

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `MongoDBStashBackupSessionFailed` / `MongoDBKubeStashBackupSessionFailed` | critical | instant | Most recent backup session failed. |
| `MongoDBStashRestoreSessionFailed` / `MongoDBKubeStashRestoreSessionFailed` | critical | instant | Most recent restore session failed. |
| `MongoDBStashNoBackupSessionForTooLong` / `MongoDBKubeStashNoBackupSessionForTooLong` | warning | instant | No recent successful backup. |
| `MongoDBStashRepositoryCorrupted` / `MongoDBKubeStashRepositoryCorrupted` | critical | 5m | Backup repository integrity check failed. |
| `MongoDBStashRepositoryStorageRunningLow` / `MongoDBKubeStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage usage is high. |
| `MongoDBStashBackupSessionPeriodTooLong` / `MongoDBKubeStashBackupSessionPeriodTooLong` | warning | instant | Backup session taking unusually long. |
| `MongoDBStashRestoreSessionPeriodTooLong` / `MongoDBKubeStashRestoreSessionPeriodTooLong` | warning | instant | Restore session taking unusually long. |

### SchemaManager Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBMongoDBSchemaPendingForTooLong` | warning | 30m | A `MongoDBDatabase` object stuck `Pending`. |
| `KubeDBMongoDBSchemaInProgressForTooLong` | warning | 30m | A `MongoDBDatabase` object stuck `InProgress`. |
| `KubeDBMongoDBSchemaTerminatingForTooLong` | warning | 30m | A `MongoDBDatabase` object stuck `Terminating`. |
| `KubeDBMongoDBSchemaFailed` | warning | instant | A `MongoDBDatabase` object failed. |
| `KubeDBMongoDBSchemaExpired` | warning | instant | A `MongoDBDatabase` object expired. |

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
          mongodbTooManyConnections:
            enabled: true
            duration: "5m"
            severity: warning
```

```bash
$ helm upgrade mongodb-alert-demo oci://ghcr.io/appscode-charts/mongodb-alerts \
    -n alert-mongodb \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

```bash
$ helm uninstall mongodb-alert-demo -n alert-mongodb
$ kubectl delete mongodb -n alert-mongodb mongodb-alert-demo
$ kubectl delete ns alert-mongodb
```

## Next Steps

- Monitor your MongoDB database with KubeDB using [built-in Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Monitor your MongoDB database with KubeDB using [Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
