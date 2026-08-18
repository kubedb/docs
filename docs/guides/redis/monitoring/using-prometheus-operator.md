---
title: Monitoring Redis using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: rd-using-prometheus-operator-monitoring
    name: Prometheus Operator
    parent: rd-monitoring-redis
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring Redis Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor Redis server deployed with KubeDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/redis/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md).

- If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

> Note: YAML files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by a `Prometheus` crd. We are going to provide these labels in `spec.monitor.prometheus.labels` field of Redis crd so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME         AGE
monitoring   prometheus   18m
```

> If you don't have any Prometheus server running in your cluster, deploy one following the guide specified in **Before You Begin** section.

Now, let's view the YAML of the available Prometheus server `prometheus` in `monitoring` namespace.

```yaml
$ kubectl get prometheus -n monitoring prometheus -o yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"monitoring.coreos.com/v1","kind":"Prometheus","metadata":{"annotations":{},"labels":{"prometheus":"prometheus"},"name":"prometheus","namespace":"monitoring"},"spec":{"replicas":1,"resources":{"requests":{"memory":"400Mi"}},"serviceAccountName":"prometheus","serviceMonitorSelector":{"matchLabels":{"release":"prometheus"}}}}
  creationTimestamp: 2019-01-03T13:41:51Z
  generation: 1
  labels:
    prometheus: prometheus
  name: prometheus
  namespace: monitoring
  resourceVersion: "44402"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/prometheuses/prometheus
  uid: 5324ad98-0f5d-11e9-b230-080027f306f3
spec:
  replicas: 1
  resources:
    requests:
      memory: 400Mi
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      release: prometheus
```

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` crd. So, we are going to use this label in `spec.monitor.prometheus.labels` field of Redis crd.

## Deploy Redis with Monitoring Enabled

At first, let's deploy an Redis server with monitoring enabled. Below is the Redis object that we are going to create.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: coreos-prom-redis
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
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Here,

- `monitor.agent:  prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.
- `monitor.prometheus.namespace: monitoring` specifies that KubeDB should create `ServiceMonitor` in `monitoring` namespace.

- `monitor.prometheus.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.

- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the Redis object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/monitoring/coreos-prom-redis.yaml
redis.kubedb.com/coreos-prom-redis created
```

Now, wait for the database to go into `Running` state.

```bash
$ kubectl get rd -n demo coreos-prom-redis
NAME                VERSION   STATUS    AGE
coreos-prom-redis   4.0-v1    Running   15s
```

KubeDB will create a separate stats service with name `{Redis crd name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=coreos-prom-redis"
NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
coreos-prom-redis         ClusterIP   10.110.70.53   <none>        6379/TCP    35s
coreos-prom-redis-stats   ClusterIP   10.99.161.76   <none>        56790/TCP   31s
```

Here, `coreos-prom-redis-stats` service has been created for monitoring purpose.

Let's describe this stats service.

```yaml
$ kubectl describe svc -n demo coreos-prom-redis-stats
Name:              coreos-prom-redis-stats
Namespace:         demo
Labels:            app.kubernetes.io/name=redises.kubedb.com
                   app.kubernetes.io/instance=coreos-prom-redis
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/name=redises.kubedb.com,app.kubernetes.io/instance=coreos-prom-redis
Type:              ClusterIP
IP:                10.99.161.76
Port:              prom-http  56790/TCP
TargetPort:        prom-http/TCP
Endpoints:         172.17.0.7:56790
Session Affinity:  None
Events:            <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use these information to target its endpoints.

KubeDB will also create a `ServiceMonitor` crd in `monitoring` namespace that select the endpoints of `coreos-prom-redis-stats` service. Verify that the `ServiceMonitor` crd has been created.

```bash
$ kubectl get servicemonitor -n demo
NAME                            AGE
kubedb-demo-coreos-prom-redis   1m
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of Redis crd.

```bash
$ kubectl get servicemonitor -n demo kubedb-demo-coreos-prom-redis -o yaml
```

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: 2019-01-03T15:55:23Z
  generation: 1
  labels:
    release: prometheus
    monitoring.appscode.com/service: coreos-prom-redis-stats.demo
  name: kubedb-demo-coreos-prom-redis
  namespace: monitoring
  resourceVersion: "54802"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/servicemonitors/kubedb-demo-coreos-prom-redis
  uid: fafceb49-0f6f-11e9-b230-080027f306f3
spec:
  endpoints:
  - honorLabels: true
    interval: 10s
    path: /metrics
    port: prom-http
  namespaceSelector:
    matchNames:
    - demo
  selector:
    matchLabels:
      app.kubernetes.io/name: redises.kubedb.com
      app.kubernetes.io/instance: coreos-prom-redis
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in Redis crd.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `coreos-prom-redis-stats` service. It also, target the `prom-http` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                      READY   STATUS    RESTARTS   AGE
prometheus-prometheus-0   3/3     Running   1          63m
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `prom-http` endpoint of `coreos-prom-redis-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/redis/monitoring/redis-coreos-prom-target.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels marked by red rectangle. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create a dashboard with the collected metrics, as shown below.

## Install Panopticon (required for full dashboard data)

The `redis_exporter` sidecar only reports Redis-native metrics (connections, memory, commands, etc.). The KubeDB Redis Grafana dashboards ([Visualize Metrics with Grafana](#visualize-metrics-with-grafana) below) also chart KubeDB's own view of the resource — database status, phase, version, deletion policy — and those come from a **separate** metric source: **Panopticon**, the Appscode operator that exports `kubedb_com_redis_*` metrics for every KubeDB object. Skip this step and the dashboards will still load, but the General Info / phase panels will show "No data".

Install Panopticon with the `prometheus.io/operator` monitoring agent — the same agent this tutorial already uses for Redis itself, so Panopticon's metrics get picked up the same way, via a `ServiceMonitor` the chart creates automatically. No manual Prometheus config editing needed, unlike the [built-in Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md) tutorial.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm upgrade --install panopticon appscode/panopticon \
  --version v2026.4.30 \
  --namespace kubeops --create-namespace \
  --set monitoring.enabled=true \
  --set monitoring.agent=prometheus.io/operator \
  --set monitoring.serviceMonitor.labels.release=prometheus \
  --set-file license=/path/to/kubedb-license.txt \
  --wait --timeout 5m0s
```

> The `monitoring.serviceMonitor.labels.release` value must match whatever label your `Prometheus` CR's `serviceMonitorSelector` expects — `release: prometheus` here, matching the label already used for Redis's own `ServiceMonitor` earlier in this tutorial. If your Prometheus server uses a different `serviceMonitorSelector` label, use that instead.

Verify Panopticon is running:

```bash
$ kubectl get pods -n kubeops
NAME                          READY   STATUS    RESTARTS   AGE
panopticon-xxxx               1/1     Running   0          1m
```

## Visualize Metrics with Grafana

### Deploy Grafana

The Prometheus operator setup used in this tutorial doesn't bundle Grafana, so install it separately with Helm.

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
3. Set the URL to the `Prometheus` operator's service (`prometheus-operated`, created automatically by the operator for the `prometheus` CR used in this tutorial):

   ```
   http://prometheus-operated.monitoring.svc:9090
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

`SYNCED: Current` confirms `grafana-operator` successfully pushed each dashboard into Grafana. Open Grafana — the three dashboards are already there under `Dashboards`, fully wired to your Prometheus data source, ready to explore in [Explore the Dashboard](#explore-the-dashboard) below.

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
| **app**        | Summary dashboard       | Name of your Redis instance (e.g., `coreos-prom-redis`)    |
| **redis**      | Pod, Shard dashboards   | Name of your Redis instance (e.g., `coreos-prom-redis`)    |
| **Pod Name**   | Pod, Shard dashboards   | A specific pod (e.g., `coreos-prom-redis-0`)               |
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
# cleanup database
kubectl delete -n demo rd/coreos-prom-redis

# cleanup panopticon
helm uninstall panopticon -n kubeops

# cleanup grafana visualization (if you used Option A)
helm uninstall kubedb-grafana-dashboards -n kubeops
helm uninstall grafana-operator -n kubeops
kubectl delete secret -n monitoring grafana-admin-token
kubectl delete appbinding -n monitoring grafana

# cleanup grafana
helm uninstall grafana -n monitoring

# cleanup prometheus resources
kubectl delete -n monitoring prometheus prometheus
kubectl delete -n monitoring clusterrolebinding prometheus
kubectl delete -n monitoring clusterrole prometheus
kubectl delete -n monitoring serviceaccount prometheus
kubectl delete -n monitoring service prometheus-operated

# cleanup prometheus operator resources
kubectl delete -n monitoring deployment prometheus-operator
kubectl delete -n dmeo serviceaccount prometheus-operator
kubectl delete clusterrolebinding prometheus-operator
kubectl delete clusterrole prometheus-operator

# delete namespace
kubectl delete ns monitoring
kubectl delete ns demo
```

## Next Steps

- Monitor your Redis server with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Detail concepts of [RedisVersion object](/docs/guides/redis/concepts/catalog.md).
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
