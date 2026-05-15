---
title: Reconfigure TLS of Oracle
menu:
  docs_{{ .version }}:
    identifier: oracle-reconfigure-tls-cluster
    name: Cluster
    parent: oracle-reconfigure-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Oracle TLS/SSL (Transport Encryption)

KubeDB supports reconfiguring TLS/SSL certificates for Oracle — adding, removing, updating, and rotating certificates via a `OracleOpsRequest`. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `cert-manager` v1.0.0 or later to your cluster to manage your SSL/TLS certificates from [here](https://cert-manager.io/docs/installation/).

- Now, install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)
  - [TLS/SSL Overview](/docs/guides/oracle/tls/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/oracle/reconfigure-tls/yamls](/docs/guides/oracle/reconfigure-tls/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Add TLS to a Oracle database

Here, we are going to create a Oracle database without TLS and then reconfigure the database to use TLS.

### Deploy Oracle without TLS

In this section, we are going to deploy a Oracle cluster without TLS. Below is the YAML of the `Oracle` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
  namespace: demo
spec:
  version: "21.3.0"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/reconfigure-tls/yamls/oracle.yaml
oracle.kubedb.com/oracle-sample created
```

Now, wait until `oracle-sample` has status `Ready`:

```bash
$ kubectl get oracle -n demo
NAME             VERSION   STATUS   AGE
oracle-sample    1.17.0    Ready    3m22s
```

### Create Issuer

Now, we are going to create an example `Issuer` that will be used to enable SSL/TLS in Oracle. Alternatively you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

1. Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=oracle/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
```

2. Create a secret using the certificate files we have just generated,

```bash
$ kubectl create secret tls oracle-ca --cert=ca.crt  --key=ca.key --namespace=demo
secret/oracle-ca created
```

3. Now we are going to create an `Issuer` using the `oracle-ca` secret that contains the CA certificate we have just created:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: oracle-issuer
  namespace: demo
spec:
  ca:
    secretName: oracle-ca
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/reconfigure-tls/yamls/issuer.yaml
issuer.cert-manager.io/oracle-issuer created
```

### Add TLS

Now, we are going to create a `OracleOpsRequest` to add TLS to the running database.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: oracle-sample
  tls:
    issuerRef:
      name: oracle-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `oracle-sample` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer to use for signing certificates.
- `spec.tls.certificates` specifies the certificate configuration.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/reconfigure-tls/yamls/add-tls.yaml
oracleopsrequest.ops.kubedb.com/qdops-add-tls created
```

Let's wait for `OracleOpsRequest` to be `Successful`:

```bash
$ watch -n 3 kubectl get OracleOpsRequest -n demo qdops-add-tls
Every 3.0s: kubectl get OracleOpsRequest -n demo qdops-add-tls

NAME             TYPE             STATUS       AGE
qdops-add-tls    ReconfigureTLS   Successful   6m30s
```

## Rotate Certificates

Now we are going to rotate the certificates of the database.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: oracle-sample
  tls:
    rotateCertificates: true
```

Here,

- `spec.tls.rotateCertificates` specifies that we are requesting to rotate the certificates of the `oracle-sample` database.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/reconfigure-tls/yamls/rotate-tls.yaml
oracleopsrequest.ops.kubedb.com/qdops-rotate-tls created
```

Let's wait for `OracleOpsRequest` to be `Successful`:

```bash
$ kubectl get oracleopsrequest -n demo qdops-rotate-tls
NAME               TYPE             STATUS       AGE
qdops-rotate-tls   ReconfigureTLS   Successful   3m8s
```

## Remove TLS from the Database

In this section, we are going to reconfigure TLS setting of the database by removing the TLS configuration.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: oracle-sample
  tls:
    remove: true
```

Here,

- `spec.tls.remove` specifies that we are removing the TLS configuration from `oracle-sample` database.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/reconfigure-tls/yamls/remove-tls.yaml
oracleopsrequest.ops.kubedb.com/qdops-remove-tls created
```

Let's wait for `OracleOpsRequest` to be `Successful`:

```bash
$ kubectl get oracleopsrequest -n demo qdops-remove-tls
NAME               TYPE             STATUS       AGE
qdops-remove-tls   ReconfigureTLS   Successful   4m20s
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracleopsrequest -n demo qdops-add-tls qdops-rotate-tls qdops-remove-tls
kubectl delete oracle -n demo oracle-sample
kubectl delete issuer -n demo oracle-issuer
kubectl delete ns demo
```