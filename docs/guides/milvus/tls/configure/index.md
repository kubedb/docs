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

- An object-storage secret named `my-release-minio` must exist in the `demo` namespace.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/tls/configure/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/tls/configure/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a cert-manager Issuer

KubeDB uses cert-manager to issue the Milvus certificates. First create a self-signed CA secret, then an `Issuer` (or `ClusterIssuer`) backed by it.

```bash
openssl genrsa -out ca.key 2048
```

```bash
openssl req -x509 -new -nodes -key ca.key -subj "/CN=milvus-ca/O=kubedb" -days 3650 -out ca.crt
```

```bash
kubectl create secret tls milvus-ca --cert=ca.crt --key=ca.key -n demo
```
secret/milvus-ca created

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
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/tls/configure/yamls/issuer.yaml
```
issuer.cert-manager.io/milvus-issuer created

```bash
kubectl get issuer -n demo
```
NAME            READY   AGE
milvus-issuer   True    5s

> A `ClusterIssuer` works the same way; a sample `cluster-issuer.yaml` (backed by secret `milvus-cluster-ca`) is included in the `yamls` folder. With a `ClusterIssuer`, set `spec.tls.issuerRef.kind: ClusterIssuer`.

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
      name: "my-release-minio"
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
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/tls/configure/yamls/standalone.yaml
```
milvus.kubedb.com/milvus-standalone created

Wait until it is `Ready`.

## Verify TLS

### Certificate Secrets

KubeDB requests the `server` and `client` certificates and stores them in secrets:

```bash
kubectl get secret -n demo | grep -E 'milvus-standalone-(server|client)-cert'
```
milvus-standalone-client-cert   kubernetes.io/tls   4   91s
milvus-standalone-server-cert   kubernetes.io/tls   3   91s

### TLS Files Mounted in the Pod

The certificates and CA are mounted at `/milvus/tls`:

```bash
kubectl exec -n demo milvus-standalone-0 -c milvus -- ls -l /milvus/tls
```
ca.pem
client.key
client.pem
server.key
server.pem

### Rendered Configuration

The rendered `milvus.yaml` wires the certificates into Milvus:

```bash
kubectl get secret <config-secret> -n demo -o jsonpath='{.data.milvus\.yaml}' | base64 -d | grep -A4 internaltls
```
internaltls:
    caPemPath: /milvus/tls/ca.pem
    serverKeyPath: /milvus/tls/server.key
    serverPemPath: /milvus/tls/server.pem
    sni: milvus-standalone

### AppBinding Scheme

Because TLS is enabled, the AppBinding connection scheme is `https`:

```bash
kubectl get appbinding milvus-standalone -n demo -o jsonpath='{.spec.clientConfig.service.scheme}'
```
https

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
      name: "my-release-minio"
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

After it becomes `Ready`, the certificate secrets exist, the certificates are mounted into every role's pods, and the AppBinding scheme is `https`:

```bash
kubectl get secret -n demo | grep -E 'milvus-cluster-(server|client)-cert'
```
milvus-cluster-client-cert   kubernetes.io/tls   4   4m
milvus-cluster-server-cert   kubernetes.io/tls   3   4m

```bash
kubectl exec -n demo milvus-cluster-mixcoord-0 -c milvus -- ls /milvus/tls
```
ca.pem
client.key
client.pem
server.key
server.pem

```bash
kubectl get appbinding milvus-cluster -n demo -o jsonpath='{.spec.clientConfig.service.scheme}'
```
https

## Cleaning up

```bash
kubectl delete milvus.kubedb.com -n demo milvus-standalone
```

```bash
kubectl delete issuer -n demo milvus-issuer
```

```bash
kubectl delete secret -n demo milvus-ca
```

```bash
kubectl delete ns demo
```

## Next Steps

- Add, rotate or remove TLS on a running database with [Reconfigure TLS](/docs/guides/milvus/reconfigure-tls/guide.md).
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
