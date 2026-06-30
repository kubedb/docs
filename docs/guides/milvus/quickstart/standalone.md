---
title: Milvus Standalone Quickstart
menu:
  docs_{{ .version }}:
    identifier: milvus-quickstart-standalone
    name: Standalone
    parent: milvus-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - Milvus Standalone

This tutorial will show you how to use KubeDB to provision a **Standalone** [Milvus](https://milvus.io) database. Milvus is an open-source vector database built to power embedding similarity search and AI applications.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md), and make sure to include the flag `--set global.featureGates.Milvus=true` to ensure the **Milvus CRD** is installed.

- Milvus requires a few **external dependencies** to be available in the cluster:
  - **Object storage** (MinIO / S3-compatible) is **mandatory**. Every example in this guide expects an object-storage configuration secret named `my-release-minio`.
  - **etcd** is used as the metadata store. When `spec.metaStorage` is omitted, KubeDB provisions and manages an internal etcd cluster for you, so an **etcd operator** must be installed and running in the cluster.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/quickstart/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available Milvus Versions

When you install the KubeDB operator, it registers a CRD named `MilvusVersion`. The installation comes with a set of built-in `MilvusVersion` objects. Let's check the available `MilvusVersion`s by:

```bash
$ kubectl get milvusversions
NAME     VERSION   DB_IMAGE                                DEPRECATED   AGE
2.6.11   2.6.11    ghcr.io/appscode-images/milvus:2.6.11                11h
2.6.7    2.6.7     ghcr.io/appscode-images/milvus:2.6.7                 11h
2.6.9    2.6.9     ghcr.io/appscode-images/milvus:2.6.9                 11h
```

## Prepare Object Storage Secret

Milvus stores its segments/logs in object storage, so an object-storage connection secret **must** exist before you create a `Milvus` object. The secret is referenced through `spec.objectStorage.configSecret`. A typical MinIO-backed secret holds three keys — `address`, `accesskey`, and `secretkey`:

```bash
$ kubectl get secret my-release-minio -n demo
NAME               TYPE     DATA   AGE
my-release-minio   Opaque   3      11h
```

> If you do not have a MinIO deployment yet, you can adapt the sample secret shipped with the Milvus operator. The exact contents depend on your storage endpoint and credentials.

## Create a Milvus Database

The KubeDB operator implements a `Milvus` CRD to define the specification of a Milvus database. Below is the `Milvus` object we will create. Notice that in addition to the database itself, the manifest already enables **Prometheus Operator monitoring** and **TLS** — these are covered in detail in the [monitoring](/docs/guides/milvus/monitoring/using-prometheus-operator.md) and [TLS](/docs/guides/milvus/tls/configure/index.md) guides; you can drop those blocks for a bare deployment.

`standalone.yaml`

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-standalone
  namespace: demo
spec:
  version: "2.6.11"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: "my-release-minio"
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    storageClassName: local-path
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
  tls:
    issuerRef:
      name: milvus-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    external:
      mode: mTLS
    internal:
      mode: TLS
```

Here,

- `spec.version` is the name of a `MilvusVersion` CRD object. `2.6.11` points to the Milvus `2.6.11` image.
- `spec.topology.mode: Standalone` deploys Milvus as a single all-in-one workload (one PetSet).
- `spec.objectStorage.configSecret` references the mandatory object-storage secret.
- `spec.storageType` can be `Durable` or `Ephemeral`. With `Durable`, the persistent volume described in `spec.storage` is used.
- `spec.storage` defines the persistent volume claim for the standalone workload.

Create the database:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/standalone.yaml
milvus.kubedb.com/milvus-standalone created
```

## Wait for the Database to be Ready

KubeDB will create the necessary resources to provision the Milvus database. Watch the `Milvus` object until its `STATUS` becomes `Ready`:

> **Note:** Because both `milvuses.kubedb.com` and `milvuses.gitops.kubedb.com` are registered, the short name `milvus` is ambiguous. Use the fully-qualified `milvuses.kubedb.com` (or `kubectl get milvus.kubedb.com`) to query the database.

```bash
$ kubectl get milvuses.kubedb.com -n demo -w
NAME                VERSION   STATUS         AGE
milvus-standalone   2.6.11    Provisioning   24s
milvus-standalone   2.6.11    Ready          39s
```

Standalone Milvus typically becomes ready within a few minutes.

## Verify the Created Resources

Once Milvus is `Ready`, KubeDB has created the following resources. For a standalone deployment there is exactly **one PetSet** named after the database (`<db-name>`):

```bash
$ kubectl get petset -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                AGE
milvus-standalone   88s

$ kubectl get pods -n demo -l app.kubernetes.io/instance=milvus-standalone -o wide
NAME                  READY   STATUS    RESTARTS   AGE   IP           NODE   NOMINATED NODE   READINESS GATES
milvus-standalone-0   1/1     Running   0          88s   10.42.0.86   urmi   <none>           <none>
```

### Services

KubeDB creates a primary client service named after the database (gRPC port `19530`) and, because monitoring is enabled, a `-stats` service exposing the metrics port `9091`:

```bash
$ kubectl get svc -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
milvus-standalone         ClusterIP   10.43.144.154   <none>        19530/TCP   91s
milvus-standalone-stats   ClusterIP   10.43.12.191    <none>        9091/TCP    91s
```

### Storage

The standalone workload mounts a single persistent volume created from `spec.storage`:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-milvus-standalone-0   Bound    pvc-a6333ee2-f0ab-4ec2-8437-599d270b9ed0   1Gi        RWO            local-path     90s
```

> The internal etcd metadata store provisions its own PVCs (`etcd-data-demo-etcd-*`), and MinIO has its own storage. Those are separate from the Milvus data volume.

### Auth Secret

Milvus authentication is enabled by default (`spec.disableSecurity` defaults to `false`). Because `spec.authSecret` was not provided, KubeDB auto-generated a basic-auth secret named `<db-name>-auth` with a `root` user and a random password:

```bash
$ kubectl get secret -n demo | grep milvus-standalone
milvus-standalone-42559a        Opaque                     2      92s
milvus-standalone-auth          kubernetes.io/basic-auth   2      92s
milvus-standalone-client-cert   kubernetes.io/tls          4      91s
milvus-standalone-server-cert   kubernetes.io/tls          3      91s

$ kubectl get secret milvus-standalone-auth -n demo -o jsonpath='{.data.username}' | base64 -d
root
```

The other secrets are the rendered configuration secret (`milvus-standalone-42559a`, holding `milvus.yaml` and `glog.conf`) and the TLS certificate secrets (`-server-cert`, `-client-cert`).

### AppBinding

KubeDB also creates an `AppBinding` — a connection descriptor pointing at the primary service, the auth secret, and the connection scheme (note `scheme: https`, because TLS is enabled):

```bash
$ kubectl get appbinding milvus-standalone -n demo -o yaml
...
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Milvus
    name: milvus-standalone
    namespace: demo
  clientConfig:
    service:
      name: milvus-standalone
      path: /
      port: 19530
      scheme: https
  secret:
    kind: Secret
    name: milvus-standalone-auth
  type: kubedb.com/milvus
  version: 2.6.11
```

## Rendered Configuration

KubeDB renders the effective `milvus.yaml` into the configuration secret. Notice that authentication and internal TLS are wired up automatically:

```bash
$ kubectl get secret milvus-standalone-42559a -n demo -o jsonpath='{.data.milvus\.yaml}' | base64 -d
common:
    msgChannelType: rocksmq
    security:
        authorizationEnabled: true
        defaultRootPassword: <redacted>
        internaltlsEnabled: "true"
        rootUsername: root
        tlsMode: 2
    storageType: remote
...
etcd:
    endpoints:
        - http://demo-etcd-0.demo-etcd.demo.svc.cluster.local:2379
        - http://demo-etcd-1.demo-etcd.demo.svc.cluster.local:2379
        - http://demo-etcd-2.demo-etcd.demo.svc.cluster.local:2379
    rootPath: by-dev
internaltls:
    caPemPath: /milvus/tls/ca.pem
    serverKeyPath: /milvus/tls/server.key
    serverPemPath: /milvus/tls/server.pem
    sni: milvus-standalone
localStorage:
    path: /var/lib/milvus/data/
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo milvus.kubedb.com milvus-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com milvus-standalone -n demo
$ kubectl delete ns demo
```

## Next Steps

- Deploy a [distributed Milvus cluster](/docs/guides/milvus/quickstart/distributed.md).
- Monitor your Milvus database with KubeDB using [Prometheus Operator](/docs/guides/milvus/monitoring/using-prometheus-operator.md).
- Secure your Milvus database with [TLS/SSL](/docs/guides/milvus/tls/configure/index.md).
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
