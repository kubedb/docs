---
title: Rest Proxy Overview
menu:
  docs_{{ .version }}:
    identifier: kf-rest-proxy-guides-overview
    name: Overview
    parent: kf-rest-proxy-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RestProxy QuickStart

This tutorial will show you how to use KubeDB to run a [Rest Proxy](https://www.karapace.io/quickstart).

<p align="center">
  <img alt="lifecycle"  src="/docs/images/kafka/restproxy/restproxy-crd-lifecycle.png">
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

> Note: YAML files used in this tutorial are stored in [examples/kafka/restproxy/](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka/restproxy) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> We have designed this tutorial to demonstrate a production setup of KubeDB managed Schema Registry. If you just want to try out KubeDB, you can bypass some safety features following the tips [here](/docs/guides/kafka/quickstart/connectcluster/index.md#tips-for-testing).

## Find Available RestProxy Versions

When you install the KubeDB operator, it registers a CRD named [SchemaRegistryVersion](/docs/guides/kafka/concepts/schemaregistryversion.md). RestProxy uses SchemaRegistryVersions which distribution is `Aiven` to create a RestProxy instance. The installation process comes with a set of tested SchemaRegistryVersion objects. Let's check available SchemaRegistryVersions by,

```bash
$ kubectl get ksrversion

NAME    VERSION   DB_IMAGE                                    DEPRECATED   AGE
NAME           VERSION   DISTRIBUTION   REGISTRY_IMAGE                                     DEPRECATED   AGE
2.5.11.final   2.5.11    Apicurio       apicurio/apicurio-registry-kafkasql:2.5.11.Final                3d
3.15.0         3.15.0    Aiven          ghcr.io/aiven-open/karapace:3.15.0                              3d
```

> **Note**: Currently Schema  is supported only for Apicurio distribution. Use distribution `Apicurio` to create Schema Registry.

Notice the `DEPRECATED` column. Here, `true` means that this SchemaRegistryVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated KafkaVersion. You can also use the short from `ksrversion` to check available SchemaRegistryVersion.

In this tutorial, we will use `3.15.0` SchemaRegistryVersion CR to create a Kafka Rest Proxy.

## Create a Kafka RestProxy

The KubeDB operator implements a RestProxy CRD to define the specification of SchemaRegistry.

The RestProxy instance used for this tutorial:

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: RestProxy
metadata:
  name: restproxy-quickstart
  namespace: demo
spec:
  version: 3.15.0
  replicas: 2
  kafkaRef:
    name: kafka-quickstart
    namespace: demo
  deletionPolicy: WipeOut
```

Here,

- `spec.version` - is the name of the SchemaRegistryVersion CR. Here, a SchemaRegistry of version `3.15.0` will be created.
- `spec.replicas` - specifies the number of rest proxy instances to run. Here, the RestProxy will run with 2 replicas.
- `spec.kafkaRef` specifies the Kafka instance that the RestProxy will connect to. Here, the RestProxy will connect to the Kafka instance named `kafka-quickstart` in the `demo` namespace. It is an appbinding reference of the Kafka instance.
- `spec.deletionPolicy` specifies what KubeDB should do when a user try to delete RestProxy CR. Deletion policy `WipeOut` will delete all the instances, secret when the RestProxy CR is deleted.

Before create RestProxy, you have to deploy a `Kafka` cluster first. To deploy kafka cluster, follow the [Kafka Quickstart](/docs/guides/kafka/quickstart/kafka/index.md) guide. Let's assume `kafka-quickstart` is already deployed using KubeDB.
Let's create the RestProxy CR that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/restproxy/restproxy-quickstart.yaml
restproxy.kafka.kubedb.com/restproxy-quickstart created
```

The RestProxy's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the RestProxy.

```bash
$ kubectl get restproxy -n demo -w
NAME                        TYPE                        VERSION   STATUS         AGE
restproxy-quickstart        kafka.kubedb.com/v1alpha1   3.6.1     Provisioning   2s
restproxy-quickstart        kafka.kubedb.com/v1alpha1   3.6.1     Provisioning   4s
.
.
restproxy-quickstart        kafka.kubedb.com/v1alpha1   3.6.1     Ready          112s
```

Describe the `RestProxy` object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe restproxy -n demo restproxy-quickstart
Name:         restproxy-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kafka.kubedb.com/v1alpha1
Kind:         RestProxy
Metadata:
  Creation Timestamp:  2024-09-02T06:27:36Z
  Finalizers:
    kafka.kubedb.com/finalizer
  Generation:        1
  Resource Version:  179508
  UID:               5defcf67-015d-4f15-a8ef-661717258f76
Spec:
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     10
    Timeout Seconds:    10
  Kafka Ref:
    Name:       kafka-quickstart
    Namespace:  demo
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  rest-proxy
        Resources:
          Limits:
            Memory:  1Gi
          Requests:
            Cpu:     500m
            Memory:  1Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Non Root:  true
          Run As User:      1001
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:  1001
  Replicas:        2
  Version:         3.15.0
Status:
  Conditions:
    Last Transition Time:  2024-09-02T06:27:36Z
    Message:               The KubeDB operator has started the provisioning of RestProxy: demo/restproxy-quickstart
    Observed Generation:   1
    Reason:                RestProxyProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-09-02T06:28:17Z
    Message:               All desired replicas are ready.
    Observed Generation:   1
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-09-02T06:28:29Z
    Message:               The RestProxy: demo/restproxy-quickstart is accepting client requests
    Observed Generation:   1
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-09-02T06:28:29Z
    Message:               The RestProxy: demo/restproxy-quickstart is ready.
    Observed Generation:   1
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-09-02T06:28:30Z
    Message:               The RestProxy: demo/restproxy-quickstart is successfully provisioned.
    Observed Generation:   1
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```

### KubeDB Operator Generated Resources

On deployment of a RestProxy CR, the operator creates the following resources:

```bash
$ kubectl get all,secret,petset -n demo -l 'app.kubernetes.io/instance=restproxy-quickstart'
NAME                         READY   STATUS    RESTARTS   AGE
pod/restproxy-quickstart-0   1/1     Running   0          117s
pod/restproxy-quickstart-1   1/1     Running   0          79s

NAME                                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/restproxy-quickstart        ClusterIP   10.96.117.46   <none>        8082/TCP   119s
service/restproxy-quickstart-pods   ClusterIP   None           <none>        8082/TCP   119s

NAME                                 TYPE     DATA   AGE
secret/restproxy-quickstart-config   Opaque   1      119s

NAME                                                AGE
petset.apps.k8s.appscode.com/restproxy-quickstart   117s
```

- `PetSet` - a PetSet named after the RestProxy instance.
- `Services` -  For a RestProxy instance headless service is created with name `{RestProxy-name}-{pods}` and a primary service created with name `{RestProxy-name}`.
- `Secrets` - default configuration secrets are generated for RestProxy.
    - `{RestProxy-Name}-config` - the default configuration secret created by the operator.

### Accessing Kafka using Rest Proxy

You can access `Kafka` using the REST API. The RestProxy REST API is available at port `8082` of the Rest Proxy service.

To access the RestProxy REST API, you can use `kubectl port-forward` command to forward the port to your local machine.

```bash
$ kubectl port-forward svc/restproxy-quickstart 8082:8082 -n demo
Forwarding from 127.0.0.1:8082 -> 8082
Forwarding from [::1]:8082 -> 8082
```

In another terminal, you can use `curl` to list topics, produce and consume messages from the Kafka cluster.

List topics:

```bash
$ curl localhost:8082/topics | jq
[
  "order_notification",
  "kafka-health",
  "__consumer_offsets",
  "kafkasql-journal"
]

```

#### Produce a message to a topic `order_notification`(replace `order_notification` with your topic name):

> Note: The topic must be created in the Kafka cluster before producing messages.

```bash
curl -X POST http://localhost:8082/topics/order_notification \
     -H "Content-Type: application/vnd.kafka.json.v2+json" \
     -d '{
  "records": [
      {"value": {"orderId": "12345", "status": "Order Placed", "customerName": "Alice Johnson", "totalAmount": 150.75, "timestamp": "2024-08-30T12:34:56Z"}},
      {"value": {"orderId": "12346", "status": "Shipped", "customerName": "Bob Smith", "totalAmount": 249.99, "timestamp": "2024-08-30T12:45:12Z"}},
      {"value": {"orderId": "12347", "status": "Delivered", "customerName": "Charlie Brown", "totalAmount": 89.50, "timestamp": "2024-08-30T13:00:22Z"}}
    ]
  }' | jq
  
{
  "key_schema_id": null,
  "offsets": [
    {
      "offset": 0,
      "partition": 0
    },
    {
      "offset": 1,
      "partition": 0
    },
    {
      "offset": 2,
      "partition": 0
    }
  ],
  "value_schema_id": null
}
```
#### Consume messages from a topic `order_notification`(replace `order_notification` with your topic name):

To consume messages from a Kafka topic using the Kafka REST Proxy, you'll need to perform the following steps:

Create a Consumer Instance

```bash
$ curl -X POST http://localhost:8082/consumers/order_consumer \
  -H "Content-Type: application/vnd.kafka.v2+json" \
  -d '{
    "name": "order_consumer_instance",
    "format": "json",
    "auto.offset.reset": "earliest"
  }' | jq
 
