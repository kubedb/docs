---
title: Configuring Kafka Combined Cluster
menu:
  docs_{{ .version }}:
    identifier: kf-configuration-combined-cluster
    name: Combined Cluster
    parent: kf-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure Kafka Combined Cluster

In Kafka combined cluster, every node can perform as broker and controller nodes simultaneously. In this tutorial, we will see how to configure a combined cluster.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install the KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka/configuration/
) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will have to provide `StorageClass` in Kafka CR specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  1h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Use Custom Configuration

Say we want to change the default log retention time and default replication factor of creating a topic. Let's create the `server.properties` file with our desire configurations.

**server.properties:**

```properties
log.retention.hours=100
default.replication.factor=2
```

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: configsecret-combined
  namespace: demo
stringData:
  server.properties: |-
    log.retention.hours=100
    default.replication.factor=2
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/configuration/configsecret-combined.yaml
secret/configsecret-combined created
```

Now that the config secret is created, it needs to be mentioned in the [Kafka](/docs/guides/kafka/concepts/kafka.md) object's yaml:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-dev
  namespace: demo
spec:
  replicas: 2
  version: 3.9.0
  configuration:
    secretName: configsecret-combined
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

Now, create the Kafka object by the following command:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/configuration/kafka-combined.yaml
kafka.kubedb.com/kafka-dev created
```

Now, wait for the Kafka to become ready:

```bash
$ kubectl get kf -n demo -w
NAME         TYPE            VERSION   STATUS         AGE
kafka-dev    kubedb.com/v1   3.9.0     Provisioning   0s
kafka-dev    kubedb.com/v1   3.9.0     Provisioning   24s
.
.
kafka-dev    kubedb.com/v1   3.9.0     Ready          92s
```

## Verify Configuration

Lets exec into one of the kafka pod that we have created and check the configurations are applied or not:

Exec into the Kafka pod:

```bash
$ kubectl exec -it -n demo kafka-dev-0 -- bash
kafka@kafka-dev-0:~$ 
```

Now, execute the following commands to see the configurations:
```bash
kafka@kafka-dev-0:~$ kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep log.retention.hours
  log.retention.hours=100 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=100, DEFAULT_CONFIG:log.retention.hours=168}
  log.retention.hours=100 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=100, DEFAULT_CONFIG:log.retention.hours=168}
kafka@kafka-dev-0:~$ kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep default.replication.factor
  default.replication.factor=2 sensitive=false synonyms={STATIC_BROKER_CONFIG:default.replication.factor=2, DEFAULT_CONFIG:default.replication.factor=1}
  default.replication.factor=2 sensitive=false synonyms={STATIC_BROKER_CONFIG:default.replication.factor=2, DEFAULT_CONFIG:default.replication.factor=1}
```
Here, we can see that our given configuration is applied to the Kafka cluster for all brokers.

## Cleanup

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete kf -n demo kafka-dev 
$ kubectl delete secret -n demo configsecret-combined 
$ kubectl delete namespace demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

