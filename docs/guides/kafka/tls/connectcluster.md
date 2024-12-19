---
title: Kafka ConnectCluster TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: kf-tls-connectcluster
    name: ConnectCluster
    parent: kf-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Kafka ConnectCluster with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption for Kafka ConnectCluster. This tutorial will show you how to use KubeDB to run a Kafka ConnectCluster with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/kafka](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in Kafka.

- `spec:`
    - `enableSSL`
    - `tls:`
        - `issuerRef`
        - `certificate`

Read about the fields in details in [kafka concept](/docs/guides/kafka/concepts/kafka.md),

`tls` is applicable for all types of Kafka (i.e., `combined` and `topology`).

Users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.crt`, `tls.crt`, `tls.key`, `keystore.jks` and `truststore.jks`.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Kafka. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=connectcluster/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls connectcluster-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: connectcluster-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: connectcluster-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/tls/connectcluster-issuer.yaml
issuer.cert-manager.io/connectcluster-ca-issuer created
```

## TLS/SSL encryption in Kafka Topology Cluster

> **Note:** Before creating Kafka ConnectCluster, make sure you have a Kafka cluster with/without TLS/SSL enabled. If you don't have a Kafka cluster, you can follow the steps [here](/docs/guides/kafka/tls/topology.md).

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: ConnectCluster
metadata:
  name: connectcluster-distributed
  namespace: demo
spec:
  version: 3.9.0
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: connectcluster-ca-issuer
  replicas: 3
  connectorPlugins:
    - postgres-3.0.5.final
    - jdbc-3.0.5.final
  kafkaRef:
    name: kafka-prod-tls
    namespace: demo
  deletionPolicy: WipeOut
```

Here,
- `spec.enableSSL` is set to `true` to enable TLS/SSL encryption.
- `spec.tls.issuerRef` refers to the `Issuer` that we have created in the previous step.
- `spec.kafkaRef` refers to the Kafka cluster that we have created from [here](/docs/guides/kafka/tls/topology.md).

### Deploy Kafka ConnectCluster with TLS/SSL

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/tls/connectcluster-tls.yaml
connectcluster.kafka.kubedb.com/connectcluster-tls created
```

Now, wait until `connectcluster-tls created` has status `Ready`. i.e,

```bash
$ watch kubectl get connectcluster -n demo

Every 2.0s: kubectl get connectcluster -n demo                                                                                                                 aadee: Fri Sep  6 14:59:32 2024

NAME                 TYPE                        VERSION   STATUS         AGE
connectcluster-tls   kafka.kubedb.com/v1alpha1   3.9.0     Provisioning   0s
connectcluster-tls   kafka.kubedb.com/v1alpha1   3.9.0     Provisioning   34s
.
.
connectcluster-tls   kafka.kubedb.com/v1alpha1   3.9.0     Ready          2m
```

### Verify TLS/SSL in Kafka ConnectCluster

```bash
$ kubectl describe secret -n demo connectcluster-tls-client-connect-cert 

Name:         connectcluster-tls-client-connect-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=kafka
              app.kubernetes.io/instance=connectcluster-tls
              app.kubernetes.io/managed-by=kafka.kubedb.com
              app.kubernetes.io/name=connectclusters.kafka.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
                *.connectcluster-tls-pods.demo.svc,*.connectcluster-tls-pods.demo.svc.cluster.local,connectcluster-tls,connectcluster-tls-pods.demo.svc,co...
              cert-manager.io/certificate-name: connectcluster-tls-client-connect-cert
              cert-manager.io/common-name: connectcluster-tls-pods.demo.svc
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: connectcluster-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:   1184 bytes
tls.crt:  1566 bytes
tls.key:  1704 bytes
```

Now, Let's exec into a ConnectCluster pod and verify the configuration that the TLS is enabled.

```bash
$ kubectl exec -it connectcluster-tls-0 -n demo -- bash
kafka@connectcluster-tls-0:~$ curl -u "$CONNECT_CLUSTER_USER:$CONNECT_CLUSTER_PASSWORD" http://localhost:8083
curl: (1) Received HTTP/0.9 when not allowed
```

From the above output, we can see that we are unable to connect to the Kafka cluster using the HTTP protocol.

```bash
kafka@connectcluster-tls-0:~$ curl -u "$CONNECT_CLUSTER_USER:$CONNECT_CLUSTER_PASSWORD" https://localhost:8083
curl: (60) SSL certificate problem: unable to get local issuer certificate
More details here: https://curl.se/docs/sslcerts.html

curl failed to verify the legitimacy of the server and therefore could not
establish a secure connection to it. To learn more about this situation and
how to fix it, please visit the web page mentioned above.
```

Here, we can see that we are unable to connect to the Kafka cluster using the HTTPS protocol. This is because the client does not have the CA certificate to verify the server certificate.

```bash
kafka@connectcluster-tls-0:~$ curl --cacert /var/private/ssl/ca.crt -u "$CONNECT_CLUSTER_USER:$CONNECT_CLUSTER_PASSWORD" https://localhost:8083
{"version":"3.9.0","commit":"5e3c2b738d253ff5","kafka_cluster_id":"11ef-8f52-c284f2efe29w"}
```

From the above output, we can see that we are able to connect to the Kafka ConnectCluster using the TLS configuration.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kafka -n demo kafka-prod-tls
kubectl delete connectcluster -n demo connectcluster-tls
kubectl delete issuer -n demo connectcluster-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Monitor your Kafka cluster with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Monitor your Kafka cluster with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/kafka/monitoring/using-builtin-prometheus.md).
- Use [kubedb cli](/docs/guides/kafka/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
