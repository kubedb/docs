---
title: Monitoring Redis Using Builtin Prometheus Discovery
menu:
  docs_{{ .version }}:
    identifier: rd-using-builtin-prometheus-monitoring
    name: Builtin Prometheus
    parent: rd-monitoring-redis
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Redis with builtin Prometheus

This tutorial will show you how to monitor Redis server using builtin [Prometheus](https://github.com/prometheus/prometheus) scraper.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- If you are not familiar with how to configure Prometheus to scrape metrics from various Kubernetes resources, please read the tutorial from [here](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/redis/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Redis with Monitoring Enabled

At first, let's deploy an Redis server with monitoring enabled. Below is the Redis object that we are going to create.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: builtin-prom-redis
  namespace: demo
spec:
  version: 8.2.2
  deletionPolicy: WipeOut
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/builtin
```

Here,

- `spec.monitor.agent: prometheus.io/builtin` specifies that we are going to monitor this server using builtin Prometheus scraper.

Let's create the Redis crd we have shown above.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/monitoring/builtin-prom-redis.yaml
redis.kubedb.com/builtin-prom-redis created
```

Now, wait for the database to go into `Running` state.

```bash
$ kubectl get rd -n demo builtin-prom-redis
NAME                 VERSION   STATUS    AGE
builtin-prom-redis   4.0-v1    Running   41s
```

KubeDB will create a separate stats service with name `{Redis crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=builtin-prom-redis"
NAME                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
builtin-prom-redis         ClusterIP   10.109.162.108   <none>        6379/TCP    59s
builtin-prom-redis-stats   ClusterIP   10.106.243.251   <none>        56790/TCP   41s
```

Here, `builtin-prom-redis-stats` service has been created for monitoring purpose. Let's describe the service.

```bash
$ kubectl describe svc -n demo builtin-prom-redis-stats
Name:              builtin-prom-redis-stats
Namespace:         demo
Labels:            app.kubernetes.io/name=redises.kubedb.com
                   app.kubernetes.io/instance=builtin-prom-redis
Annotations:       monitoring.appscode.com/agent: prometheus.io/builtin
                   prometheus.io/path: /metrics
                   prometheus.io/port: 56790
                   prometheus.io/scrape: true
Selector:          app.kubernetes.io/name=redises.kubedb.com,app.kubernetes.io/instance=builtin-prom-redis
Type:              ClusterIP
IP:                10.106.243.251
Port:              prom-http  56790/TCP
TargetPort:        prom-http/TCP
Endpoints:         172.17.0.14:56790
Session Affinity:  None
Events:            <none>
```

You can see that the service contains following annotations.

```bash
prometheus.io/path: /metrics
prometheus.io/port: 56790
prometheus.io/scrape: true
```

The Prometheus server will discover the service endpoint using these specifications and will scrape metrics from the exporter.

## Configure Prometheus Server

Now, we have to configure a Prometheus scraping job to scrape the metrics using this service. We are going to configure scraping job similar to this [kubernetes-service-endpoints](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin#kubernetes-service-endpoints) job that scrapes metrics from endpoints of a service.

Let's configure a Prometheus scraping job to collect metrics from this service.

```yaml
- job_name: 'kubedb-databases'
  honor_labels: true
  scheme: http
  kubernetes_sd_configs:
  - role: endpoints
  # by default Prometheus server select all Kubernetes services as possible target.
  # relabel_config is used to filter only desired endpoints
  relabel_configs:
  # keep only those services that has "prometheus.io/scrape","prometheus.io/path" and "prometheus.io/port" anootations
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape, __meta_kubernetes_service_annotation_prometheus_io_port]
    separator: ;
    regex: true;(.*)
    action: keep
  # currently KubeDB supported databases uses only "http" scheme to export metrics. so, drop any service that uses "https" scheme.
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
    action: drop
    regex: https
  # only keep the stats services created by KubeDB for monitoring purpose which has "-stats" suffix
  - source_labels: [__meta_kubernetes_service_name]
    separator: ;
    regex: (.*-stats)
    action: keep
  # service created by KubeDB will have "app.kubernetes.io/name" and "app.kubernetes.io/instance" annotations. keep only those services that have these annotations.
  - source_labels: [__meta_kubernetes_service_label_app_kubernetes_io_name]
    separator: ;
    regex: (.*)
    action: keep
  # read the metric path from "prometheus.io/path: <path>" annotation
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
    action: replace
    target_label: __metrics_path__
    regex: (.+)
  # read the port from "prometheus.io/port: <port>" annotation and update scraping address accordingly
  - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
    action: replace
    target_label: __address__
    regex: ([^:]+)(?::\d+)?;(\d+)
    replacement: $1:$2
  # add service namespace as label to the scraped metrics
  - source_labels: [__meta_kubernetes_namespace]
    separator: ;
    regex: (.*)
    target_label: namespace
    replacement: $1
    action: replace
  # add service name as a label to the scraped metrics
  - source_labels: [__meta_kubernetes_service_name]
    separator: ;
    regex: (.*)
    target_label: service
    replacement: $1
    action: replace
  # add stats service's labels to the scraped metrics
  - action: labelmap
    regex: __meta_kubernetes_service_label_(.+)
```

### Configure Existing Prometheus Server

If you already have a Prometheus server running, you have to add above scraping job in the `ConfigMap` used to configure the Prometheus server. Then, you have to restart it for the updated configuration to take effect.

>If you don't use a persistent volume for Prometheus storage, you will lose your previously scraped data on restart.

### Deploy New Prometheus Server

If you don't have any existing Prometheus server running, you have to deploy one. In this section, we are going to deploy a Prometheus server in `monitoring` namespace to collect metrics using this stats service.

**Create ConfigMap:**

At first, create a ConfigMap with the scraping configuration. Bellow, the YAML of ConfigMap that we are going to create in this tutorial.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  labels:
    app: prometheus-demo
  namespace: monitoring
data:
  prometheus.yml: |-
    global:
      scrape_interval: 5s
      evaluation_interval: 5s
    scrape_configs:
    - job_name: 'kubedb-databases'
      honor_labels: true
      scheme: http
      kubernetes_sd_configs:
      - role: endpoints
      # by default Prometheus server select all Kubernetes services as possible target.
      # relabel_config is used to filter only desired endpoints
      relabel_configs:
      # keep only those services that has "prometheus.io/scrape","prometheus.io/path" and "prometheus.io/port" anootations
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape, __meta_kubernetes_service_annotation_prometheus_io_port]
        separator: ;
        regex: true;(.*)
        action: keep
      # currently KubeDB supported databases uses only "http" scheme to export metrics. so, drop any service that uses "https" scheme.
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: drop
        regex: https
      # only keep the stats services created by KubeDB for monitoring purpose which has "-stats" suffix
      - source_labels: [__meta_kubernetes_service_name]
        separator: ;
        regex: (.*-stats)
        action: keep
      # service created by KubeDB will have "app.kubernetes.io/name" and "app.kubernetes.io/instance" annotations. keep only those services that have these annotations.
      - source_labels: [__meta_kubernetes_service_label_app_kubernetes_io_name]
        separator: ;
        regex: (.*)
        action: keep
      # read the metric path from "prometheus.io/path: <path>" annotation
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      # read the port from "prometheus.io/port: <port>" annotation and update scraping address accordingly
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      # add service namespace as label to the scraped metrics
      - source_labels: [__meta_kubernetes_namespace]
        separator: ;
        regex: (.*)
        target_label: namespace
        replacement: $1
        action: replace
      # add service name as a label to the scraped metrics
      - source_labels: [__meta_kubernetes_service_name]
        separator: ;
        regex: (.*)
        target_label: service
        replacement: $1
        action: replace
      # add stats service's labels to the scraped metrics
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
```

Let's create above `ConfigMap`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/monitoring/builtin-prometheus/prom-config.yaml
configmap/prometheus-config created
```

**Create RBAC:**

If you are using an RBAC enabled cluster, you have to give necessary RBAC permissions for Prometheus. Let's create necessary RBAC stuffs for Prometheus,

```bash
$ kubectl apply -f https://github.com/appscode/third-party-tools/raw/master/monitoring/prometheus/builtin/artifacts/rbac.yaml
clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created
```

>YAML for the RBAC resources created above can be found [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/artifacts/rbac.yaml).

**Deploy Prometheus:**

Now, we are ready to deploy Prometheus server. We are going to use following [deployment](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/artifacts/deployment.yaml) to deploy Prometheus server.

Let's deploy the Prometheus server.

```bash
$ kubectl apply -f https://github.com/appscode/third-party-tools/raw/master/monitoring/prometheus/builtin/artifacts/deployment.yaml
deployment.apps/prometheus created
```

### Verify Monitoring Metrics

Prometheus server is listening to port `9090`. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

At first, let's check if the Prometheus pod is in `Running` state.

```bash
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                          READY   STATUS    RESTARTS   AGE
prometheus-8568c86d86-95zhn   1/1     Running   0          77s
```

Now, run following command on a separate terminal to forward 9090 port of `prometheus-8568c86d86-95zhn` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-8568c86d86-95zhn 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see the endpoint of `builtin-prom-redis-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" height="100%" src="/docs/images/redis/monitoring/redis-builtin-prom-target.png" style="padding:10px">
</p>

Check the labels marked with red rectangle. These labels confirm that the metrics are coming from `Redis` server `builtin-prom-redis` through stats service `builtin-prom-redis-stats`.

Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create a dashboard with the collected metrics, as shown below.

## Visualize Metrics with Grafana

### Expose the Prometheus Server

Grafana needs to reach Prometheus over an in-cluster address, so expose the `prometheus` deployment through a `Service` first.

```bash
$ kubectl expose deployment prometheus -n monitoring --port=9090 --target-port=9090 --name=prometheus
service/prometheus exposed
```

### Deploy Grafana

This tutorial's Prometheus was deployed manually (not via `kube-prometheus-stack`), so Grafana doesn't come bundled — install it separately with Helm.

```bash
$ helm repo add grafana https://grafana.github.io/helm-charts
$ helm repo update

$ helm upgrade --install grafana grafana/grafana -n monitoring --set persistence.enabled=false
```

Wait for the Grafana pod to be ready:

```bash
$ kubectl get pods -n monitoring -l=app.kubernetes.io/name=grafana
NAME                       READY   STATUS    RESTARTS   AGE
grafana-6b9c8f9c7d-x2n4p   1/1     Running   0          45s
```

Port-forward the Grafana service:

```bash
$ kubectl port-forward -n monitoring svc/grafana 3000:80
Forwarding from 127.0.0.1:3000 -> 3000
```

Open [http://localhost:3000](http://localhost:3000). The username is `admin`. Retrieve the auto-generated password from the secret:

```bash
$ kubectl get secret -n monitoring grafana -o jsonpath='{.data.admin-password}' | base64 -d
```

| Field    | Value                       |
|----------|-----------------------------|
| Username | `admin`                     |
| Password | output of the command above |

<p align="center">
  <img alt="Grafana Login" src="/docs/images/redis/monitoring/rd-grafana-login.png" style="padding:10px">
</p>

After a successful login you will see the Grafana home page:

<p align="center">
  <img alt="Grafana Home" src="/docs/images/redis/monitoring/rd-grafana-home.png" style="padding:10px">
</p>

### Configure Prometheus as a Data Source

1. Go to **Connections** → **Data sources** → **Add new data source**.
2. Select **Prometheus**.
3. Set the URL to the Prometheus service exposed above:

   ```
   http://prometheus.monitoring.svc:9090
   ```

4. Click **Save & test**. You should see `Data source is working`.

### Import KubeDB Redis Dashboard — Option A: Automatically, via the `kubedb-grafana-dashboards` chart

Rather than downloading and uploading each JSON file by hand (Option B below), KubeDB ships a chart that creates all matching dashboards for you as `GrafanaDashboard` custom resources. A separate controller, **`grafana-operator`**, watches these resources and pushes the actual dashboard JSON into your Grafana instance — both pieces are required.

**1. Install `grafana-operator`** (skip if it's already running in your cluster):

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm upgrade --install grafana-operator appscode/grafana-operator \
    --version v2026.6.12 \
    --namespace kubeops --create-namespace
```

**2. Register your Grafana instance as an `AppBinding`.** `grafana-operator` needs to know where to push dashboards and how to authenticate — it reads this from an `AppBinding` object, not from the chart install command itself. Create an API key for the Grafana deployed above, store it in a Secret, and reference both in an `AppBinding`:

```bash
$ kubectl port-forward -n monitoring svc/grafana 3000:80 &

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
    url: http://grafana.monitoring.svc:80
  secret:
    name: grafana-admin-token
EOF
```

**3. Install the dashboards:**

```bash
$ helm upgrade -i kubedb-grafana-dashboards appscode/kubedb-grafana-dashboards \
    -n kubeops --create-namespace --version=v2026.8.14-rc.0 \
    --set featureGates.Redis=true \
    --set grafana.name=grafana \
    --set grafana.namespace=monitoring
```

`featureGates.Redis` already defaults to `true` — set explicitly above for clarity. `grafana.name`/`grafana.namespace` point the chart's `GrafanaDashboard` resources at the `AppBinding` created in step 2 (omitting them falls back to whichever `AppBinding` in your cluster is labeled as the cluster-default Grafana, if any — explicit is safer on a shared cluster).

This single command creates every dashboard this chart ships for Redis — `KubeDB / Redis / Summary`, `KubeDB / Redis / Pod`, `KubeDB / Redis / Shard` (plus the `RedisSentinel` variants) — which `grafana-operator` then pushes into your Grafana instance automatically. No manual JSON download or upload needed.

Verify they landed:

```bash
$ kubectl get grafanadashboards.openviz.dev -n kubeops | grep -i redis
NAME                    TITLE                     SYNCED    AGE
kubedb-redis-pod        KubeDB / Redis / Pod       Current   30s
kubedb-redis-shard      KubeDB / Redis / Shard     Current   30s
kubedb-redis-summary    KubeDB / Redis / Summary   Current   30s
```

`SYNCED: Current` confirms `grafana-operator` successfully pushed each dashboard into Grafana. Open Grafana — the three dashboards are already there under `Dashboards`, fully wired to your Prometheus data source, ready to explore in [Step 6](#explore-the-dashboard) below.

### Import KubeDB Redis Dashboard — Option B: Manually, by uploading JSON files

If you'd rather not run `grafana-operator`, or want fine-grained control over exactly which dashboards get imported, upload the same dashboard JSON files by hand instead.

The KubeDB Redis dashboards are distributed as JSON files. Each JSON file is a complete dashboard definition — panels, queries, variables, and layout — that Grafana loads in one shot. Without importing, you would have to build every panel and write every PromQL query by hand. Importing lets you skip that entirely.

Three dashboards are available. Download the JSON files from the [opnpulse/dashboards](https://github.com/opnpulse/dashboards/tree/master/redis) repository (`redis/` folder):

| File | Dashboard |
|------|-----------|
| `redis_summary_dashboard.json` | KubeDB / Redis / Summary |
| `redis_pod_dashboard.json` | KubeDB / Redis / Pod |
| `redis_shards_dashboard.json` | KubeDB / Redis / Shard |

> The Shard dashboard is relevant for Redis Cluster mode (`spec.mode: Cluster`); its panels stay empty for a standalone (non-cluster) Redis instance like the one deployed in this tutorial.

**Import steps (repeat for each file you need):**

1. In Grafana, click the `+` icon in the left sidebar.
2. Select `Import` from the menu.
3. Click `Upload JSON file` and select one of the downloaded `.json` files.
4. In the `Prometheus` dropdown that appears, select your Prometheus data source.
5. Click `Import`.

The import page looks like this — click **Upload dashboard JSON file** to select the file:

<p align="center">
  <img alt="Grafana Import Dashboard" src="/docs/images/redis/monitoring/rd-grafana-import.png" style="padding:10px">
</p>

After importing the files you need, they will appear under `Dashboards` in the left sidebar.

| Dashboard Name | Description |
|---|---|
| KubeDB / Redis / Summary | Instance overview: status, version, mode, node count, resource requests/limits, CPU usage |
| KubeDB / Redis / Pod | Per-pod role, master/slaves, connected clients, memory, commands/sec, network I/O, CPU/memory |
| KubeDB / Redis / Shard | Cluster shard slot health, node/slave count, per-slave status, cluster mode |

### Explore the Dashboard

After opening a dashboard, use the dropdown filters at the top to focus on a specific instance.

| Variable       | Applies to              | What to select                                             |
|----------------|--------------------------|--------------------------------------------------------------|
| **datasource** | All dashboards          | Your Prometheus data source                                |
| **Namespace**  | All dashboards          | Namespace where your Redis is deployed (e.g., `demo`)      |
| **app**        | Summary dashboard       | Name of your Redis instance (e.g., `builtin-prom-redis`)   |
| **redis**      | Pod, Shard dashboards   | Name of your Redis instance (e.g., `builtin-prom-redis`)   |
| **Pod Name**   | Pod, Shard dashboards   | A specific pod (e.g., `builtin-prom-redis-0`)              |
| **Filters**    | Shard dashboard         | Additional label filters for the selected shard             |

**KubeDB / Redis / Summary** — start here for an instance overview:
- **General Info** — database status, version, max clients, Redis mode, deletion policy, total nodes
- **Resource Requests / Limits** — configured CPU, memory, and storage requests and limits
- **CPU Info / CPU Quota** — per-pod CPU usage over time and quota utilization

<p align="center">
  <img alt="KubeDB Redis Summary Dashboard" src="/docs/images/redis/monitoring/rd-grafana-summary.png" style="padding:10px">
</p>

**KubeDB / Redis / Pod** — drill into a specific pod:
- **General Counters And File Descriptor Stats** — status, role (master/slave), my master, my slaves, connected clients, Go routines
- **Uptime / Memory Usage / Commands Executed / Hits-Misses** — pod uptime, memory usage, command execution rate, cache hit/miss rate
- **Network I/O / Command Calls / Connected Clients** — network throughput, per-command call breakdown, connected client count over time
- **CPU And Memory Usage Stats** — total memory usage, average CPU usage, average memory usage

<p align="center">
  <img alt="KubeDB Redis Pod Dashboard" src="/docs/images/redis/monitoring/rd-grafana-pod.png" style="padding:10px">
</p>

**KubeDB / Redis / Shard** — cluster shard health for Cluster mode:
- **Cluster Shard Slots / Cluster Shard Slots Failed** — hash slot coverage and any failed slots
- **Cluster Nodes / Cluster Masters** — total nodes and master count in the cluster
- **Connected Slaves / My Slaves** — number of connected slaves and their IP, port, and online status
- **Mode** — confirms the instance is running in `cluster` mode

<p align="center">
  <img alt="KubeDB Redis Shard Dashboard" src="/docs/images/redis/monitoring/rd-grafana-shard.png" style="padding:10px">
</p>

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run following commands

```bash
$ kubectl delete -n demo rd/builtin-prom-redis

# If you used Option A (automatic dashboard import)
$ helm uninstall kubedb-grafana-dashboards -n kubeops
$ helm uninstall grafana-operator -n kubeops
$ kubectl delete secret -n monitoring grafana-admin-token
$ kubectl delete appbinding -n monitoring grafana

$ helm uninstall grafana -n monitoring
$ kubectl delete -n monitoring svc/prometheus
$ kubectl delete -n monitoring deployment.apps/prometheus

$ kubectl delete -n monitoring clusterrole.rbac.authorization.k8s.io/prometheus
$ kubectl delete -n monitoring serviceaccount/prometheus
$ kubectl delete -n monitoring clusterrolebinding.rbac.authorization.k8s.io/prometheus

$ kubectl delete ns demo
$ kubectl delete ns monitoring
```

> `kubeops` is left in place since `grafana-operator` (and other KubeDB operator components) may be shared with other tutorials on the same cluster — delete it manually only if you're sure nothing else depends on it.

## Next Steps

- Monitor your Redis server with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
