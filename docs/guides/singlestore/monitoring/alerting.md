---
title: SingleStore Alerting with Prometheus
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-monitoring-alerting
    name: Alerting
    parent: guides-sdb-monitoring
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStore Alerting with Prometheus

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed SingleStore cluster using the `singlestore-alerts` Helm chart.

## Before You Begin

* Ensure you have a Kubernetes cluster and that `kubectl` is configured to communicate with it. If you do not already have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Install the KubeDB operator by following the steps [here](/docs/setup/README.md).

* Deploy the database in the `alert-singlestore` namespace:

  ```bash
  $ kubectl create ns alert-singlestore
  namespace/alert-singlestore created
  ```

* SingleStore requires a license. Create a secret with your license before deploying:

  ```bash
  $ kubectl create secret generic -n alert-singlestore license-secret \
      --from-literal=username=license \
      --from-literal=password='your-license-key-here'
  secret/license-secret created
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

* To learn more about how Prometheus monitoring works with KubeDB, see the overview [here](/docs/guides/singlestore/monitoring/overview/index.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/singlestore](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/singlestore) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

- **KubeDB** deploys SingleStore without a separate exporter image — metrics are obtained via the `memsql-admin` binary built into the SingleStore container itself, which KubeDB's operator configures automatically when `spec.monitor` is set.
- **ServiceMonitor** (named `{singlestore-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape metrics every 10 seconds.
- **PrometheusRule** is created by the `singlestore-alerts` chart and contains alert definitions grouped by concern: database health, provisioner, and KubeStash backup/restore.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).
- **Grafana** dashboards for SingleStore are covered separately — see [Grafana Dashboard](grafana-dashboard.md) rather than duplicated here.

---

## Deploy SingleStore with Monitoring Enabled

SingleStore is always deployed as a cluster of `aggregator` and `leaf` nodes. Below is the smallest viable topology for this tutorial — one aggregator, one leaf.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: singlestore-alert-demo
  namespace: alert-singlestore
spec:
  version: "8.5.7"
  storageType: Durable
  topology:
    aggregator:
      replicas: 1
      storage:
        storageClassName: "local-path"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 1
      storage:
        storageClassName: "local-path"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  licenseSecret:
    name: license-secret
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/singlestore/monitoring/singlestore-alert-demo.yaml
singlestore.kubedb.com/singlestore-alert-demo created
```

Wait for the cluster to go into `Ready` state.

```bash
$ kubectl get singlestore -n alert-singlestore singlestore-alert-demo
NAME                      VERSION   STATUS   AGE
singlestore-alert-demo    8.5.7     Ready    5m
```

KubeDB creates a dedicated stats service with the `-stats` suffix for monitoring.

```bash
$ kubectl get svc -n alert-singlestore --selector="app.kubernetes.io/instance=singlestore-alert-demo"
NAME                                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
singlestore-alert-demo              ClusterIP   10.43.10.20    <none>        3306/TCP    5m
singlestore-alert-demo-pods         ClusterIP   None           <none>        3306/TCP    5m
singlestore-alert-demo-stats        ClusterIP   10.43.10.21    <none>        56790/TCP   5m
```

KubeDB also creates a `ServiceMonitor` that tells Prometheus where to scrape.

```bash
$ kubectl get servicemonitor -n alert-singlestore
NAME                          AGE
singlestore-alert-demo-stats  5m

$ kubectl get servicemonitor -n alert-singlestore singlestore-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Install singlestore-alerts

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression (via `job="{release-name}-stats"` / `app="{release-name}"`) from the **Helm release name** — so the release name must match the SingleStore object's name (`singlestore-alert-demo`).

### Install

```bash
$ helm upgrade -i singlestore-alert-demo oci://ghcr.io/appscode-charts/singlestore-alerts \
    -n alert-singlestore \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus
```

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-singlestore
NAME                       AGE
singlestore-alert-demo     30s

$ kubectl get prometheusrule -n alert-singlestore singlestore-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Confirm Prometheus loaded the rules

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Open `http://localhost:9090/rules` and locate the `singlestore.database`, `singlestore.provisioner`, and `singlestore.kubeStash` groups.

<p align="center">
  <img alt="Prometheus Rule Health" src="/docs/images/singlestore/monitoring/singlestore-alerting-prom-rules.png" style="padding:10px">
</p>

All groups should show **OK**. Unlike several other `*-alerts` charts, `singlestore-alerts` v2026.7.14 has no `opsManager` group at all — its `values.yaml` only declares `database`, `provisioner`, and `kubeStash`.

---

## Verify End-to-End

### 1. Check the Prometheus target is UP

Open `http://localhost:9090/query?g0.expr=up%7Bnamespace%3D%22alert-singlestore%22%7D&g0.tab=1`.

<p align="center">
  <img alt="Prometheus up query — singlestore-alert-demo nodes UP" src="/docs/images/singlestore/monitoring/singlestore-alerting-prom-target.png" style="padding:10px">
</p>

Both the aggregator and leaf pods should report `up == 1`.

### 2. Confirm the SingleStore alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — SingleStore groups inactive" src="/docs/images/singlestore/monitoring/singlestore-alerting-prom-alerts.png" style="padding:10px">
</p>

