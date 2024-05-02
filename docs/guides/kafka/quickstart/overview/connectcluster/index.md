---
title: ConnectCluster Quickstart
menu:
  docs_{{ .version }}:
    identifier: kf-kafka-overview-connectcluster
    name: connectcluster
    parent: kf-overview-kafka
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Kafka QuickStart

This tutorial will show you how to use KubeDB to run an [Apache Kafka](https://kafka.apache.org/).

<p align="center">
  <img alt="lifecycle"  src="/docs/images/kafka/connectcluster/connectcluster-crd-lifecycle.png">
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

## Find Available ConnectCluster Versions

When you install the KubeDB operator, it registers a CRD named [KafkaVersion](/docs/guides/kafka/concepts/catalog.md). ConnectCluster Version is using the KafkaVersion CR to define the specification of ConnectCluster. The installation process comes with a set of tested KafkaVersion objects. Let's check available KafkaVersions by,

```bash
NAME    VERSION   DB_IMAGE                                    DEPRECATED   AGE
3.3.2   3.3.2     ghcr.io/appscode-images/kafka-kraft:3.3.2                24m
3.4.1   3.4.1     ghcr.io/appscode-images/kafka-kraft:3.4.1                24m
3.5.1   3.5.1     ghcr.io/appscode-images/kafka-kraft:3.5.1                24m
3.5.2   3.5.2     ghcr.io/appscode-images/kafka-kraft:3.5.2                24m
3.6.0   3.6.0     ghcr.io/appscode-images/kafka-kraft:3.6.0                24m
3.6.1   3.6.1     ghcr.io/appscode-images/kafka-kraft:3.6.1                24m

```

Notice the `DEPRECATED` column. Here, `true` means that this KafkaVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated KafkaVersion. You can also use the short from `kfversion` to check available KafkaVersions.

In this tutorial, we will use `3.6.1` KafkaVersion CR to create a Kafka Connect cluster.

## Find Available KafkaConnector Versions

When you install the KubeDB operator, it registers a CRD named [KafkaVersion](/docs/guides/kafka/concepts/catalog.md). KafkaConnector Version use to load connector-plugins to run ConnectCluster worker node(ex. mongodb-source/sink). The installation process comes with a set of tested KafkaConnectorVersion objects. Let's check available KafkaConnectorVersions by,

```bash
NAME                   VERSION   CONNECTOR_IMAGE                                                DEPRECATED   AGE
gcs-0.13.0             0.13.0    ghcr.io/appscode-images/kafka-connector-gcs:0.13.0                          10m
jdbc-2.6.1.final       2.6.1     ghcr.io/appscode-images/kafka-connector-jdbc:2.6.1.final                    10m
mongodb-1.11.0         1.11.0    ghcr.io/appscode-images/kafka-connector-mongodb:1.11.0                      10m
mysql-2.4.2.final      2.4.2     ghcr.io/appscode-images/kafka-connector-mysql:2.4.2.final                   10m
postgres-2.4.2.final   2.4.2     ghcr.io/appscode-images/kafka-connector-postgres:2.4.2.final                10m
s3-2.15.0              2.15.0    ghcr.io/appscode-images/kafka-connector-s3:2.15.0                           10m
```

Notice the `DEPRECATED` column. Here, `true` means that this KafkaConnectorVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated KafkaConnectorVersion. You can also use the short from `kcversion` to check available KafkaConnectorVersions.


## Create a Kafka Connect Cluster

The KubeDB operator implements a ConnectCluster CRD to define the specification of ConnectCluster.

The ConnectCluster instance used for this tutorial:

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: ConnectCluster
metadata:
  name: connectcluster-quickstart
  namespace: demo
spec:
  version: 3.6.1
  replicas: 3
  connectorPlugins:
    - mongodb-1.11.0
    - mysql-2.4.2.final
    - postgres-2.4.2.final
    - jdbc-2.6.1.final
  kafkaRef:
    name: kafka-quickstart
    namespace: demo
  terminationPolicy: WipeOut
```

Here,

- `spec.version` - is the name of the KafkaVersion CR. Here, a Kafka of version `3.6.1` will be created.
- `spec.replicas` - specifies the number of ConnectCluster workers.
- `spec.connectorPlugins` - is the name of the KafkaConnectorVersion CR. Here, mongodb, mysql, postgres, and jdbc connector-plugins will be loaded to the ConnectCluster worker nodes.
- `spec.kafkaRef` specifies the Kafka instance that the ConnectCluster will connect to. Here, the ConnectCluster will connect to the Kafka instance named `kafka-quickstart` in the `demo` namespace.
- `spec.terminationPolicy` specifies what KubeDB should do when a user try to delete Kafka CR. Termination policy `Delete` will delete the database pods, secret when the Kafka CR is deleted.

Let's create the Kafka CR that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/Kafka/quickstart/overview/connectcluster/yamls/connectcluster.yaml
connectcluster.kafka.kubedb.com/connectcluster-quickstart created
```

The ConnectCluster's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the ConnectCluster.

```bash
$ kubectl get connectcluster -n demo -w
NAME                        TYPE                        VERSION   STATUS         AGE
connectcluster-quickstart   kafka.kubedb.com/v1alpha1   3.6.1     Provisioning   2s
connectcluster-quickstart   kafka.kubedb.com/v1alpha1   3.6.1     Provisioning   4s
.
.
connectcluster-quickstart   kafka.kubedb.com/v1alpha1   3.6.1     Ready          112s

```

Describe the connectcluster object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe connectcluster -n demo connectcluster-quickstart
Name:         connectcluster-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kafka.kubedb.com/v1alpha1
Kind:         ConnectCluster
Metadata:
  Creation Timestamp:  2024-05-02T07:06:07Z
  Finalizers:
    kafka.kubedb.com/finalizer
  Generation:        2
  Resource Version:  8824
  UID:               bbf4669c-db7a-46c0-a1f4-c93a5e24592e
Spec:
  Auth Secret:
    Name:  connectcluster-quickstart-connect-cred
  Connector Plugins:
    mongodb-1.11.0
    mysql-2.4.2.final
    postgres-2.4.2.final
    jdbc-2.6.1.final
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     20
    Timeout Seconds:    10
  Kafka Ref:
    Name:       kafka-quickstart
    Namespace:  demo
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Env:
          Name:   CONNECT_CLUSTER_MODE
          Value:  distributed
        Name:     connect-cluster
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
          Run As Group:     1001
          Run As Non Root:  true
          Run As User:      1001
          Seccomp Profile:
            Type:  RuntimeDefault
      Init Containers:
        Name:  mongodb
        Resources:
          Limits:
            Memory:  512Mi
          Requests:
            Cpu:     200m
            Memory:  512Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     1001
          Run As Non Root:  true
          Run As User:      1001
          Seccomp Profile:
            Type:  RuntimeDefault
        Name:      mysql
        Resources:
          Limits:
            Memory:  512Mi
          Requests:
            Cpu:     200m
            Memory:  512Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     1001
          Run As Non Root:  true
          Run As User:      1001
          Seccomp Profile:
            Type:  RuntimeDefault
        Name:      postgres
        Resources:
          Limits:
            Memory:  512Mi
          Requests:
            Cpu:     200m
            Memory:  512Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     1001
          Run As Non Root:  true
          Run As User:      1001
          Seccomp Profile:
            Type:  RuntimeDefault
        Name:      jdbc
        Resources:
          Limits:
            Memory:  512Mi
          Requests:
            Cpu:     200m
            Memory:  512Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     1001
          Run As Non Root:  true
          Run As User:      1001
          Seccomp Profile:
            Type:  RuntimeDefault
      Security Context:
        Fs Group:      1001
  Replicas:            3
  Termination Policy:  WipeOut
  Version:             3.6.1
Status:
  Conditions:
    Last Transition Time:  2024-05-02T08:04:29Z
    Message:               The KubeDB operator has started the provisioning of ConnectCluster: demo/connectcluster-quickstart
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-05-02T08:06:20Z
    Message:               All desired replicas are ready.
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-05-02T08:06:45Z
    Message:               The ConnectCluster: demo/connectcluster-quickstart is accepting client requests
    Observed Generation:   2
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-05-02T08:06:45Z
    Message:               The ConnectCluster: demo/connectcluster-quickstart is ready.
    Observed Generation:   2
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-05-02T08:06:46Z
    Message:               The ConnectCluster: demo/connectcluster-quickstart is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>

```

### KubeDB Operator Generated Resources

On deployment of a ConnectCluster CR, the operator creates the following resources:

```bash
$ kubectl get all,secret -n demo -l 'app.kubernetes.io/instance=connectcluster-quickstart'
NAME                              READY   STATUS    RESTARTS   AGE
pod/connectcluster-quickstart-0   1/1     Running   0          3m50s
pod/connectcluster-quickstart-1   1/1     Running   0          3m7s
pod/connectcluster-quickstart-2   1/1     Running   0          2m36s

NAME                                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/connectcluster-quickstart        ClusterIP   10.128.221.44   <none>        8083/TCP   3m55s
service/connectcluster-quickstart-pods   ClusterIP   None            <none>        8083/TCP   3m55s

NAME                                         READY   AGE
statefulset.apps/connectcluster-quickstart   3/3     3m50s

NAME                                                           TYPE                              VERSION   AGE
appbinding.appcatalog.appscode.com/connectcluster-quickstart   kafka.kubedb.com/connectcluster   3.6.1     3m50s

NAME                                            TYPE                       DATA   AGE
secret/connectcluster-quickstart-config         Opaque                     1      3m55s
secret/connectcluster-quickstart-connect-cred   kubernetes.io/basic-auth   2      3m56s

```

- `StatefulSet` - a StatefulSet named after the ConnectCluster instance.
- `Services` -  For a ConnectCluster instance headless service is created with name `{ConnectCluster-name}-{pods}` and a primary service created with name `{ConnectCluster-name}`.
- `AppBinding` - an [AppBinding](/docs/guides/kafka/concepts/appbinding.md) which hold to connect information for the ConnectCluster worker nodes. It is also named after the ConnectCluster instance.
- `Secrets` - 3 types of secrets are generated for each Connect cluster.
    - `{ConnectCluster-Name}-connect-cred` - the auth secrets which hold the `username` and `password` for the Kafka users. Operator generates credentials for `admin` user if not provided and creates a secret for authentication.
    - `{ConnectCluster-Name}-{alias}-cert` - the certificate secrets which hold `tls.crt`, `tls.key`, and `ca.crt` for configuring the ConnectCluster instance if tls enabled.
    - `{ConnectCluster-Name}-config` - the default configuration secret created by the operator.


## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo connectcluster connectcluster-quickstart -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
connectcluster.kafka.kubedb.com/connectcluster-quickstart patched

$ kubectl delete kf connectcluster-quickstart  -n demo
connectcluster.kafka.kubedb.com "connectcluster-quickstart" deleted

$  kubectl delete namespace demo
namespace "demo" deleted
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for the production environment. You can follow these tips to avoid them.

1 **Use `terminationPolicy: Delete`**. It is nice to be able to resume the cluster from the previous one. So, we preserve auth `Secrets`. If you don't want to resume the cluster, you can just use `spec.terminationPolicy: WipeOut`. It will clean up every resource that was created with the ConnectCluster CR. For more details, please visit [here](/docs/guides/kafka/concepts/kafka.md#specterminationpolicy).

## Next Steps

- [Quickstart ConnectCluster](/docs/guides/kafka/quickstart/overview/connectcluster/index.md) with KubeDB Operator.
- Kafka Clustering supported by KubeDB
  - [Combined Clustering](/docs/guides/kafka/clustering/combined-cluster/index.md)
  - [Topology Clustering](/docs/guides/kafka/clustering/topology-cluster/index.md)
- Use [kubedb cli](/docs/guides/kafka/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
