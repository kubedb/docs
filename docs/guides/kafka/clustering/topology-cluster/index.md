---
title: Kafka Topology Cluster
menu:
  docs_{{ .version }}:
    identifier: kf-topology-cluster
    name: Topology Cluster
    parent: kf-clustering
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Topology Cluster

A Kafka topology cluster is a comprised of two groups of kafka nodes (eg. pods) where one group of nodes are assigned to controller role which manages cluster metadata & participates in leader election. Outer group of nodes are assigned to dedicated broker roles that only act as Kafka broker for message publishing and subscribing. Topology mode clustering is suitable for production deployment.

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

## Create Topology Kafka Cluster

Here, we are going to create a TLS secured Kafka topology cluster in Kraft mode.

### Create Issuer/ ClusterIssuer

At first, make sure you have cert-manager installed on your k8s for enabling TLS. KubeDB operator uses cert manager to inject certificates into kubernetes secret & uses them for secure `SASL` encrypted communication among kafka brokers and controllers. We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Kafka. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you CA certificates using openssl.

```bash
openssl req -newkey rsa:2048 -keyout ca.key -nodes -x509 -days 3650 -out ca.crt
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

### Provision TLS secure Kafka

For this demo, we are going to provision kafka version `3.9.0` with 3 controllers and 3 brokers. To learn more about Kafka CR, visit [here](/docs/guides/kafka/concepts/kafka.md). visit [here](/docs/guides/kafka/concepts/kafkaversion.md) to learn more about KafkaVersion CR.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 4.0.0
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: kafka-ca-issuer
      kind: Issuer
  topology:
    broker:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 3
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/clustering/kf-topology.yaml
kafka.kubedb.com/kafka-prod created
```

Watch the bootstrap progress:

```bash
$ kubectl get kf -n demo -w
NAME         TYPE                  VERSION   STATUS         AGE
kafka-prod   kubedb.com/v1alpha2   3.9.0     Provisioning   6s
kafka-prod   kubedb.com/v1alpha2   3.9.0     Provisioning   14s
kafka-prod   kubedb.com/v1alpha2   3.9.0     Provisioning   50s
kafka-prod   kubedb.com/v1alpha2   3.9.0     Ready          68s
```

Hence, the cluster is ready to use.
Let's check the k8s resources created by the operator on the deployment of Kafka CRO:

```bash
$ kubectl get all,petset,secret,pvc -n demo -l 'app.kubernetes.io/instance=kafka-prod'
NAME                          READY   STATUS    RESTARTS        AGE
pod/kafka-prod-broker-0       1/1     Running   0               4m10s
pod/kafka-prod-broker-1       1/1     Running   0               4m4s
pod/kafka-prod-broker-2       1/1     Running   0               3m57s
pod/kafka-prod-controller-0   1/1     Running   0               4m8s
pod/kafka-prod-controller-1   1/1     Running   0               4m
pod/kafka-prod-controller-2   1/1     Running   0               3m53s

NAME                      TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                       AGE
service/kafka-prod-pods   ClusterIP   None         <none>        9092/TCP,9093/TCP,29092/TCP   4m14s

NAME                                     READY   AGE
petset.apps.k8s.appscode.com/kafka-prod-broker       3/3     4m10s
petset.apps.k8s.appscode.com/kafka-prod-controller   3/3     4m8s

NAME                                            TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/kafka-prod   kubedb.com/kafka   3.9.0     4m8s

NAME                                  TYPE                       DATA   AGE
secret/kafka-prod-auth                kubernetes.io/basic-auth   2      4m14s
secret/kafka-prod-broker-config       Opaque                     3      4m14s
secret/kafka-prod-client-cert         kubernetes.io/tls          3      4m14s
secret/kafka-prod-controller-config   Opaque                     3      4m10s
secret/kafka-prod-keystore-cred       Opaque                     3      4m14s
secret/kafka-prod-server-cert         kubernetes.io/tls          5      4m14s

NAME                                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/kafka-prod-data-kafka-prod-broker-0       Bound    pvc-1ce9bf24-8d2d-4cae-9453-28df9f52ac44   1Gi        RWO            standard       4m10s
persistentvolumeclaim/kafka-prod-data-kafka-prod-broker-1       Bound    pvc-5e2dc46b-0947-4de1-adb0-307fc881b2ba   1Gi        RWO            standard       4m4s
persistentvolumeclaim/kafka-prod-data-kafka-prod-broker-2       Bound    pvc-b7ef2986-db7d-4089-95a2-474cd14c6282   1Gi        RWO            standard       3m57s
persistentvolumeclaim/kafka-prod-data-kafka-prod-controller-0   Bound    pvc-8e3ae399-fb87-4906-91d8-3f5a09014d2a   1Gi        RWO            standard       4m8s
persistentvolumeclaim/kafka-prod-data-kafka-prod-controller-1   Bound    pvc-faf53264-e125-430a-9a73-c2c73da1b97e   1Gi        RWO            standard       4m
persistentvolumeclaim/kafka-prod-data-kafka-prod-controller-2   Bound    pvc-d962a03b-7af7-41ba-9d53-044e8ffa03f2   1Gi        RWO            standard       3m53s
```

## Publish & Consume messages with Kafka

We will create a Kafka topic using `kafka-topics.sh` script which is provided by kafka container itself. We will use `kafka console producer` and `kafka console consumer` as clients for publishing messages to the topic and then consume those messages. Exec into one of the kafka broker pods in interactive mode first.

```bash
$ kubectl exec -it -n demo  kafka-prod-broker-0 -- bash
root@kafka-prod-broker-0:~# pwd
/opt/kafka
```

