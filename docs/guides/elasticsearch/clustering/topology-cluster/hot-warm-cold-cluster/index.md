---
title: Elasticsearch Hot-Warm-Cold Cluster
menu:
  docs_{{ .version }}:
    identifier: es-hot-warm-cold-cluster
    name: Hot-Warm-Cold Cluster
    parent: es-topology-cluster
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Hot-Warm-Cold Cluster

Hot-warm-cold architectures are common for time series data such as logging or metrics and it also has various use cases too. For example, assume Elasticsearch is being used to aggregate log files from multiple systems. Logs from today are actively being indexed and this week's logs are the most heavily searched (hot). Last week's logs may be searched but not as much as the current week's logs (warm). Last month's logs may or may not be searched often, but are good to keep around just in case (cold).

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install the KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   14s
```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/elasticsearch/clustering/topology-cluster/hot-warm-cold-cluster/yamls) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will have to provide `StorageClass` in Elasticsearch CR specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                                    PROVISIONER               RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)                      rancher.io/local-path     Delete          WaitForFirstConsumer   false                  10m
linode-block-storage                    linodebs.csi.linode.com   Delete          Immediate              true                   10m
linode-block-storage-retain (default)   linodebs.csi.linode.com   Retain          Immediate              true                   10m
```

Here, we use `linode-block-storage` as StorageClass in this demo.

## Create Elasticsearch Hot-Warm-Cold Cluster

