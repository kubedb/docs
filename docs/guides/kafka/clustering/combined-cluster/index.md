---
title: Kafka Combined Cluster
menu:
  docs_{{ .version }}:
    identifier: kf-combined-cluster
    name: Combined Cluster
    parent: kf-clustering
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Combined Cluster

A Kafka combined cluster is a group of kafka brokers where each broker also acts as a controller and participates in leader election as a voter. Combined mode can be used in development environment, but it should be avoided in critical deployment environments.

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

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka/clustering) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create Standalone Kafka Cluster

Here, we are going to create a standalone (i.e. `replicas: 1`) Kafka cluster in Kraft mode. For this demo, we are going to provision kafka version `3.6.1`. To learn more about Kafka CR, visit [here](/docs/guides/kafka/concepts/kafka.md). visit [here](/docs/guides/kafka/concepts/kafkaversion.md) to learn more about KafkaVersion CR.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-standalone
  namespace: demo
spec:
  replicas: 1
  version: 3.6.1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: DoNotTerminate
```

Let's deploy the above example by the following command:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/clustering/kf-standalone.yaml
kafka.kubedb.com/kafka-standalone created
```

Watch the bootstrap progress:

```bash
$ kubectl get kf -n demo -w
NAME               TYPE                  VERSION   STATUS         AGE
kafka-standalone   kubedb.com/v1alpha2   3.6.1     Provisioning   8s
kafka-standalone   kubedb.com/v1alpha2   3.6.1     Provisioning   14s
kafka-standalone   kubedb.com/v1alpha2   3.6.1     Provisioning   35s
kafka-standalone   kubedb.com/v1alpha2   3.6.1     Provisioning   35s
kafka-standalone   kubedb.com/v1alpha2   3.6.1     Provisioning   36s
kafka-standalone   kubedb.com/v1alpha2   3.6.1     Ready          41s
```

Hence, the cluster is ready to use.
Let's check the k8s resources created by the operator on the deployment of Kafka CRO:

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=kafka-standalone'
NAME                     READY   STATUS    RESTARTS   AGE
pod/kafka-standalone-0   1/1     Running   0          8m56s

NAME                            TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                       AGE
service/kafka-standalone-pods   ClusterIP   None         <none>        9092/TCP,9093/TCP,29092/TCP   8m59s

NAME                                READY   AGE
statefulset.apps/kafka-standalone   1/1     8m56s

NAME                                                  TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/kafka-standalone   kubedb.com/kafka   3.6.1     8m56s

NAME                                 TYPE                       DATA   AGE
secret/kafka-standalone-admin-cred   kubernetes.io/basic-auth   2      8m59s
secret/kafka-standalone-config       Opaque                     2      8m59s

NAME                                                             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/kafka-standalone-data-kafka-standalone-0   Bound    pvc-56f8284a-249e-4444-ab3d-31e01662a9a0   1Gi        RWO            standard       8m56s
```

## Create Multi-Node Combined Kafka Cluster

Here, we are going to create a multi-node (say `replicas: 3`) Kafka cluster. We will use the KafkaVersion `3.4.0` for this demo. To learn more about kafka CR, visit [here](/docs/guides/kafka/concepts/kafka.md).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-multinode
  namespace: demo
spec:
  replicas: 3
  version: 3.6.1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: DoNotTerminate
```