{
  "base_uri": "http://restproxy-quickstart-0:8082/consumers/order_consumer/instances/order_consumer_instance",
  "instance_id": "order_consumer_instance"
}
```

Subscribe the Consumer to a Topic

```bash
$ curl -X POST http://localhost:8082/consumers/order_consumer/instances/order_consumer_instance/subscription \
  -H "Content-Type: application/vnd.kafka.v2+json" \
  -d '{
    "topics": ["order_notification"]
  }'
```

Consume Messages

```bash
$ curl -X GET http://localhost:8082/consumers/order_consumer/instances/order_consumer_instance/records \
  -H "Accept: application/vnd.kafka.json.v2+json" | jq
  
[
  {
    "key": null,
    "offset": 0,
    "partition": 0,
    "timestamp": 1725259256610,
    "topic": "order_notification",
    "value": {
      "customerName": "Alice Johnson",
      "orderId": "12345",
      "status": "Order Placed",
      "timestamp": "2024-08-30T12:34:56Z",
      "totalAmount": 150.75
    }
  },
  {
    "key": null,
    "offset": 1,
    "partition": 0,
    "timestamp": 1725259256610,
    "topic": "order_notification",
    "value": {
      "customerName": "Bob Smith",
      "orderId": "12346",
      "status": "Shipped",
      "timestamp": "2024-08-30T12:45:12Z",
      "totalAmount": 249.99
    }
  },
  {
    "key": null,
    "offset": 2,
    "partition": 0,
    "timestamp": 1725259256610,
    "topic": "order_notification",
    "value": {
      "customerName": "Charlie Brown",
      "orderId": "12347",
      "status": "Delivered",
      "timestamp": "2024-08-30T13:00:22Z",
      "totalAmount": 89.5
    }
  }
]
```

Delete the Consumer Instance

```bash
$ curl -X DELETE http://localhost:8082/consumers/order_consumer/instances/order_consumer_instance
```

You can also list brokers, describe topics and more using the Kafka RestProxy.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo restproxy restproxy-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
restproxy.kafka.kubedb.com/restproxy-quickstart patched

$ kubectl delete krp restproxy-quickstart  -n demo
restproxy.kafka.kubedb.com "restproxy-quickstart" deleted

$ kubectl delete kafka kafka-quickstart -n demo
kafka.kubedb.com "kafka-quickstart" deleted

$  kubectl delete namespace demo
namespace "demo" deleted
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for the production environment. You can follow these tips to avoid them.

1 **Use `deletionPolicy: Delete`**. It is nice to be able to resume the cluster from the previous one. So, we preserve auth `Secrets`. If you don't want to resume the cluster, you can just use `spec.deletionPolicy: WipeOut`. It will clean up every resource that was created with the SchemaRegistry CR. For more details, please visit [here](/docs/guides/kafka/concepts/schemaregistry.md#specdeletionpolicy).

## Next Steps

- [Quickstart Kafka](/docs/guides/kafka/quickstart/kafka/index.md) with KubeDB Operator.
- [Quickstart ConnectCluster](/docs/guides/kafka/quickstart/connectcluster/index.md) with KubeDB Operator.
- Use [kubedb cli](/docs/guides/kafka/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [ConnectCluster object](/docs/guides/kafka/concepts/connectcluster.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
