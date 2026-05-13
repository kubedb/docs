---
title: Reconfigure TLS of Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-reconfigure-tls-cluster
    name: Cluster
    parent: qdrant-reconfigure-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Qdrant TLS/SSL (Transport Encryption)

KubeDB supports reconfiguring TLS/SSL certificates for Qdrant — adding, removing, updating, and rotating certificates via a `QdrantOpsRequest`. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `cert-manager` v1.0.0 or later to your cluster to manage your SSL/TLS certificates from [here](https://cert-manager.io/docs/installation/).

- Now, install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)
  - [TLS/SSL Overview](/docs/guides/qdrant/tls/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/qdrant/reconfigure-tls/yamls](/docs/guides/qdrant/reconfigure-tls/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Add TLS to a Qdrant database

Here, we are going to create a Qdrant database without TLS and then reconfigure the database to use TLS.

### Deploy Qdrant without TLS

In this section, we are going to deploy a Qdrant cluster without TLS. Below is the YAML of the `Qdrant` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Qdrant` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/reconfigure-tls/yamls/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

Now, wait until `qdrant-sample` has status `Ready`:

```bash
$ kubectl get qdrant -n demo
NAME             VERSION   STATUS   AGE
qdrant-sample    1.17.0    Ready    3m22s
```

### Create Issuer

Now, we are going to create an example `Issuer` that will be used to enable SSL/TLS in Qdrant. Alternatively you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

1. Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=qdrant/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
```

2. Create a secret using the certificate files we have just generated,

```bash
$ kubectl create secret tls qdrant-ca --cert=ca.crt  --key=ca.key --namespace=demo
secret/qdrant-ca created
```

3. Now we are going to create an `Issuer` using the `qdrant-ca` secret that contains the CA certificate we have just created:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: qdrant-issuer
  namespace: demo
spec:
  ca:
    secretName: qdrant-ca
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/reconfigure-tls/yamls/issuer.yaml
issuer.cert-manager.io/qdrant-issuer created
```

### Add TLS

Now, we are going to create a `QdrantOpsRequest` to add TLS to the running database.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: qdrant-sample
  tls:
    issuerRef:
      name: qdrant-issuer
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
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `qdrant-sample` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer to use for signing certificates.
- `spec.tls.certificates` specifies the certificate configuration.
- `spec.timeout` specifies the timeout for the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#spectimeout)).
- `spec.apply` specifies when to apply the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#specapply)).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/reconfigure-tls/yamls/add-tls.yaml
qdrantopsrequest.ops.kubedb.com/qdops-add-tls created
```

Let's wait for `QdrantOpsRequest` to be `Successful`:

```bash
$ watch -n 3 kubectl get QdrantOpsRequest -n demo qdops-add-tls
Every 3.0s: kubectl get QdrantOpsRequest -n demo qdops-add-tls

NAME             TYPE             STATUS       AGE
qdops-add-tls    ReconfigureTLS   Successful   6m30s
```

## Rotate Certificates

Now we are going to rotate the certificates of the database.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: qdrant-sample
  tls:
    rotateCertificates: true
```

Here,

- `spec.tls.rotateCertificates` specifies that we are requesting to rotate the certificates of the `qdrant-sample` database.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/reconfigure-tls/yamls/rotate-tls.yaml
qdrantopsrequest.ops.kubedb.com/qdops-rotate-tls created
```

Let's wait for `QdrantOpsRequest` to be `Successful`:

```bash
$ kubectl get qdrantopsrequest -n demo qdops-rotate-tls
NAME               TYPE             STATUS       AGE
qdops-rotate-tls   ReconfigureTLS   Successful   3m8s
```

## Remove TLS from the Database

In this section, we are going to reconfigure TLS setting of the database by removing the TLS configuration.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: qdrant-sample
  tls:
    remove: true
```

Here,

- `spec.tls.remove` specifies that we are removing the TLS configuration from `qdrant-sample` database.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/reconfigure-tls/yamls/remove-tls.yaml
qdrantopsrequest.ops.kubedb.com/qdops-remove-tls created
```

Let's wait for `QdrantOpsRequest` to be `Successful`:

```bash
$ kubectl get qdrantopsrequest -n demo qdops-remove-tls
NAME               TYPE             STATUS       AGE
qdops-remove-tls   ReconfigureTLS   Successful   4m20s
```

## Next Steps

- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrantopsrequest -n demo qdops-add-tls qdops-rotate-tls qdops-remove-tls
kubectl delete qdrant -n demo qdrant-sample
kubectl delete issuer -n demo qdrant-issuer
kubectl delete ns demo
```