---
title: RabbitMQ TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: rm-tls-describe
    name: Configure TLS
    parent: rm-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run RabbitMQ with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption (via, `.spec.enableSSL`) for RabbitMQ. This tutorial will show you how to use KubeDB to run a RabbitMQ database with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/RabbitMQ](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/RabbitMQ) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in RabbitMQ.

- `spec:`
  - `tls:`
    - `issuerRef`
    - `certificate`
  - `enableSSL`

Read about the fields in details in [RabbitMQ concept](/docs/guides/rabbitmq/concepts/rabbitmq.md),

When, SSLMode is anything other than `disabled`, users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.crt`, `tls.crt` and `tls.key`.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in RabbitMQ. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=rabbitmq/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls rabbitmq-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: rabbitmq-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: rabbitmq-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/RabbitMQ/tls/issuer.yaml
issuer.cert-manager.io/rabbitmq-ca-issuer created
```

## TLS/SSL encryption in RabbitMQ Standalone

Below is the YAML for RabbitMQ Standalone. Here, [`spec.sslMode`](/docs/guides/rabbitmq/concepts/rabbitmq.md#spectls) specifies tls configurations required for operator to create corresponding resources.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq-tls
  namespace: demo
spec:
  version: "3.13.2"
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: rabbitmq-ca-issuer
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

### Deploy RabbitMQ Standalone

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/tls/rm-standalone-ssl.yaml
rabbitmq.kubedb.com/rabbitmq-tls created
```

Now, wait until `rabbitmq-tls created` has status `Ready`. i.e,

```bash
$ watch kubectl get rm -n demo
Every 2.0s: kubectl get rm -n demo
NAME            VERSION     STATUS     AGE
rabbitmq-tls    3.13.2      Ready      14s
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete rabbitmq -n demo rabbitmq-tls
kubectl delete issuer -n demo rabbitmq-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [RabbitMQ object](/docs/guides/rabbitmq/concepts/rabbitmq.md).
(/docs/guides/RabbitMQ/monitoring/using-prometheus-operator.md).
- Monitor your RabbitMQ database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/rabbitmq/monitoring/using-builtin-prometheus.md).
- Detail concepts of [RabbitMQ object](/docs/guides/rabbitmq/concepts/rabbitmq.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
