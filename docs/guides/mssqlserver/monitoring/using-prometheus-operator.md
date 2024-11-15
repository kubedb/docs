---
title: Monitor SQL Server using Prometheus Operator
menu:
  docs_{{ .version }}:
    identifier: ms-monitoring-prometheus-operator
    name: Prometheus Operator
    parent: ms-monitoring
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Monitoring MSSQLServer Using Prometheus operator

[Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use Prometheus operator to monitor MSSQLServer  deployed with KubeDB.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation. 

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- To learn how Prometheus monitoring works with KubeDB in general, please visit [here](/docs/guides/mssqlserver/monitoring/overview.md).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources. We are going to deploy database in `demo` namespace.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```

- We need a [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator) instance running. If you don't already have a running instance, you can deploy one using this helm chart [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack).


> Note: YAML files used in this tutorial are stored in [docs/examples/mssqlserver/monitoring](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/monitoring) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find out required labels for ServiceMonitor

We need to know the labels used to select `ServiceMonitor` by `Prometheus` Operator. We are going to provide these labels in `spec.monitor.prometheus.labels` field of MSSQLServer CR so that KubeDB creates `ServiceMonitor` object accordingly.

At first, let's find out the available Prometheus server in our cluster.

```bash
$ kubectl get prometheus --all-namespaces
NAMESPACE    NAME                                    VERSION   DESIRED   READY   RECONCILED   AVAILABLE   AGE
monitoring   prometheus-kube-prometheus-prometheus   v2.54.1   1         1       True         True        16d
```

> If you don't have any Prometheus server running in your cluster, deploy one following the guide specified in **Before You Begin** section.

Now, let's view the YAML of the available Prometheus server `prometheus-kube-prometheus-prometheus` in `monitoring` namespace.

```bash
$ kubectl get prometheus -n monitoring prometheus-kube-prometheus-prometheus -oyaml
```
```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  annotations:
    meta.helm.sh/release-name: prometheus
    meta.helm.sh/release-namespace: monitoring
  creationTimestamp: "2024-10-14T10:14:36Z"
  generation: 1
  labels:
    app: kube-prometheus-stack-prometheus
    app.kubernetes.io/instance: prometheus
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/part-of: kube-prometheus-stack
    app.kubernetes.io/version: 65.2.0
    chart: kube-prometheus-stack-65.2.0
    heritage: Helm
    release: prometheus
  name: prometheus-kube-prometheus-prometheus
  namespace: monitoring
  resourceVersion: "1004097"
  uid: b7879d3e-e4bb-4425-8d78-f917561d95f7
spec:
  alerting:
    alertmanagers:
      - apiVersion: v2
        name: prometheus-kube-prometheus-alertmanager
        namespace: monitoring
        pathPrefix: /
        port: http-web
  automountServiceAccountToken: true
  enableAdminAPI: false
  evaluationInterval: 30s
  externalUrl: http://prometheus-kube-prometheus-prometheus.monitoring:9090
  hostNetwork: false
  image: quay.io/prometheus/prometheus:v2.54.1
  listenLocal: false
  logFormat: logfmt
  logLevel: info
  paused: false
  podMonitorNamespaceSelector: {}
  podMonitorSelector:
    matchLabels:
      release: prometheus
  portName: http-web
  probeNamespaceSelector: {}
  probeSelector:
    matchLabels:
      release: prometheus
  replicas: 1
  retention: 10d
  routePrefix: /
  ruleNamespaceSelector: {}
  ruleSelector:
    matchLabels:
      release: prometheus
  scrapeConfigNamespaceSelector: {}
  scrapeConfigSelector:
    matchLabels:
      release: prometheus
  scrapeInterval: 30s
  securityContext:
    fsGroup: 2000
    runAsGroup: 2000
    runAsNonRoot: true
    runAsUser: 1000
    seccompProfile:
      type: RuntimeDefault
  serviceAccountName: prometheus-kube-prometheus-prometheus
  serviceMonitorNamespaceSelector: {}
  serviceMonitorSelector:
    matchLabels:
      release: prometheus
  shards: 1
  tsdb:
    outOfOrderTimeWindow: 0s
  version: v2.54.1
  walCompression: true
status:
  availableReplicas: 1
  conditions:
    - lastTransitionTime: "2024-10-31T07:38:36Z"
      message: ""
      observedGeneration: 1
      reason: ""
      status: "True"
      type: Available
    - lastTransitionTime: "2024-10-31T07:38:36Z"
      message: ""
      observedGeneration: 1
      reason: ""
      status: "True"
      type: Reconciled
  paused: false
  replicas: 1
  selector: app.kubernetes.io/instance=prometheus-kube-prometheus-prometheus,app.kubernetes.io/managed-by=prometheus-operator,app.kubernetes.io/name=prometheus,operator.prometheus.io/name=prometheus-kube-prometheus-prometheus,prometheus=prometheus-kube-prometheus-prometheus
  shardStatuses:
    - availableReplicas: 1
      replicas: 1
      shardID: "0"
      unavailableReplicas: 0
      updatedReplicas: 1
  shards: 1
  unavailableReplicas: 0
  updatedReplicas: 1
```

Notice the `spec.serviceMonitorSelector` section. Here, `release: prometheus` label is used to select `ServiceMonitor` CR. So, we are going to use this label in `spec.monitor.prometheus.labels` field of MSSQLServer CR.

## Deploy MSSQLServer with Monitoring Enabled

First, an issuer needs to be created, even if TLS is not enabled for SQL Server. The issuer will be used to configure the TLS-enabled Wal-G proxy server, which is required for the SQL Server backup and restore operations.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=MSSQLServer/O=kubedb"
```
- Create a secret using the certificate files we have just generated,
```bash
$ kubectl create secret tls mssqlserver-ca --cert=ca.crt  --key=ca.key --namespace=demo 
secret/mssqlserver-ca created
```
Now, we are going to create an `Issuer` using the `mssqlserver-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` CR that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: mssqlserver-ca-issuer
 namespace: demo
spec:
 ca:
   secretName: mssqlserver-ca
```

Let’s create the `Issuer` CR we have shown above,
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/ag-cluster/mssqlserver-ca-issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer created
```

Now, let's deploy an MSSQLServer with monitoring enabled. Below is the MSSQLServer object that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-monitoring
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation # Change it 
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9399
        resources:
          limits:
            memory: 512Mi
          requests:
            cpu: 200m
            memory: 256Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
              - ALL
          runAsGroup: 10001
          runAsNonRoot: true
          runAsUser: 10001
          seccompProfile:
            type: RuntimeDefault
      serviceMonitor:
        interval: 10s
        labels:
          release: prometheus
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Here,

- `monitor.agent:  prometheus.io/operator` indicates that we are going to monitor this server using Prometheus operator.

- `monitor.prometheus.serviceMonitor.labels` specifies that KubeDB should create `ServiceMonitor` with these labels.

- `monitor.prometheus.interval` indicates that the Prometheus server should scrape metrics from this database with 10 seconds interval.

Let's create the MSSQLServer object that we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/monitoring/mssql-monitoring.yaml
mssqlserverql.kubedb.com/mssql-monitoring created
```

Now, wait for the database to go into `Ready` state.

```bash
$ kubectl get ms -n demo mssql-monitoring
NAME               VERSION     STATUS   AGE
mssql-monitoring   2022-cu12   Ready    108m
```

KubeDB will create a separate stats service with name `{mssqlserver cr name}-stats` for monitoring purpose.

```bash
$ kubectl get svc -n demo --selector="app.kubernetes.io/instance=mssql-monitoring"
NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
mssql-monitoring         ClusterIP   10.96.225.130   <none>        1433/TCP   108m
mssql-monitoring-pods    ClusterIP   None            <none>        1433/TCP   108m
mssql-monitoring-stats   ClusterIP   10.96.147.93    <none>        9399/TCP   108m
```

Here, `mssql-monitoring-stats` service has been created for monitoring purpose.

Let's describe this stats service.


```bash
$ kubectl describe svc -n demo mssql-monitoring-stats
```
```yaml
Name:              mssql-monitoring-stats
Namespace:         demo
Labels:            app.kubernetes.io/component=database
  app.kubernetes.io/instance=mssql-monitoring
  app.kubernetes.io/managed-by=kubedb.com
  app.kubernetes.io/name=mssqlservers.kubedb.com
  kubedb.com/role=stats
Annotations:       monitoring.appscode.com/agent: prometheus.io/operator
Selector:          app.kubernetes.io/instance=mssql-monitoring,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mssqlservers.kubedb.com
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                10.96.147.93
IPs:               10.96.147.93
Port:              metrics  9399/TCP
TargetPort:        metrics/TCP
Endpoints:         10.244.0.47:9399
Session Affinity:  None
Events:            <none>
```

Notice the `Labels` and `Port` fields. `ServiceMonitor` will use these information to target its endpoints.

KubeDB will also create a `ServiceMonitor` CR in `demo` namespace that select the endpoints of `mssql-monitoring-stats` service. Verify that the `ServiceMonitor` CR has been created.

```bash
$ kubectl get servicemonitor -n demo
NAME                     AGE
mssql-monitoring-stats   110m
```

Let's verify that the `ServiceMonitor` has the label that we had specified in `spec.monitor` section of MSSQLServer CR.

```bash
$ kubectl get servicemonitor -n demo mssql-monitoring-stats -o yaml
```

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: "2024-10-31T07:38:36Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: mssql-monitoring
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mssqlservers.kubedb.com
    release: prometheus
  name: mssql-monitoring-stats
  namespace: demo
  ownerReferences:
    - apiVersion: v1
      blockOwnerDeletion: true
      controller: true
      kind: Service
      name: mssql-monitoring-stats
      uid: 99193679-301b-41fd-aae5-a732b3070d19
  resourceVersion: "1004080"
  uid: 87635ad4-dfb2-4544-89af-e48b40783205
spec:
  endpoints:
    - honorLabels: true
      interval: 10s
      path: /metrics
      port: metrics
  namespaceSelector:
    matchNames:
      - demo
  selector:
    matchLabels:
      app.kubernetes.io/component: database
      app.kubernetes.io/instance: mssql-monitoring
      app.kubernetes.io/managed-by: kubedb.com
      app.kubernetes.io/name: mssqlservers.kubedb.com
      kubedb.com/role: stats
```

Notice that the `ServiceMonitor` has label `release: prometheus` that we had specified in MSSQLServer CR.

Also notice that the `ServiceMonitor` has selector which match the labels we have seen in the `mssql-monitoring-stats` service. It also, target the `metrics` port that we have seen in the stats service.

## Verify Monitoring Metrics

At first, let's find out the respective Prometheus pod for `prometheus-kube-prometheus-prometheus` Prometheus server.

```bash
$ kubectl get pod -n monitoring -l=app.kubernetes.io/name=prometheus
NAME                                                 READY   STATUS    RESTARTS         AGE
prometheus-prometheus-kube-prometheus-prometheus-0   2/2     Running   1                16d
```

Prometheus server is listening to port `9090` of `prometheus-prometheus-kube-prometheus-prometheus-0` pod. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

Run following command on a separate terminal to forward the port 9090 of `prometheus-prometheus-0` pod,

```bash
$ kubectl port-forward -n monitoring prometheus-prometheus-kube-prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `metrics` endpoint of `mssql-monitoring-stats` service as one of the targets.

<p align="center">
  <img alt="Prometheus Target" src="/docs/images/mssqlserver/monitoring/mssql-monitoring-targets.png" style="padding:10px">
</p>

Check the `endpoint` and `service` labels. It verifies that the target is our expected database. Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboards with collected metrics.

# Grafana Dashboards

There are three dashboards to monitor Microsoft SQL Server Databases managed by KubeDB.

- KubeDB / MSSQLServer / Summary: Shows overall summary of Microsoft SQL Server instance.
- KubeDB / MSSQLServer / Pod: Shows individual pod-level information.
- KubeDB / MSSQLServer / Database: Shows Microsoft SQL Server internal metrics for an instance.
> Note: These dashboards are developed in Grafana version 7.5.5


To use KubeDB `Grafana Dashboards` to monitor Microsoft SQL Server Databases managed by `KubeDB`, Check out [mssqlserver-dashboards](https://github.com/ops-center/grafana-dashboards/tree/master/mssqlserver)

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run following commands

```bash
kubectl delete -n demo ms/mssql-monitoring
kubectl delete ns demo

helm uninstall prometheus -n monitoring
kubectl delete ns monitoring
```

## Next Steps
- Learn about [backup and restore](/docs/guides/mssqlserver/backup/overview/index.md) SQL Server using KubeStash.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
