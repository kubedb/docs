---
title: Kafka Quickstart
menu:
docs_{{ .version }}:
identifier: kf-quickstart-quickstart
name: Overview
parent: es-quickstart-kafka
weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Kafka QuickStart

This tutorial will show you how to use KubeDB to run an [Apache Kafka](https://kafka.apache.org/).

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/kafka/images/Kafka-CRD-Lifecycle.png">
</p>

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install the KubeDB operator in your cluster following the steps [here](/docs/setup/install/_index.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [guides/kafka/quickstart/overview/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/kafka/quickstart/overview/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> We have designed this tutorial to demonstrate a production setup of KubeDB managed Apache Kafka. If you just want to try out KubeDB, you can bypass some safety features following the tips [here](/docs/guides/kafka/quickstart/overview/index.md#tips-for-testing).

## Find Available StorageClass

We will have to provide `StorageClass` in Kafka CRD specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  14h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Find Available KafkaVersion

When you install the KubeDB operator, it registers a CRD named [KafkaVersion](/docs/guides/kafka/concepts/catalog/index.md). The installation process comes with a set of tested KafkaVersion objects. Let's check available KafkaVersions by,

```bash
NAME    VERSION   DB_IMAGE                   DEPRECATED   AGE
3.3.0   3.3.0     kubedb/kafka-kraft:3.3.0                6d
```

Notice the `DEPRECATED` column. Here, `true` means that this KafkaVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated KafkaVersion. You can also use the short from `kfversion` to check available KafkaVersions.

In this tutorial, we will use `3.3.0` KafkaVersion CR to create a Kafka cluster.

## Create a Kafka Cluster

The KubeDB operator implements a Kafka CRD to define the specification of Kafka.

The Kafka instance used for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-quickstart
  namespace: demo
spec:
  replicas: 3
  version: 3.3.0
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: DoNotTerminate
```

Here,

- `spec.version` - is the name of the KafkaVersion CR. Here, a Kafka of version `3.3.0` will be created.
- `spec.replicas` - specifies the number of Kafka brokers.
- `spec.storageType` - specifies the type of storage that will be used for Kafka. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the Kafka using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this Kafka instance. This storage spec will be passed to the StatefulSet created by the KubeDB operator to run Kafka pods. You can specify any StorageClass available in your cluster with appropriate resource requests. If you don't specify `spec.storageType: Ephemeral`, then this field is required.
- `spec.terminationPolicy` specifies what KubeDB should do when a user try to delete Kafka CR. Termination policy `Delete` will delete the database pods, secret and PVC when the Kafka CR is deleted.

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in the `storage.resources.requests` field. Don't specify `limits` here. PVC does not get resized automatically.

Let's create the Kafka CR that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/Kafka/quickstart/overview/yamls/kafka.yaml
kafka.kubedb.com/kafka-quickstart created
```

The Kafka's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the Kafka.

```bash
$ kubectl get kafka -n demo -w
NAME               TYPE                  VERSION   STATUS   AGE
kafka-quickstart   kubedb.com/v1alpha2   3.3.0     Provisioning   2s
kafka-quickstart   kubedb.com/v1alpha2   3.3.0     Provisioning   4s
.
.
kafka-quickstart   kubedb.com/v1alpha2   3.3.0     Ready          112s

```

Describe the kafka object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe kafka -n demo kafka-quickstart
Name:         kafka-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Kafka
Metadata:
  Creation Timestamp:  2023-01-04T10:13:12Z
  Finalizers:
    kubedb.com
  Generation:  2
  Managed Fields:
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:healthChecker:
          .:
          f:failureThreshold:
          f:periodSeconds:
          f:timeoutSeconds:
        f:replicas:
        f:storage:
          .:
          f:accessModes:
          f:resources:
            .:
            f:requests:
              .:
              f:storage:
          f:storageClassName:
        f:storageType:
        f:terminationPolicy:
        f:version:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2023-01-04T10:13:12Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:finalizers:
          .:
          v:"kubedb.com":
      f:spec:
        f:authSecret:
    Manager:      kubedb-provisioner
    Operation:    Update
    Time:         2023-01-04T10:13:12Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:phase:
    Manager:         kubedb-provisioner
    Operation:       Update
    Subresource:     status
    Time:            2023-01-04T10:13:14Z
  Resource Version:  192231
  UID:               8a1eb48b-75f3-4b3d-b8ff-0634780a9f09
Spec:
  Auth Secret:
    Name:  kafka-quickstart-admin-cred
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     20
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
  Replicas:        3
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Termination Policy:    DoNotTerminate
  Version:               3.3.0
Status:
  Conditions:
    Last Transition Time:  2023-01-04T10:13:14Z
    Message:               The KubeDB operator has started the provisioning of Kafka: demo/kafka-quickstart
    Observed Generation:   2
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2023-01-04T10:13:20Z
    Message:               All desired replicas are ready.
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2023-01-04T10:13:52Z
    Message:               The Kafka: demo/kafka-quickstart is accepting client requests
    Observed Generation:   2
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2023-01-04T10:15:00Z
    Message:               The Kafka: demo/kafka-quickstart is ready.
    Observed Generation:   2
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2023-01-04T10:15:02Z
    Message:               The Kafka: demo/kafka-quickstart is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>

```

### KubeDB Operator Generated Resources

On deployment of a Kafka CR, the operator creates the following resources:

```bash
$ kubectl get all,secret -n demo -l 'app.kubernetes.io/instance=kafka-quickstart'
NAME                     READY   STATUS    RESTARTS   AGE
pod/kafka-quickstart-0   1/1     Running   0          8m50s
pod/kafka-quickstart-1   1/1     Running   0          8m48s
pod/kafka-quickstart-2   1/1     Running   0          8m46s

NAME                            TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                       AGE
service/kafka-quickstart-pods   ClusterIP   None         <none>        9092/TCP,9093/TCP,29092/TCP   8m52s

NAME                                READY   AGE
statefulset.apps/kafka-quickstart   3/3     8m50s

NAME                                                  TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/kafka-quickstart   kubedb.com/kafka   3.3.0     8m50s

NAME                                 TYPE                       DATA   AGE
secret/kafka-quickstart-admin-cred   kubernetes.io/basic-auth   2      8m52s
secret/kafka-quickstart-config       Opaque                     2      8m52s
```

- `StatefulSet` - a StatefulSet named after the Kafka instance. In topology mode, the operator creates 3 statefulSets with name `{Kafka-Name}-{Sufix}`.
- `Services` -  3 services are generated for each Kafka database.
    - `{Kafka-Name}` - the client service which is used to connect to the database. It points to the `ingest` nodes.
    - `{Kafka-Name}-master` - the master service which is used to connect to the master nodes. It is a headless service.
    - `{Kafka-Name}-pods` - the node discovery service which is used by the Elasticsearch nodes to communicate each other. It is a headless service.
- `AppBinding` - an [AppBinding](/docs/guides/kafka/concepts/appbinding/index.md) which hold to connect information for the Kafka brokers. It is also named after the Kafka instance.
- `Secrets` - 3 types of secrets are generated for each Elasticsearch database.
    - `{Kafka-Name}-{username}-cred` - the auth secrets which hold the `username` and `password` for the Elasticsearch users.
    - `{Kafka-Name}-{alias}-cert` - the certificate secrets which hold `tls.crt`, `tls.key`, and `ca.crt` for configuring the Kafka instance.
    - `{Kafka-Name}-config` - the default configuration secret created by the operator.
dsxcz
## Connect with Kafka

We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our Elasticsearch database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our Elasticsearch database is working well.

Let's port-forward the port `9200` to local machine:

```bash
$ kubectl port-forward -n demo svc/es-quickstart 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, our Elasticsearch cluster is accessible at `localhost:9200`.

**Connection information:**

- Address: `localhost:9200`
- Username:

  ```bash
  $ kubectl get secret -n demo es-quickstart-elastic-cred -o jsonpath='{.data.username}' | base64 -d
  elastic
  ```

- Password:

  ```bash
  $ kubectl get secret -n demo es-quickstart-elastic-cred -o jsonpath='{.data.password}' | base64 -d
  vIHoIfHn=!Z8F4gP
  ```

Now let's check the health of our Elasticsearch database.

```bash
$ curl -XGET -k -u 'elastic:vIHoIfHn=!Z8F4gP' "https://localhost:9200/_cluster/health?pretty"

{
  "cluster_name" : "es-quickstart",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 3,
  "active_shards" : 6,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}
```

From the health information above, we can see that our Elasticsearch cluster's status is `green` which means the cluster is healthy.

## Halt Elasticsearch

KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` termination policy. If admission webhook is enabled, it prevents the user from deleting the database as long as the `spec.terminationPolicy` is set `DoNotTerminate`.

To halt the database, we have to set `spec.terminationPolicy:` to `Halt` by updating it,

```bash
$ kubectl edit elasticsearch -n demo es-quickstart

>> spec:
>>   terminationPolicy: Halt
```

Now, if you delete the Elasticsearch object, the KubeDB operator will delete every resource created for this Elasticsearch CR, but leaves the auth secrets, and PVCs.

```bash
$  kubectl delete elasticsearch -n demo es-quickstart 
elasticsearch.kubedb.com "es-quickstart" deleted
```

Check resources:

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=es-quickstart'
NAME                                               TYPE                       DATA   AGE
secret/es-quickstart-apm-system-cred               kubernetes.io/basic-auth   2      5m39s
secret/es-quickstart-beats-system-cred             kubernetes.io/basic-auth   2      5m39s
secret/es-quickstart-elastic-cred                  kubernetes.io/basic-auth   2      5m39s
secret/es-quickstart-kibana-system-cred            kubernetes.io/basic-auth   2      5m39s
secret/es-quickstart-logstash-system-cred          kubernetes.io/basic-auth   2      5m39s
secret/es-quickstart-remote-monitoring-user-cred   kubernetes.io/basic-auth   2      5m39s

NAME                                         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-es-quickstart-0   Bound    pvc-5b657e2a-6c32-4631-bac9-eefebbcb129a   1Gi        RWO            standard       5m29s
persistentvolumeclaim/data-es-quickstart-1   Bound    pvc-e44d7ab8-fc2b-4cfe-9bef-74f2a2d875f5   1Gi        RWO            standard       5m23s
persistentvolumeclaim/data-es-quickstart-2   Bound    pvc-dad75b1b-37ed-4318-a82a-5e38f04d36bc   1Gi        RWO            standard       5m18s

```

## Resume Elasticsearch

Say, the Elasticsearch CR was deleted with `spec.terminationPolicy` to `Halt` and you want to re-create the Elasticsearch cluster using the existing auth secrets and the PVCs.

You can do it by simpily re-deploying the original Elasticsearch object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/quickstart/overview/yamls/elasticsearch.yaml
elasticsearch.kubedb.com/es-quickstart created
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo elasticsearch es-quickstart -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
elasticsearch.kubedb.com/es-quickstart patched

$ kubectl delete -n demo es/quick-elasticsearch
elasticsearch.kubedb.com "es-quickstart" deleted

$  kubectl delete namespace demo
namespace "demo" deleted
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for the production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if the database pod fails. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purposes, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `terminationPolicy: WipeOut`**. It is nice to be able to resume the database from the previous one. So, we preserve all your `PVCs` and auth `Secrets`. If you don't want to resume the database, you can just use `spec.terminationPolicy: WipeOut`. It will clean up every resouce that was created with the Elasticsearch CR. For more details, please visit [here](/docs/guides/elasticsearch/concepts/elasticsearch/index.md#specterminationpolicy).

## Next Steps

- Learn about [backup & restore](/docs/guides/elasticsearch/backup/overview/index.md) Elasticsearch database using Stash.
- Learn how to configure [Elasticsearch Topology Cluster](/docs/guides/elasticsearch/clustering/topology-cluster/simple-dedicated-cluster/index.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
