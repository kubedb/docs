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

This guide sets up both dependencies in the `demo` namespace and clarifies when KubeDB manages etcd for you and when you instead want to point Milvus at an **external etcd cluster**.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured to talk to it.
- Install KubeDB with the Milvus and Etcd feature gates enabled. Milvus's internally-managed metadata store is a [KubeDB `Etcd`](/docs/guides/etcd/README.md) database, so the `Etcd` feature gate (still **alpha**) must be on alongside `Milvus`'s:

  ```bash
  helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
    --namespace kubedb --create-namespace \
    --set global.featureGates.Milvus=true \
    --set global.featureGates.Etcd=true
  ```

  If you install the operator components separately, the equivalent operator flag is `--feature-gates=Etcd=true`, set on the provisioner, the ops-manager, the autoscaler and the crd-manager alike.

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

## The etcd Operator

Milvus always uses etcd as its metadata store. In KubeDB, there are two supported patterns:

1. **KubeDB-managed etcd**: omit `spec.metaStorage` (or set it without `externallyManaged`). KubeDB creates an internal [`Etcd`](/docs/guides/etcd/concepts/etcd.md) database for you, managed by the **KubeDB etcd operator** - the same operator, and the same `Etcd` CRD, used to run etcd as a standalone KubeDB database.
2. **Externally managed etcd**: set `spec.metaStorage.externallyManaged: true` and provide endpoints yourself.

For the KubeDB-managed path, the KubeDB etcd operator must already be installed in the cluster - enabling the `Etcd` feature gate in [Before You Begin](#before-you-begin) does exactly that; there is no separate upstream operator to install.

Verify the `Etcd` CRD and the etcd controller are present:

```bash
$ kubectl get crd etcds.kubedb.com
NAME               CREATED AT
etcds.kubedb.com   2026-07-08T...

$ kubectl get pods -n kubedb -l app.kubernetes.io/name=kubedb-provisioner
NAME                                  READY   STATUS    RESTARTS   AGE
kubedb-provisioner-6c8d9f6b7c-abcde   1/1     Running   0          1m
```

## Default Path: KubeDB-Managed etcd

For the [standalone](/docs/guides/milvus/quickstart/standalone.md) and [distributed](/docs/guides/milvus/quickstart/distributed.md) quickstarts, this is the default and recommended path.

If you **omit** `spec.metaStorage` from the `Milvus` manifest:

- KubeDB creates an internal `Etcd` CR (named `<milvus-name>-etcd`) and lets the KubeDB etcd operator reconcile it
- KubeDB wires Milvus to that internal etcd automatically
- You do **not** need to apply any external etcd YAML yourself

So for the default quickstarts, having the `Etcd` feature gate enabled is enough - KubeDB takes care of the rest, including TLS, password auth and storage for the internal etcd cluster if you ask for them. See [Securing the internal etcd cluster](#securing-the-internal-etcd-cluster) below.

### Securing the internal etcd cluster

<!-- This section is drafted from the Milvus/Etcd CRD schemas and controller source, not yet verified against a live cluster. -->

The internal `Etcd` cluster Milvus creates can be configured the same way any other KubeDB-managed etcd can, through fields on `spec.metaStorage` (only meaningful when `externallyManaged` is not `true`):

- `spec.metaStorage.tls` - issues server/client/peer certificates for the internal etcd cluster from a cert-manager `Issuer`/`ClusterIssuer`, same shape as `Etcd.spec.tls`. Requires `tls.issuerRef`.
- `spec.metaStorage.authSecret` - the etcd root credential. Omit it to let the etcd operator auto-generate one, or point at your own `kubernetes.io/basic-auth` secret.
- `spec.metaStorage.storage.storageClassName` - point this at an encrypted-volume StorageClass (e.g. an encrypted EBS/PD class, or an encrypted Longhorn/Ceph class) to get encryption at rest for the internal etcd's data - no other configuration is needed for that.

```yaml
metaStorage:
  size: 3
  storage:
    storageClassName: encrypted-ssd
    resources:
      requests:
        storage: 10Gi
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: milvus-meta-etcd-issuer
  authSecret:
    name: milvus-meta-etcd-auth
```

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
- `spec.metaStorage.tls`/`spec.metaStorage.authSecret` do not apply to this path - they configure the KubeDB-managed cluster only. Point your external etcd's own TLS/auth at Milvus through the endpoint scheme/credentials it expects instead.

Only choose this path if you intentionally want Milvus to use an external etcd cluster. The default quickstarts do not require any external etcd manifest.

## Optional Controllers

These are not required for the base quickstarts:

- Install [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator) only if you want to follow the [monitoring guide](/docs/guides/milvus/monitoring/using-prometheus-operator.md).
- Install [cert-manager](https://cert-manager.io/docs/installation/) only if you want to follow the [TLS guide](/docs/guides/milvus/tls/guide.md) or the [TLS reconfiguration guide](/docs/guides/milvus/reconfigure-tls/guide.md), or if you want to set `spec.metaStorage.tls` to secure the internal etcd cluster (see [Securing the internal etcd cluster](#securing-the-internal-etcd-cluster)).

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

   For KubeDB-managed etcd, the PVC names follow the pattern `data-<milvus-name>-etcd-<ordinal>`.

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
