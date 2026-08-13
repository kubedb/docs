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

- Object storage, exposed through a secret named `milvus-storage-config`.
- etcd for metadata.

This guide sets up both dependencies in the `demo` namespace and clarifies when you need only the **etcd operator** and when you instead want to point Milvus at an **external etcd cluster**.

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

Milvus stores its segments and logs in object storage. We recommend using the MinIO Operator to deploy a MinIO Tenant.

### 1. Install the MinIO Operator

```bash
$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1
```

### 2. Deploy a MinIO Tenant

```bash
$ helm upgrade --install --namespace "demo" milvus-minio minio/tenant \
  --set tenant.name=milvus-minio \
  --set tenant.pools[0].servers=1 \
  --set tenant.pools[0].volumesPerServer=1 \
  --set tenant.pools[0].size=1Gi \
  --set tenant.certificate.requestAutoCert=false \
  --set tenant.buckets[0].name="mlv-release" \
  --set tenant.pools[0].name="default"
```

Wait for the Tenant to become initialized:

```bash
$ kubectl get tenant -n demo milvus-minio -w
NAME           STATE         HEALTH   AGE
milvus-minio   Initialized   green    1m
```

### 3. Create the Storage Config Secret

Create a secret with your MinIO connection details. The keys must use the milvus.yaml config key names directly:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: milvus-storage-config
  namespace: demo
type: Opaque
stringData:
  address: "milvus-minio-hl.demo.svc.cluster.local"
  accessKeyID: "minio"
  secretAccessKey: "minio123"
  bucketName: "mlv-release"
  rootPath: "files"
  port: "9000"
```

Apply the secret:

```bash
$ kubectl apply -f milvus-storage-config.yaml
secret/milvus-storage-config created
```

Verify:

```bash
$ kubectl get secret milvus-storage-config -n demo
NAME                     TYPE     DATA   AGE
milvus-storage-config    Opaque   6      1m
```

### If You Already Have S3 or MinIO

You do not need to use the MinIO Operator. You can instead create only the secret and point Milvus at your existing object store:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: milvus-storage-config
  namespace: demo
type: Opaque
stringData:
  address: "existing-minio.demo.svc.cluster.local"
  accessKeyID: "minioadmin"
  secretAccessKey: "minioadmin"
  bucketName: "milvus"
  rootPath: "files-1"
  port: "9000"
```

If you use a different secret name, update `spec.objectStorage.configSecret.name` in the `Milvus` manifest accordingly.

## Install the etcd Operator

Milvus always uses etcd as its metadata store. In KubeDB, there are two supported patterns:

1. **KubeDB-managed etcd**: omit `spec.metaStorage`. KubeDB creates the internal etcd cluster for you.
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

## Default Path: KubeDB-Managed etcd

For the [standalone](/docs/guides/milvus/quickstart/standalone.md) and [distributed](/docs/guides/milvus/quickstart/distributed.md) quickstarts, this is the default and recommended path.

If you **omit** `spec.metaStorage` from the `Milvus` manifest:

- KubeDB creates an internal etcd cluster
- KubeDB wires Milvus to that internal etcd automatically
- You do **not** need to apply any external etcd YAML yourself

So for the default quickstarts, **having the etcd operator running is enough**.

## Optional Path: Use External etcd

If you already manage etcd yourself, do not let KubeDB create an internal metadata cluster. Instead, set `spec.metaStorage.externallyManaged: true` and provide your own etcd endpoints:

```yaml
metaStorage:
  externallyManaged: true
  endpoints:
    - http://etcd-0.example.svc.cluster.local:2379
    - http://etcd-1.example.svc.cluster.local:2379
    - http://etcd-2.example.svc.cluster.local:2379
```

Requirements for external etcd:

- The endpoints must be reachable from the Milvus pods.
- The etcd cluster must already be healthy before you create the `Milvus` object.
- The etcd operator is still required in the cluster for the default KubeDB-managed path, but this external-endpoint configuration does not require any sample external etcd YAML from these docs.

Only choose this path if you intentionally want Milvus to use an external etcd cluster. The default quickstarts do not require any external etcd manifest.

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

1. Delete your external etcd resources using whatever workflow manages them:

   ```bash
   # Example only:
   # kubectl delete <your-etcd-resources>
   ```

2. Delete leftover etcd PVCs if your etcd management workflow leaves them behind:

   ```bash
   # Example only:
   # kubectl delete pvc -n demo <your-etcd-pvc-names>
   ```

   For KubeDB-managed etcd, the PVC names follow the pattern `etcd-data-<milvus-name>-etcd-<ordinal>`.

3. Delete the MinIO Tenant:

   ```bash
   $ helm uninstall milvus-minio -n demo
   ```

4. Delete the storage config secret:

   ```bash
   $ kubectl delete secret -n demo milvus-storage-config
   ```

## Next Steps

- [Deploy standalone Milvus](/docs/guides/milvus/quickstart/standalone.md)
- [Deploy distributed Milvus](/docs/guides/milvus/quickstart/distributed.md)