We are going to create a Elasticsearch Hot-Warm-Cold cluster in topology mode. Our cluster will be consist of 2 master nodes, 2 ingest nodes, 1 data content node, 3 data hot nodes, 2 data warm node, and 2 data cold nodes. Here, we are using Elasticsearch version (`xpack-7.16.2`) of ElasticStack distribution for this demo. To learn more about the Elasticsearch CR, visit [here](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-cluster
  namespace: demo
spec:
  enableSSL: true
  version: xpack-7.16.2
  topology:
      master:
        replicas: 2
        storage:
          resources:
            requests:
              storage: 1Gi
          storageClassName: "linode-block-storage"
      ingest:
        replicas: 2
        storage:
          resources:
            requests:
              storage: 1Gi
          storageClassName: "linode-block-storage"
      dataContent:
        replicas: 1
        storage:
          resources:
            requests:
              storage: 5Gi
          storageClassName: "linode-block-storage"
      dataHot:
        replicas: 3
        storage:
          resources:
            requests:
              storage: 3Gi
          storageClassName: "linode-block-storage"
      dataWarm:
        replicas: 2
        storage:
          resources:
            requests:
              storage: 5Gi
          storageClassName: "linode-block-storage"
      dataCold:
        replicas: 2
        storage:
          resources:
            requests:
              storage: 5Gi
          storageClassName: "linode-block-storage"

```

Here,

- `spec.version` - is the name of the ElasticsearchVersion CR. Here, we are using Elasticsearch version `xpack-7.16.2` of ElasticStack distribution.
- `spec.enableSSL` - specifies whether the HTTP layer is secured with certificates or not.
- `spec.storageType` - specifies the type of storage that will be used for Elasticsearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the Elasticsearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.topology` - specifies the node-specific properties for the Elasticsearch cluster.
  - `topology.master` - specifies the properties of [master](https://www.elastic.co/guide/en/elasticsearch/reference/7.16/modules-node.html#master-node) nodes.
    - `master.replicas` - specifies the number of master nodes.
    - `master.storage` - specifies the master node storage information that passed to the StatefulSet.
  - `topology.dataContent` - specifies the properties of [data content](https://www.elastic.co/guide/en/elasticsearch/reference/7.16/modules-node.html#data-content-node) node.
    - `dataContent.replicas` - specifies the number of data content node.
    - `dataContent.storage` - specifies the data content node storage information that passed to the StatefulSet.
  - `topology.ingest` - specifies the properties of [ingest](https://www.elastic.co/guide/en/elasticsearch/reference/7.16/modules-node.html#node-ingest-node) nodes.
    - `ingest.replicas` - specifies the number of ingest nodes.
    - `ingest.storage` - specifies the ingest node storage information that passed to the StatefulSet.
  - `topology.dataHot` - specifies the properties of [dataHot](https://www.elastic.co/guide/en/elasticsearch/reference/7.16/modules-node.html#data-hot-node) nodes.
    - `dataHot.replicas` - specifies the number of dataHot nodes.
    - `dataHot.storage` - specifies the dataHot node storage information that passed to the StatefulSet.
  - `topology.dataWarm` - specifies the properties of [dataWarm](https://www.elastic.co/guide/en/elasticsearch/reference/7.16/modules-node.html#data-warm-node) nodes.
    - `dataWarm.replicas` - specifies the number of dataWarm nodes.
    - `dataWarm.storage` - specifies the dataWarm node storage information that passed to the StatefulSet.
  - `topology.dataCold` - specifies the properties of [dataCold](https://www.elastic.co/guide/en/elasticsearch/reference/7.16/modules-node.html#data-cold-node) nodes.
    - `dataCold.replicas` - specifies the number of dataCold nodes.
    - `dataCold.storage` - specifies the dataCold node storage information that passed to the StatefulSet.
> Here, we use `linode-block-storage` as storage for every node. But it is recommended to prioritize faster storage for `dataHot` node then `dataWarm` and finally `dataCold`.

Let's deploy the above example by the following command:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/clustering/topology-cluster/hot-warm-cold-cluster/yamls/es-cluster.yaml
elasticsearch.kubedb.com/es-cluster created
```

KubeDB will create the necessary resources to deploy the Elasticsearch cluster according to the above specification. Let’s wait until the database to be ready to use,

```bash
$ watch kubectl get elasticsearch -n demo
NAME         VERSION        STATUS   AGE
es-cluster   xpack-7.16.2   Ready    2m48s
```
Here, Elasticsearch is in `Ready` state. It means the database is ready to accept connections.

Describe the Elasticsearch object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe elasticsearch -n demo es-cluster
Name:         es-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Elasticsearch
Metadata:
  Creation Timestamp:  2022-03-14T06:33:20Z
  Finalizers:
    kubedb.com
  Generation:  2
  Resource Version:  20467655
  UID:               236fd414-9d94-4fce-93d3-7891fcf7f6a4
Spec:
  Auth Secret:
    Name:                es-cluster-elastic-cred
  Enable SSL:            true
  Heap Size Percentage:  50
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
                Match Expressions:
                  Key:       ${NODE_ROLE}
                  Operator:  Exists
                Match Labels:
                  app.kubernetes.io/instance:    es-cluster
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        elasticsearches.kubedb.com
              Namespaces:
                demo
              Topology Key:  kubernetes.io/hostname
            Weight:          100
            Pod Affinity Term:
              Label Selector:
                Match Expressions:
                  Key:       ${NODE_ROLE}
                  Operator:  Exists
                Match Labels:
                  app.kubernetes.io/instance:    es-cluster
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
      Service Account Name:  es-cluster
  Storage Type:              Durable
  Termination Policy:        Delete
  Tls:
    Certificates:
      Alias:  ca
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-cluster-ca-cert
      Subject:
        Organizations:
          kubedb
      Alias:  transport
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-cluster-transport-cert
      Subject:
        Organizations:
          kubedb
      Alias:  http
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-cluster-http-cert
      Subject:
        Organizations:
          kubedb
      Alias:  archiver
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-cluster-archiver-cert
      Subject:
        Organizations:
          kubedb
  Topology:
    Data Cold:
      Replicas:  2
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Storage:
        Resources:
          Requests:
            Storage:         5Gi
        Storage Class Name:  linode-block-storage
      Suffix:                data-cold
    Data Content:
      Replicas:  1
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Storage:
        Resources:
          Requests:
            Storage:         5Gi
        Storage Class Name:  linode-block-storage
      Suffix:                data-content
    Data Hot:
      Replicas:  3
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Storage:
        Resources:
          Requests:
            Storage:         3Gi
        Storage Class Name:  linode-block-storage
      Suffix:                data-hot
    Data Warm:
      Replicas:  2
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Storage:
        Resources:
          Requests:
            Storage:         5Gi
        Storage Class Name:  linode-block-storage
      Suffix:                data-warm
    Ingest:
      Replicas:  2
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Storage:
        Resources:
          Requests:
            Storage:         1Gi
        Storage Class Name:  linode-block-storage
      Suffix:                ingest
    Master:
      Replicas:  2
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Storage:
        Resources:
          Requests:
            Storage:         1Gi
        Storage Class Name:  linode-block-storage
      Suffix:                master
  Version:                   xpack-7.16.2
Status:
  Conditions:
    Last Transition Time:  2022-03-14T06:33:20Z
    Message:               The KubeDB operator has started the provisioning of Elasticsearch: demo/es-cluster
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2022-03-14T06:34:55Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2022-03-14T06:35:17Z
    Message:               The Elasticsearch: demo/es-cluster is accepting client requests.
    Observed Generation:   2
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2022-03-14T06:35:27Z
    Message:               The Elasticsearch: demo/es-cluster is ready.
    Observed Generation:   2
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2022-03-14T06:35:28Z
    Message:               The Elasticsearch: demo/es-cluster is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     2
  Phase:                   Ready
Events:
  Type    Reason      Age    From             Message
  ----    ------      ----   ----             -------
  Normal  Successful  3m29s  KubeDB Operator  Successfully created governing service
  Normal  Successful  3m29s  KubeDB Operator  Successfully created Service
  Normal  Successful  3m29s  KubeDB Operator  Successfully created Service
  Normal  Successful  3m27s  KubeDB Operator  Successfully created Elasticsearch
  Normal  Successful  3m26s  KubeDB Operator  Successfully created appbinding
  Normal  Successful  3m26s  KubeDB Operator  Successfully governing service
```
- Here, in `Status.Conditions` 
  - `Conditions.Status` is `True` for the `Condition.Type:ProvisioningStarted` which means database provisioning has been started successfully.
  - `Conditions.Status` is `True` for the `Condition.Type:ReplicaReady` which specifies all replicas are ready in the cluster.
  - `Conditions.Status` is `True` for the `Condition.Type:AcceptingConnection` which means database has been accepting connection request.
  - `Conditions.Status` is `True` for the `Condition.Type:Ready` which defines database is ready to use.
  - `Conditions.Status` is `True` for the `Condition.Type:Provisioned` which specifies Database has been successfully provisioned.

### KubeDB Operator Generated Resources

Let's check the Kubernetes resources created by the operator on the deployment of Elasticsearch CRO:

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=es-cluster'
NAME                            READY   STATUS    RESTARTS   AGE
pod/es-cluster-data-cold-0      1/1     Running   0          5m46s
pod/es-cluster-data-cold-1      1/1     Running   0          4m51s
pod/es-cluster-data-content-0   1/1     Running   0          5m46s
pod/es-cluster-data-hot-0       1/1     Running   0          5m46s
pod/es-cluster-data-hot-1       1/1     Running   0          5m9s
pod/es-cluster-data-hot-2       1/1     Running   0          4m41s
pod/es-cluster-data-warm-0      1/1     Running   0          5m46s
pod/es-cluster-data-warm-1      1/1     Running   0          4m52s
pod/es-cluster-ingest-0         1/1     Running   0          5m46s
pod/es-cluster-ingest-1         1/1     Running   0          5m14s
pod/es-cluster-master-0         1/1     Running   0          5m46s
pod/es-cluster-master-1         1/1     Running   0          4m50s

NAME                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/es-cluster          ClusterIP   10.128.132.28   <none>        9200/TCP   5m50s
service/es-cluster-master   ClusterIP   None            <none>        9300/TCP   5m50s
service/es-cluster-pods     ClusterIP   None            <none>        9200/TCP   5m50s

NAME                                       READY   AGE
statefulset.apps/es-cluster-data-cold      2/2     5m48s
statefulset.apps/es-cluster-data-content   1/1     5m48s
statefulset.apps/es-cluster-data-hot       3/3     5m48s
statefulset.apps/es-cluster-data-warm      2/2     5m48s
statefulset.apps/es-cluster-ingest         2/2     5m48s
statefulset.apps/es-cluster-master         2/2     5m48s

NAME                                            TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-cluster   kubedb.com/elasticsearch   7.16.2    5m49s

NAME                               TYPE                       DATA   AGE
secret/es-cluster-archiver-cert    kubernetes.io/tls          3      5m51s
secret/es-cluster-ca-cert          kubernetes.io/tls          2      5m51s
secret/es-cluster-config           Opaque                     1      5m50s
secret/es-cluster-elastic-cred     kubernetes.io/basic-auth   2      5m51s
secret/es-cluster-http-cert        kubernetes.io/tls          3      5m51s
secret/es-cluster-transport-cert   kubernetes.io/tls          3      5m51s

NAME                                                   STATUS   VOLUME                 CAPACITY   ACCESS MODES   STORAGECLASS           AGE
persistentvolumeclaim/data-es-cluster-data-cold-0      Bound    pvc-47585d52c11a4a52   10Gi       RWO            linode-block-storage   5m50s
persistentvolumeclaim/data-es-cluster-data-cold-1      Bound    pvc-66aaa122c5774713   10Gi       RWO            linode-block-storage   4m55s
persistentvolumeclaim/data-es-cluster-data-content-0   Bound    pvc-d51361e9352b4e9f   10Gi       RWO            linode-block-storage   5m50s
persistentvolumeclaim/data-es-cluster-data-hot-0       Bound    pvc-3712187a3c6540da   10Gi       RWO            linode-block-storage   5m50s
persistentvolumeclaim/data-es-cluster-data-hot-1       Bound    pvc-2318d4eacb4b453f   10Gi       RWO            linode-block-storage   5m13s
persistentvolumeclaim/data-es-cluster-data-hot-2       Bound    pvc-c309c7058b114578   10Gi       RWO            linode-block-storage   4m45s
persistentvolumeclaim/data-es-cluster-data-warm-0      Bound    pvc-d5950f5b075c4d3f   10Gi       RWO            linode-block-storage   5m50s
persistentvolumeclaim/data-es-cluster-data-warm-1      Bound    pvc-3f6b99d11b1d46ea   10Gi       RWO            linode-block-storage   4m56s
persistentvolumeclaim/data-es-cluster-ingest-0         Bound    pvc-081be753a20a45da   10Gi       RWO            linode-block-storage   5m50s
persistentvolumeclaim/data-es-cluster-ingest-1         Bound    pvc-1bea5a3b5be24817   10Gi       RWO            linode-block-storage   5m18s
persistentvolumeclaim/data-es-cluster-master-0         Bound    pvc-2c49a2ccb4644d6e   10Gi       RWO            linode-block-storage   5m50s
persistentvolumeclaim/data-es-cluster-master-1         Bound    pvc-cb1d970febff498f   10Gi       RWO            linode-block-storage   4m54s

```

- `StatefulSet` - 6 StatefulSets are created for 6 types Elasticsearch nodes. The StatefulSets are named after the Elasticsearch instance with given suffix: `{Elasticsearch-Name}-{Sufix}`.
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

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get service -n demo
NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
es-cluster             ClusterIP   10.128.132.28   <none>        9200/TCP   10m
es-cluster-dashboard   ClusterIP   10.128.99.51    <none>        5601/TCP   10m
es-cluster-master      ClusterIP   None            <none>        9300/TCP   10m
es-cluster-pods        ClusterIP   None            <none>        9200/TCP   10m
```
Here, we are going to use `es-cluster` Service to connect with the database. Now, let’s port-forward the `es-cluster` Service to the port `9200` to local machine:

```bash
$ kubectl port-forward -n demo svc/es-cluster 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```
Now, our Elasticsearch cluster is accessible at `localhost:9200`.

#### Export the Credentials

KubeDB also create some Secrets for the database. Let’s check which Secrets have been created by KubeDB for our `es-cluster`.

```bash
$ kubectl get secret -n demo | grep es-cluster
es-cluster-archiver-cert                  kubernetes.io/tls                     3      12m
es-cluster-ca-cert                        kubernetes.io/tls                     2      12m
es-cluster-config                         Opaque                                1      12m
es-cluster-dashboard-ca-cert              kubernetes.io/tls                     2      12m
es-cluster-dashboard-config               Opaque                                1      12m
es-cluster-dashboard-kibana-server-cert   kubernetes.io/tls                     3      12m
es-cluster-elastic-cred                   kubernetes.io/basic-auth              2      12m
es-cluster-http-cert                      kubernetes.io/tls                     3      12m
es-cluster-token-v97c7                    kubernetes.io/service-account-token   3      12m
es-cluster-transport-cert                 kubernetes.io/tls                     3      12m
```
Now, we can connect to the database with `es-cluster-elastic-cred` which contains the admin level credentials to connect with the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

```bash
$ kubectl get secret -n demo es-cluster-elastic-cred -o jsonpath='{.data.username}' | base64 -d
elastic
$ kubectl get secret -n demo es-cluster-elastic-cred -o jsonpath='{.data.password}' | base64 -d
YQB)~K6M9U)d_yVu
```

Now, let's check the health of our Elasticsearch cluster

```bash
# curl -XGET -k -u 'username:password' https://localhost:9200/_cluster/health?pretty"
$ curl -XGET -k -u 'elastic:YQB)~K6M9U)d_yVu' "https://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "es-cluster",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 12,
  "number_of_data_nodes" : 8,
  "active_primary_shards" : 9,
  "active_shards" : 10,
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

### Verify Node Role

As we have assigned a dedicated role to each type of node, let's verify them by following command,

```bash
$ curl -XGET -k -u 'elastic:YQB)~K6M9U)d_yVu' "https://localhost:9200/_cat/nodes?v"
ip        heap.percent ram.percent cpu load_1m load_5m load_15m node.role master name
10.2.2.30           41          90   3    0.22    0.31     0.34 s         -      es-cluster-data-content-0
10.2.1.28           70          76   3    0.00    0.03     0.07 h         -      es-cluster-data-hot-0
10.2.0.28           45          87   4    0.09    0.20     0.26 i         -      es-cluster-ingest-0
10.2.2.29           33          75   3    0.22    0.31     0.34 w         -      es-cluster-data-warm-0
10.2.0.29           65          76   3    0.09    0.20     0.26 h         -      es-cluster-data-hot-1
10.2.0.30           46          75   3    0.09    0.20     0.26 c         -      es-cluster-data-cold-1
10.2.1.29           56          77   3    0.00    0.03     0.07 m         *      es-cluster-master-0
10.2.3.50           52          74   3    0.02    0.06     0.11 c         -      es-cluster-data-cold-0
10.2.2.31           34          75   3    0.22    0.31     0.34 m         -      es-cluster-master-1
10.2.1.30           21          74   3    0.00    0.03     0.07 w         -      es-cluster-data-warm-1
10.2.3.49           23          85   3    0.02    0.06     0.11 i         -      es-cluster-ingest-1
10.2.3.51           72          75   3    0.02    0.06     0.11 h         -      es-cluster-data-hot-2

```

- `node.role` field specifies the dedicated role that we have assigned for each type of node. Where `h` refers to the hot node, `w` refers to the warm node, `c` refers to the cold node, `i` refers to the ingest node, `m` refers to the master node, and `s` refers to the content node.
- `master` field specifies the acive master node. Here, we can see a `*` in the `es-cluster-master-0` which shows that it is the active master node now.



## Cleaning Up

To cleanup the k8s resources created by this tutorial, run:

```bash
$ kubectl patch -n demo elasticsearch es-cluster -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"

$ kubectl delete elasticsearch -n demo es-cluster 

# Delete namespace
$ kubectl delete namespace demo
```

## Next Steps

- Learn about [taking backup](/docs/guides/elasticsearch/backup/overview/index.md) of Elasticsearch database using Stash.
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).