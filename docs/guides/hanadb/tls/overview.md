---
title: HanaDB TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-tls-overview
    name: TLS Overview
    parent: guides-hanadb-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDB TLS/SSL Encryption

KubeDB supports providing TLS/SSL encryption for HanaDB using [cert-manager](https://cert-manager.io/).
This guide shows how to deploy a HanaDB with TLS enabled, and how to add, rotate, and remove TLS on a
running database with a `HanaDBOpsRequest`.

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/tls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/tls) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Install [cert-manager](https://cert-manager.io/docs/installation/) in your cluster (it provisions the
  certificates).
- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- Create a namespace:

```bash
$ kubectl create ns demo
namespace/demo created
```

## How TLS works in HanaDB

When `spec.tls` is set on a `HanaDB`, KubeDB asks cert-manager to issue three certificates, identified by
their **alias**:

| Alias              | Secret name                  | Used for                                          |
|--------------------|------------------------------|---------------------------------------------------|
| `server`           | `<db>-server-cert`           | The HANA SQL server certificate (mounted into the pod at `/etc/hanadb-tls/server`). |
| `client`           | `<db>-client-cert`           | Client authentication used by KubeDB.             |
| `metrics-exporter` | `<db>-metrics-exporter-cert` | The Prometheus exporter's client connection.      |

`spec.tls.issuerRef` is required and must reference a cert-manager `Issuer` or `ClusterIssuer`.

## Create an Issuer

These guides use a self-signed CA `Issuer`. First create a CA key pair and a `Secret`, then an `Issuer`
that signs with it.

```bash
# generate a CA
$ openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout ca.key -out ca.crt -subj "/CN=ca/O=kubedb"
# store it as a Secret
$ kubectl create secret tls hdb-ca --cert=ca.crt --key=ca.key -n demo
```

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: hdb-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: hdb-ca
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/tls/hdb-ca-issuer.yaml
issuer.cert-manager.io/hdb-ca-issuer created
```

## Option A — Deploy a HanaDB with TLS enabled

Set `spec.tls.issuerRef` directly on the `HanaDB`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 2
  storageType: Durable
  topology:
    mode: SystemReplication
    systemReplication:
      replicationMode: fullsync
      operationMode: logreplay_readaccess
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: hdb-ca-issuer
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/tls/system-replication-tls.yaml
hanadb.kubedb.com/hanadb-cluster created
```

## Option B — Add TLS to a running HanaDB (ReconfigureTLS)

If the database already exists without TLS, add TLS with a `HanaDBOpsRequest` of type `ReconfigureTLS`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hanadb-cluster
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: hdb-ca-issuer
  timeout: 30m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/tls/reconfigure-add-tls.yaml
hanadbopsrequest.ops.kubedb.com/hdbops-add-tls created
```

> `ReconfigureTLS` performs a **rolling restart** of the HANA pods to load the new certificates. Request
> success alone is not enough — verify the database returns to `Ready` and that TLS connections work.

Wait for the ops request to succeed:

```bash
$ kubectl get hdbops -n demo hdbops-add-tls
NAME             TYPE             STATUS       AGE
hdbops-add-tls   ReconfigureTLS   Successful   7m
```

```bash
$ kubectl describe hdbops -n demo hdbops-add-tls
...
Status:
  Conditions:
    Message:  HanaDBOpsRequest has started to reconfigure TLS
    Reason:   ReconfigureTLS
    Type:     ReconfigureTLS
    Message:  Successfully reconciled HanaDB with TLS configuration
    Reason:   UpdatePetSets
    Type:     UpdatePetSets
    Message:  Successfully restarted HanaDB nodes
    Reason:   RestartNodes
    Type:     RestartNodes
    Message:  Successfully resumed database
    Reason:   DatabaseResumeSucceeded
    Type:     DatabaseResumeSucceeded
    Message:  HanaDB is ready after reconfigureTLS
    Reason:   DatabaseReady
    Type:     DatabaseReady
    Message:  Successfully completed reconfigureTLS for HanaDB.
    Reason:   Successful
    Type:     Successful
  Phase:      Successful
