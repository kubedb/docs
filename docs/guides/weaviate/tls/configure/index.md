---
title: Configure TLS in Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-tls-configure
    name: Configure TLS
    parent: weaviate-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS/SSL in Weaviate

`KubeDB` provides support for TLS/SSL encryption for `Weaviate`. This tutorial will show you how to use `KubeDB` to deploy a `Weaviate` database with TLS/SSL configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manager`](https://cert-manager.io/docs/installation/) v1.4.0 or later to your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [TLS Overview](/docs/guides/weaviate/tls/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/weaviate/tls/configure/yamls](/docs/guides/weaviate/tls/configure/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Deploy Weaviate database with TLS/SSL configuration

As a pre-requisite, we are going to create an Issuer/ClusterIssuer. This Issuer/ClusterIssuer is used to create certificates. Then we are going to deploy a Weaviate with TLS/SSL configuration.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=weaviate/O=kubedb"
```

- Create a secret using the certificate files we have just generated:

```bash
$ kubectl create secret tls weaviate-ca --cert=ca.crt --key=ca.key --namespace=demo
secret/weaviate-ca created
```

Now, we are going to create an `Issuer` using the `weaviate-ca` secret that contains the CA certificate we have just created. Below is the YAML of the `Issuer` CR that we are going to create:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: weaviate-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: weaviate-ca
```

Let's create the `Issuer` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/weaviate/tls/configure/yamls/issuer.yaml
issuer.cert-manager.io/weaviate-ca-issuer created
```

### Deploy Weaviate cluster with TLS/SSL configuration

Here, our issuer `weaviate-ca-issuer` is ready to deploy a `Weaviate` cluster with TLS/SSL configuration. Below is the YAML for the Weaviate cluster that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: "1.33.1"
  replicas: 3
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: weaviate-ca-issuer
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

- `spec.tls.issuerRef` refers to the `weaviate-ca-issuer` issuer.
- `spec.tls.certificates` provides options to configure certificate renewal and keep-alive. You can find more details from [here](/docs/guides/weaviate/concepts/weaviate.md#tls).

**Deploy Weaviate Cluster:**

Let's create the `Weaviate` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/weaviate/tls/configure/yamls/tls-weaviate.yaml
weaviate.kubedb.com/weaviate-sample created
```

**Wait for the database to be ready:**

Now, watch `Weaviate` going to `Running` state and also watch `PetSet` and its pods going to `Running` state:

```bash
$ watch kubectl get weaviate -n demo weaviate-sample
NAME              VERSION   STATUS   AGE
weaviate-sample   1.33.1    Ready    62s

$ watch -n 3 kubectl get petset -n demo weaviate-sample
NAME              READY   AGE
weaviate-sample   3/3     2m30s

$ watch -n 3 kubectl get pod -n demo -l app.kubernetes.io/instance=weaviate-sample
NAME                READY   STATUS    RESTARTS   AGE
weaviate-sample-0   1/1     Running   0          3m5s
weaviate-sample-1   1/1     Running   0          2m40s
weaviate-sample-2   1/1     Running   0          2m20s
```

**Verify TLS/SSL configuration:**

Now, let's verify the TLS/SSL configuration by checking the secrets created for the Weaviate database:

```bash
$ kubectl get secrets -n demo | grep weaviate-sample
weaviate-sample-server-cert   kubernetes.io/tls   3      3m
weaviate-sample-client-cert   kubernetes.io/tls   3      3m
```

The TLS certificates have been created and the Weaviate cluster is now configured to use TLS/SSL for both client connections and peer-to-peer communication.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo weaviate-sample
kubectl delete issuer -n demo weaviate-ca-issuer
kubectl delete ns demo
```
