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

This tutorial shows you how to configure Prometheus-based alerting for a KubeDB-managed SingleStore cluster using the `singlestore-alerts` Helm chart. This chart also bundles a Grafana dashboard that it imports automatically through a post-install Job — no separate dashboard chart is required.

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

* You will also need a Grafana API key / token with **Editor** permission so the chart's dashboard-import Job can push the dashboard. See [Step 1](#step-1--create-a-grafana-api-key) below.

> Note: YAML files used in this tutorial are stored in [docs/examples/singlestore](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/singlestore) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

<p align="center">
  <img alt="SingleStore Alerting Architecture" src="/docs/images/singlestore/monitoring/singlestore-alerting-overview.svg">
</p>

- **KubeDB** deploys SingleStore without a separate exporter image — metrics are obtained via the `memsql-admin` binary built into the SingleStore container itself, which KubeDB's operator configures automatically when `spec.monitor` is set.
- **ServiceMonitor** (named `{singlestore-name}-stats`) is created automatically by KubeDB and tells Prometheus to scrape metrics every 10 seconds.
- **PrometheusRule** is created by the `singlestore-alerts` chart and contains alert definitions grouped by concern: database health, provisioner, and KubeStash backup/restore.
- **Dashboard-import Job** — when `grafana.enabled` is `true`, the chart also creates a one-shot `Job` that `POST`s a bundled dashboard JSON straight to your Grafana instance's `/api/dashboards/import` endpoint.
- **Prometheus Operator** evaluates every rule expression every 30 seconds and fires matching alerts to AlertManager.
- **AlertManager** groups, inhibits, and silences alerts, then routes them to configured receivers (Slack, email, PagerDuty, webhook, etc.).

---

## Deploy SingleStore with Monitoring Enabled

SingleStore is always deployed as a cluster of `aggregator` and `leaf` nodes. Below is the topology used in this tutorial — one aggregator, two leaves.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: singlestore-alert-demo
  namespace: alert-singlestore
spec:
  version: "8.9.3"
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
            storage: 5Gi
    leaf:
      replicas: 2
      storage:
        storageClassName: "local-path"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
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

Here,

