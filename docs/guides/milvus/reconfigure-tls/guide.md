---
title: Reconfigure Milvus TLS/SSL
menu:
  docs_{{ .version }}:
    identifier: milvus-reconfigure-tls-guide
    name: Guide
    parent: milvus-reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Milvus TLS/SSL

This guide will show you how to use the `KubeDB` Ops-manager operator to add, rotate, change the issuer of, and remove TLS for a Milvus database using a `MilvusOpsRequest`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)
  - [Reconfigure TLS Overview](/docs/guides/milvus/reconfigure-tls/overview.md)

- Install [cert-manager](https://cert-manager.io/docs/installation/) in your cluster — KubeDB uses it to manage Milvus certificates.

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO, creates the `my-release-minio` secret, and installs the etcd operator required by Milvus.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/reconfigure-tls/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/reconfigure-tls/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a cert-manager Issuer

All TLS operations need an `Issuer` (or `ClusterIssuer`). First create a CA secret, then an `Issuer` backed by it:

```bash
# generate a self-signed CA
$ openssl genrsa -out ca.key 2048
$ openssl req -x509 -new -nodes -key ca.key -subj "/CN=milvus-ca/O=kubedb" -days 3650 -out ca.crt
$ kubectl create secret tls milvus-ca --cert=ca.crt --key=ca.key -n demo
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/reconfigure-tls/yamls/issuer.yaml
issuer.cert-manager.io/milvus-issuer created
```

## Reconfigure TLS — Standalone Milvus

The four TLS operations below are demonstrated on the standalone database `milvus-standalone`. (For the distributed database, point `spec.databaseRef.name` at `milvus-cluster` — see the [distributed section](#reconfigure-tls--distributed-milvus).)

### 1. Add TLS

Add TLS to a database that does not currently have it:

`reconfigureTls-add-standalone.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: mvops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: milvus-standalone
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

- `spec.tls.external.mode` controls client-facing traffic (`Disabled`/`TLS`/`mTLS`).
- `spec.tls.internal.mode` controls inter-component traffic (`Disabled`/`TLS`).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/reconfigure-tls/yamls/reconfigureTls-add-standalone.yaml
milvusopsrequest.ops.kubedb.com/mvops-add-tls created

$ kubectl get milvusopsrequest mvops-add-tls -n demo
NAME            TYPE             STATUS       AGE
mvops-add-tls   ReconfigureTLS   Successful   82s
```

```bash
$ kubectl describe milvusopsrequest mvops-add-tls -n demo
...
  Normal   CertificateSynced  Successfully synced all certificates
  Normal   UpdatePetSets      successfully reconciled the Milvus with tls configuration
  Normal   RestartNodes       Successfully restarted all nodes
  Normal   Successful         Successfully resumed Milvus database: demo/milvus-standalone for MilvusOpsRequest: mvops-add-tls
```

After adding TLS, the certificate secrets exist, the AppBinding scheme becomes `https`, and the certificates are mounted in the pod:

```bash
$ kubectl get secret -n demo | grep -E 'milvus-standalone-(server|client)-cert'
milvus-standalone-client-cert   kubernetes.io/tls   4   91s
milvus-standalone-server-cert   kubernetes.io/tls   3   91s

$ kubectl get appbinding milvus-standalone -n demo -o jsonpath='{.spec.clientConfig.service.scheme}'
https

$ kubectl exec -n demo milvus-standalone-0 -c milvus -- ls /milvus/tls
ca.pem
client.key
client.pem
server.key
server.pem
```

### 2. Rotate Certificates

Re-issue the certificates from the same issuer:

`reconfigureTls-rotate-standalone.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: mvops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: milvus-standalone
  tls:
    rotateCertificates: true
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/reconfigure-tls/yamls/reconfigureTls-rotate-standalone.yaml
milvusopsrequest.ops.kubedb.com/mvops-rotate created

$ kubectl get milvusopsrequest mvops-rotate -n demo
NAME           TYPE             STATUS       AGE
mvops-rotate   ReconfigureTLS   Successful   52s
```

The server certificate serial number changes, confirming the certificate was re-issued:

```bash
# before rotation
serial=7E91C774E29D1C7F9EF578F956480A55C09DAC27
# after rotation
serial=391AEF6EC65BFF56119784DFB629C080C4FB5FB3
```

### 3. Change the Issuer

Point the database at a different issuer (here, `mv-new-issuer`, backed by a different CA). Create the new issuer first (same procedure as above, with its own CA secret):

`reconfigureTls-add-new-issuer-standalone.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: mv-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: milvus-standalone
  tls:
    issuerRef:
      name: mv-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/reconfigure-tls/yamls/reconfigureTls-add-new-issuer-standalone.yaml
milvusopsrequest.ops.kubedb.com/mv-change-issuer created

$ kubectl get milvusopsrequest mv-change-issuer -n demo
NAME               TYPE             STATUS       AGE
mv-change-issuer   ReconfigureTLS   Successful   62s
```

The database's issuer reference is updated and the new certificate chains to the new CA:

```bash
$ kubectl get milvuses.kubedb.com milvus-standalone -n demo -o jsonpath='{.spec.tls.issuerRef}'
{"apiGroup":"cert-manager.io","kind":"Issuer","name":"mv-new-issuer"}

# certificate issuer before:  issuer=CN=milvus-ca, O=kubedb
# certificate issuer after:   issuer=CN=mvnew-ca, O=kubedb
```

### 4. Remove TLS

Remove TLS from the database:

`reconfigureTls-remove-standalone.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: mvops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: milvus-standalone
  tls:
    remove: true
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/reconfigure-tls/yamls/reconfigureTls-remove-standalone.yaml
milvusopsrequest.ops.kubedb.com/mvops-remove created

$ kubectl get milvusopsrequest mvops-remove -n demo
NAME           TYPE             STATUS       AGE
mvops-remove   ReconfigureTLS   Successful   82s
```

After removal, the `tls` block is gone, the certificate secrets are removed, and the AppBinding scheme reverts to `http`:

```bash
$ kubectl get milvuses.kubedb.com milvus-standalone -n demo -o jsonpath='{.spec.tls}'
# (empty)

$ kubectl get appbinding milvus-standalone -n demo -o jsonpath='{.spec.clientConfig.service.scheme}'
http

$ kubectl get secret -n demo | grep -E 'milvus-standalone-(server|client)-cert'
# (no cert secrets)
```

## Reconfigure TLS — Distributed Milvus

The same four operations apply to a distributed Milvus; only `spec.databaseRef.name` differs (`milvus-cluster`). The operator issues/rotates the `server` and `client` certificates and propagates them to every distributed role.

For example, adding TLS to the distributed database:

`reconfigureTls-add-distributed.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: mvops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: milvus-cluster
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

The `reconfigureTls-rotate-distributed.yaml`, `reconfigureTls-add-new-issuer-distributed.yaml` and `reconfigureTls-remove-distributed.yaml` files mirror the standalone ones, targeting `milvus-cluster`.

On the distributed database the operator drives each flow exactly as for standalone, restarting every role's pods (`mixcoord`, `datanode`, `querynode`, `streamingnode`, `proxy`, and any extra replicas) one at a time — so each flow takes proportionally longer.

**Remove TLS** (distributed):

```bash
$ kubectl get milvusopsrequest mvops-remove -n demo
NAME           TYPE             STATUS       AGE
mvops-remove   ReconfigureTLS   Successful   11m

# after removal: certificate secrets are deleted and the AppBinding scheme reverts to http
$ kubectl get appbinding milvus-cluster -n demo -o jsonpath='{.spec.clientConfig.service.scheme}'
http
$ kubectl get secret -n demo | grep -E 'milvus-cluster-(server|client)-cert'
# (no cert secrets)
```

**Add TLS** (distributed) — the `server`/`client` certificate secrets are recreated and mounted into every role's pods:

```bash
$ kubectl get milvusopsrequest mvops-add-tls -n demo
NAME            TYPE             STATUS       AGE
mvops-add-tls   ReconfigureTLS   Successful   9m

$ kubectl get secret -n demo | grep -E 'milvus-cluster-(server|client)-cert'
milvus-cluster-client-cert   kubernetes.io/tls   4   2m
milvus-cluster-server-cert   kubernetes.io/tls   3   2m

$ kubectl get appbinding milvus-cluster -n demo -o jsonpath='{.spec.clientConfig.service.scheme}'
https
```

**Rotate certificates** and **Change issuer** behave identically to the standalone flows shown above — applying `reconfigureTls-rotate-distributed.yaml` re-issues the `server`/`client` certificates (the serial numbers change), and `reconfigureTls-add-new-issuer-distributed.yaml` repoints `spec.tls.issuerRef` to `mv-new-issuer` so new certificates chain to the new CA.

> On a single-node test cluster, each distributed `ReconfigureTLS` flow can take several minutes because every role (plus any scaled-out replicas) is restarted sequentially.

## Cleaning up

```bash
$ kubectl delete milvusopsrequest -n demo mvops-add-tls mvops-rotate mv-change-issuer mvops-remove
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone
$ kubectl delete ns demo
```

## Next Steps

- Deploy a [TLS-secured Milvus](/docs/guides/milvus/tls/guide.md) from the start.
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
