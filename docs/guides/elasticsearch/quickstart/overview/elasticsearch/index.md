---
title: Elasticsearch Quickstart
menu:
  docs_{{ .version }}:
    identifier: es-elasticsearch-overview-elasticsearch
    name: Elasticsearch
    parent: es-overview-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch QuickStart

This tutorial will show you how to use KubeDB to run an Elasticsearch database.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/elasticsearch/quickstart/overview/elasticsearch/images/Lifecycle-of-an-Elasticsearch-CRD.png">
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

> Note: YAML files used in this tutorial are stored in [guides/elasticsearch/quickstart/overview/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/elasticsearch/quickstart/overview/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> We have designed this tutorial to demonstrate a production setup of KubeDB managed Elasticsearch. If you just want to try out KubeDB, you can bypass some safety features following the tips [here](/docs/guides/elasticsearch/quickstart/overview/elasticsearch/index.md#tips-for-testing).

## Find Available StorageClass

We will have to provide `StorageClass` in Elasticsearch CRD specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  14h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Find Available ElasticsearchVersion

When you install the KubeDB operator, it registers a CRD named [ElasticsearchVersion](/docs/guides/elasticsearch/concepts/catalog/index.md). The installation process comes with a set of tested ElasticsearchVersion objects. Let's check available ElasticsearchVersions by,

```bash
$ kubectl get elasticsearchversions
NAME                        VERSION   DISTRIBUTION   DB_IMAGE                                              DEPRECATED   AGE
kubedb-searchguard-5.6.16   5.6.16    KubeDB         kubedb/elasticsearch:5.6.16-searchguard-v2022.02.22                4h24m
kubedb-xpack-7.12.0         7.12.0    KubeDB         kubedb/elasticsearch:7.12.0-xpack-v2021.08.23                      4h24m
kubedb-xpack-7.13.2         7.13.2    KubeDB         kubedb/elasticsearch:7.13.2-xpack-v2021.08.23                      4h24m
xpack-8.11.1         7.14.0    KubeDB         kubedb/elasticsearch:7.14.0-xpack-v2021.08.23                      4h24m
kubedb-xpack-8.11.1         7.16.2    KubeDB         kubedb/elasticsearch:7.16.2-xpack-v2021.12.24                      4h24m
kubedb-xpack-7.9.1          7.9.1     KubeDB         kubedb/elasticsearch:7.9.1-xpack-v2021.08.23                       4h24m
kubedb-xpack-8.2.3          8.2.0     KubeDB         kubedb/elasticsearch:8.2.0-xpack-v2022.05.24                       4h24m
opendistro-1.0.2            7.0.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.0.2                          4h24m
opendistro-1.0.2-v1         7.0.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.0.2                          4h24m
opendistro-1.1.0            7.1.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.1.0                          4h24m
opendistro-1.1.0-v1         7.1.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.1.0                          4h24m
opendistro-1.10.1           7.9.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.10.1                         4h24m
opensearch-2.8.0           7.9.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.10.1                         4h24m
opensearch-2.8.0           7.10.0    OpenDistro     amazon/opendistro-for-elasticsearch:1.12.0                         4h24m
opendistro-1.13.2           7.10.2    OpenDistro     amazon/opendistro-for-elasticsearch:1.13.2                         4h24m
opendistro-1.2.1            7.2.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.2.1                          4h24m
opendistro-1.2.1-v1         7.2.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.2.1                          4h24m
opendistro-1.3.0            7.3.2     OpenDistro     amazon/opendistro-for-elasticsearch:1.3.0                          4h24m
opendistro-1.3.0-v1         7.3.2     OpenDistro     amazon/opendistro-for-elasticsearch:1.3.0                          4h24m
opendistro-1.4.0            7.4.2     OpenDistro     amazon/opendistro-for-elasticsearch:1.4.0                          4h24m
opendistro-1.4.0-v1         7.4.2     OpenDistro     amazon/opendistro-for-elasticsearch:1.4.0                          4h24m
opendistro-1.6.0            7.6.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.6.0                          4h24m
opendistro-1.6.0-v1         7.6.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.6.0                          4h24m
opendistro-1.7.0            7.6.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.7.0                          4h24m
opendistro-1.7.0-v1         7.6.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.7.0                          4h24m
opendistro-1.8.0            7.7.0     OpenDistro     amazon/opendistro-for-elasticsearch:1.8.0                          4h24m
opendistro-1.8.0-v1         7.7.0     OpenDistro     amazon/opendistro-for-elasticsearch:1.8.0                          4h24m
opendistro-1.9.0            7.8.0     OpenDistro     amazon/opendistro-for-elasticsearch:1.9.0                          4h24m
opendistro-1.9.0-v1         7.8.0     OpenDistro     amazon/opendistro-for-elasticsearch:1.9.0                          4h24m
opensearch-1.1.0            1.1.0     OpenSearch     opensearchproject/opensearch:1.1.0                                 4h24m
opensearch-2.8.0            1.2.2     OpenSearch     opensearchproject/opensearch:1.2.2                                 4h24m
opensearch-2.8.0            1.3.2     OpenSearch     opensearchproject/opensearch:1.3.2                                 4h24m
searchguard-6.8.1           6.8.1     SearchGuard    floragunncom/sg-elasticsearch:6.8.1-oss-25.1                       4h24m
searchguard-6.8.1-v1        6.8.1     SearchGuard    floragunncom/sg-elasticsearch:6.8.1-oss-25.1                       4h24m
searchguard-7.0.1           7.0.1     SearchGuard    floragunncom/sg-elasticsearch:7.0.1-oss-35.0.0                     4h24m
searchguard-7.0.1-v1        7.0.1     SearchGuard    floragunncom/sg-elasticsearch:7.0.1-oss-35.0.0                     4h24m
searchguard-7.1.1           7.1.1     SearchGuard    floragunncom/sg-elasticsearch:7.1.1-oss-35.0.0                     4h24m
searchguard-7.1.1-v1        7.1.1     SearchGuard    floragunncom/sg-elasticsearch:7.1.1-oss-35.0.0                     4h24m
searchguard-7.10.2          7.10.2    SearchGuard    floragunncom/sg-elasticsearch:7.10.2-oss-49.0.0                    4h24m
xpack-8.11.1          7.14.2    SearchGuard    floragunncom/sg-elasticsearch:7.14.2-52.3.0                        4h24m
searchguard-7.3.2           7.3.2     SearchGuard    floragunncom/sg-elasticsearch:7.3.2-oss-37.0.0                     4h24m
searchguard-7.5.2           7.5.2     SearchGuard    floragunncom/sg-elasticsearch:7.5.2-oss-40.0.0                     4h24m
xpack-8.11.1        7.5.2     SearchGuard    floragunncom/sg-elasticsearch:7.5.2-oss-40.0.0                     4h24m
searchguard-7.8.1           7.8.1     SearchGuard    floragunncom/sg-elasticsearch:7.8.1-oss-43.0.0                     4h24m
xpack-8.11.1           7.9.3     SearchGuard    floragunncom/sg-elasticsearch:7.9.3-oss-47.1.0                     4h24m
xpack-6.8.10-v1             6.8.10    ElasticStack   elasticsearch:6.8.10                                               4h24m
xpack-6.8.16                6.8.16    ElasticStack   elasticsearch:6.8.16                                               4h24m
xpack-6.8.22                6.8.22    ElasticStack   elasticsearch:6.8.22                                               4h24m
xpack-7.0.1-v1              7.0.1     ElasticStack   elasticsearch:7.0.1                                                4h24m
xpack-7.1.1-v1              7.1.1     ElasticStack   elasticsearch:7.1.1                                                4h24m
xpack-7.12.0                7.12.0    ElasticStack   elasticsearch:7.12.0                                               4h24m
xpack-7.12.0-v1             7.12.0    ElasticStack   elasticsearch:7.12.0                                               4h24m
xpack-7.13.2                7.13.2    ElasticStack   elasticsearch:7.13.2                                               4h24m
xpack-8.11.1                7.14.0    ElasticStack   elasticsearch:7.14.0                                               4h24m
xpack-8.11.1                7.16.2    ElasticStack   elasticsearch:7.16.2                                               4h24m
xpack-7.17.3                7.17.3    ElasticStack   elasticsearch:7.17.3                                               4h24m
xpack-7.2.1-v1              7.2.1     ElasticStack   elasticsearch:7.2.1                                                4h24m
xpack-7.3.2-v1              7.3.2     ElasticStack   elasticsearch:7.3.2                                                4h24m
xpack-7.4.2-v1              7.4.2     ElasticStack   elasticsearch:7.4.2                                                4h24m
xpack-7.5.2-v1              7.5.2     ElasticStack   elasticsearch:7.5.2                                                4h24m
xpack-7.6.2-v1              7.6.2     ElasticStack   elasticsearch:7.6.2                                                4h24m
xpack-7.7.1-v1              7.7.1     ElasticStack   elasticsearch:7.7.1                                                4h24m
xpack-7.8.0-v1              7.8.0     ElasticStack   elasticsearch:7.8.0                                                4h24m
xpack-8.11.1              7.9.1     ElasticStack   elasticsearch:7.9.1                                                4h24m
xpack-7.9.1-v2              7.9.1     ElasticStack   elasticsearch:7.9.1                                                4h24m
xpack-8.2.3                 8.2.0     ElasticStack   elasticsearch:8.2.0                                                4h24m
xpack-8.5.2                 8.5.2     ElasticStack   elasticsearch:8.5.2                                                4h24m
```

Notice the `DEPRECATED` column. Here, `true` means that this ElasticsearchVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated ElasticsearchVersion.

In this tutorial, we will use `xpack-8.2.3` ElasticsearchVersion CR to create an Elasticsearch cluster.

> Note: An image with a higher modification tag will have more features and fixes than an image with a lower modification tag. Hence, it is recommended to use ElasticsearchVersion CRD with the highest modification tag to take advantage of the latest features. For example, use `xpack-8.11.1` over `7.9.1-xpack`.

## Create an Elasticsearch Cluster

The KubeDB operator implements an Elasticsearch CRD to define the specification of an Elasticsearch database.

The Elasticsearch instance used for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-quickstart
  namespace: demo
spec:
  version: xpack-8.2.3
  enableSSL: true
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

Here,

- `spec.version` - is the name of the ElasticsearchVersion CR. Here, an Elasticsearch of version `8.2.0` will be created with `x-pack` security plugin.
- `spec.enableSSL` - specifies whether the HTTP layer is secured with certificates or not.
- `spec.replicas` - specifies the number of Elasticsearch nodes.
- `spec.storageType` - specifies the type of storage that will be used for Elasticsearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the Elasticsearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by the KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. If you don't specify `spec.storageType: Ephemeral`, then this field is required.
- `spec.deletionPolicy` specifies what KubeDB should do when a user try to delete Elasticsearch CR. Termination policy `Delete` will delete the database pods, secret and PVC when the Elasticsearch CR is deleted.

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in the `storage.resources.requests` field. Don't specify `limits` here. PVC does not get resized automatically.

Let's create the Elasticsearch CR that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/quickstart/overview/elasticsearch/yamls/elasticsearch.yaml
elasticsearch.kubedb.com/es-quickstart created
```

The Elasticsearch's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the database.

```bash
$ kubectl get elasticsearch -n demo -w
NAME            VERSION       STATUS         AGE
es-quickstart   xpack-8.2.3   Provisioning   7s
... ...
es-quickstart   xpack-8.2.3   Ready          39s
```

Describe the Elasticsearch object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe elasticsearch -n demo  es-quickstart
Name:         es-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Elasticsearch
Metadata:
  Creation Timestamp:  2022-12-27T05:25:39Z
  Finalizers:
    kubedb.com
  Generation:  1
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
        f:enableSSL:
        f:healthChecker:
          .:
          f:failureThreshold:
          f:periodSeconds:
          f:timeoutSeconds:
        f:heapSizePercentage:
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
        f:deletionPolicy:
        f:version:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-12-27T05:25:39Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:finalizers:
          .:
          v:"kubedb.com":
    Manager:      kubedb-provisioner
    Operation:    Update
    Time:         2022-12-27T05:25:39Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-provisioner
    Operation:       Update
    Subresource:     status
    Time:            2022-12-27T05:25:39Z
  Resource Version:  313887
  UID:               cf37390a-ab9f-4886-9f7e-1a5bedc975e7
Spec:
  Auth Secret:
    Name:  es-quickstart-elastic-cred
  Auto Ops:
  Enable SSL:  true
  Health Checker:
    Failure Threshold:   1
    Period Seconds:      10
    Timeout Seconds:     10
  Heap Size Percentage:  50
  Internal Users:
    apm_system:
      Backend Roles:
        apm_system
      Secret Name:  es-quickstart-apm-system-cred
    beats_system:
      Backend Roles:
        beats_system
      Secret Name:  es-quickstart-beats-system-cred
    Elastic:
      Backend Roles:
        superuser
      Secret Name:  es-quickstart-elastic-cred
    kibana_system:
      Backend Roles:
        kibana_system
      Secret Name:  es-quickstart-kibana-system-cred
    logstash_system:
      Backend Roles:
        logstash_system
      Secret Name:  es-quickstart-logstash-system-cred
    remote_monitoring_user:
      Backend Roles:
        remote_monitoring_collector
        remote_monitoring_agent
      Secret Name:  es-quickstart-remote-monitoring-user-cred
  Kernel Settings:
    Privileged:  true
    Sysctls:
      Name:   vm.max_map_count
      Value:  262144
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Affinity:
        Pod Anti Affinity:
          Preferred During Scheduling Ignored During Execution:
            Pod Affinity Term:
              Label Selector:
                Match Labels:
                  app.kubernetes.io/instance:    es-quickstart
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        elasticsearches.kubedb.com
              Namespaces:
                demo
              Topology Key:  kubernetes.io/hostname
            Weight:          100
            Pod Affinity Term:
              Label Selector:
                Match Labels:
                  app.kubernetes.io/instance:    es-quickstart
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        elasticsearches.kubedb.com
              Namespaces:
                demo
              Topology Key:  failure-domain.beta.kubernetes.io/zone
            Weight:          50
      Container Security Context:
        Capabilities:
          Add:
            IPC_LOCK
            SYS_RESOURCE
        Privileged:   false
        Run As User:  1000
      Resources:
        Limits:
          Memory:  1536Mi
        Requests:
          Cpu:               500m
          Memory:            1536Mi
      Service Account Name:  es-quickstart
  Replicas:                  3
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Termination Policy:    Delete
  Tls:
    Certificates:
      Alias:  ca
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-quickstart-ca-cert
      Subject:
        Organizations:
          kubedb
      Alias:  transport
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-quickstart-transport-cert
      Subject:
        Organizations:
          kubedb
      Alias:  http
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-quickstart-http-cert
      Subject:
        Organizations:
          kubedb
      Alias:  client
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-quickstart-client-cert
      Subject:
        Organizations:
          kubedb
  Version:  xpack-8.2.3
Status:
  Conditions:
    Last Transition Time:  2022-12-27T05:25:39Z
    Message:               The KubeDB operator has started the provisioning of Elasticsearch: demo/es-quickstart
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2022-12-27T05:25:41Z
    Message:               Internal Users for Elasticsearch: demo/es-quickstart is ready.
    Observed Generation:   1
    Reason:                InternalUsersCredentialsSyncedSuccessfully
    Status:                True
    Type:                  InternalUsersSynced
    Last Transition Time:  2022-12-27T05:28:48Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2022-12-27T05:29:05Z
    Message:               The Elasticsearch: demo/es-quickstart is accepting client requests.
    Observed Generation:   1
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2022-12-27T05:29:05Z
    Message:               The Elasticsearch: demo/es-quickstart is ready.
    Observed Generation:   1
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2022-12-27T05:29:06Z
    Message:               The Elasticsearch: demo/es-quickstart is accepting write requests.
    Observed Generation:   1
    Reason:                DatabaseWriteAccessCheckSucceeded
    Status:                True
    Type:                  DatabaseWriteAccess
    Last Transition Time:  2022-12-27T05:29:13Z
    Message:               The Elasticsearch: demo/es-quickstart is successfully provisioned.
    Observed Generation:   1
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
    Last Transition Time:  2022-12-27T05:29:15Z
    Message:               The Elasticsearch: demo/es-quickstart is accepting read requests.
    Observed Generation:   1
    Reason:                DatabaseReadAccessCheckSucceeded
    Status:                True
    Type:                  DatabaseReadAccess
  Observed Generation:     1
  Phase:                   Ready
Events:
  Type    Reason      Age    From             Message
  ----    ------      ----   ----             -------
  Normal  Successful  4m48s  KubeDB Operator  Successfully created governing service
  Normal  Successful  4m47s  KubeDB Operator  Successfully created Service
  Normal  Successful  4m47s  KubeDB Operator  Successfully created Service
  Normal  Successful  4m40s  KubeDB Operator  Successfully created Elasticsearch
  Normal  Successful  4m40s  KubeDB Operator  Successfully created appbinding
  Normal  Successful  4m40s  KubeDB Operator  Successfully  governing service
  Normal  Successful  4m32s  KubeDB Operator  Successfully  governing service
  Normal  Successful  99s    KubeDB Operator  Successfully  governing service
  Normal  Successful  82s    KubeDB Operator  Successfully  governing service
  Normal  Successful  74s    KubeDB Operator  Successfully  governing service
  Normal  Successful  66s    KubeDB Operator  Successfully  governing service
```

### KubeDB Operator Generated Resources

On deployment of an Elasticsearch CR, the operator creates the following resources:

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=es-quickstart'
NAME                  READY   STATUS    RESTARTS   AGE
pod/es-quickstart-0   1/1     Running   0          8m2s
pod/es-quickstart-1   1/1     Running   0          5m15s
pod/es-quickstart-2   1/1     Running   0          5m8s

NAME                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/es-quickstart          ClusterIP   10.96.209.204   <none>        9200/TCP   8m9s
service/es-quickstart-master   ClusterIP   None            <none>        9300/TCP   8m9s
service/es-quickstart-pods     ClusterIP   None            <none>        9200/TCP   8m10s

NAME                             READY   AGE
statefulset.apps/es-quickstart   3/3     8m2s

NAME                                               TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-quickstart   kubedb.com/elasticsearch   8.2.0     8m2s

NAME                                               TYPE                       DATA   AGE
secret/es-quickstart-apm-system-cred               kubernetes.io/basic-auth   2      8m8s
secret/es-quickstart-beats-system-cred             kubernetes.io/basic-auth   2      8m8s
secret/es-quickstart-ca-cert                       kubernetes.io/tls          2      8m9s
secret/es-quickstart-client-cert                   kubernetes.io/tls          3      8m8s
secret/es-quickstart-config                        Opaque                     1      8m8s
secret/es-quickstart-elastic-cred                  kubernetes.io/basic-auth   2      8m8s
secret/es-quickstart-http-cert                     kubernetes.io/tls          3      8m9s
secret/es-quickstart-kibana-system-cred            kubernetes.io/basic-auth   2      8m8s
secret/es-quickstart-logstash-system-cred          kubernetes.io/basic-auth   2      8m8s
secret/es-quickstart-remote-monitoring-user-cred   kubernetes.io/basic-auth   2      8m8s
secret/es-quickstart-transport-cert                kubernetes.io/tls          3      8m9s

NAME                                         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-es-quickstart-0   Bound    pvc-e5227633-2fc0-4a50-a599-57cba8b31d14   1Gi        RWO            standard       8m2s
persistentvolumeclaim/data-es-quickstart-1   Bound    pvc-fbacd36c-4132-4e2a-a5c5-91149054044c   1Gi        RWO            standard       5m15s
persistentvolumeclaim/data-es-quickstart-2   Bound    pvc-9f9c6eaf-1ba6-4167-a37d-86eaf1f7e103   1Gi        RWO            standard       5m8s
```

- `StatefulSet` - a StatefulSet named after the Elasticsearch instance. In topology mode, the operator creates 3 petSets with name `{Elasticsearch-Name}-{Sufix}`.
- `Services` -  3 services are generated for each Elasticsearch database.
  - `{Elasticsearch-Name}` - the client service which is used to connect to the database. It points to the `ingest` nodes.
  - `{Elasticsearch-Name}-master` - the master service which is used to connect to the master nodes. It is a headless service.
  - `{Elasticsearch-Name}-pods` - the node discovery service which is used by the Elasticsearch nodes to communicate each other. It is a headless service.
- `AppBinding` - an [AppBinding](/docs/guides/elasticsearch/concepts/appbinding/index.md) which hold to connect information for the database. It is also named after the Elastics
- `Secrets` - 3 types of secrets are generated for each Elasticsearch database.
  - `{Elasticsearch-Name}-{username}-cred` - the auth secrets which hold the `username` and `password` for the Elasticsearch users. The auth secret `es-quickstart-elastic-cred` holds the `username` and `password` for `elastic` user which lets administrative access.
  - `{Elasticsearch-Name}-{alias}-cert` - the certificate secrets which hold `tls.crt`, `tls.key`, and `ca.crt` for configuring the Elasticsearch database.
  - `{Elasticsearch-Name}-config` - the default configuration secret created by the operator.
  - `data-{Elasticsearch-node-name}` - the persistent volume claims created by the StatefulSet.

## Connect with Elasticsearch Database

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

KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` termination policy. If admission webhook is enabled, it prevents the user from deleting the database as long as the `spec.deletionPolicy` is set `DoNotTerminate`.

To halt the database, we have to set `spec.deletionPolicy:` to `Halt` by updating it,

```bash
$ kubectl edit elasticsearch -n demo es-quickstart

>> spec:
>>   deletionPolicy: Halt
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

Say, the Elasticsearch CR was deleted with `spec.deletionPolicy` to `Halt` and you want to re-create the Elasticsearch cluster using the existing auth secrets and the PVCs.

You can do it by simpily re-deploying the original Elasticsearch object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/quickstart/overview/elasticsearch/yamls/elasticsearch.yaml
elasticsearch.kubedb.com/es-quickstart created
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo elasticsearch es-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
elasticsearch.kubedb.com/es-quickstart patched

$ kubectl delete -n demo es/quick-elasticsearch
elasticsearch.kubedb.com "es-quickstart" deleted

$  kubectl delete namespace demo
namespace "demo" deleted
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for the production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if the database pod fails. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purposes, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `deletionPolicy: WipeOut`**. It is nice to be able to resume the database from the previous one. So, we preserve all your `PVCs` and auth `Secrets`. If you don't want to resume the database, you can just use `spec.deletionPolicy: WipeOut`. It will clean up every resouce that was created with the Elasticsearch CR. For more details, please visit [here](/docs/guides/elasticsearch/concepts/elasticsearch/index.md#specdeletionpolicy).

## Next Steps

- [Quickstart Kibana](/docs/guides/elasticsearch/elasticsearch-dashboard/kibana/index.md) with KubeDB Operator.
- Learn how to configure [Elasticsearch Topology Cluster](/docs/guides/elasticsearch/clustering/topology-cluster/simple-dedicated-cluster/index.md).
- Learn about [backup & restore](/docs/guides/elasticsearch/backup/overview/index.md) Elasticsearch database using Stash.
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
