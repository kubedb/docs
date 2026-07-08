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

This tutorial shows how to use KubeDB to provision a **standalone** [Milvus](https://milvus.io) database.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured to talk to it.
- Install KubeDB with `--set global.featureGates.Milvus=true`.
- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). That guide installs MinIO, creates the `my-release-minio` secret, and installs the etcd operator required by Milvus.
- This quickstart intentionally uses the smallest working manifest. It does **not** require Prometheus Operator or cert-manager.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/quickstart/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available Milvus Versions

When you install the KubeDB operator, it registers a CRD named `MilvusVersion`. The installation comes with a set of built-in `MilvusVersion` objects:

```bash
$ kubectl get milvusversions
NAME     VERSION   DB_IMAGE                                DEPRECATED   AGE
2.6.11   2.6.11    ghcr.io/appscode-images/milvus:2.6.11                11h
2.6.7    2.6.7     ghcr.io/appscode-images/milvus:2.6.7                 11h
2.6.9    2.6.9     ghcr.io/appscode-images/milvus:2.6.9                 11h
```

## Create a Standalone Milvus

The following manifest is the smallest durable standalone deployment:

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
```

Here,

- `spec.version` is the name of a `MilvusVersion` object.
- `spec.topology.mode: Standalone` deploys Milvus as a single all-in-one workload.
- `spec.objectStorage.configSecret` points to the required MinIO/object-storage secret.
- `spec.metaStorage` is omitted, so KubeDB creates and manages an internal etcd cluster through the installed etcd operator.
- `spec.storageType: Durable` tells KubeDB to provision persistent storage for the standalone workload.

Create the database:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/standalone.yaml
milvus.kubedb.com/milvus-standalone created
```

## Wait for the Database to be Ready

Watch the `Milvus` object until its `STATUS` becomes `Ready`:

> Because both `milvuses.kubedb.com` and `milvuses.gitops.kubedb.com` are registered, the short name `milvus` is ambiguous. Use `milvuses.kubedb.com`.

```bash
$ kubectl get milvuses.kubedb.com -n demo -w
NAME                VERSION   STATUS         AGE
milvus-standalone   2.6.11    Provisioning   24s
milvus-standalone   2.6.11    Ready          39s
```

Standalone Milvus typically becomes ready within a few minutes.

## Verify the Created Resources

### PetSet and Pod

```bash
$ kubectl get petset -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                AGE
milvus-standalone   88s

$ kubectl get pods -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                  READY   STATUS    RESTARTS   AGE
milvus-standalone-0   1/1     Running   0          88s
```

### Service

KubeDB creates a primary client service named after the database on gRPC port `19530`:

```bash
$ kubectl get svc -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
milvus-standalone   ClusterIP   10.43.144.154   <none>        19530/TCP   91s
```

> If you later enable [Prometheus Operator monitoring](/docs/guides/milvus/monitoring/using-prometheus-operator.md), KubeDB also creates a `milvus-standalone-stats` service on port `9091`.

### Storage

The standalone workload mounts a single persistent volume created from `spec.storage`:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-milvus-standalone-0   Bound    pvc-a6333ee2-f0ab-4ec2-8437-599d270b9ed0   1Gi        RWO            local-path     90s
```

The internal etcd metadata store provisions its own PVCs (`etcd-data-<milvus-name>-etcd-*`), and MinIO has separate PVCs as well.

### Auth Secret

Milvus authentication is enabled by default. Because `spec.authSecret` was not provided, KubeDB auto-generates a basic-auth secret named `<db-name>-auth` with a `root` user and a random password:

```bash
$ kubectl get secret -n demo | grep milvus-standalone
milvus-standalone-42559a   Opaque                     2      92s
milvus-standalone-auth     kubernetes.io/basic-auth   2      92s

$ kubectl get secret milvus-standalone-auth -n demo -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secret milvus-standalone-auth -n demo -o jsonpath='{.data.password}' | base64 -d
<generated-password>
```

The other secret (`milvus-standalone-42559a`) is the rendered configuration secret holding `milvus.yaml` and `glog.conf`.

### AppBinding

KubeDB also creates an `AppBinding` pointing at the primary service and the auth secret. Because this quickstart does not enable TLS, the connection scheme is `http`:

```bash
$ kubectl get appbinding milvus-standalone -n demo -o yaml
...
spec:
  clientConfig:
    service:
      name: milvus-standalone
      path: /
      port: 19530
      scheme: http
  secret:
    kind: Secret
    name: milvus-standalone-auth
  type: kubedb.com/milvus
  version: 2.6.11
```

## Use an External etcd Cluster Instead

The default quickstart omits `spec.metaStorage`, so KubeDB manages etcd for you. If you want Milvus to use an external `EtcdCluster` instead:

1. Install the etcd operator.
2. Create an external `EtcdCluster` by following [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#optional-path-create-an-external-etcdcluster).
3. Add `spec.metaStorage.externallyManaged: true` and the external endpoints to your `Milvus` manifest.

## Cleaning up

```bash
$ kubectl patch -n demo milvus.kubedb.com milvus-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com milvus-standalone -n demo
```

If you want to remove the dependencies too, either delete the whole `demo` namespace or follow the cleanup steps in [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#cleanup).

## Next Steps

- [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md) for another cluster.
- [Deploy a Distributed Milvus](/docs/guides/milvus/quickstart/distributed.md).
- [Enable Prometheus Operator monitoring](/docs/guides/milvus/monitoring/using-prometheus-operator.md).
- [Enable TLS](/docs/guides/milvus/tls/guide.md).
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
