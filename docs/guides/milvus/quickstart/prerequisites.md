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

- Object storage, exposed through a secret (one per Milvus instance; `milvus-storage-config` and `milvus-cluster-storage-config` in these guides).
- etcd for metadata.

This guide sets up object storage in the `demo` namespace and explains the two etcd options: an internal etcd that KubeDB creates for you (the default), or an **external etcd cluster** you manage yourself.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured to talk to it.
- Install KubeDB with the Milvus and Etcd feature gates enabled. The internal metadata store that KubeDB creates for each Milvus is a KubeDB `Etcd` object, so the Etcd feature gate is required too:

  ```bash
  helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
    --namespace kubedb --create-namespace \
    --set global.featureGates.Milvus=true \
    --set global.featureGates.Etcd=true
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

The tenant chart's default image (`quay.io/minio/minio`) can no longer be pulled anonymously (the pull fails with `401 UNAUTHORIZED` and the pod stays in `ImagePullBackOff`), so the command below overrides it with a pullable MinIO image:

```bash
$ helm upgrade --install --namespace "demo" milvus-minio minio/tenant \
  --set tenant.name=milvus-minio \
  --set tenant.image.repository=docker.io/bitnamilegacy/minio \
  --set tenant.image.tag=2025.7.23-debian-12-r3 \
  --set tenant.pools[0].servers=1 \
  --set tenant.pools[0].volumesPerServer=1 \
  --set tenant.pools[0].size=1Gi \
  --set tenant.certificate.requestAutoCert=false \
  --set tenant.buckets[0].name="mlv-release" \
  --set tenant.pools[0].name="default"
```

> The MinIO Operator allows only one Tenant per namespace. If the namespace already has a Tenant, reuse it as described in [If You Already Have S3 or MinIO](#if-you-already-have-s3-or-minio). Also note that `tenant.name` does not rename the chart's config secret (`myminio-env-configuration`), so a second tenant release in the same namespace collides with the first one.

Wait for the Tenant to become initialized:

```bash
$ kubectl get tenant -n demo milvus-minio -w
NAME           STATE         HEALTH   AGE
milvus-minio   Initialized   green    1m
```

### 3. Create the Storage Config Secrets

Create a secret with your MinIO connection details. The keys must use the milvus.yaml config key names directly (KubeDB copies them into the `minio` section of `milvus.yaml`).

**Use a distinct `rootPath` for every Milvus instance.** `rootPath` is the prefix under which a Milvus stores all of its object-storage data (segments, logs and its write-ahead log). Two Milvus instances on the same bucket and `rootPath` read and write the same objects: one instance's garbage collection can delete the other's files, and their write-ahead-log lock objects collide. A secret holds one `rootPath`, so create one secret per Milvus instance.

Deleting a Milvus (with any `deletionPolicy`) does **not** delete its objects under `<bucketName>/<rootPath>/`. Reusing that `rootPath` for a **new** Milvus is not supported: the new instance gets a new metadata etcd, collides with the old write-ahead-log lock objects under `<rootPath>/wp/`, and can hang in `Provisioning` with `writer lock session has expired` in its logs. Before you create a new database on a `rootPath` that was used before, either change `rootPath` in its secret or [delete the old prefix](#delete-a-milvus-instances-object-storage-data) first.

The guides deploy a standalone Milvus named `milvus-standalone`, which uses `milvus-storage-config`, and a distributed Milvus named `milvus-cluster`, which uses `milvus-cluster-storage-config`. Both can run at the same time:

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
  rootPath: "milvus-standalone"
  port: "9000"
---
apiVersion: v1
kind: Secret
metadata:
  name: milvus-cluster-storage-config
  namespace: demo
type: Opaque
stringData:
  address: "milvus-minio-hl.demo.svc.cluster.local"
  accessKeyID: "minio"
  secretAccessKey: "minio123"
  bucketName: "mlv-release"
  rootPath: "milvus-cluster"
  port: "9000"
```

Apply the secrets:

```bash
$ kubectl apply -f milvus-storage-config.yaml
secret/milvus-storage-config created
secret/milvus-cluster-storage-config created
```

Verify:

```bash
$ kubectl get secret -n demo milvus-storage-config milvus-cluster-storage-config
NAME                            TYPE     DATA   AGE
milvus-storage-config           Opaque   6      1m
milvus-cluster-storage-config   Opaque   6      1m
```

### If You Already Have S3 or MinIO

You do not need to use the MinIO Operator. You can instead create only the secrets and point Milvus at your existing object store (again, one secret with its own `rootPath` per Milvus instance):

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

## etcd for Metadata

Milvus always uses etcd as its metadata store. In KubeDB, there are two supported patterns:

1. **KubeDB-managed etcd**: omit `spec.metaStorage`. KubeDB creates an internal etcd for you as a KubeDB `Etcd` object. This needs the `Etcd` feature gate (`--set global.featureGates.Etcd=true`) in the KubeDB install; no separate etcd operator is needed.
2. **Externally managed etcd**: set `spec.metaStorage.externallyManaged: true` and provide endpoints yourself.

Verify that the KubeDB `Etcd` CRD is registered:

```bash
$ kubectl get crd etcds.kubedb.com
```

## Default Path: KubeDB-Managed etcd

For the [standalone](/docs/guides/milvus/quickstart/standalone.md) and [distributed](/docs/guides/milvus/quickstart/distributed.md) quickstarts, this is the default and recommended path.

If you **omit** `spec.metaStorage` from the `Milvus` manifest:

- KubeDB creates an internal `Etcd` object named `<milvus-name>-etcd` (3 members) and an auth secret `<milvus-name>-etcd-auth` for it
- KubeDB wires Milvus to that internal etcd automatically
- You do **not** need to apply any etcd YAML yourself

So for the default quickstarts, **the `Etcd` feature gate in the KubeDB install is enough**.

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
- If the external etcd requires authentication or TLS, reference its credentials with `spec.metaStorage.authSecret` and its client certificate with `spec.metaStorage.clientTLSSecret` (see the [Milvus concepts](/docs/guides/milvus/concepts/milvus.md#specmetastorage)).

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

1. Delete your Milvus objects first (see the cleanup section of each guide). With `deletionPolicy: Delete` or `WipeOut`, KubeDB also deletes the internal etcd and its PVCs.

2. Delete what the internal etcd leaves behind. The `<milvus-name>-etcd-auth` secret is kept even after a `WipeOut` delete. If any etcd PVCs remain (for example, after `deletionPolicy: Halt`), delete them too. Their names follow the pattern `data-<milvus-name>-etcd-<ordinal>`:

   ```bash
   $ kubectl delete pvc -n demo data-milvus-standalone-etcd-0 data-milvus-standalone-etcd-1 data-milvus-standalone-etcd-2
   $ kubectl delete secret -n demo milvus-standalone-etcd-auth
   ```

   If you use an external etcd, clean it up with whatever workflow manages it.

3. Delete each Milvus instance's object-storage data, as shown in [Delete a Milvus Instance's Object-Storage Data](#delete-a-milvus-instances-object-storage-data).

4. Delete the MinIO Tenant:

   ```bash
   $ helm uninstall milvus-minio -n demo
   ```

5. Delete the storage config secrets:

   ```bash
   $ kubectl delete secret -n demo milvus-storage-config milvus-cluster-storage-config
   ```

## Delete a Milvus Instance's Object-Storage Data

KubeDB does not delete a Milvus instance's objects from object storage, even with `deletionPolicy: WipeOut`. Once the `Milvus` object is gone, remove its `rootPath` prefix with the MinIO client (`mc`) before reusing that `rootPath` for a new Milvus. For the tenant created above, port-forward the MinIO service in a separate terminal:

```bash
$ kubectl port-forward svc/milvus-minio-hl -n demo 9000:9000
```

Then register an alias and delete the prefix of the deleted database, `milvus-standalone` in this example:

```bash
$ mc alias set milvus-minio http://localhost:9000 minio minio123
$ mc rm -r --force milvus-minio/mlv-release/milvus-standalone
```

For your own object store, use its endpoint, credentials, `bucketName` and `rootPath` instead (`mc rm -r --force <alias>/<bucket>/<rootPath>`).

## Next Steps

- [Deploy standalone Milvus](/docs/guides/milvus/quickstart/standalone.md)
- [Deploy distributed Milvus](/docs/guides/milvus/quickstart/distributed.md)
