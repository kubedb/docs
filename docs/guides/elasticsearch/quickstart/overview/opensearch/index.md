---
title: OpenSearch Quickstart
menu:
  docs_{{ .version }}:
    identifier: es-opensearch-overview-elasticsearch
    name: OpenSearch
    parent: es-overview-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](https://kubedb.com/docs/v2021.12.21/welcome/).

# OpenSearch QuickStart

This tutorial will show you how to use KubeDB to run an OpenSearch database.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/elasticsearch/quickstart/overview/opensearch/images/Lifecycle-of-an-Opensearch-CRD.png">
</p>

## Before You Begin

* At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

* Now, install the KubeDB operator in your cluster following the steps [here](https://kubedb.com/docs/v2021.12.21/setup/).

* Elasticsearch has many distributions like `ElasticStack`, `OpenSearch`, `SearchGuard`, `OpenDistro` etc. KubeDB provides all of these distribution’s support under the Elasticsearch CR of KubeDB. So, in this tutorial we will deploy OpenSearch with the help of KubeDB managed Elasticsearch CR.

* [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required for CRD specification. Check the available StorageClass in cluster.

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  11h
```
Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).


To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [guides/elasticsearch/quickstart/overview/opensearch/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/elasticsearch/quickstart/overview/opensearch/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> We have designed this tutorial to demonstrate a production setup of KubeDB managed OpenSearch. If you just want to try out KubeDB, you can bypass some of the safety features following the tips [here](/docs/guides/elasticsearch/quickstart/overview/opensearch/index.md#tips-for-testing).


## Find Available Versions

When you install the KubeDB operator, it registers a CRD named [ElasticsearchVersion](/docs/guides/elasticsearch/concepts/catalog/index.md). The installation process comes with a set of tested ElasticsearchVersion objects. Let's check available ElasticsearchVersions by following command,

```bash
$ kubectl get elasticsearchversions
NAME                   VERSION   DISTRIBUTION   DB_IMAGE                                          DEPRECATED   AGE
kubedb-xpack-7.12.0    7.12.0    KubeDB         kubedb/elasticsearch:7.12.0-xpack-v2021.08.23                  17h
kubedb-xpack-7.13.2    7.13.2    KubeDB         kubedb/elasticsearch:7.13.2-xpack-v2021.08.23                  17h
xpack-8.11.1    7.14.0    KubeDB         kubedb/elasticsearch:7.14.0-xpack-v2021.08.23                  17h
kubedb-xpack-8.11.1    7.16.2    KubeDB         kubedb/elasticsearch:7.16.2-xpack-v2021.12.24                  17h
kubedb-xpack-7.9.1     7.9.1     KubeDB         kubedb/elasticsearch:7.9.1-xpack-v2021.08.23                   17h
opendistro-1.0.2       7.0.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.0.2                      17h
opendistro-1.0.2-v1    7.0.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.0.2                      17h
opendistro-1.1.0       7.1.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.1.0                      17h
opendistro-1.1.0-v1    7.1.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.1.0                      17h
opendistro-1.10.1      7.9.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.10.1                     17h
opensearch-2.8.0      7.9.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.10.1                     17h
opensearch-2.8.0      7.10.0    OpenDistro     amazon/opendistro-for-elasticsearch:1.12.0                     17h
opendistro-1.13.2      7.10.2    OpenDistro     amazon/opendistro-for-elasticsearch:1.13.2                     17h
opendistro-1.2.1       7.2.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.2.1                      17h
opendistro-1.2.1-v1    7.2.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.2.1                      17h
opendistro-1.3.0       7.3.2     OpenDistro     amazon/opendistro-for-elasticsearch:1.3.0                      17h
opendistro-1.3.0-v1    7.3.2     OpenDistro     amazon/opendistro-for-elasticsearch:1.3.0                      17h
opendistro-1.4.0       7.4.2     OpenDistro     amazon/opendistro-for-elasticsearch:1.4.0                      17h
opendistro-1.4.0-v1    7.4.2     OpenDistro     amazon/opendistro-for-elasticsearch:1.4.0                      17h
opendistro-1.6.0       7.6.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.6.0                      17h
opendistro-1.6.0-v1    7.6.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.6.0                      17h
opendistro-1.7.0       7.6.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.7.0                      17h
opendistro-1.7.0-v1    7.6.1     OpenDistro     amazon/opendistro-for-elasticsearch:1.7.0                      17h
opendistro-1.8.0       7.7.0     OpenDistro     amazon/opendistro-for-elasticsearch:1.8.0                      17h
opendistro-1.8.0-v1    7.7.0     OpenDistro     amazon/opendistro-for-elasticsearch:1.8.0                      17h
opendistro-1.9.0       7.8.0     OpenDistro     amazon/opendistro-for-elasticsearch:1.9.0                      17h
opendistro-1.9.0-v1    7.8.0     OpenDistro     amazon/opendistro-for-elasticsearch:1.9.0                      17h
opensearch-1.1.0       1.1.0     OpenSearch     opensearchproject/opensearch:1.1.0                             17h
opensearch-2.8.0       1.2.2     OpenSearch     opensearchproject/opensearch:1.2.2                             17h
searchguard-6.8.1      6.8.1     SearchGuard    floragunncom/sg-elasticsearch:6.8.1-oss-25.1                   17h
searchguard-6.8.1-v1   6.8.1     SearchGuard    floragunncom/sg-elasticsearch:6.8.1-oss-25.1                   17h
searchguard-7.0.1      7.0.1     SearchGuard    floragunncom/sg-elasticsearch:7.0.1-oss-35.0.0                 17h
searchguard-7.0.1-v1   7.0.1     SearchGuard    floragunncom/sg-elasticsearch:7.0.1-oss-35.0.0                 17h
searchguard-7.1.1      7.1.1     SearchGuard    floragunncom/sg-elasticsearch:7.1.1-oss-35.0.0                 17h
searchguard-7.1.1-v1   7.1.1     SearchGuard    floragunncom/sg-elasticsearch:7.1.1-oss-35.0.0                 17h
searchguard-7.10.2     7.10.2    SearchGuard    floragunncom/sg-elasticsearch:7.10.2-oss-49.0.0                17h
xpack-8.11.1     7.14.2    SearchGuard    floragunncom/sg-elasticsearch:7.14.2-52.3.0                    17h
searchguard-7.3.2      7.3.2     SearchGuard    floragunncom/sg-elasticsearch:7.3.2-oss-37.0.0                 17h
searchguard-7.5.2      7.5.2     SearchGuard    floragunncom/sg-elasticsearch:7.5.2-oss-40.0.0                 17h
xpack-8.11.1   7.5.2     SearchGuard    floragunncom/sg-elasticsearch:7.5.2-oss-40.0.0                 17h
searchguard-7.8.1      7.8.1     SearchGuard    floragunncom/sg-elasticsearch:7.8.1-oss-43.0.0                 17h
xpack-8.11.1      7.9.3     SearchGuard    floragunncom/sg-elasticsearch:7.9.3-oss-47.1.0                 17h
xpack-6.8.10-v1        6.8.10    ElasticStack   elasticsearch:6.8.10                                           17h
xpack-6.8.16           6.8.16    ElasticStack   elasticsearch:6.8.16                                           17h
xpack-6.8.22           6.8.22    ElasticStack   elasticsearch:6.8.22                                           17h
xpack-7.0.1-v1         7.0.1     ElasticStack   elasticsearch:7.0.1                                            17h
xpack-7.1.1-v1         7.1.1     ElasticStack   elasticsearch:7.1.1                                            17h
xpack-7.12.0           7.12.0    ElasticStack   elasticsearch:7.12.0                                           17h
xpack-7.12.0-v1        7.12.0    ElasticStack   elasticsearch:7.12.0                                           17h
xpack-7.13.2           7.13.2    ElasticStack   elasticsearch:7.13.2                                           17h
xpack-8.11.1           7.14.0    ElasticStack   elasticsearch:7.14.0                                           17h
xpack-8.11.1           7.16.2    ElasticStack   elasticsearch:7.16.2                                           17h
xpack-7.2.1-v1         7.2.1     ElasticStack   elasticsearch:7.2.1                                            17h
xpack-7.3.2-v1         7.3.2     ElasticStack   elasticsearch:7.3.2                                            17h
xpack-7.4.2-v1         7.4.2     ElasticStack   elasticsearch:7.4.2                                            17h
xpack-7.5.2-v1         7.5.2     ElasticStack   elasticsearch:7.5.2                                            17h
xpack-7.6.2-v1         7.6.2     ElasticStack   elasticsearch:7.6.2                                            17h
xpack-7.7.1-v1         7.7.1     ElasticStack   elasticsearch:7.7.1                                            17h
xpack-7.8.0-v1         7.8.0     ElasticStack   elasticsearch:7.8.0                                            17h
xpack-8.11.1         7.9.1     ElasticStack   elasticsearch:7.9.1                                            17h
xpack-7.9.1-v2         7.9.1     ElasticStack   elasticsearch:7.9.1                                            17h
```

Notice the `DEPRECATED` column. Here, `true` means that this ElasticsearchVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated ElasticsearchVersion.

In this tutorial, we will use `opensearch-2.8.0` ElasticsearchVersion CR to create an OpenSearch cluster.

> Note: An image with a higher modification tag will have more features and fixes than an image with a lower modification tag. Hence, it is recommended to use ElasticsearchVersion CRD with the highest modification tag to take advantage of the latest features. For example, we are using `opensearch-2.8.0` over `opensearch-1.1.0`.

## Create an OpenSearch Cluster

The KubeDB operator implements an Elasticsearch CRD to define the specification of an OpenSearch database.

Here is the yaml we will use for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: sample-opensearch
  namespace: demo
spec:
  version: opensearch-2.8.0
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
  terminationPolicy: DoNotTerminate
```

Here,

- `spec.version` - is the name of the ElasticsearchVersion CR. Here, we are using `opensearch-2.8.0` version.
- `spec.enableSSL` - specifies whether the HTTP layer is secured with certificates or not.
- `spec.replicas` - specifies the number of OpenSearch nodes.
- `spec.storageType` - specifies the type of storage that will be used for OpenSearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the OpenSearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by the KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. If you don't specify `spec.storageType: Ephemeral`, then this field is required.
- `spec.terminationPolicy` specifies what KubeDB should do when a user try to delete the operation of Elasticsearch CR. Termination policy `DoNotTerminate` prevents a user from deleting this object if the admission webhook is enabled.

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in the `storage.resources.requests` field. Don't specify `limits` here. PVC does not get resized automatically.

Let's apply the yaml that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/quickstart/overview/opensearch/yamls/opensearch.yaml
elasticsearch.kubedb.com/es-quickstart created
```

Wait for few minutes until the `STATUS` will go from `Provisioning` to `Ready`. Once the `STATUS` is `Ready`, you are ready to use the database.

```bash
$ kubectl get elasticsearch -n demo -w
NAME                VERSION            STATUS         AGE
sample-opensearch   opensearch-2.8.0   Provisioning   49s
... ...
$ kubectl get elasticsearch -n demo -w
NAME                VERSION            STATUS   AGE
sample-opensearch   opensearch-2.8.0   Ready    5m4s
```

Describe the object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe elasticsearch -n demo sample-opensearch
Name:         sample-opensearch
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Elasticsearch
Metadata:
  Creation Timestamp:  2022-02-15T07:00:21Z
  Finalizers:
    kubedb.com
  Generation:  1
  Resource Version:  84343
  UID:               20c388a6-54b1-4c0d-891b-879ec8e2a8c6
Spec:
  Auth Secret:
    Name:      sample-opensearch-admin-cred
  Enable SSL:  true
  Internal Users:
    Admin:
      Backend Roles:
        admin
      Reserved:     true
      Secret Name:  sample-opensearch-admin-cred
    Kibanaro:
      Secret Name:  sample-opensearch-kibanaro-cred
    Kibanaserver:
      Reserved:     true
      Secret Name:  sample-opensearch-kibanaserver-cred
    Logstash:
      Secret Name:  sample-opensearch-logstash-cred
    Readall:
      Secret Name:  sample-opensearch-readall-cred
    Snapshotrestore:
      Secret Name:  sample-opensearch-snapshotrestore-cred
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
                  app.kubernetes.io/instance:    sample-opensearch
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        elasticsearches.kubedb.com
              Namespaces:
                demo
              Topology Key:  kubernetes.io/hostname
            Weight:          100
            Pod Affinity Term:
              Label Selector:
                Match Labels:
                  app.kubernetes.io/instance:    sample-opensearch
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
        Privileged:  false
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:               500m
          Memory:            1Gi
      Service Account Name:  sample-opensearch
  Replicas:                  3
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Termination Policy:    DoNotTerminate
  Tls:
    Certificates:
      Alias:  ca
      Private Key:
        Encoding:   PKCS8
      Secret Name:  sample-opensearch-ca-cert
      Subject:
        Organizations:
          kubedb
      Alias:  transport
      Private Key:
        Encoding:   PKCS8
      Secret Name:  sample-opensearch-transport-cert
      Subject:
        Organizations:
          kubedb
      Alias:  admin
      Private Key:
        Encoding:   PKCS8
      Secret Name:  sample-opensearch-admin-cert
      Subject:
        Organizations:
          kubedb
      Alias:  http
      Private Key:
        Encoding:   PKCS8
      Secret Name:  sample-opensearch-http-cert
      Subject:
        Organizations:
          kubedb
      Alias:  archiver
      Private Key:
        Encoding:   PKCS8
      Secret Name:  sample-opensearch-archiver-cert
      Subject:
        Organizations:
          kubedb
  Version:  opensearch-2.8.0
Status:
  Conditions:
    Last Transition Time:  2022-02-15T07:00:21Z
    Message:               The KubeDB operator has started the provisioning of Elasticsearch: demo/sample-opensearch
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2022-02-15T07:00:44Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2022-02-15T07:01:35Z
    Message:               The Elasticsearch: demo/sample-opensearch is accepting client requests.
    Observed Generation:   1
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2022-02-15T07:01:35Z
    Message:               The Elasticsearch: demo/sample-opensearch is ready.
    Observed Generation:   1
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2022-02-15T07:01:35Z
    Message:               The Elasticsearch: demo/sample-opensearch is successfully provisioned.
    Observed Generation:   1
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     1
  Phase:                   Ready
Events:
  Type    Reason      Age    From             Message
  ----    ------      ----   ----             -------
  Normal  Successful  56m    KubeDB Operator  Successfully  governing service
  Normal  Successful  56m    KubeDB Operator  Successfully  governing service
```

### KubeDB Operator Generated Resources

after the deployment, the operator creates the following resources:

```bash
$ kubectl get all,secret -n demo -l 'app.kubernetes.io/instance=sample-opensearch'
NAME                      READY   STATUS    RESTARTS   AGE
pod/sample-opensearch-0   1/1     Running   0          23m
pod/sample-opensearch-1   1/1     Running   0          23m
pod/sample-opensearch-2   1/1     Running   0          23m

NAME                               TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/sample-opensearch          ClusterIP   10.96.29.157   <none>        9200/TCP   23m
service/sample-opensearch-master   ClusterIP   None           <none>        9300/TCP   23m
service/sample-opensearch-pods     ClusterIP   None           <none>        9200/TCP   23m

NAME                                 READY   AGE
statefulset.apps/sample-opensearch   3/3     23m

NAME                                                   TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/sample-opensearch   kubedb.com/elasticsearch   1.2.2     23m

NAME                                            TYPE                       DATA   AGE
secret/sample-opensearch-admin-cert             kubernetes.io/tls          3      23m
secret/sample-opensearch-admin-cred             kubernetes.io/basic-auth   2      23m
secret/sample-opensearch-archiver-cert          kubernetes.io/tls          3      23m
secret/sample-opensearch-ca-cert                kubernetes.io/tls          2      23m
secret/sample-opensearch-config                 Opaque                     3      23m
secret/sample-opensearch-http-cert              kubernetes.io/tls          3      23m
secret/sample-opensearch-kibanaro-cred          kubernetes.io/basic-auth   2      23m
secret/sample-opensearch-kibanaserver-cred      kubernetes.io/basic-auth   2      23m
secret/sample-opensearch-logstash-cred          kubernetes.io/basic-auth   2      23m
secret/sample-opensearch-readall-cred           kubernetes.io/basic-auth   2      23m
secret/sample-opensearch-snapshotrestore-cred   kubernetes.io/basic-auth   2      23m
secret/sample-opensearch-transport-cert         kubernetes.io/tls          3      23m

```

- `StatefulSet` - a StatefulSet named after the OpenSearch instance.
- `Services` -  3 services are generated for each OpenSearch database.
  - `{OpenSearch-Name}` - the client service which is used to connect to the database. It points to the `ingest` nodes.
  - `{OpenSearch-Name}-master` - the master service which is used to connect to the master nodes. It is a headless service.
  - `{OpenSearch-Name}-pods` - the node discovery service which is used by the OpenSearch nodes to communicate each other. It is a headless service.
- `AppBinding` - an [AppBinding](/docs/guides/elasticsearch/concepts/appbinding/index.md) which hold to connect information for the database.
- `Secrets` - 3 types of secrets are generated for each OpenSearch database.
  - `{OpenSearch-Name}-{username}-cred` - the auth secrets which hold the `username` and `password` for the OpenSearch users.
  - `{OpenSearch-Name}-{alias}-cert` - the certificate secrets which hold `tls.crt`, `tls.key`, and `ca.crt` for configuring the OpenSearch database.
  - `{OpenSearch-Name}-config` - the default configuration secret created by the operator.

### Insert Sample Data

In this section, we are going to create few indexes in the deployed OpenSearch. At first, we are going to port-forward the respective Service so that we can connect with the database from our local machine. Then, we are going to insert some data into the OpenSearch.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s see the Services created by KubeDB for our OpenSearch,

```bash
$ kubectl get service -n demo
NAME                               TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
sample-opensearch                  ClusterIP   10.48.14.99   <none>        9200/TCP   4m33s
sample-opensearch-master           ClusterIP   None          <none>        9300/TCP   4m33s
sample-opensearch-pods             ClusterIP   None          <none>        9200/TCP   4m33s
```
Here, we are going to use the `sample-opensearch` Service to connect with the database. Now, let’s port-forward the `sample-opensearch` Service.

```bash
# Port-forward the service to local machine
$ kubectl port-forward -n demo svc/sample-opensearch 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

#### Export the Credentials

KubeDB will create some Secrets for the database. Let’s check which Secrets have been created by KubeDB for our `sample-opensearch`.

```bash
$ kubectl get secret -n demo | grep sample-opensearch
sample-opensearch-admin-cert             kubernetes.io/tls                     3      10m
sample-opensearch-admin-cred             kubernetes.io/basic-auth              2      10m
sample-opensearch-ca-cert                kubernetes.io/tls                     2      10m
sample-opensearch-config                 Opaque                                3      10m
sample-opensearch-kibanaro-cred          kubernetes.io/basic-auth              2      10m
sample-opensearch-kibanaserver-cred      kubernetes.io/basic-auth              2      10m
sample-opensearch-logstash-cred          kubernetes.io/basic-auth              2      10m
sample-opensearch-readall-cred           kubernetes.io/basic-auth              2      10m
sample-opensearch-snapshotrestore-cred   kubernetes.io/basic-auth              2      10m
sample-opensearch-token-zbn46            kubernetes.io/service-account-token   3      10m
sample-opensearch-transport-cert         kubernetes.io/tls                     3      10m
```
Now, we can connect to the database with any of these secret that have the prefix `cred`. Here, we are using `sample-opensearch-admin-cred` which contains the admin level credentials to connect with the database.


### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

```bash
$ kubectl get secret -n demo sample-opensearch-admin-cred -o jsonpath='{.data.username}' | base64 -d
admin
$ kubectl get secret -n demo sample-opensearch-admin-cred -o jsonpath='{.data.password}' | base64 -d
9aHT*ZhEK_qjPS~v
```

Then login and check the health of our OpenSearch database.

```bash
$ curl -XGET -k -u 'admin:9aHT*ZhEK_qjPS~v' "https://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "sample-opensearch",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "discovered_master" : true,
  "active_primary_shards" : 1,
  "active_shards" : 3,
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

Now, insert some data into OpenSearch:

```bash
$ curl -XPOST -k --user 'admin:9aHT*ZhEK_qjPS~v' "https://localhost:9200/bands/_doc?pretty" -H 'Content-Type: application/json' -d'
{
    "Name": "Backstreet Boys",
    "Album": "Millennium",
    "Song": "Show Me The Meaning"
}
'
```

Let’s verify that the index have been created successfully.

```bash
$ curl -XGET -k --user 'admin:9aHT*ZhEK_qjPS~v' "https://localhost:9200/_cat/indices?v&s=index&pretty"
health status index                        uuid                   pri rep docs.count docs.deleted store.size pri.store.size
green  open   .opendistro_security         ARYAKuVwQsKel2_0Fl3H2w   1   2          9            0    150.3kb         59.9kb
green  open   bands                        1z6Moj6XS12tpDwFPZpqYw   1   1          1            0     10.4kb          5.2kb
green  open   security-auditlog-2022.02.10 j8-mj4o_SKqCD1g-Nz2PAA   1   1          5            0    183.2kb         91.6kb
```
Also, let’s verify the data in the indexes:

```bash
$ curl -XGET -k --user 'admin:9aHT*ZhEK_qjPS~v' "https://localhost:9200/bands/_search?pretty"
{
  "took" : 183,
  "timed_out" : false,
  "_shards" : {
    "total" : 1,
    "successful" : 1,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 1,
      "relation" : "eq"
    },
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "bands",
        "_type" : "_doc",
        "_id" : "V1xW4n4BfiOqQRjndUdv",
        "_score" : 1.0,
        "_source" : {
          "Name" : "Backstreet Boys",
          "Album" : "Millennium",
          "Song" : "Show Me The Meaning"
        }
      }
    ]
  }
}

```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo elasticsearch sample-opensearch -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
elasticsearch.kubedb.com/sample-opensearch patched

$ kubectl delete -n demo es/sample-opensearch
elasticsearch.kubedb.com "sample-opensearch" deleted

$ kubectl delete namespace demo
namespace "demo" deleted
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for the production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if the database pod fails. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purposes, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `terminationPolicy: WipeOut`**. It is nice to be able to resume the database from the previous one. So, we preserve all your `PVCs` and auth `Secrets`. If you don't want to resume the database, you can just use `spec.terminationPolicy: WipeOut`. It will clean up every resouce that was created with the Elasticsearch CR. For more details, please visit [here](/docs/guides/elasticsearch/concepts/elasticsearch/index.md#specterminationpolicy).

## Next Steps

- Learn about [backup & restore](/docs/guides/elasticsearch/backup/overview/index.md) OpenSearch database using Stash.
- [Quickstart OpenSearch-Dashboards](/docs/guides/elasticsearch/elasticsearch-dashboard/opensearch-dashboards/index.md) with KubeDB Operator.
- Monitor your OpenSearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your OpenSearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).