Let's deploy the above example by the following command:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/clustering/kf-multinode.yaml
kafka.kubedb.com/kafka-multinode created
```

Watch the bootstrap progress:

```bash
$ kubectl get kf -n demo -w
kafka-multinode   kubedb.com/v1alpha2   3.6.1     Provisioning   9s
kafka-multinode   kubedb.com/v1alpha2   3.6.1     Provisioning   14s
kafka-multinode   kubedb.com/v1alpha2   3.6.1     Provisioning   18s
kafka-multinode   kubedb.com/v1alpha2   3.6.1     Provisioning   2m6s
kafka-multinode   kubedb.com/v1alpha2   3.6.1     Provisioning   2m8s
kafka-multinode   kubedb.com/v1alpha2   3.6.1     Ready          2m14s
```

Hence, the cluster is ready to use.
Let's check the k8s resources created by the operator on the deployment of Kafka CRO:

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=kafka-multinode'
NAME                    READY   STATUS    RESTARTS   AGE
pod/kafka-multinode-0   1/1     Running   0          6m2s
pod/kafka-multinode-1   1/1     Running   0          5m56s
pod/kafka-multinode-2   1/1     Running   0          5m51s

NAME                           TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                       AGE
service/kafka-multinode-pods   ClusterIP   None         <none>        9092/TCP,9093/TCP,29092/TCP   6m7s

NAME                               READY   AGE
statefulset.apps/kafka-multinode   3/3     6m2s

NAME                                                 TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/kafka-multinode   kubedb.com/kafka   3.6.1     6m2s

NAME                                TYPE                       DATA   AGE
secret/kafka-multinode-admin-cred   kubernetes.io/basic-auth   2      6m7s
secret/kafka-multinode-config       Opaque                     2      6m7s

NAME                                                           STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/kafka-multinode-data-kafka-multinode-0   Bound    pvc-15cc2329-15ba-4781-8b7f-f0fe6cf81614   1Gi        RWO            standard       6m2s
persistentvolumeclaim/kafka-multinode-data-kafka-multinode-1   Bound    pvc-bc3773cc-dff0-458c-b71a-7ef6aa877549   1Gi        RWO            standard       5m56s
persistentvolumeclaim/kafka-multinode-data-kafka-multinode-2   Bound    pvc-e4829946-b2bb-473e-84d9-c5f9c360f3f0   1Gi        RWO            standard       5m51s
```

## Publish & Consume messages with Kafka

We will create a Kafka topic using `kafka-topics.sh` script which is provided by kafka container itself. We will use `kafka console producer` and `kafka console consumer` as clients for publishing messages to the topic and then consume those messages. Exec into one of the kafka brokers in interactive mode first.

```bash
$ kubectl exec -it -n demo  kafka-multinode-0 -- bash
root@kafka-multinode-0:~# pwd
/opt/kafka
```

You will find a file named `clientauth.properties` in the config directory. This file is generated by the operator which contains necessary authentication/authorization configurations that are required during publishing or subscribing messages to a kafka topic.

```bash
root@kafka-multinode-0:~# cat config/clientauth.properties
security.protocol=SASL_PLAINTEXT
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="************";
```

Now, we have to use a bootstrap server to perform operations in a kafka broker. For this demo, we are going to use the http endpoint of the headless service `kafka-multinode-pods` as bootstrap server for publishing & consuming messages to kafka brokers. These endpoints are pointing to all the kafka broker pods. We will set an environment variable for the `clientauth.properties` filepath as well. At first, describe the service to get the http endpoints.

```bash
$ kubectl describe svc -n demo kafka-multinode-pods
Name:              kafka-multinode-pods
Namespace:         demo
Labels:            app.kubernetes.io/component=database
                   app.kubernetes.io/instance=kafka-multinode
                   app.kubernetes.io/managed-by=kubedb.com
                   app.kubernetes.io/name=kafkas.kubedb.com
Annotations:       <none>
Selector:          app.kubernetes.io/instance=kafka-multinode,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                None
IPs:               None
Port:              http  9092/TCP
TargetPort:        http/TCP
Endpoints:         10.244.0.69:9092,10.244.0.71:9092,10.244.0.73:9092
Port:              controller  9093/TCP
TargetPort:        controller/TCP
Endpoints:         10.244.0.69:9093,10.244.0.71:9093,10.244.0.73:9093
Port:              internal  29092/TCP
TargetPort:        internal/TCP
Endpoints:         10.244.0.69:29092,10.244.0.71:29092,10.244.0.73:29092
Session Affinity:  None
Events:            <none>
```

Use the `http endpoints` and `clientauth.properties` file to set environment variables. These environment variables will be useful for handling console command operations easily.

