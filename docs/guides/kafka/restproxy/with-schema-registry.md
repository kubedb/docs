---
title: Rest Proxy With Schema Registry
menu:
  docs_{{ .version }}:
    identifier: kf-rest-proxy-with-schema-registry
    name: With SchemaRegistry
    parent: kf-rest-proxy-guides
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RestProxy with Schema Registry

This tutorial will show you how to use KubeDB to run a [Rest Proxy](https://www.karapace.io/quickstart) with Schema Registry.

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

> We have designed this tutorial to demonstrate a production setup of KubeDB managed RestProxy with Schema Registry. If you just want to try out KubeDB, you can bypass some safety features following the tips [here](/docs/guides/kafka/restproxy/overview.md#tips-for-testing).

## Find Available RestProxy Versions

When you install the KubeDB operator, it registers a CRD named [SchemaRegistryVersion](/docs/guides/kafka/concepts/schemaregistryversion.md). RestProxy uses SchemaRegistryVersions which distribution is `Aiven` to create a RestProxy instance. The installation process comes with a set of tested SchemaRegistryVersion objects. Let's check available SchemaRegistryVersions by,

```bash
$ kubectl get ksrversion

NAME           VERSION   DISTRIBUTION   REGISTRY_IMAGE                                     DEPRECATED   AGE
2.5.11.final   2.5.11    Apicurio       apicurio/apicurio-registry-kafkasql:2.5.11.Final                3d
3.15.0         3.15.0    Aiven          ghcr.io/aiven-open/karapace:3.15.0                              3d
```

> **Note**: Currently RestProxy is supported only for Aiven distribution. Use version with distribution `Aiven` to create Kafka Rest Proxy.

Notice the `DEPRECATED` column. Here, `true` means that this SchemaRegistryVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated KafkaVersion. You can also use the short from `ksrversion` to check available SchemaRegistryVersion.

In this tutorial, we will use `3.15.0` SchemaRegistryVersion CR to create a Kafka Rest Proxy.

## Create a Kafka RestProxy with SchemaRegistry

There are two ways to use RestProxy with SchemaRegistry.

- **Using RestProxy Internal SchemaRegistry**: We have used `Aiven` distribution of RestProxy here. It comes with an internal SchemaRegistry. You can use this internal SchemaRegistry to store and manage Avro/Json schemas and also validation sending messages through RestProxy.
- **Using External SchemaRegistry**: You can also use an external SchemaRegistry with RestProxy. You can use run `Apicurio` distribution SchemaRegistry from [here](/docs/guides/kafka/schemaregistry/overview.md). Then use the SchemaRegistry reference in the RestProxy CR.

### Create a RestProxy CR with Internal SchemaRegistry
The KubeDB operator implements a RestProxy CRD to define the specification of Restproxy.

The RestProxy instance used for this tutorial:

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: RestProxy
metadata:
  name: restproxy-interal-sr
  namespace: demo
spec:
  version: 3.15.0
  replicas: 2
  schemaRegistryRef:
    internallyManaged: true
  kafkaRef:
    name: kafka-quickstart
    namespace: demo
  deletionPolicy: WipeOut
```

Here,

- `spec.version` - is the name of the SchemaRegistryVersion CR. Here, a SchemaRegistry of version `3.15.0` will be created.
- `spec.replicas` - specifies the number of rest proxy instances to run. Here, the RestProxy will run with 2 replicas.
- `spec.kafkaRef` specifies the Kafka instance that the RestProxy will connect to. Here, the RestProxy will connect to the Kafka instance named `kafka-quickstart` in the `demo` namespace. It is an appbinding reference of the Kafka instance.
- `spec.schemaRegistryRef` - specifies the SchemaRegistry instance that the RestProxy will use. Here, the RestProxy will use an internal SchemaRegistry. `internallyManaged: true` means that the RestProxy will use an internal SchemaRegistry. If you want to use an external SchemaRegistry, you can use `schemaRegistryRef` field to specify the SchemaRegistry instance.
- `spec.deletionPolicy` specifies what KubeDB should do when a user try to delete RestProxy CR. Deletion policy `WipeOut` will delete all the instances, secret when the RestProxy CR is deleted.

Before create RestProxy, you have to deploy a `Kafka` cluster first. To deploy kafka cluster, follow the [Kafka Quickstart](/docs/guides/kafka/quickstart/kafka/index.md) guide. Let's assume `kafka-quickstart` is already deployed using KubeDB.
Let's create the RestProxy CR that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/restproxy/restproxy-interal-sr.yaml
restproxy.kafka.kubedb.com/restproxy-interal-sr created
```

The RestProxy's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the RestProxy.

```bash
$ kubectl get restproxy -n demo -w
NAME                        TYPE                        VERSION   KAFKA             STATUS         AGE
restproxy-internal-sr       kafka.kubedb.com/v1alpha1   3.15.0    kafka-quickstart  Provisioning   2s
restproxy-internal-sr       kafka.kubedb.com/v1alpha1   3.15.0    kafka-quickstart  Provisioning   4s
.
.
restproxy-internal-sr       kafka.kubedb.com/v1alpha1   3.15.0    kafka-quickstart  Ready          112s
```

Describe the `RestProxy` object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe restproxy -n demo restproxy-interal-sr
Name:         restproxy-interal-sr
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
  Schema Registry Ref:
    Internally Managed:  true
  Version:         3.15.0
Status:
  Conditions:
    Last Transition Time:  2024-09-02T06:27:36Z
    Message:               The KubeDB operator has started the provisioning of RestProxy: demo/restproxy-interal-sr
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
    Message:               The RestProxy: demo/restproxy-interal-sr is accepting client requests
    Observed Generation:   1
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-09-02T06:28:29Z
    Message:               The RestProxy: demo/restproxy-interal-sr is ready.
    Observed Generation:   1
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-09-02T06:28:30Z
    Message:               The RestProxy: demo/restproxy-interal-sr is successfully provisioned.
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
$ kubectl get all,secret,petset -n demo -l 'app.kubernetes.io/instance=restproxy-interal-sr'
NAME                         READY   STATUS    RESTARTS   AGE
pod/restproxy-interal-sr-0   1/1     Running   0          117s
pod/restproxy-interal-sr-1   1/1     Running   0          79s

NAME                                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/restproxy-interal-sr        ClusterIP   10.96.117.46   <none>        8082/TCP   119s
service/restproxy-interal-sr-pods   ClusterIP   None           <none>        8082/TCP   119s

NAME                                 TYPE     DATA   AGE
secret/restproxy-interal-sr-config   Opaque   1      119s

NAME                                                AGE
petset.apps.k8s.appscode.com/restproxy-interal-sr   117s
```

- `PetSet` - a PetSet named after the RestProxy instance.
- `Services` -  For a RestProxy instance headless service is created with name `{RestProxy-name}-{pods}` and a primary service created with name `{RestProxy-name}`.
- `Secrets` - default configuration secrets are generated for RestProxy.
    - `{RestProxy-Name}-config` - the default configuration secret created by the operator.

### Create RestProxy CR with External SchemaRegistry

Here, We will create a RestProxy CR with an external SchemaRegistry. We have already deployed a SchemaRegistry `Apicurio` instance named `schemaregistry-qucikstart` in the `demo` namespace. If you have not deployed a SchemaRegistry instance yet, you can follow the [SchemaRegistry Quickstart](/docs/guides/kafka/schemaregistry/overview.md) guide to deploy a SchemaRegistry instance. Let's create a RestProxy CR that uses this external SchemaRegistry.

The RestProxy instance used for this tutorial:

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: RestProxy
metadata:
  name: restproxy-external-sr
  namespace: demo
spec:
  version: 3.15.0
  replicas: 2
  schemaRegistryRef:
    name: schemaregistry-qucikstart
    namespace: demo
  kafkaRef:
    name: kafka-quickstart
    namespace: demo
  deletionPolicy: WipeOut
```

Here,

- `spec.version` - is the name of the SchemaRegistryVersion CR. Here, a SchemaRegistry of version `3.15.0` will be created.
- `spec.replicas` - specifies the number of rest proxy instances to run. Here, the RestProxy will run with 2 replicas.
- `spec.kafkaRef` specifies the Kafka instance that the RestProxy will connect to. Here, the RestProxy will connect to the Kafka instance named `kafka-quickstart` in the `demo` namespace. It is an appbinding reference of the Kafka instance.
- `spec.schemaRegistryRef` - specifies the SchemaRegistry instance that the RestProxy will use. Here, the RestProxy will use an external SchemaRegistry. `name` and `namespace` fields are used to specify the SchemaRegistry instances appbinding reference.
- `spec.deletionPolicy` specifies what KubeDB should do when a user try to delete RestProxy CR. Deletion policy `WipeOut` will delete all the instances, secret when the RestProxy CR is deleted.

Before create RestProxy, you have to deploy a `Kafka` cluster first. To deploy kafka cluster, follow the [Kafka Quickstart](/docs/guides/kafka/quickstart/kafka/index.md) guide. Let's assume `kafka-quickstart` is already deployed using KubeDB.
Let's create the RestProxy CR that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/restproxy/restproxy-external-sr.yaml
restproxy.kafka.kubedb.com/restproxy-external-sr created
```

The RestProxy's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the RestProxy.

```bash
$ kubectl get restproxy -n demo -w
NAME                        TYPE                        VERSION   KAFKA             STATUS         AGE
restproxy-external-sr       kafka.kubedb.com/v1alpha1   3.15.0    kafka-quickstart  Provisioning   2s
restproxy-external-sr       kafka.kubedb.com/v1alpha1   3.15.0    kafka-quickstart  Provisioning   4s
.
.
restproxy-external-sr       kafka.kubedb.com/v1alpha1   3.15.0    kafka-quickstart  Ready          112s
```

Describe the `RestProxy` object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe restproxy -n demo restproxy-external-sr 
Name:         restproxy-external-sr
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kafka.kubedb.com/v1alpha1
Kind:         RestProxy
Metadata:
  Creation Timestamp:  2025-03-27T08:06:00Z
  Finalizers:
    kafka.kubedb.com/finalizer
  Generation:        1
  Resource Version:  70832
  UID:               3fe0e135-201e-4faf-ac83-f02de0c92bf0
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
  Replicas:        1
  Schema Registry Ref:
    Name:       schemaregistry-quickstart
    Namespace:  demo
  Version:      3.15.0
Status:
  Conditions:
    Last Transition Time:  2025-03-27T08:06:00Z
    Message:               The KubeDB operator has started the provisioning of RestProxy: demo/restproxy-external-sr
    Observed Generation:   1
    Reason:                RestProxyProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2025-03-27T08:06:02Z
    Message:               All desired replicas are ready.
    Observed Generation:   1
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2025-03-27T08:06:13Z
    Message:               The RestProxy: demo/restproxy-external-sr is accepting client requests
    Observed Generation:   1
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2025-03-27T08:06:13Z
    Message:               The RestProxy: demo/restproxy-external-sr is ready.
    Observed Generation:   1
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2025-03-27T08:06:14Z
    Message:               The RestProxy: demo/restproxy-external-sr is successfully provisioned.
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
$ kubectl get all,secret,petset -n demo -l 'app.kubernetes.io/instance=restproxy-external-sr'
NAME                          READY   STATUS    RESTARTS   AGE
pod/restproxy-external-sr-0   1/1     Running   0          24s

NAME                                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/restproxy-external-sr        ClusterIP   10.96.87.193   <none>        8082/TCP   25s
service/restproxy-external-sr-pods   ClusterIP   None           <none>        8082/TCP   25s

NAME                                  TYPE     DATA   AGE
secret/restproxy-external-sr-config   Opaque   1      25s

NAME                                                 AGE
petset.apps.k8s.appscode.com/restproxy-external-sr   24s
```

- `PetSet` - a PetSet named after the RestProxy instance.
- `Services` -  For a RestProxy instance headless service is created with name `{RestProxy-name}-{pods}` and a primary service created with name `{RestProxy-name}`.
- `Secrets` - default configuration secrets are generated for RestProxy.
  - `{RestProxy-Name}-config` - the default configuration secret created by the operator.

### Create schema

We are going to create schema first for message validation. There are many ways to create schema. Here, we are going to create schema using `curl` command.

* If you are using internal SchemaRegistry, port forward the `restproxy-internal-sr` service to port `8082`, and export `SCHEMA_BASE_URL` like below:

```bash
$ kubectl port-forward svc/restproxy-interal-sr 8082:8082 -n demo
Forwarding from 127.0.0.1:8082 -> 8082
Forwarding from [::1]:8082 -> 8082
Handling connection for 8082
```
```bash
$ export SCHEMA_BASE_URL=http://localhost:8082
```

* If you are using external SchemaRegistry, port forward the `schemaregistry-quickstart` service to port `8080`, and export `SCHEMA_BASE_URL` like below:

```bash
$ kubectl port-forward svc/schemaregistry-quickstart 8080:8080 -n demo
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
Handling connection for 8080
````

```bash
$ export SCHEMA_BASE_URL=http://localhost:8080/apis/ccompat/v7
```

Schema:

```json
{
  "type": "record",
  "name": "User",
  "fields": [
    {"name": "id", "type": "int"},
    {"name": "name", "type": "string"},
    {"name": "email", "type": ["null", "string"], "default": null}
  ]
}
```

Create schema:

```bash
curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
  --data '{"schema": "{\"type\":\"record\",\"name\":\"User\",\"fields\":[{\"name\":\"id\",\"type\":\"int\"},{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"email\",\"type\":[\"null\",\"string\"],\"default\":null}]}"}' \
  $SCHEMA_BASE_URL/subjects/rest-demo-value/versions
  {"id":1}
```
Schema is created successfully and the response with schema id `1`.

### Accessing Kafka using Rest Proxy with Schema Registry

You can access `Kafka` using the REST API. The RestProxy REST API is available at port `8082` of the Rest Proxy service.

To access the RestProxy REST API, you can use `kubectl port-forward` command to forward the port to your local machine.

```bash
$ kubectl port-forward svc/restproxy-interal-sr 8082:8082 -n demo
Forwarding from 127.0.0.1:8082 -> 8082
Forwarding from [::1]:8082 -> 8082
```

In another terminal, you can use `curl` to list topics, produce and consume messages from the Kafka cluster.

List topics:

```bash
$ curl localhost:8082/topics | jq
[
  "rest-demo",
  "kafka-health",
  "_restproxy_schemas",
  "__consumer_offsets",
  "kafkasql-journal"
]
```

#### Produce a message to a topic `rest-demo`(replace `rest-demo` with your topic name):

> Note: The topic must be created in the Kafka cluster before producing messages.

```bash
$ curl -X POST http://localhost:8082/topics/rest-demo \
             -H "Content-Type: application/vnd.kafka.avro.v2+json" \
             -d '{
           "value_schema_id": 1,
           "records": [
             {"value": {"id": 1, "name": "Alice", "email": {"string": "alice@example.com"}}},
             {"value": {"id": 2, "name": "Bob", "email": {"string": "bob@example.com"}}},
             {"value": {"id": 3, "name": "Charlie", "email": {"string": "charlie@example.com"}}}
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
      "offset": 0,
      "partition": 1
    },
    {
      "offset": 1,
      "partition": 0
    }
  ],
  "value_schema_id": 1
}
```
Here, we have produced 3 messages to the topic `rest-demo` and used the value schema id `1` that we created earlier.

#### Consume messages from a topic `order_notification`(replace `order_notification` with your topic name):

To consume messages from a Kafka topic using the Kafka REST Proxy, you'll need to perform the following steps:

Create a Consumer Instance

```bash
$  curl -X POST http://localhost:8082/consumers/my_consumer_group \
          -H "Content-Type: application/vnd.kafka.v2+json" \
          -d '{
        "name": "my_consumer",
        "format": "avro",
        "auto.offset.reset": "earliest"
      }' | jq
{
  "base_uri": "http://restproxy-interal-sr-0:8082/consumers/my_consumer_group/instances/my_consumer",
  "instance_id": "my_consumer"
}
```

Subscribe the Consumer to a Topic

```bash
$ curl -X POST http://localhost:8082/consumers/my_consumer_group/instances/my_consumer/subscription \
          -H "Content-Type: application/vnd.kafka.v2+json" \
          -d '{
        "topics": ["rest-demo"]
      }'