- `spec.topology.aggregator` / `spec.topology.leaf` define the aggregator and leaf node groups that make up the SingleStore cluster. The aggregator is a single point of coordination for the cluster — crashing it takes the whole cluster `NotReady`, while a single leaf going down only degrades the cluster (more leaves survive).
- `spec.licenseSecret.name` references the license secret created in [Before You Begin](#before-you-begin).
- `spec.monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` resource managed by the Prometheus operator.
- `spec.monitor.prometheus.serviceMonitor.labels.release: prometheus` adds the `release: prometheus` label to the created `ServiceMonitor`, matching the Prometheus `serviceMonitorSelector` so the target is discovered automatically.

Let's create the SingleStore resource.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/singlestore/monitoring/singlestore-alert-demo.yaml
singlestore.kubedb.com/singlestore-alert-demo created
```

Wait for the cluster to go into `Ready` state.

```bash
$ kubectl get singlestore -n alert-singlestore singlestore-alert-demo
NAME                      VERSION   STATUS   AGE
singlestore-alert-demo    8.9.3     Ready    5m
```

KubeDB brings up one aggregator pod and two leaf pods:

```bash
$ kubectl get pods -n alert-singlestore
NAME                                  READY   STATUS    RESTARTS   AGE
singlestore-alert-demo-aggregator-0   2/2     Running   0          5m
singlestore-alert-demo-leaf-0         2/2     Running   0          5m
singlestore-alert-demo-leaf-1         2/2     Running   0          4m
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
```

Verify that the `ServiceMonitor` carries the `release: prometheus` label so Prometheus discovers it.

```bash
$ kubectl get servicemonitor -n alert-singlestore singlestore-alert-demo-stats \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

---

## Step 1 — Create a Grafana API Key

The chart's dashboard-import Job authenticates to Grafana with a bearer token, so create one first.

* **Grafana 9+**: **Administration → Service accounts → Add service account** → role **Editor** → **Add token**. Copy the token.
* **Grafana 8.x and earlier** (no Service Accounts UI, e.g. the bundled `kube-prometheus-stack` Grafana 7.5.5): use the legacy **API Keys** endpoint instead:

  ```bash
  # Port-forward Grafana
  $ kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80

  # Retrieve the admin password
  $ kubectl get secret -n monitoring prometheus-grafana \
      -o jsonpath='{.data.admin-password}' | base64 -d && echo

  # Create an API key with Editor role
  $ curl -s -X POST -H "Content-Type: application/json" \
      -u admin:<grafana_password> \
      http://localhost:3000/api/auth/keys \
      -d '{"name":"singlestore-alerts-demo","role":"Editor"}'
  # Note the returned "key"

  # Stop the port-forward
  $ kill %1
  ```

Either way, you end up with a bearer token to use as `grafana.apikey` below.

## Step 2 — Install singlestore-alerts

The `singlestore-alerts` chart creates a `PrometheusRule` resource containing all SingleStore alert definitions.

### Why the Helm release name matters

The chart derives the `PrometheusRule` name and scopes every PromQL expression (via `job="{release-name}-stats"` / `app="{release-name}"`) from the **Helm release name** — so the release name must match the SingleStore object's name (`singlestore-alert-demo`).

### Install

```bash
$ helm upgrade -i singlestore-alert-demo appscode/singlestore-alerts \
    -n alert-singlestore \
    --create-namespace \
    --version=v2026.7.14 \
    --set form.alert.labels.release=prometheus \
    --set grafana.enabled=true \
    --set grafana.url="http://prometheus-grafana.monitoring.svc:80" \
    --set grafana.apikey="<token-from-above>" \
    --set grafana.jobName=singlestore-alert-demo-stats
```

| Flag | Value | Purpose |
|------|-------|---------|
| `grafana.url` | in-cluster Grafana URL | The dashboard-import Job runs **inside the cluster**, so this must be a cluster-internal address, not `localhost` |
| `grafana.apikey` | token from Step 1 | Authenticates the dashboard-import `POST` request |
| `grafana.jobName` | `singlestore-alert-demo-stats` | **Required** — the chart's default (`kubedb-databases`) doesn't match any real Prometheus job, so most of the dashboard's panels show "No data" unless you override it to your instance's actual stats-service name |

> To install **alerts only, without the dashboard**, omit the `grafana.*` flags (or set `--set grafana.enabled=false`).

### Verify the PrometheusRule is created

```bash
$ kubectl get prometheusrule -n alert-singlestore
NAME                       AGE
singlestore-alert-demo     30s

$ kubectl get prometheusrule -n alert-singlestore singlestore-alert-demo \
    -o jsonpath='{.metadata.labels.release}'
prometheus
```

### Verify the dashboard-import Job

```bash
$ kubectl get job -n alert-singlestore
NAME                             STATUS     COMPLETIONS   AGE
singlestore-alert-demo-post-job  Complete   1/1           17s

$ kubectl logs -n alert-singlestore job/singlestore-alert-demo-post-job
{"pluginId":"","title":"kubedb.com / Singlestore / alert-singlestore / singlestore-alert-demo","imported":true, ...}
```

A `"imported":true` response confirms the dashboard `kubedb.com / Singlestore / alert-singlestore / singlestore-alert-demo` now exists in Grafana.

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

Only **one** series is returned, scoped to `pod="singlestore-alert-demo-aggregator-0"`. This is not a query mistake — the `singlestore-alert-demo-stats` Service selects all three pods, but its Endpoints only ever resolve to the aggregator (`kubectl get endpoints -n alert-singlestore singlestore-alert-demo-stats` shows a single address). In this chart, **only the aggregator node exposes the Prometheus metrics endpoint** — the leaf pods don't. Every database-group alert (`SinglestoreInstanceDown`, `SinglestoreHighQPS`, etc.) is therefore effectively scoped to the aggregator only, not the leaves, regardless of how many leaf replicas you run.

### 2. Confirm the SingleStore alerts are inactive

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — SingleStore groups inactive" src="/docs/images/singlestore/monitoring/singlestore-alerting-prom-alerts.png" style="padding:10px">
</p>

All 10 rules in the `singlestore.database` group and all 7 rules in the `singlestore.kubeStash` group show **INACTIVE**. `singlestore.kubeStash` rules stay INACTIVE with no data unless KubeStash backups are configured.

### 3. Check AlertManager

```bash
$ kubectl port-forward -n monitoring \
    svc/prometheus-kube-prometheus-alertmanager 9093:9093
```

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager" src="/docs/images/singlestore/monitoring/singlestore-alerting-alertmanager.png" style="padding:10px">
</p>

---

## Simulating a Firing Alert

This section deliberately triggers `KubeDBSinglestorePhaseNotReady` by crashing the **aggregator** node. With `spec.topology.aggregator.replicas: 1`, the aggregator is a single point of coordination for the whole cluster — crashing it repeatedly holds the CR's `status.phase` at `NotReady` long enough for the provisioner-group alert to fire, unlike crashing a single leaf (which only degrades the cluster, since a second leaf survives).

### 1. Crash the aggregator process repeatedly

```bash
$ end=$(( $(date +%s) + 120 ))
  while [ $(date +%s) -lt $end ]; do
    kubectl exec -n alert-singlestore singlestore-alert-demo-aggregator-0 -c singlestore -- sh -c 'pid=$(pgrep -x memsqld | head -1); [ -n "$pid" ] && kill -9 "$pid"' >/dev/null 2>&1
    sleep 5
  done
```

Watch the CR phase move to `NotReady`:

```bash
$ kubectl get singlestore -n alert-singlestore singlestore-alert-demo -o jsonpath='{.status.phase}'
NotReady
```

`KubeDBSinglestorePhaseNotReady` keys off `kubedb_com_singlestore_status_phase` (a metric emitted by the KubeDB operator itself, via panopticon), so what matters is the CR's `status.phase` staying at `NotReady` for the full `for: 1m` window, not the aggregator's own scrape health.

### 2. Watch the alert fire in Prometheus

Open `http://localhost:9090/alerts`.

<p align="center">
  <img alt="Prometheus Alerts — KubeDBSinglestorePhaseNotReady Firing" src="/docs/images/singlestore/monitoring/singlestore-alerting-prom-alerts-firing.png" style="padding:10px">
</p>

`KubeDBSinglestorePhaseNotReady` (in the `singlestore.provisioner` group) transitions from **INACTIVE** to **FIRING**, while `singlestore.database` and `singlestore.kubeStash` stay **INACTIVE** — crashing the aggregator's `memsqld` process doesn't take the metrics scrape target down (the pod's `singlestore` container restarts quickly), so `SinglestoreInstanceDown` never fires here; it's specifically the operator's own phase tracking that reacts.

### 3. Check the AlertManager dashboard

Open `http://localhost:9093`.

<p align="center">
  <img alt="AlertManager — KubeDBSinglestorePhaseNotReady Firing" src="/docs/images/singlestore/monitoring/singlestore-alerting-alertmanager-firing.png" style="padding:10px">
</p>

AlertManager shows the `KubeDBSinglestorePhaseNotReady` alert. The alert card displays labels including:

- **alertname**: `KubeDBSinglestorePhaseNotReady`
- **severity**: `critical`
- **app**: `singlestore-alert-demo`, **app_namespace**: `alert-singlestore`
- **phase**: `NotReady`
- **k8s_kind**: `Singlestore`

Note that the `instance`/`pod`/`job` labels point at the KubeDB operator's **panopticon** component (`job="panopticon"`), not at the SingleStore pod itself — because this alert is derived from the operator's own status metric.

### 4. Restore SingleStore

Stop the loop from step 1.

```bash
$ kubectl get singlestore -n alert-singlestore singlestore-alert-demo -w
NAME                      VERSION   STATUS   AGE
singlestore-alert-demo    8.9.3     Ready    24m
```

If the aggregator does not recover on its own within a minute or two, force a clean restart: `kubectl delete pod -n alert-singlestore singlestore-alert-demo-aggregator-0`.

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
$ helm upgrade singlestore-alert-demo appscode/singlestore-alerts \
    -n alert-singlestore \
    --version=v2026.7.14 \
    -f custom-alerts.yaml
```

---

## Cleaning up

To remove all resources created in this tutorial, run the following commands.

```bash
# Remove the singlestore-alerts release (PrometheusRule + dashboard-import Job)
$ helm uninstall singlestore-alert-demo -n alert-singlestore

# Remove the imported Grafana dashboard (it is not removed by helm uninstall)
$ curl -s -X DELETE -H "Authorization: Bearer <grafana-token>" \
    http://localhost:3000/api/dashboards/uid/<uid>

$ kubectl delete singlestore -n alert-singlestore singlestore-alert-demo
$ kubectl delete secret -n alert-singlestore license-secret
$ kubectl delete ns alert-singlestore
```

## Next Steps

- Monitor your SingleStore cluster with KubeDB using [built-in Prometheus](/docs/guides/singlestore/monitoring/builtin-prometheus/index.md).
- Monitor your SingleStore cluster with KubeDB using [Prometheus operator](/docs/guides/singlestore/monitoring/prometheus-operator/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
