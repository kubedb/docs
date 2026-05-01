---
title: Configure TLS in Oracle
menu:
  docs_{{ .version }}:
    identifier: oracle-tls-configure
    name: Configure TLS
    parent: oracle-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS/SSL in Oracle

`KubeDB` provides support for TLS/SSL encryption for `Oracle`. This tutorial will show you how to use `KubeDB` to deploy a `Oracle` database with TLS/SSL configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manager`](https://cert-manager.io/docs/installation/) v1.4.0 or later to your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [TLS Overview](/docs/guides/oracle/tls/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/oracle/tls/configure/yamls](/docs/guides/oracle/tls/configure/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Deploy Oracle database with TLS/SSL configuration

As a pre-requisite, we are going to create an Issuer/ClusterIssuer. This Issuer/ClusterIssuer is used to create certificates. Then we are going to deploy a Oracle with TLS/SSL configuration.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=oracle/O=kubedb"
```

- Create a secret using the certificate files we have just generated:

```bash
$ kubectl create secret tls oracle-ca --cert=ca.crt --key=ca.key --namespace=demo
secret/oracle-ca created
```

Now, we are going to create an `Issuer` using the `oracle-ca` secret that contains the CA certificate we have just created. Below is the YAML of the `Issuer` CR that we are going to create:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: oracle-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: oracle-ca
```

Let's create the `Issuer` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/tls/configure/yamls/issuer.yaml
issuer.cert-manager.io/oracle-ca-issuer created
```

### Deploy Oracle cluster with TLS/SSL configuration

Here, our issuer `oracle-ca-issuer` is ready to deploy a `Oracle` cluster with TLS/SSL configuration. Below is the YAML for the Oracle cluster that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-tls
  namespace: demo
spec:
  version: "21.3.0"
  replicas: 3
  tcpsConfig:
    tls:
      issuerRef:
        apiGroup: cert-manager.io
        name: oracle-ca-issuer
        kind: Issuer
      certificates:
      - alias: server
        subject:
          organizations:
          - kubedb:server
        dnsNames:
        - localhost
        ipAddresses:
        - "127.0.0.1"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Here,

- `spec.tcpsConfig.tls.issuerRef` refers to the `oracle-ca-issuer` issuer.
- `spec.tcpsConfig.tls.certificates` provides options to configure certificate renewal and keep-alive. You can find more details from [here](/docs/guides/oracle/concepts/oracle.md#tls).

**Deploy Oracle Cluster:**

Let's create the `Oracle` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/tls/configure/yamls/tls-oracle.yaml
oracle.kubedb.com/oracle-tls created
```

**Wait for the database to be ready:**

Now, watch `Oracle` going to `Running` state and also watch `PetSet` and its pods going to `Running` state:

```bash
$ watch kubectl get oracle -n demo oracle-tls
NAME          VERSION   STATUS   AGE
oracle-tls    1.17.0    Ready    62s

$ watch -n 3 kubectl get petset -n demo oracle-tls
NAME          READY   AGE
oracle-tls    3/3     2m30s

$ watch -n 3 kubectl get pod -n demo -l app.kubernetes.io/instance=oracle-tls
NAME              READY   STATUS    RESTARTS   AGE
oracle-tls-0      1/1     Running   0          3m5s
oracle-tls-1      1/1     Running   0          2m40s
oracle-tls-2      1/1     Running   0          2m20s
```

**Verify TLS/SSL configuration:**

Now, let's verify the TLS/SSL configuration by checking the secrets created for the Oracle database:

```bash
$ kubectl get secrets -n demo | grep oracle-tls
oracle-tls-server-cert   kubernetes.io/tls   3      3m
oracle-tls-client-cert   kubernetes.io/tls   3      3m
```

The TLS certificates have been created and the Oracle cluster is now configured to use TLS/SSL for both client connections and peer-to-peer communication.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracle -n demo oracle-tls
kubectl delete issuer -n demo oracle-ca-issuer
kubectl delete ns demo
```