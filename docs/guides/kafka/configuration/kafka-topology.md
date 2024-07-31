---
title: Configuring Kafka Topology Cluster
menu:
  docs_{{ .version }}:
    identifier: kf-configuration-topology-cluster
    name: Topology Cluster
    parent: kf-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure Kafka Topology Cluster

In Kafka topology cluster, broker and controller nodes run separately. In this tutorial, we will see how to configure a topology cluster.

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

Say we want to change the default log retention time and default replication factor of creating a topic of brokers. Let's create the `broker.properties` file with our desire configurations.

**broker.properties:**

```properties
log.retention.hours=100
default.replication.factor=2
```

and we also want to change the metadata.log.dir of the all controller nodes. Let's create the `controller.properties` file with our desire configurations.

**controller.properties:**

```properties
metadata.log.dir=/var/log/kafka/metadata-custom
```

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: configsecret-topology
  namespace: demo
stringData:
  broker.properties: |-
    log.retention.hours=100
    default.replication.factor=2
  controller.properties: |-
    metadata.log.dir=/var/log/kafka/metadata-custom
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/configuration/configsecret-topology.yaml
secret/configsecret-topology created
```

Now that the config secret is created, it needs to be mention in the [Kafka](/docs/guides/kafka/concepts/kafka.md) object's yaml:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 3.6.1
  configSecret:
    name: configsecret-topology
  topology:
    broker:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                requests:
                  cpu: "500m"
                  memory: "1Gi"
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                requests:
                  cpu: "500m"
                  memory: "1Gi"
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/configuration/kafka-topology.yaml
kafka.kubedb.com/kafka-prod created
```

Now, wait for the Kafka to become ready:

```bash
$ kubectl get kf -n demo -w
NAME         TYPE            VERSION   STATUS         AGE
kafka-prod   kubedb.com/v1   3.6.1     Provisioning   5s
kafka-prod   kubedb.com/v1   3.6.1     Provisioning   7s
.
.
kafka-prod   kubedb.com/v1   3.6.1     Ready          2m
```

## Verify Configuration

Let's exec into one of the kafka broker pod that we have created and check the configurations are applied or not:

Exec into the Kafka broker:

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- bash
kafka@kafka-prod-broker-0:~$ 
```

Now, execute the following commands to see the configurations:
```bash
kafka@kafka-prod-broker-0:~$ kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep log.retention.hours
  log.retention.hours=100 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=100, DEFAULT_CONFIG:log.retention.hours=168}
  log.retention.hours=100 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=100, DEFAULT_CONFIG:log.retention.hours=168}
kafka@kafka-prod-broker-0:~$ kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep default.replication.factor
  default.replication.factor=2 sensitive=false synonyms={STATIC_BROKER_CONFIG:default.replication.factor=2, DEFAULT_CONFIG:default.replication.factor=1}
  default.replication.factor=2 sensitive=false synonyms={STATIC_BROKER_CONFIG:default.replication.factor=2, DEFAULT_CONFIG:default.replication.factor=1}
```
Here, we can see that our given configuration is applied to the Kafka cluster for all brokers.

Now, let's exec into one of the kafka controller pod that we have created and check the configurations are applied or not:

Exec into the Kafka controller:

```bash
$ kubectl exec -it -n demo kafka-prod-controller-0 -- bash
kafka@kafka-prod-controller-0:~$ 
```

Now, execute the following commands to see the metadata storage directory:
```bash
kafka@kafka-prod-controller-0:~$ ls /var/log/kafka/
1000  cluster_id  metadata-custom
```

Here, we can see that our given configuration is applied to the controller. Metadata log directory is changed to `/var/log/kafka/metadata-custom`.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

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

