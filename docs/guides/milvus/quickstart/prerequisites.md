---
title: Prepare Dependencies
menu:
  docs_{{ .version }}:
    identifier: milvus-quickstart-prerequisites
    name: Prepare Dependencies
    parent: milvus-quickstart
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Prepare Milvus Dependencies

Milvus will not start from a bare `Milvus` manifest alone. Every Milvus deployment in KubeDB needs:

- Object storage, exposed through a secret named `my-release-minio`.
- etcd for metadata.

This guide sets up both dependencies in the `demo` namespace and clarifies when you need only the **etcd operator** and when you also need an **external `EtcdCluster`**.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured to talk to it.
- Install KubeDB with the Milvus feature gate enabled:

  ```bash
  helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
    --namespace kubedb --create-namespace \
    --set global.featureGates.Milvus=true
  ```

## Create the Demo Namespace

All Milvus examples in this guide use the `demo` namespace:

```bash
$ kubectl create namespace demo
namespace/demo created
```

## Install MinIO

Milvus stores its segments and logs in object storage. For this quickstart, we provide a **sample MinIO deployment** that creates:

- A MinIO `StatefulSet`
- The `my-release-minio` service
- The `my-release-minio` secret expected by all Milvus examples

This is only an example. You may change the namespace, image tag, replica count, PVC size, credentials, or even use an existing S3-compatible object store instead. The only requirement from the Milvus side is that `spec.objectStorage.configSecret` points to a secret containing:

- `address`
- `accesskey`
- `secretkey`

If you use the sample as-is, apply it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/minio.yaml
serviceaccount/my-release-minio created
secret/my-release-minio created
configmap/my-release-minio created
service/my-release-minio created
service/my-release-minio-svc created
statefulset.apps/my-release-minio created
```

Verify it:

```bash
$ kubectl get secret my-release-minio -n demo
NAME               TYPE     DATA   AGE
my-release-minio   Opaque   3      1m

$ kubectl get statefulset my-release-minio -n demo
NAME               READY   AGE
my-release-minio   4/4     1m
```

The full example manifest lives in:

- [docs/guides/milvus/quickstart/yamls/minio.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/minio.yaml)

### If You Already Have S3 or MinIO

You do not need to use the sample MinIO deployment. You can instead create only the secret and point Milvus at your existing object store:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-release-minio
  namespace: demo
type: Opaque
stringData:
  address: existing-minio.demo.svc.cluster.local:9000
  accesskey: minioadmin
  secretkey: minioadmin
```

If you use a different secret name, update `spec.objectStorage.configSecret.name` in the `Milvus` manifest accordingly.

## Install the etcd Operator

Milvus always uses etcd as its metadata store. In KubeDB, there are two supported patterns:

1. **KubeDB-managed etcd**: omit `spec.metaStorage`. KubeDB creates the `EtcdCluster` for you.
2. **Externally managed etcd**: set `spec.metaStorage.externallyManaged: true` and provide endpoints yourself.

In both cases, the **etcd operator must already be installed** in the cluster.

The simplest installation path is the upstream install manifest:

```bash
$ kubectl apply -f https://raw.githubusercontent.com/etcd-io/etcd-operator/refs/heads/main/dist/install-v0.1.0.yaml
namespace/etcd-operator-system created
customresourcedefinition.apiextensions.k8s.io/etcdclusters.operator.etcd.io created
deployment.apps/etcd-operator-controller-manager created
...
```

Verify it:

```bash
$ kubectl get deployment -n etcd-operator-system
NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
etcd-operator-controller-manager   1/1     1            1           1m

$ kubectl get crd etcdclusters.operator.etcd.io
NAME                           CREATED AT
etcdclusters.operator.etcd.io  2026-07-08T...
```

> If your environment cannot pull the default etcd-operator image, use the source-build workflow from the official repo: build and push your own image, then run `make install` and `make deploy IMG=<your-image>`.

## Default Path: KubeDB-Managed etcd

For the [standalone](/docs/guides/milvus/quickstart/standalone.md) and [distributed](/docs/guides/milvus/quickstart/distributed.md) quickstarts, this is the default and recommended path.

If you **omit** `spec.metaStorage` from the `Milvus` manifest:

- KubeDB creates an internal `EtcdCluster`
- KubeDB wires Milvus to that internal etcd automatically
- You do **not** need to apply any `EtcdCluster` YAML yourself

So for the default quickstarts, **having the etcd operator running is enough**.

## Optional Path: Create an External EtcdCluster

If you want to bring your own etcd, apply the sample `EtcdCluster` manifest:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/etcdcluster.yaml
etcdcluster.operator.etcd.io/etcdcluster-sample created
```

Wait for the cluster:

```bash
$ kubectl get pods,svc -n demo | grep etcdcluster-sample
pod/etcdcluster-sample-0     1/1   Running   0   2m
pod/etcdcluster-sample-1     1/1   Running   0   2m
pod/etcdcluster-sample-2     1/1   Running   0   2m
service/etcdcluster-sample   ClusterIP   None   <none>   <none>   2m
```

Use these endpoints in `spec.metaStorage`:

```yaml
metaStorage:
  externallyManaged: true
  endpoints:
    - http://etcdcluster-sample-0.etcdcluster-sample.demo.svc.cluster.local:2379
    - http://etcdcluster-sample-1.etcdcluster-sample.demo.svc.cluster.local:2379
    - http://etcdcluster-sample-2.etcdcluster-sample.demo.svc.cluster.local:2379
```

Only choose this path if you intentionally want Milvus to use an external etcd cluster. The default quickstarts do not require it.

## Optional Controllers

These are not required for the base quickstarts:

- Install [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator) only if you want to follow the [monitoring guide](/docs/guides/milvus/monitoring/using-prometheus-operator.md).
- Install [cert-manager](https://cert-manager.io/docs/installation/) only if you want to follow the [TLS guide](/docs/guides/milvus/tls/guide.md) or the [TLS reconfiguration guide](/docs/guides/milvus/reconfigure-tls/guide.md).

## Cleanup

If you delete the whole `demo` namespace, Kubernetes removes the Milvus, MinIO, and etcd resources in that namespace together:

```bash
$ kubectl delete namespace demo
namespace "demo" deleted
```

If you want to keep the namespace and clean up dependencies separately:

1. Delete the external etcd cluster if you created one:

   ```bash
   $ kubectl delete etcdcluster -n demo etcdcluster-sample
   etcdcluster.operator.etcd.io "etcdcluster-sample" deleted
   ```

2. Delete leftover etcd PVCs if the etcd operator leaves them behind:

   ```bash
   $ kubectl delete pvc -n demo etcd-data-etcdcluster-sample-0 etcd-data-etcdcluster-sample-1 etcd-data-etcdcluster-sample-2
   ```

   For KubeDB-managed etcd, the PVC names follow the pattern `etcd-data-<milvus-name>-etcd-<ordinal>`.

3. Delete MinIO:

   ```bash
   $ kubectl delete -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/minio.yaml
   ```

4. Delete leftover MinIO PVCs if you are keeping the namespace:

   ```bash
   $ kubectl delete pvc -n demo export-my-release-minio-0 export-my-release-minio-1 export-my-release-minio-2 export-my-release-minio-3
   ```

## Next Steps

- [Deploy standalone Milvus](/docs/guides/milvus/quickstart/standalone.md)
- [Deploy distributed Milvus](/docs/guides/milvus/quickstart/distributed.md)
