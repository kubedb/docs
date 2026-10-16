---
title: RabbitMQ Grafana Dashboard
menu:
  docs_{{ .version }}:
    identifier: rm-grafana-dashboard-monitoring
    name: Grafana Dashboard
    parent: rm-monitoring-guides
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RabbitMQ Grafana Dashboard

KubeDB exposes RabbitMQ metrics through a built-in Prometheus plugin, and its own view of each resource (status, phase, version) through Panopticon. Once Prometheus is scraping both, you can visualize them in Grafana using a pre-built KubeDB dashboard. This tutorial walks through the full setup: deploying the monitoring stack, enabling monitoring on a RabbitMQ instance, and importing the Grafana dashboard.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- KubeDB must be installed in your cluster with `kubedb-metrics` enabled. Follow the setup guide [here](/docs/setup/README.md) and make sure to include the flag below during installation:

  ```bash
  --set kubedb-metrics.enabled=true
  ```

  `kubedb-metrics` creates `MetricsConfiguration` objects for each database type, which Panopticon (see [Configuration](/docs/guides/rabbitmq/monitoring/using-prometheus-operator.md#configuration)) uses to expose metrics to Prometheus.

- To keep monitoring resources isolated, we use a separate `monitoring` namespace and deploy the database in the `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

* Before proceeding, complete the [Configuration](/docs/guides/rabbitmq/monitoring/using-prometheus-operator.md#configuration) steps to deploy **kube-prometheus-stack** and **Panopticon**.

> Note: YAML files used in this tutorial are stored in [docs/examples/rabbitmq/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/rabbitmq/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Setup

## Step 1: Deploy RabbitMQ

Below is the RabbitMQ object with monitoring configured to use Prometheus Operator.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rmq-grafana-demo
  namespace: demo
spec:
  version: "4.0.4"
  replicas: 1
  deletionPolicy: WipeOut
  storage:
    storageClassName: "standard"
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

- `monitor.agent: prometheus.io/operator` tells KubeDB to create a `ServiceMonitor` for this instance.
- `monitor.prometheus.serviceMonitor.labels` must match the `serviceMonitorSelector` label of your Prometheus (`release: prometheus`).
- `monitor.prometheus.serviceMonitor.interval` sets the scrape interval to 10 seconds.

Create the RabbitMQ instance:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/monitoring/coreos-prom-rabbitmq.yaml
rabbitmq.kubedb.com/rmq-grafana-demo created
```

Wait for it to be `Ready`:

```bash
$ kubectl get rabbitmq -n demo rmq-grafana-demo
NAME               VERSION   STATUS   AGE
rmq-grafana-demo   4.0.4     Ready    2m
```

KubeDB creates a stats service named `{rabbitmq-name}-stats` for monitoring:

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=rmq-grafana-demo"
NAME                     TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)      AGE
rmq-grafana-demo         ClusterIP   10.96.10.1     <none>        5672/TCP     2m
rmq-grafana-demo-stats   ClusterIP   10.96.10.2     <none>        15692/TCP    2m
```

KubeDB also creates a `ServiceMonitor` in the `demo` namespace:

```bash
$ kubectl get servicemonitor -n demo
NAME                     AGE
rmq-grafana-demo-stats   2m
```

Verify it carries the correct label:

```bash
$ kubectl get servicemonitor -n demo rmq-grafana-demo-stats -o jsonpath='{.metadata.labels}'
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

Open [http://localhost:9090/targets](http://localhost:9090/targets) in your browser. Look for an entry whose `service` label matches `rmq-grafana-demo-stats`. Its state should be **UP**.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/rabbitmq/monitoring/rmq-prom-targets.png" style="padding:10px">
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
  <img alt="Grafana Login" src="/docs/images/rabbitmq/monitoring/rmq-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/rabbitmq/monitoring/rmq-grafana-home.png" style="padding:10px">
</p>

## Step 4: Configure Prometheus as a Data Source

If you installed Grafana via `kube-prometheus-stack`, Prometheus is already configured as the default data source — skip to Step 5.

If you're using a different Grafana instance than the one installed in the Configuration prerequisite (linked in "Before You Begin" above), add Prometheus as a data source manually:

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

The chart packages dashboard JSON for every database it supports, so install it from a copy trimmed down to just RabbitMQ's dashboards (this also keeps `grafana-operator` from creating dashboards for databases you don't run):

```bash
$ helm pull appscode/kubedb-grafana-dashboards --version v2026.7.10 --untar

$ cd kubedb-grafana-dashboards/dashboards
$ ls | grep -v '^rabbitmq$' | xargs rm -rf   # keep only dashboards/rabbitmq
$ cd ../..

$ helm package kubedb-grafana-dashboards
Successfully packaged chart and saved it to: kubedb-grafana-dashboards-v2026.7.10.tgz

$ helm upgrade -i kubedb-grafana-dashboards-rabbitmq ./kubedb-grafana-dashboards-v2026.7.10.tgz \
    -n kubeops --create-namespace \
    --set grafana.name=grafana \
    --set grafana.namespace=monitoring
```

> Use a release name unique to this database (`kubedb-grafana-dashboards-rabbitmq`), not the plain `kubedb-grafana-dashboards` name — if you also follow another DB's Grafana Dashboard guide on this same cluster, each `helm upgrade -i` under a *shared* release name would prune the previous DB's dashboards (Helm removes anything not in the new release's manifest). A per-DB release name lets them coexist.

`grafana.name`/`grafana.namespace` point the chart's `GrafanaDashboard` resources at the `AppBinding` created in step 2 (omitting them falls back to whichever `AppBinding` in your cluster is labeled as the cluster-default Grafana, if any — explicit is safer on a shared cluster). No need to touch `featureGates` — with every other database's `dashboards/` folder removed, their gates simply match zero files and render nothing, regardless of being `true` by default.

This creates every dashboard the chart ships for RabbitMQ — `KubeDB / RabbitMQ / Summary`, `KubeDB / RabbitMQ / Pod`, `KubeDB / RabbitMQ / Database` — which `grafana-operator` then pushes into your Grafana instance automatically. No manual JSON download or upload needed.

Verify they landed:

```bash
$ kubectl get grafanadashboards.openviz.dev -n kubeops | grep -i rabbitmq
NAME                                TITLE                           SYNCED    AGE
grafana-kubedb-rabbitmq-summary     KubeDB / RabbitMQ / Summary     Current   30s
grafana-kubedb-rabbitmq-pod         KubeDB / RabbitMQ / Pod         Current   30s
grafana-kubedb-rabbitmq-database    KubeDB / RabbitMQ / Database    Current   30s
```

`SYNCED: Current` confirms `grafana-operator` successfully pushed each dashboard into Grafana. Open Grafana — the dashboard are already there under `Dashboards`, fully wired to your Prometheus data source, ready to explore in Step 6 below.

If the `SYNCED` column is missing entirely from your output (not just showing a non-`Current` value), `grafana-operator` most likely never processed the resource at all. Check that the operator pod is actually running (`kubectl get pods -n kubeops -l app.kubernetes.io/name=grafana-operator`), inspect its logs for `AppBinding`/auth errors (`kubectl logs -n kubeops deploy/grafana-operator`), and check `kubectl get grafanadashboards.openviz.dev -n kubeops <name> -o yaml` for `status.conditions` — the CR existing only means `kubectl` accepted it, not that it reached Grafana.

After importing them, they will appear under `Dashboards` in the left sidebar as well:

| Dashboard Name | Description |
|---|---|
| KubeDB / RabbitMQ / Summary | Node/queue/connection overview, message rates, memory/disk alarms, CPU/memory/storage |
| KubeDB / RabbitMQ / Pod | Per-pod message rates, file descriptors, memory usage, CPU/memory |
| KubeDB / RabbitMQ / Database | Queue depth, consumer utilisation, message age, publish/deliver/ack rates per queue |

## Step 5: Import Dashboard — Option B: Manual (JSON upload)

If you'd rather not run `grafana-operator`, or want fine-grained control over exactly which dashboards get imported, upload the same dashboard JSON files by hand instead.

The KubeDB RabbitMQ dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download all three JSON files from the [appscode/grafana-dashboards](https://github.com/appscode/grafana-dashboards/tree/master/rabbitmq) repository (`rabbitmq/` folder):

| File | Dashboard |
|------|-----------|
| `rabbitmq_summary_dashboard.json` | KubeDB / RabbitMQ / Summary |
| `rabbitmq_pods_dashboard.json` | KubeDB / RabbitMQ / Pod |
| `rabbitmq_databases_dashboard.json` | KubeDB / RabbitMQ / Database |

**Import steps (repeat for each of the three files):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

The import page looks like this — click **Upload dashboard JSON file** to select the file:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/rabbitmq/monitoring/rmq-grafana-import.png" style="padding:10px">
</p>

After importing all three files, they will appear under `Dashboards` in the left sidebar.

## Step 6: Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable      | Applies to              | What to select                                              |
|---------------|-------------------------|-------------------------------------------------------------|
| **namespace** | All dashboards          | Namespace where your RabbitMQ is deployed (e.g., `demo`)   |
| **app**       | All dashboards          | Name of your RabbitMQ instance (e.g., `rmq-grafana-demo`)  |
| **pod**       | Pod, Database dashboards | A specific pod, or `All` for an aggregated view            |
| **vhost**     | Database dashboard only | A specific virtual host, or `All`                          |

**KubeDB / RabbitMQ / Summary** — start here for a node and cluster overview:
- **Node Health** — running/stopped/disk alarm/memory alarm status per node
- **Queue Count** — total queues in the cluster
- **Message Rates** — publish rate, deliver rate, acknowledge rate
- **Messages Ready / Unacknowledged** — total backlog depth
- **Connection / Channel Count** — active connections and open channels
- **CPU / Memory / Disk Free** — resource consumption per node

<p align="center">
  <img alt="KubeDB RabbitMQ Summary Dashboard" src="/docs/images/rabbitmq/monitoring/rmq-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / RabbitMQ / Pod** — drill into a specific node:
- **Erlang Process Count** — number of Erlang processes (high counts indicate load)
- **Memory Breakdown** — code, heap, binaries, ETS table memory
- **Socket Descriptors** — used vs. available file descriptors for connections
- **GC** — garbage collection runs and bytes reclaimed per second
- **CPU / Memory** — per-pod resource usage over time

<p align="center">
  <img alt="KubeDB RabbitMQ Pod Dashboard" src="/docs/images/rabbitmq/monitoring/rmq-grafana-pod.png" style="padding:10px">
</p>

**KubeDB / RabbitMQ / Database** — per-queue and per-vhost metrics:
- **Queue Depth** — messages ready + unacknowledged per queue
- **Publish / Deliver Rate** — throughput per queue
- **Consumer Count** — active consumers per queue
- **Oldest Unacknowledged Message** — age of the oldest pending message (latency indicator)

<p align="center">
  <img alt="KubeDB RabbitMQ Database Dashboard" src="/docs/images/rabbitmq/monitoring/rmq-grafana-database.png" style="padding:10px">
</p>

## Cleaning up

```bash
# Remove the RabbitMQ instance
kubectl delete rabbitmq -n demo rmq-grafana-demo

# Remove namespaces
kubectl delete ns demo

# Uninstall the Grafana dashboards chart, if you used Option A
helm uninstall kubedb-grafana-dashboards-rabbitmq -n kubeops

# Uninstall grafana-operator (optional — skip if other DB guides on this cluster still use it)
helm uninstall grafana-operator -n kubeops

# Uninstall monitoring stack (optional)
helm uninstall prometheus -n monitoring
helm uninstall panopticon -n kubeops
kubectl delete ns monitoring kubeops
```

## Next Steps

- Monitor your RabbitMQ instance with KubeDB using [built-in Prometheus](/docs/guides/rabbitmq/monitoring/using-builtin-prometheus.md).
- Monitor your RabbitMQ instance with KubeDB using [Prometheus Operator](/docs/guides/rabbitmq/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