```

Consume Messages

```bash
$ curl -X GET -H "Accept: application/vnd.kafka.avro.v2+json" http://localhost:8082/consumers/my_consumer_group/instances/my_consumer/records | jq
  
[
  {
    "key": null,
    "offset": 0,
    "partition": 1,
    "timestamp": 1743065191936,
    "topic": "rest-demo",
    "value": {
      "email": "bob@example.com",
      "id": 2,
      "name": "Bob"
    }
  },
  {
    "key": null,
    "offset": 0,
    "partition": 0,
    "timestamp": 1743065191936,
    "topic": "rest-demo",
    "value": {
      "email": "alice@example.com",
      "id": 1,
      "name": "Alice"
    }
  },
  {
    "key": null,
    "offset": 1,
    "partition": 0,
    "timestamp": 1743065191936,
    "topic": "rest-demo",
    "value": {
      "email": "charlie@example.com",
      "id": 3,
      "name": "Charlie"
    }
  }
]
```

Delete the Consumer Instance

```bash
$ curl -X DELETE -H "Content-Type: application/vnd.kafka.v2+json" http://localhost:8082/consumers/my_consumer_group/instances/my_consumer
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo restproxy restproxy-interal-sr -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
restproxy.kafka.kubedb.com/restproxy-interal-sr patched

$ kubectl delete krp restproxy-interal-sr,restproxy-external-sr  -n demo
restproxy.kafka.kubedb.com "restproxy-interal-sr,restproxy-external-sr" deleted

$ kubectl delete ksr schemaregistry-quickstart -n demo
schemaregistry.kafka.kubedb.com "schemaregistry-quickstart" deleted

$ kubectl delete kafka kafka-quickstart -n demo
kafka.kubedb.com "kafka-quickstart" deleted

$  kubectl delete namespace demo
namespace "demo" deleted
```

## Next Steps

- [Quickstart Kafka](/docs/guides/kafka/quickstart/kafka/index.md) with KubeDB Operator.
- [Quickstart ConnectCluster](/docs/guides/kafka/connectcluster/quickstart.md) with KubeDB Operator.
- Use [kubedb cli](/docs/guides/kafka/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [ConnectCluster object](/docs/guides/kafka/concepts/connectcluster.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