```

## Verify TLS

Confirm the certificates and secrets exist, and that the server cert is mounted:

```bash
$ kubectl get issuer,certificate -n demo
NAME                                  READY   AGE
issuer.cert-manager.io/hdb-ca-issuer   True    10m

NAME                                                               READY   SECRET                                 AGE
certificate.cert-manager.io/hanadb-cluster-client-cert             True    hanadb-cluster-client-cert             3m
certificate.cert-manager.io/hanadb-cluster-metrics-exporter-cert   True    hanadb-cluster-metrics-exporter-cert   3m
certificate.cert-manager.io/hanadb-cluster-server-cert             True    hanadb-cluster-server-cert             3m

$ kubectl exec -n demo hanadb-cluster-1 -c hanadb -- /bin/sh -lc 'ls -l /etc/hanadb-tls/server'
total 0
lrwxrwxrwx 1 root root ... ca.crt -> ..data/ca.crt
lrwxrwxrwx 1 root root ... tls.crt -> ..data/tls.crt
lrwxrwxrwx 1 root root ... tls.key -> ..data/tls.key
```

Verify the TLS handshake against the SQL port (`39017`) using `openssl s_client`:

```bash
$ kubectl run hdb-tls-check -n demo --rm -i --restart=Never --image=alpine:3.20 -- \
  sh -lc "apk add --no-cache openssl >/dev/null && echo | openssl s_client -connect hanadb-cluster.demo.svc:39017 -servername hanadb-cluster.demo.svc"
...
New, TLSv1.3, Cipher is TLS_AES_256_GCM_SHA384
...
```

The handshake completes over TLS 1.3, confirming the SQL port now requires TLS. (SAP HANA serves the
SQL endpoint from its own internal PKI keystore, so the certificate subject shown by `openssl` is HANA's
`ClientPKI` rather than the cert-manager leaf; the cert-manager material is supplied to the pod at
`/etc/hanadb-tls/server`.)

## Rotate Certificates

To force cert-manager to re-issue the certificates (for example before expiry), apply a `ReconfigureTLS`
request with `tls.rotateCertificates: true`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hanadb-cluster
  tls:
    rotateCertificates: true
  timeout: 30m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/tls/reconfigure-rotate-tls.yaml
hanadbopsrequest.ops.kubedb.com/hdbops-rotate-tls created
```

KubeDB re-issues all three certificates from the same issuer and performs a rolling restart so the pods
pick up the new material. Track the request and confirm the database returns to `Ready`:

```bash
$ kubectl get hdbops -n demo hdbops-rotate-tls
$ kubectl get hanadb.kubedb.com -n demo hanadb-cluster
```

> After a certificate rotation on a System Replication cluster, verify that a `primary` role is
> re-elected (`kubectl get pods -n demo -l app.kubernetes.io/instance=hanadb-cluster -L kubedb.com/role`)
> and that the raft coordinator can connect to HANA over the new certificate. If the coordinator logs
> show `x509: certificate signed by unknown authority`, the rotated CA is not yet trusted by the
> sidecar — inspect `kubectl logs <pod> -c hanadb-coordinator` before retrying.

## Remove TLS

To disable TLS, apply a `ReconfigureTLS` request with `tls.remove: true`. For a System Replication
cluster, KubeDB keeps HANA's `sslclientpki` disabled after removal (per SAP guidance); for a standalone
database it restores HANA's built-in ClientPKI. Either way the pods are restarted.

## Cleaning Up

```bash
$ kubectl delete hdbops -n demo hdbops-add-tls hdbops-rotate-tls
$ kubectl delete hanadb.kubedb.com -n demo hanadb-cluster
$ kubectl delete issuer -n demo hdb-ca-issuer
$ kubectl delete ns demo
```

## Next Steps

- Set up [monitoring](/docs/guides/hanadb/monitoring/using-builtin-prometheus.md).
- Review the [HanaDBOpsRequest CRD](/docs/guides/hanadb/concepts/opsrequest.md).
