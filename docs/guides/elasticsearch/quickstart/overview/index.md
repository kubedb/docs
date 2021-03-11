---
title: Elasticsearch Quickstart
menu:
  docs_{{ .version }}:
    identifier: es-quickstart-quickstart
    name: Overview
    parent: es-quickstart-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch QuickStart

This tutorial will show you how to use KubeDB to run an Elasticsearch database.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/elasticsearch/quickstart/overview/images/Lifecycle-of-an-Elasticsearch-CRD.svg">
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

> We have designed this tutorial to demonstrate a production setup of KubeDB managed Elasticsearch. If you just want to try out KubeDB, you can bypass some of the safety features following the tips [here](/docs/guides/elasticsearch/quickstart/overview/index.md#tips-for-testing).

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
NAME                   VERSION   DB_IMAGE                                 AUTH_PLUGIN   DEPRECATED   AGE
opendistro-1.0.2       7.0.1     kubedb/elasticsearch:1.0.2-opendistro    OpenDistro                 5m17s
opendistro-1.0.2-v1    7.0.1     kubedb/elasticsearch:1.0.2-opendistro    OpenDistro                 5m17s
opendistro-1.1.0       7.1.1     kubedb/elasticsearch:1.1.0-opendistro    OpenDistro                 5m17s
opendistro-1.1.0-v1    7.1.1     kubedb/elasticsearch:1.1.0-opendistro    OpenDistro                 5m17s
opendistro-1.10.1      7.9.1     kubedb/elasticsearch:1.10.1-opendistro   OpenDistro                 5m17s
opendistro-1.11.0      7.9.1     kubedb/elasticsearch:1.11.0-opendistro   OpenDistro                 5m17s
opendistro-1.12.0      7.10.0    kubedb/elasticsearch:1.12.0-opendistro   OpenDistro                 5m17s
opendistro-1.2.1       7.2.1     kubedb/elasticsearch:1.2.1-opendistro    OpenDistro                 5m17s
opendistro-1.2.1-v1    7.2.1     kubedb/elasticsearch:1.2.1-opendistro    OpenDistro                 5m17s
opendistro-1.3.0       7.3.2     kubedb/elasticsearch:1.3.0-opendistro    OpenDistro                 5m17s
opendistro-1.3.0-v1    7.3.2     kubedb/elasticsearch:1.3.0-opendistro    OpenDistro                 5m17s
opendistro-1.4.0       7.4.2     kubedb/elasticsearch:1.4.0-opendistro    OpenDistro                 5m17s
opendistro-1.4.0-v1    7.4.2     kubedb/elasticsearch:1.4.0-opendistro    OpenDistro                 5m17s
opendistro-1.6.0       7.6.1     kubedb/elasticsearch:1.6.0-opendistro    OpenDistro                 5m17s
opendistro-1.6.0-v1    7.6.1     kubedb/elasticsearch:1.6.0-opendistro    OpenDistro                 5m17s
opendistro-1.7.0       7.6.1     kubedb/elasticsearch:1.7.0-opendistro    OpenDistro                 5m17s
opendistro-1.7.0-v1    7.6.1     kubedb/elasticsearch:1.7.0-opendistro    OpenDistro                 5m17s
opendistro-1.8.0       7.7.0     kubedb/elasticsearch:1.8.0-opendistro    OpenDistro                 5m17s
opendistro-1.8.0-v1    7.7.0     kubedb/elasticsearch:1.8.0-opendistro    OpenDistro                 5m17s
opendistro-1.9.0       7.8.0     kubedb/elasticsearch:1.9.0-opendistro    OpenDistro                 5m17s
opendistro-1.9.0-v1    7.8.0     kubedb/elasticsearch:1.9.0-opendistro    OpenDistro                 5m17s
searchguard-6.8.1      6.8.1     kubedb/elasticsearch:6.8.1-searchguard   SearchGuard                5m17s
searchguard-6.8.1-v1   6.8.1     kubedb/elasticsearch:6.8.1-searchguard   SearchGuard                5m17s
searchguard-7.0.1      7.0.1     kubedb/elasticsearch:7.0.1-searchguard   SearchGuard                5m17s
searchguard-7.0.1-v1   7.0.1     kubedb/elasticsearch:7.0.1-searchguard   SearchGuard                5m17s
searchguard-7.1.1      7.1.1     kubedb/elasticsearch:7.1.1-searchguard   SearchGuard                5m17s
searchguard-7.1.1-v1   7.1.1     kubedb/elasticsearch:7.1.1-searchguard   SearchGuard                5m17s
searchguard-7.3.2      7.3.2     kubedb/elasticsearch:7.3.2-searchguard   SearchGuard                5m17s
searchguard-7.5.2      7.5.2     kubedb/elasticsearch:7.5.2-searchguard   SearchGuard                5m17s
searchguard-7.5.2-v1   7.5.2     kubedb/elasticsearch:7.5.2-searchguard   SearchGuard                5m17s
searchguard-7.8.1      7.8.1     kubedb/elasticsearch:7.8.1-searchguard   SearchGuard                5m17s
searchguard-7.9.3      7.9.3     kubedb/elasticsearch:7.9.3-searchguard   SearchGuard                5m17s
xpack-6.8.10           6.8.10    kubedb/elasticsearch:6.8.10-xpack        X-Pack                     5m17s
xpack-6.8.10-v1        6.8.10    kubedb/elasticsearch:6.8.10-xpack        X-Pack                     5m17s
xpack-7.0.1            7.0.1     kubedb/elasticsearch:7.0.1-xpack         X-Pack                     5m17s
xpack-7.0.1-v1         7.0.1     kubedb/elasticsearch:7.0.1-xpack         X-Pack                     5m17s
xpack-7.1.1            7.1.1     kubedb/elasticsearch:7.1.1-xpack         X-Pack                     5m17s
xpack-7.1.1-v1         7.1.1     kubedb/elasticsearch:7.1.1-xpack         X-Pack                     5m17s
xpack-7.2.1            7.2.1     kubedb/elasticsearch:7.2.1-xpack         X-Pack                     5m17s
xpack-7.2.1-v1         7.2.1     kubedb/elasticsearch:7.2.1-xpack         X-Pack                     5m17s
xpack-7.3.2            7.3.2     kubedb/elasticsearch:7.3.2-xpack         X-Pack                     5m17s
xpack-7.3.2-v1         7.3.2     kubedb/elasticsearch:7.3.2-xpack         X-Pack                     5m17s
xpack-7.4.2            7.4.2     kubedb/elasticsearch:7.4.2-xpack         X-Pack                     5m17s
xpack-7.4.2-v1         7.4.2     kubedb/elasticsearch:7.4.2-xpack         X-Pack                     5m17s
xpack-7.5.2            7.5.2     kubedb/elasticsearch:7.5.2-xpack         X-Pack                     5m17s
xpack-7.5.2-v1         7.5.2     kubedb/elasticsearch:7.5.2-xpack         X-Pack                     5m17s
xpack-7.6.2            7.6.2     kubedb/elasticsearch:7.6.2-xpack         X-Pack                     5m17s
xpack-7.6.2-v1         7.6.2     kubedb/elasticsearch:7.6.2-xpack         X-Pack                     5m17s
xpack-7.7.1            7.7.1     kubedb/elasticsearch:7.7.1-xpack         X-Pack                     5m17s
xpack-7.7.1-v1         7.7.1     kubedb/elasticsearch:7.7.1-xpack         X-Pack                     5m17s
xpack-7.8.0            7.8.0     kubedb/elasticsearch:7.8.0-xpack         X-Pack                     5m17s
xpack-7.8.0-v1         7.8.0     kubedb/elasticsearch:7.8.0-xpack         X-Pack                     5m17s
xpack-7.9.1            7.9.1     kubedb/elasticsearch:7.9.1-xpack         X-Pack                     5m17s
xpack-7.9.1-v1         7.9.1     kubedb/elasticsearch:7.9.1-xpack         X-Pack                     5m17s
```

Notice the `DEPRECATED` column. Here, `true` means that this ElasticsearchVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated ElasticsearchVersion.

In this tutorial, we will use `xpack-7.9.1-v1` ElasticsearchVersion CR to create an Elasticsearch cluster.

> Note: An image with a higher modification tag will have more features and fixes than an image with a lower modification tag. Hence, it is recommended to use ElasticsearchVersion CRD with the highest modification tag to take advantage of the latest features. For example, we are using `xpack-7.9.1-v1` over `7.9.1-xpack`.

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
  version: xpack-7.9.1-v1
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

- `spec.version` - is the name of the ElasticsearchVersion CR. Here, an Elasticsearch of version `7.9.1` will be created with `x-pack` security plugin.
- `spec.enableSSL` - specifies whether the HTTP layer is secured with certificates or not.
- `spec.replicas` - specifies the number of Elasticsearch nodes.
- `spec.storageType` - specifies the type of storage that will be used for Elasticsearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the Elasticsearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by the KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. If you don't specify `spec.storageType: Ephemeral`, then this field is required.
- `spec.terminationPolicy` specifies what KubeDB should do when a user try to delete Elasticsearch CR. Termination policy `DoNotTerminate` prevents a user from deleting this object if the admission webhook is enabled.

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in the `storage.resources.requests` field. Don't specify `limits` here. PVC does not get resized automatically.

Let's create the Elasticsearch CR that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/quickstart/overview/yamls/elasticsearch.yaml
elasticsearch.kubedb.com/es-quickstart created
```

The Elasticsearch's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the database.

```bash
$ kubectl get elasticsearch -n demo -w
NAME            VERSION          STATUS         AGE
es-quickstart   xpack-7.9.1-v1   Provisioning   1m34s
... ...
es-quickstart   xpack-7.9.1-v1   Ready          2m6s
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
  Creation Timestamp:  2021-02-26T05:29:30Z
  Finalizers:
    kubedb.com
  Generation:  2
  Resource Version:  13279
  UID:               04715366-cda8-4e44-b63b-8e722b986bbd
Spec:
  Auth Secret:
    Name:      es-quickstart-elastic-cred
  Enable SSL:  true
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
      Resources:
        Limits:
          Cpu:     500m
          Memory:  1Gi
        Requests:
          Cpu:               500m
          Memory:            1Gi
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
  Termination Policy:    DoNotTerminate
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
      Alias:  archiver
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-quickstart-archiver-cert
      Subject:
        Organizations:
          kubedb
  Version:  xpack-7.9.1-v1
Status:
  Conditions:
    Last Transition Time:  2021-02-26T05:29:30Z
    Message:               The KubeDB operator has started the provisioning of Elasticsearch: demo/es-quickstart
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2021-02-26T05:32:44Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2021-02-26T05:34:35Z
    Message:               The Elasticsearch: demo/es-quickstart is accepting client requests.
    Observed Generation:   2
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2021-02-26T05:34:36Z
    Message:               The Elasticsearch: demo/es-quickstart is ready.
    Observed Generation:   2
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2021-02-26T05:34:36Z
    Message:               The Elasticsearch: demo/es-quickstart is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     2
  Phase:                   Ready
Events:
  Type    Reason      Age    From                    Message
  ----    ------      ----   ----                    -------
  Normal  Successful  54m    Elasticsearch operator  Successfully  governing service
  Normal  Successful  54m    Elasticsearch operator  Successfully patched Elasticsearch
```

### KubeDB Operator Generated Resources

On deployment of an Elasticsearch CR, the operator creates the following resources:

```bash
$ kubectl get all,secret -n demo -l 'app.kubernetes.io/instance=es-quickstart'
NAME                  READY   STATUS    RESTARTS   AGE
pod/es-quickstart-0   1/1     Running   0          3h42m
pod/es-quickstart-1   1/1     Running   0          3h39m
pod/es-quickstart-2   1/1     Running   0          3h39m

NAME                           TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/es-quickstart          ClusterIP   10.96.239.14   <none>        9200/TCP   3h42m
service/es-quickstart-master   ClusterIP   None           <none>        9300/TCP   3h42m
service/es-quickstart-pods     ClusterIP   None           <none>        9200/TCP   3h42m

NAME                             READY   AGE
statefulset.apps/es-quickstart   3/3     3h42m

NAME                                               TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-quickstart   kubedb.com/elasticsearch   7.9.1     3h42m

NAME                                  TYPE                       DATA   AGE
secret/es-quickstart-archiver-cert    kubernetes.io/tls          3      3h42m
secret/es-quickstart-ca-cert          kubernetes.io/tls          2      3h42m
secret/es-quickstart-config           Opaque                     1      3h42m
secret/es-quickstart-elastic-cred     kubernetes.io/basic-auth   2      3h42m
secret/es-quickstart-http-cert        kubernetes.io/tls          3      3h42m
secret/es-quickstart-transport-cert   kubernetes.io/tls          3      3h42m
```

- `StatefulSet` - a StatefulSet named after the Elasticsearch instance. In topology mode, the operator creates 3 statefulSets with name `{Elasticsearch-Name}-{Sufix}`.
- `Services` -  3 services are generated for each Elasticsearch database.
  - `{Elasticsearch-Name}` - the client service which is used to connect to the database. It points to the `ingest` nodes.
  - `{Elasticsearch-Name}-master` - the master service which is used to connect to the master nodes. It is a headless service.
  - `{Elasticsearch-Name}-pods` - the node discovery service which is used by the Elasticsearch nodes to communicate each other. It is a headless service.
- `AppBinding` - an [AppBinding](/docs/guides/elasticsearch/concepts/appbinding/index.md) which hold the connect information for the database. It is also named after the Elastics
- `Secrets` - 3 types of secrets are generated for each Elasticsearch database.
  - `{Elasticsearch-Name}-{username}-cred` - the auth secrets which hold the `username` and `password` for the Elasticsearch users.
  - `{Elasticsearch-Name}-{alias}-cert` - the certificate secrets which hold `tls.crt`, `tls.key`, and `ca.crt` for configuring the Elasticsearch database.
  - `{Elasticsearch-Name}-config` - the default configuration secret created by the operator.

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
  q6XreFWkWi$;BsQy
  ```

Now let's check the health of our Elasticsearch database.

```bash
$ curl -XGET -k -u 'elastic:q6XreFWkWi$;BsQy' "https://localhost:9200/_cluster/health?pretty"

{
  "cluster_name" : "es-quickstart",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 0,
  "active_shards" : 0,
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

In this tutorial, Elasticsearch `es-quickstart` is created with `spec.terminationPolicy: DoNotTerminate`. So if you try to delete this Elasticsearch object, the admission webhook will nullify the delete operation.

```bash
$ kubectl delete elasticsearch -n demo es-quickstart 
Error from server (BadRequest): admission webhook "elasticsearch.validators.kubedb.com" denied the request: elasticsearch "demo/es-quickstart" can't be terminated. To delete, change spec.terminationPolicy
```

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
NAME                                TYPE                       DATA   AGE
secret/es-quickstart-elastic-cred   kubernetes.io/basic-auth   2      6h48m

NAME                                         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-es-quickstart-0   Bound    pvc-9208f770-7308-45b3-a23e-233590087f45   1Gi        RWO            standard       6h48m
persistentvolumeclaim/data-es-quickstart-1   Bound    pvc-0f12a74e-ba80-4e67-bece-3a00c3fcd28f   1Gi        RWO            standard       6h45m
persistentvolumeclaim/data-es-quickstart-2   Bound    pvc-6609582b-8988-4efb-8a4b-5b2757fd6066   1Gi        RWO            standard       6h45m
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

- Learn about [backup & restore](/docs/guides/elasticsearch/backup/stash.md) Elasticsearch database using Stash.
- Learn how to configure [Elasticsearch Topology](/docs/guides/elasticsearch/clustering/topology-cluster/index.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