You will find a file named `clientauth.properties` in the config directory. This file is generated by the operator which contains necessary authentication/authorization configurations that are required during publishing or subscribing messages to a kafka topic.

```bash
root@kafka-prod-broker-0:~# cat config/clientauth.properties
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="*************";
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
ssl.truststore.location=/var/private/ssl/server.truststore.jks
ssl.truststore.password=***********
```

Now, we have to use a bootstrap server to perform operations in a kafka broker. For this demo, we are going to use the http endpoint of the headless service `kafka-prod-broker` as bootstrap server for publishing & consuming messages to kafka brokers. These endpoints are pointing to all the kafka broker pods. We will set an environment variable for the `clientauth.properties` filepath as well. At first, describe the service to get the http endpoints.

```bash
$ kubectl describe svc -n demo kafka-prod-pods
Name:              kafka-prod-pods
Namespace:         demo
Labels:            app.kubernetes.io/component=database
                   app.kubernetes.io/instance=kafka-prod
                   app.kubernetes.io/managed-by=kubedb.com
                   app.kubernetes.io/name=kafkas.kubedb.com
Annotations:       <none>
Selector:          app.kubernetes.io/instance=kafka-prod,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                None
IPs:               None
Port:              broker  9092/TCP
TargetPort:        broker/TCP
Endpoints:         10.244.0.33:9092,10.244.0.37:9092,10.244.0.41:9092
Port:              controller  9093/TCP
TargetPort:        controller/TCP
Endpoints:         10.244.0.16:9093,10.244.0.20:9093,10.244.0.24:9093
Port:              local  29092/TCP
TargetPort:        local/TCP
Endpoints:         10.244.0.33:29092,10.244.0.37:29092,10.244.0.41:29092
Session Affinity:  None
Events:            <none>
```

Use the `http endpoints` and `clientauth.properties` file to set environment variables. These environment variables will be useful for handling console command operations easily.

```bash
root@kafka-prod-broker-0:~# export SERVER=" 10.244.0.33:9092,10.244.0.37:9092,10.244.0.41:9092"
root@kafka-prod-broker-0:~# export CLIENTAUTHCONFIG="$HOME/config/clientauth.properties"
```

Let's describe the broker metadata for the quorum.

```bash
root@kafka-prod-broker-0:~# kafka-metadata-quorum.sh --command-config $CLIENTAUTHCONFIG --bootstrap-server localhost:9092 describe --status
ClusterId:              11ed-82ed-2a2abab96b3w
LeaderId:               2
LeaderEpoch:            15
HighWatermark:          1820
MaxFollowerLag:         0
MaxFollowerLagTimeMs:   159
CurrentVoters:          [1000,1001,1002]
CurrentObservers:       [0,1,2]
```

It will show you important metadata information like clusterID, current leader ID, broker IDs which are participating in leader election voting and IDs of those brokers who are observers. It is important to mention that each broker is assigned a numeric ID which is called its broker ID. The ID is assigned sequentially with respect to the host pod name. In this case, The pods assigned broker IDs are as follows:

| Pods                | Broker ID | 
|---------------------|:---------:|
| kafka-prod-broker-0 |     0     |
| kafka-prod-broker-1 |     1     |
| kafka-prod-broker-2 |     2     |

Let's create a topic named `sample` with 1 partitions and a replication factor of 1. Describe the topic once it's created. You will see the leader ID for each partition and their replica IDs along with in-sync-replicas(ISR).

```bash
root@kafka-prod-broker-0:~# kafka-topics.sh --command-config $CLIENTAUTHCONFIG --create --topic sample --partitions 1 --replication-factor 1 --bootstrap-server localhost:9092
Created topic sample.


root@kafka-prod-broker-0:~# kafka-topics.sh --command-config $CLIENTAUTHCONFIG --describe --topic sample --bootstrap-server localhost:9092
Topic: sample	TopicId: mqlupmBhQj6OQxxG9m51CA	PartitionCount: 1	ReplicationFactor: 1	Configs: segment.bytes=1073741824
	Topic: sample	Partition: 0	Leader: 1	Replicas: 1	Isr: 1
```

Now, we are going to start a producer and a consumer for topic `sample` using console. Let's use this current terminal for producing messages and open a new terminal for consuming messages. Let's set the environment variables for bootstrap server and the configuration file in consumer terminal also.

From the topic description we can see that the leader partition for partition 0 is 1 that is `kafka-prod-broker-1`. If we produce messages to `kafka-prod-broker-1` broker(brokerID=1) it will store those messages in partition 0. Let's produce messages in the producer terminal and consume them from the consumer terminal.

```bash
root@kafka-prod-broker-1:~# kafka-console-producer.sh --producer.config $CLIENTAUTHCONFIG  --topic sample --request-required-acks all --bootstrap-server localhost:9092
>hello
>hi
>this is a message from console producer client
>I hope it's received by console consumer
>
```

```bash
root@kafka-prod-broker-1:~# kafka-console-consumer.sh --consumer.config $CLIENTAUTHCONFIG --topic sample --from-beginning --bootstrap-server localhost:9092 --partition 0
hello
hi
this is a message from console producer client
I hope it's received by console consumer
```

Notice that, messages are coming to the consumer as you continue sending messages via producer. So, we have created a kafka topic and used kafka console producer and consumer to test message publishing and consuming successfully.

## Cleaning Up

TO clean up the k8s resources created by this tutorial, run:

```bash
# standalone cluster
$ kubectl patch -n demo kf kafka-prod -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete kf -n demo kafka-prod

# multinode cluster
$ kubectl patch -n demo kf kafka-prod -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete kf -n demo kafka-prod

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