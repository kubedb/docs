---
title: Configure TLS/SSL for Milvus
menu:
  docs_{{ .version }}:
    identifier: milvus-tls-configure
    name: Configure TLS/SSL
    parent: milvus-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS/SSL for Milvus

This guide will show you how to deploy a Milvus database with TLS/SSL enabled from the start, using cert-manager.

## Before You Begin

- Install [cert-manager](https://cert-manager.io/docs/installation/) in your cluster.

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [TLS Overview](/docs/guides/milvus/tls/overview.md)

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO and creates the object-storage secrets (`milvus-storage-config` for `milvus-standalone`, `milvus-cluster-storage-config` for `milvus-cluster`).

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/tls/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/tls/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a cert-manager Issuer

KubeDB uses cert-manager to issue the Milvus certificates. First create a self-signed CA secret, then an `Issuer` (or `ClusterIssuer`) backed by it.

```bash
$ openssl genrsa -out ca.key 2048
$ openssl req -x509 -new -nodes -key ca.key -subj "/CN=milvus-ca/O=kubedb" -days 3650 -out ca.crt
$ kubectl create secret tls milvus-ca --cert=ca.crt --key=ca.key -n demo
secret/milvus-ca created
```

`issuer.yaml`

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: milvus-issuer
  namespace: demo
spec:
  ca:
    secretName: milvus-ca
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/tls/yamls/issuer.yaml
issuer.cert-manager.io/milvus-issuer created

$ kubectl get issuer -n demo
NAME            READY   AGE
milvus-issuer   True    5s
```

### Use a ClusterIssuer Instead

A `ClusterIssuer` works the same way, with one difference: cert-manager reads a `ClusterIssuer`'s CA secret from its cluster resource namespace, which is `cert-manager` by default, not from the database namespace. If the secret is in `demo`, the `ClusterIssuer` never becomes Ready (`Error getting keypair for CA issuer: secrets "milvus-cluster-ca" not found`). Create the CA secret in `cert-manager`, then apply the sample `cluster-issuer.yaml`:

```bash
$ kubectl create secret tls milvus-cluster-ca --cert=ca.crt --key=ca.key -n cert-manager
secret/milvus-cluster-ca created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/tls/yamls/cluster-issuer.yaml
clusterissuer.cert-manager.io/milvus-cluster-issuer created

$ kubectl get clusterissuer milvus-cluster-issuer
```

Then reference it from the `Milvus` manifest with `kind: ClusterIssuer`:

```yaml
spec:
  tls:
    issuerRef:
      name: milvus-cluster-issuer
      kind: ClusterIssuer
      apiGroup: "cert-manager.io"
```

The rest of this guide uses the namespaced `milvus-issuer`.

## Deploy a TLS-Secured Standalone Milvus

The standalone manifest enables TLS via `spec.tls`:

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
      name: "milvus-storage-config"
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    storageClassName: local-path
    resources:
      requests:
        storage: 1Gi
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

- `spec.tls.external.mode: mTLS` requires clients to present a certificate.
- `spec.tls.internal.mode: TLS` encrypts inter-component traffic.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/tls/yamls/standalone.yaml
milvus.kubedb.com/milvus-standalone created
```

Wait until it is `Ready`:

```bash
$ kubectl get milvuses.kubedb.com -n demo milvus-standalone -w
```

> If you deleted a Milvus of the same name in an earlier guide, delete its object-storage prefix first, as shown in [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#delete-a-milvus-instances-object-storage-data). Otherwise the new database can hang in `Provisioning`.

## Verify TLS

### Certificate Secrets

KubeDB requests the `server` and `client` certificates and stores them in secrets:

```bash
$ kubectl get secret -n demo | grep -E 'milvus-standalone-(server|client)-cert'
milvus-standalone-client-cert   kubernetes.io/tls   4   91s
milvus-standalone-server-cert   kubernetes.io/tls   3   91s
```

### TLS Files Mounted in the Pod

The certificates and CA are mounted at `/milvus/tls`:

```bash
$ kubectl exec -n demo milvus-standalone-0 -c milvus -- ls -l /milvus/tls
ca.pem
client.key
client.pem
server.key
server.pem
```

### Rendered Configuration

The rendered `milvus.yaml` wires the certificates into Milvus. It is stored in a secret named `<db-name>-<last 6 characters of the Milvus object's UID>`, for example `milvus-standalone-f93293`:

```bash
$ DB_UID=$(kubectl get milvuses.kubedb.com milvus-standalone -n demo -o jsonpath='{.metadata.uid}')
$ CFG=milvus-standalone-${DB_UID: -6}
$ kubectl get secret $CFG -n demo -o jsonpath='{.data.milvus\.yaml}' | base64 -d | grep -A4 internaltls
internaltls:
    caPemPath: /milvus/tls/ca.pem
    serverKeyPath: /milvus/tls/server.key
    serverPemPath: /milvus/tls/server.pem
    sni: milvus-standalone
```

> The rendered `milvus.yaml` also contains the root password in plain text (`common.security.defaultRootPassword`), so treat this secret as sensitive.

### AppBinding Scheme

Because TLS is enabled, the AppBinding connection scheme is `https`:

```bash
$ kubectl get appbinding milvus-standalone -n demo -o jsonpath='{.spec.clientConfig.service.scheme}'
https
```

## Connect Through External mTLS

This guide uses `spec.tls.external.mode: mTLS`, so clients must present the client certificate.

### Extract the CA and client certificate

```bash
$ kubectl get secret milvus-standalone-server-cert -n demo \
    -o jsonpath='{.data.ca\.crt}' | base64 -d > /tmp/ca.crt

$ kubectl get secret milvus-standalone-client-cert -n demo \
    -o jsonpath='{.data.tls\.crt}' | base64 -d > /tmp/client.crt

$ kubectl get secret milvus-standalone-client-cert -n demo \
    -o jsonpath='{.data.tls\.key}' | base64 -d > /tmp/client.key

$ PASSWORD=$(kubectl get secret milvus-standalone-auth -n demo -o jsonpath='{.data.password}' | base64 -d)
```

### Port-forward the REST port

Run this in a separate terminal:

```bash
$ kubectl port-forward svc/milvus-standalone -n demo 8080:8080
Forwarding from 127.0.0.1:8080 -> 8080
```

### Create a collection over HTTPS

```bash
$ curl -s -X POST "https://localhost:8080/v2/vectordb/collections/create" \
    --cacert /tmp/ca.crt \
    --cert /tmp/client.crt \
    --key /tmp/client.key \
    -H "Authorization: Bearer root:${PASSWORD}" \
    -H "Content-Type: application/json" \
    -d '{
      "collectionName": "tls_test_collection",
      "dbName": "default",
      "dimension": 4,
      "metricType": "L2"
    }' | jq .
```

### Insert and search sample vectors

```bash
$ curl -s -X POST "https://localhost:8080/v2/vectordb/entities/insert" \
    --cacert /tmp/ca.crt \
    --cert /tmp/client.crt \
    --key /tmp/client.key \
    -H "Authorization: Bearer root:${PASSWORD}" \
    -H "Content-Type: application/json" \
    -d '{
      "collectionName": "tls_test_collection",
      "dbName": "default",
      "data": [
        {"id": 1, "vector": [0.9, 0.8, 0.7, 0.6]},
        {"id": 2, "vector": [0.5, 0.4, 0.3, 0.2]}
      ]
    }' | jq .

$ curl -s -X POST "https://localhost:8080/v2/vectordb/entities/search" \
    --cacert /tmp/ca.crt \
    --cert /tmp/client.crt \
    --key /tmp/client.key \
    -H "Authorization: Bearer root:${PASSWORD}" \
    -H "Content-Type: application/json" \
    -d '{
      "collectionName": "tls_test_collection",
      "dbName": "default",
      "data": [[0.5, 0.4, 0.3, 0.2]],
      "topK": 2,
      "metricType": "L2"
    }' | jq .
```

If you omit `--cert` and `--key`, the request should fail because client certificate authentication is required.

## TLS-Secured Distributed Milvus

The distributed manifest enables TLS the same way. The `server`/`client` certificates are issued once and propagated to every distributed role (`mixcoord`, `datanode`, `querynode`, `streamingnode`, `proxy`).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-cluster
  namespace: demo
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: "milvus-cluster-storage-config"
  topology:
    mode: Distributed
    distributed:
      streamingnode:
        storageType: Durable
        storage:
          accessModes:
            - ReadWriteOnce
          storageClassName: local-path
          resources:
            requests:
              storage: 1Gi
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/tls/yamls/distributed.yaml
milvus.kubedb.com/milvus-cluster created
```

After it becomes `Ready`, the certificate secrets exist, the certificates are mounted into every role's pods, and the AppBinding scheme is `https`:

```bash
$ kubectl get secret -n demo | grep -E 'milvus-cluster-(server|client)-cert'
milvus-cluster-client-cert   kubernetes.io/tls   4   4m
milvus-cluster-server-cert   kubernetes.io/tls   3   4m

$ kubectl exec -n demo milvus-cluster-mixcoord-0 -c milvus -- ls /milvus/tls
ca.pem
client.key
client.pem
server.key
server.pem

$ kubectl get appbinding milvus-cluster -n demo -o jsonpath='{.spec.clientConfig.service.scheme}'
https
```

### Verify internal TLS between distributed roles

You can verify inter-component TLS from the `proxy` pod:

```bash
$ kubectl exec -it milvus-cluster-proxy-0 -n demo -- \
    openssl s_client \
    -connect milvus-cluster-querynode.demo.svc.cluster.local:21123 \
    -CAfile /milvus/tls/ca.pem \
    -servername milvus-cluster \
    -brief
```

You should see a successful TLS handshake.

## TLS to the Object Storage

The TLS above secures connections to Milvus. Milvus also connects to the object storage, and if that serves a certificate signed by a private CA, Milvus must trust the CA. Give KubeDB the CA with `spec.objectStorage.tls`.

Create a Secret with the CA certificate (use a ConfigMap and `caConfigMap` if the CA is published that way):

```bash
$ kubectl create secret generic -n demo s3-ca --from-file=ca.crt=./ca.crt
```

Create the object storage secret for the TLS endpoint and the Milvus object:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-s3-tls
  namespace: demo
spec:
  version: "2.6.15"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: milvus-s3
    tls:
      caSecret:
        name: s3-ca
        key: ca.crt
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

The full manifest, with the `milvus-s3` secret, is in [object-storage-tls.yaml](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/tls/yamls/object-storage-tls.yaml).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/tls/yamls/object-storage-tls.yaml
```

### Independent TLS for backup and restore

`spec.objectStorage.tls` works even when `spec.tls` is omitted. The latter controls connections to Milvus itself; it does not enable or disable HTTPS to the object storage.

Deploy the updated provisioner and webhook API together with Milvus archiver v0.2.1 or later, backup plugin v0.1.1 or later, and CSI plugin v0.2.1 or later. These versions must be published before deploying the corresponding catalog; older worker images do not contain this support. The backup workers read the operator-generated `milvus.yaml` and load the selected CA through the Kubernetes API. The storage-check Job receives a read-only CA mount. No KubeDB-issued listener certificate is needed for object-storage verification.

The backup destination uses the existing KubeStash configuration, independently of the live store:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: backup-storage-credentials
  namespace: demo
stringData:
  AWS_ACCESS_KEY_ID: <backup-access-key>
  AWS_SECRET_ACCESS_KEY: <backup-secret-key>
  CA_CERT_DATA: |
    <PEM CA bundle for the backup endpoint>
---
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: milvus-backups
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      endpoint: https://backup.example.com
      bucket: milvus-backups
      region: us-east-1
      secretName: backup-storage-credentials
  usagePolicy:
    allowedNamespaces:
      from: Same
  deletionPolicy: Delete
```

The live endpoint and backup endpoint can use different CAs. Physical backup and PITR clients keep separate trust pools. The logical backup plugin gives its external `milvus-backup` process a bundle containing both CAs and the container's normal public roots.

For restore into another namespace or object store, create the target credential and CA resources first. The target's CA is used to write restored data. Existing customer-managed resources are not overwritten from the source backup.

For a physical restore, set the source repository namespace explicitly:

```yaml
spec:
  init:
    archiver:
      fullDBRepository:
        name: source-full
        namespace: source-namespace
      recoveryTimestamp: "2026-11-04T12:00:00Z"
      encryptionSecret:
        name: backup-encryption
        namespace: source-namespace
```

For a logical `RestoreSession`, use `spec.dataSource.namespace` for the source repository and `spec.target.namespace` for the target Milvus. Configure KubeStash usage policies and access for the namespaces involved. The target's object-storage Secret and selected CA must exist in the target namespace; the source CA is not copied over them.

Physical backup, PITR and CSI workers currently require object-storage access keys. IAM-only authentication is not supported by those workers.


### Certificate renewal

New backup and restore jobs load the current CA when they start. The continuous archiver checks configuration and CA changes every minute and reconnects from its last published checkpoint. A bad replacement certificate stops progress rather than disabling verification.

To renew a private CA, first publish a bundle containing both the old and new CA. Restart Milvus pods through the existing restart workflow so they reload trust, then change the storage server certificate. Remove the old CA after clients have switched and restart Milvus again. The mounted file alone does not guarantee that a running Milvus process reloads it.

For OpenShift service certificates, use a ConfigMap annotated with `service.beta.openshift.io/inject-cabundle: "true"` and select its `service-ca.crt` key. The OpenShift service-CA controller owns that bundle.

### Verify

The CA is mounted in the pod and the rendered configuration enables TLS to the object storage:

```bash
$ kubectl exec -n demo milvus-s3-tls-0 -- ls /milvus/object-storage-ca
ca.crt

$ CS=$(kubectl get petset -n demo milvus-s3-tls -o jsonpath='{.spec.template.spec.volumes[?(@.name=="milvus-config")].secret.secretName}')
$ kubectl get secret -n demo $CS -o jsonpath='{.data.milvus\.yaml}' | base64 -d | grep -A12 '^minio:'
minio:
  ...
  ssl:
    tlsCACert: /milvus/object-storage-ca/ca.crt
  useSSL: true
```

The Milvus pod log has no `x509` errors, and the database becomes `Ready`.

## Cleaning up

```bash
$ kubectl patch -n demo milvus.kubedb.com milvus-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl patch -n demo milvus.kubedb.com milvus-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone milvus-cluster
$ kubectl delete milvuses.kubedb.com -n demo milvus-s3-tls --ignore-not-found
$ kubectl delete secret -n demo s3-ca milvus-s3 --ignore-not-found
$ kubectl delete issuer -n demo milvus-issuer
$ kubectl delete secret -n demo milvus-ca
$ kubectl delete clusterissuer milvus-cluster-issuer
$ kubectl delete secret -n cert-manager milvus-cluster-ca
$ kubectl delete ns demo
```

Skip the two `ClusterIssuer` commands if you did not create it.

## Next Steps

- Add, rotate or remove TLS on a running database with [Reconfigure TLS](/docs/guides/milvus/reconfigure-tls/guide.md).
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