All rules should show **INACTIVE**. `singlestore.kubeStash` rules stay INACTIVE with no data unless KubeStash backups are configured.

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/singlestore/monitoring/singlestore-alerting-alertmanager.png" style="padding:10px">
</p>

### 4. Grafana dashboard

See [Grafana Dashboard](grafana-dashboard.md) for how to provision and explore the SingleStore dashboards (via the `kubedb-grafana-dashboards` chart, `--set featureGates.Singlestore=true`).

---

## Simulating a Firing Alert

This section deliberately triggers `SinglestoreInstanceDown` (instant, `for: 0m`) by crashing the main process on a leaf node.

### 1. Crash a SingleStore leaf process

```bash
$ kubectl get pods -n alert-singlestore -l singlestore.com/node-type=leaf
$ kubectl exec -n alert-singlestore <leaf-pod-name> -c singlestore -- sh -c '
    end=$(( $(date +%s) + 30 ));
    while [ $(date +%s) -lt $end ]; do
      pid=$(pgrep -x memsqld | head -1);
      [ -n "$pid" ] && kill -9 "$pid" 2>/dev/null;
      sleep 1;
    done'
```

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — SinglestoreInstanceDown Firing" src="/docs/images/singlestore/monitoring/singlestore-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`SinglestoreInstanceDown` (`memsql_up == 0`) should transition straight to **FIRING** for the affected pod.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — SinglestoreInstanceDown Firing" src="/docs/images/singlestore/monitoring/singlestore-alerting-alertmanager-firing.png" style="padding:10px">
</p>

### 4. Restore SingleStore

Stop the loop from step 1.

```bash
$ kubectl get singlestore -n alert-singlestore singlestore-alert-demo -w
NAME                      VERSION   STATUS   AGE
singlestore-alert-demo    8.5.7     Ready    24m
```

If the node does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-singlestore <leaf-pod-name>`.

---

## Alert Reference

All database-group alerts are scoped to the `singlestore-alert-demo` instance via the PromQL label filters `job="singlestore-alert-demo-stats"` / `namespace="alert-singlestore"`; provisioner/kubeStash alerts use `app="singlestore-alert-demo"` / `namespace="alert-singlestore"`.

### Database Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `SinglestoreInstanceDown` | critical | instant | `memsql_up == 0` on a node. |
| `SinglestoreServiceDown` | critical | instant | No replica behind the service is answering. |
| `SinglestoreTooManyConnections` | warning | 2m | Connection count is high. |
| `SinglestoreHighThreadsRunning` | warning | 2m | Too many threads actively running. |
| `SinglestoreRestarted` | warning | instant | Uptime indicates a recent restart. |
| `SinglestoreHighQPS` | critical | instant | Query rate is unusually high. |
| `SinglestoreHighIncomingBytes` | critical | instant | Inbound network traffic is unusually high. |
| `SinglestoreHighOutgoingBytes` | critical | instant | Outbound network traffic is unusually high. |
| `DiskUsageHigh` | warning | 1m | Persistent volume usage exceeds 80%. |
| `DiskAlmostFull` | critical | 1m | Persistent volume usage exceeds 95%. |

### Provisioner Group

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `KubeDBSinglestorePhaseNotReady` | critical | 1m | KubeDB marked the SingleStore resource `NotReady`. |
| `KubeDBSinglestorePhaseCritical` | warning | 15m | SingleStore is degraded but not fully unavailable. |

### KubeStash Group

Only meaningful once KubeStash backup/restore is configured.

| Alert | Severity | For | What It Means |
|-------|----------|-----|---------------|
| `SinglestoreKubeStashBackupSessionFailed` | critical | instant | Most recent backup session failed. |
| `SinglestoreKubeStashRestoreSessionFailed` | critical | instant | Most recent restore session failed. |
| `SinglestoreKubeStashNoBackupSessionForTooLong` | warning | instant | No recent successful backup. |
| `SinglestoreKubeStashRepositoryCorrupted` | critical | 5m | Backup repository integrity check failed. |
| `SinglestoreKubeStashRepositoryStorageRunningLow` | warning | 5m | Backup repository storage usage is high. |
| `SinglestoreKubeStashBackupSessionPeriodTooLong` | warning | instant | A backup session is taking unusually long. |
| `SinglestoreKubeStashRestoreSessionPeriodTooLong` | warning | instant | A restore session is taking unusually long. |

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
          singlestoreTooManyConnections:
            enabled: true
            duration: "5m"
            severity: warning
      kubeStash:
        enabled: "none"    # disable if you don't use KubeStash
```

```bash
$ helm upgrade singlestore-alert-demo oci://ghcr.io/appscode-charts/singlestore-alerts \
    -n alert-singlestore \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

```bash
$ helm uninstall singlestore-alert-demo -n alert-singlestore
$ kubectl delete singlestore -n alert-singlestore singlestore-alert-demo
$ kubectl delete secret -n alert-singlestore license-secret
$ kubectl delete ns alert-singlestore
```

## Next Steps

- Monitor your SingleStore cluster with KubeDB using [built-in Prometheus](/docs/guides/singlestore/monitoring/builtin-prometheus/index.md).
- Monitor your SingleStore cluster with KubeDB using [Prometheus operator](/docs/guides/singlestore/monitoring/prometheus-operator/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