```bash
root@kafka-multinode-0:~# export SERVER="10.244.0.69:9092,10.244.0.71:9092,10.244.0.73:9092"
root@kafka-multinode-0:~# export CLIENTAUTHCONFIG="$HOME/config/clientauth.properties"
```

Let's describe the broker metadata for the quorum.

```bash
root@kafka-multinode-0:~# kafka-metadata-quorum.sh --command-config $CLIENTAUTHCONFIG --bootstrap-server $SERVER describe --status
ClusterId:              11ed-957c-625c6a5f47bw
LeaderId:               0
LeaderEpoch:            19
HighWatermark:          2601
MaxFollowerLag:         0
MaxFollowerLagTimeMs:   0
CurrentVoters:          [0,1,2]
CurrentObservers:       []
```

It will show you important metadata information like clusterID, current leader ID, broker IDs which are participating in leader election voting and IDs of those brokers who are observers. It is important to mention that each broker is assigned a numeric ID which is called its broker ID. The ID is assigned sequentially with respect to the host pod name. In this case, The pods assigned broker IDs are as follows:

| Pods              | Broker ID | 
|-------------------|:---------:|
| kafka-multinode-0 |     0     |
| kafka-multinode-1 |     1     |
| kafka-multinode-2 |     2     |

Let's create a topic named `sample` with 1 partitions and a replication factor of 1. Describe the topic once it's created. You will see the leader ID for each partition and their replica IDs along with in-sync-replicas(ISR).

```bash
root@kafka-multinode-0:~# kafka-topics.sh --command-config $CLIENTAUTHCONFIG --create --topic sample --partitions 1 --replication-factor 1 --bootstrap-server $SERVER
Created topic sample.

root@kafka-multinode-0:~# kafka-topics.sh --command-config $CLIENTAUTHCONFIG --describe --topic sample --bootstrap-server $SERVER
Topic: sample	TopicId: KVpw_JXfRjaeUHfoXLPBvQ	PartitionCount: 1	ReplicationFactor: 1	Configs: segment.bytes=1073741824
	Topic: sample	Partition: 0	Leader: 0	Replicas: 0	Isr: 0
```

Now, we are going to start a producer and a consumer for topic `sample` using console. Let's use this current terminal for producing messages and open a new terminal for consuming messages. Let's set the environment variables for bootstrap server and the configuration file in consumer terminal also.

From the topic description we can see that the leader partition for partition 0 is 0 (the broker that we are on). If we produce messages to `kafka-multinode-0` broker(brokerID=0) it will store those messages in partition 0. Let's produce messages in the producer terminal and consume them from the consumer terminal.

```bash
root@kafka-quickstart-0:~#  kafka-console-producer.sh --producer.config $CLIENTAUTHCONFIG  --topic sample --request-required-acks all --bootstrap-server $SERVER
>message one
>message two
>message three
>
```

```bash
root@kafka-quickstart-0:/# kafka-console-consumer.sh --consumer.config $CLIENTAUTHCONFIG --topic sample --from-beginning --bootstrap-server $SERVER --partition 0
message one
message two
message three

```

Notice that, messages are coming to the consumer as you continue sending messages via producer. So, we have created a kafka topic and used kafka console producer and consumer to test message publishing and consuming successfully.


## Cleaning Up

TO clean up the k8s resources created by this tutorial, run:

```bash
# standalone cluster
$ kubectl patch -n demo kf kafka-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete kf -n demo kafka-standalone

# multinode cluster
$ kubectl patch -n demo kf kafka-multinode -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete kf -n demo kafka-multinode

# delete namespace
$ kubectl delete namespace demo
```

## Next Steps

- Deploy [dedicated topology cluster](/docs/guides/kafka/clustering/topology-cluster/index.md) for Apache Kafka
- Monitor your Kafka cluster with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Detail concepts of [KafkaVersion object](/docs/guides/kafka/concepts/kafkaversion.md).
- Learn to use KubeDB managed Kafka objects using [CLIs](/docs/guides/kafka/cli/cli.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).