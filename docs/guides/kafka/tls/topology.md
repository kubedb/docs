---
title: Kafka Combined TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: kf-tls-topology
    name: Topology Cluster
    parent: kf-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Kafka with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption for Kafka. This tutorial will show you how to use KubeDB to run a Kafka cluster with TLS/SSL encryption.

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
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=kafka/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls kafka-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: kafka-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: kafka-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/tls/kf-issuer.yaml
issuer.cert-manager.io/kafka-ca-issuer created
```

## TLS/SSL encryption in Kafka Topology Cluster

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod-tls
  namespace: demo
spec:
  version: 4.0.0
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: kafka-ca-issuer
  topology:
    broker:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```

### Deploy Kafka Topology Cluster with TLS/SSL

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/tls/kafka-prod-tls.yaml
kafka.kubedb.com/kafka-prod-tls created
```

Now, wait until `kafka-prod-tls created` has status `Ready`. i.e,

```bash
$ watch kubectl get kafka -n demo

Every 2.0s: kubectl get kafka -n demo                                                                                                                          aadee: Fri Sep  6 12:34:51 2024
NAME             TYPE            VERSION   STATUS         AGE
kafka-prod-tls   kubedb.com/v1   3.9.0     Provisioning   17s
kafka-prod-tls   kubedb.com/v1   3.9.0     Provisioning   12s
.
.
kafka-prod-tls   kubedb.com/v1   3.9.0     Ready          2m1s
```

### Verify TLS/SSL in Kafka Topology Cluster

```bash
$ kubectl describe secret kafka-prod-tls-client-cert -n demo

Name:         kafka-prod-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=kafka-prod-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=kafkas.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
                *.kafka-prod-tls-pods.demo.svc.cluster.local,kafka-prod-tls-pods,kafka-prod-tls-pods.demo.svc,kafka-prod-tls-pods.demo.svc.cluster.local,l...
              cert-manager.io/certificate-name: kafka-prod-tls-client-cert
              cert-manager.io/common-name: kafka-prod-tls-pods.demo.svc
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: kafka-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:          1184 bytes
keystore.jks:    3254 bytes
tls.crt:         1460 bytes
tls.key:         1708 bytes
truststore.jks:  891 bytes
```

Now, Let's exec into a kafka broker pod and verify the configuration that the TLS is enabled.

```bash
$ kubectl exec -it -n demo kafka-prod-tls-broker-0 -- kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep 'ssl.keystore'
  ssl.keystore.certificate.chain=null sensitive=true synonyms={}
  ssl.keystore.key=null sensitive=true synonyms={}
  ssl.keystore.location=/var/private/ssl/server.keystore.jks sensitive=false synonyms={STATIC_BROKER_CONFIG:ssl.keystore.location=/var/private/ssl/server.keystore.jks}
  ssl.keystore.password=null sensitive=true synonyms={STATIC_BROKER_CONFIG:ssl.keystore.password=null}
  ssl.keystore.type=JKS sensitive=false synonyms={DEFAULT_CONFIG:ssl.keystore.type=JKS}
  zookeeper.ssl.keystore.location=null sensitive=false synonyms={}
  zookeeper.ssl.keystore.password=null sensitive=true synonyms={}
  zookeeper.ssl.keystore.type=null sensitive=false synonyms={}
  ssl.keystore.certificate.chain=null sensitive=true synonyms={}
  ssl.keystore.key=null sensitive=true synonyms={}
  ssl.keystore.location=/var/private/ssl/server.keystore.jks sensitive=false synonyms={STATIC_BROKER_CONFIG:ssl.keystore.location=/var/private/ssl/server.keystore.jks}
  ssl.keystore.password=null sensitive=true synonyms={STATIC_BROKER_CONFIG:ssl.keystore.password=null}
  ssl.keystore.type=JKS sensitive=false synonyms={DEFAULT_CONFIG:ssl.keystore.type=JKS}
  zookeeper.ssl.keystore.location=null sensitive=false synonyms={}
  zookeeper.ssl.keystore.password=null sensitive=true synonyms={}
  zookeeper.ssl.keystore.type=null sensitive=false synonyms={}
```

We can see from the above output that, keystore location is `/var/private/ssl/server.keystore.jks` which means that TLS is enabled.

You will find a file named `clientauth.properties` in the config directory. This file is generated by the operator which contains necessary authentication/authorization/certificate configurations that are required during connect to the Kafka cluster.

```bash
root@kafka-prod-broker-tls-0:~# cat config/clientauth.properties
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="*************";
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
ssl.truststore.location=/var/private/ssl/server.truststore.jks
ssl.truststore.password=***********
```

Now, let's exec into the kafka pod and connect using this configuration to verify the TLS is enabled.

```bash
$ kubectl exec -it -n demo kafka-prod-broker-tls-0 -- bash
kafka@kafka-prod-broker-tls-0:~$ kafka-metadata-quorum.sh --command-config config/clientauth.properties --bootstrap-server localhost:9092 describe --status
ClusterId:              11ef-921c-f2a07f85765w
LeaderId:               1001
LeaderEpoch:            17
HighWatermark:          390
MaxFollowerLag:         0
MaxFollowerLagTimeMs:   18
CurrentVoters:          [1000,1001]
CurrentObservers:       [0,1]
```

From the above output, we can see that we are able to connect to the Kafka cluster using the TLS configuration.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kafka -n demo kafka-prod-tls
kubectl delete issuer -n demo kafka-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Monitor your Kafka cluster with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Monitor your Kafka cluster with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/kafka/monitoring/using-builtin-prometheus.md).
- Use [kubedb cli](/docs/guides/kafka/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
