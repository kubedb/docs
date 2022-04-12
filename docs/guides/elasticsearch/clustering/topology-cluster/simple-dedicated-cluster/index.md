---
title: Elasticsearch Simple Dedicated Cluster
menu:
  docs_{{ .version }}:
    identifier: es-simple-dedicated-cluster
    name: Simple Dedicated Cluster
    parent: es-topology-cluster
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Simple Dedicated Cluster

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install the KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   7s
```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/elasticsearch/clustering/topology-cluster/simple-dedicated-cluster/yamls) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will have to provide `StorageClass` in Elasticsearch CR specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  1h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Create Elasticsearch Simple Dedicated Cluster

We are going to create a Elasticsearch Simple Dedicated Cluster in topology mode. Our cluster will be consist of 2 master nodes, 3 data nodes, 2 ingest nodes. Here, we are using Elasticsearch version ( `searchguard-7.14.2` ) of SearchGuard distribution for this demo. To learn more about the Elasticsearch CR, visit [here](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-cluster
  namespace: demo
spec:
  enableSSL: true 
  version: searchguard-7.14.2
  storageType: Durable
  topology:
    master:
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Here,

- `spec.version` - is the name of the ElasticsearchVersion CR. Here, we are using Elasticsearch version `searchguard-7.14.2` of SearchGuard distribution.
- `spec.enableSSL` - specifies whether the HTTP layer is secured with certificates or not.
- `spec.storageType` - specifies the type of storage that will be used for Elasticsearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the Elasticsearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.topology` - specifies the node-specific properties for the Elasticsearch cluster.
  - `topology.master` - specifies the properties of master nodes.
    - `master.replicas` - specifies the number of master nodes.
    - `master.storage` - specifies the master node storage information that passed to the StatefulSet.
  - `topology.data` - specifies the properties of data nodes.
    - `data.replicas` - specifies the number of data nodes.
    - `data.storage` - specifies the data node storage information that passed to the StatefulSet.
  - `topology.ingest` - specifies the properties of ingest nodes.
    - `ingest.replicas` - specifies the number of ingest nodes.
    - `ingest.storage` - specifies the ingest node storage information that passed to the StatefulSet.

Let's deploy the above example by the following command:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/clustering/topology-cluster/simple-dedicated-cluster/yamls/es-cluster.yaml
elasticsearch.kubedb.com/es-cluster created
```
KubeDB will create the necessary resources to deploy the Elasticsearch cluster according to the above specification. Let’s wait until the database to be ready to use,

```bash
$ watch kubectl get elasticsearch -n demo
NAME         VERSION              STATUS   AGE
es-cluster   searchguard-7.14.2   Ready    3m32s
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
  Creation Timestamp:  2022-04-07T09:48:51Z
  Finalizers:
    kubedb.com
  Generation:  2
  Resource Version:  406999
  UID:               1dff00c8-5a90-4916-bf8a-ed28f19dd433
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
    Data:
      Replicas:  3
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:         1Gi
        Storage Class Name:  standard
      Suffix:                data
    Ingest:
      Replicas:  2
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:         1Gi
        Storage Class Name:  standard
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
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:         1Gi
        Storage Class Name:  standard
      Suffix:                master
  Version:                   searchguard-7.14.2
Status:
  Conditions:
    Last Transition Time:  2022-04-07T09:48:51Z
    Message:               The KubeDB operator has started the provisioning of Elasticsearch: demo/es-cluster
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2022-04-07T09:51:28Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2022-04-07T09:53:29Z
    Message:               The Elasticsearch: demo/es-cluster is accepting client requests.
    Observed Generation:   2
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2022-04-07T09:54:02Z
    Message:               The Elasticsearch: demo/es-cluster is ready.
    Observed Generation:   2
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2022-04-07T09:53:41Z
    Message:               The Elasticsearch: demo/es-cluster is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     2
  Phase:                   Ready
Events:
  Type    Reason      Age   From             Message
  ----    ------      ----  ----             -------
  Normal  Successful  30m   KubeDB Operator  Successfully created governing service
  Normal  Successful  30m   KubeDB Operator  Successfully created Service
  Normal  Successful  30m   KubeDB Operator  Successfully created Service
  Normal  Successful  30m   KubeDB Operator  Successfully created Elasticsearch
  Normal  Successful  30m   KubeDB Operator  Successfully created appbinding
  Normal  Successful  30m   KubeDB Operator  Successfully  governing service
```
- Here, in `Status.Conditions` 
  - `Conditions.Status` is `True` for the `Condition.Type:ProvisioningStarted` which means database provisioning has been started successfully.
  - `Conditions.Status` is `True` for the `Condition.Type:ReplicaReady` which specifies all replicas are ready in the cluster.
  - `Conditions.Status` is `True` for the `Condition.Type:AcceptingConnection` which means database has been accepting connection request.
  - `Conditions.Status` is `True` for the `Condition.Type:Ready` which defines database is ready to use.
  - `Conditions.Status` is `True` for the `Condition.Type:Provisioned` which specifies Database has been successfully provisioned.

### KubeDB Operator Generated Resources

Let's check the kubernetes resources created by the operator on the deployment of Elasticsearch CRO:

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=es-cluster'
NAME                      READY   STATUS    RESTARTS   AGE
pod/es-cluster-data-0     1/1     Running   0          31m
pod/es-cluster-data-1     1/1     Running   0          29m
pod/es-cluster-data-2     1/1     Running   0          29m
pod/es-cluster-ingest-0   1/1     Running   0          31m
pod/es-cluster-ingest-1   1/1     Running   0          29m
pod/es-cluster-master-0   1/1     Running   0          31m
pod/es-cluster-master-1   1/1     Running   0          29m

NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/es-cluster          ClusterIP   10.96.67.225   <none>        9200/TCP   31m
service/es-cluster-master   ClusterIP   None           <none>        9300/TCP   31m
service/es-cluster-pods     ClusterIP   None           <none>        9200/TCP   31m

NAME                                 READY   AGE
statefulset.apps/es-cluster-data     3/3     31m
statefulset.apps/es-cluster-ingest   2/2     31m
statefulset.apps/es-cluster-master   2/2     31m

NAME                                            TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-cluster   kubedb.com/elasticsearch   7.14.2    31m

NAME                               TYPE                       DATA   AGE
secret/es-cluster-archiver-cert    kubernetes.io/tls          3      31m
secret/es-cluster-ca-cert          kubernetes.io/tls          2      31m
secret/es-cluster-config           Opaque                     1      31m
secret/es-cluster-elastic-cred     kubernetes.io/basic-auth   2      31m
secret/es-cluster-http-cert        kubernetes.io/tls          3      31m
secret/es-cluster-transport-cert   kubernetes.io/tls          3      31m

NAME                                             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-es-cluster-data-0     Bound    pvc-b55f67b3-7c2a-4b16-8cf0-77bbaafef3f7   1Gi        RWO            standard       31m
persistentvolumeclaim/data-es-cluster-data-1     Bound    pvc-62176b5a-5136-450b-afec-f483f041506f   1Gi        RWO            standard       29m
persistentvolumeclaim/data-es-cluster-data-2     Bound    pvc-c36f2ca3-466f-4314-81d7-7c1c0b4acf4f   1Gi        RWO            standard       29m
persistentvolumeclaim/data-es-cluster-ingest-0   Bound    pvc-96a081a1-90ff-4b82-bbf5-3bdf349b7de4   1Gi        RWO            standard       31m
persistentvolumeclaim/data-es-cluster-ingest-1   Bound    pvc-18420ed8-8455-4b18-864b-f13637dade38   1Gi        RWO            standard       29m
persistentvolumeclaim/data-es-cluster-master-0   Bound    pvc-6892422b-e399-44e1-9fdb-884b68fc66b5   1Gi        RWO            standard       31m
persistentvolumeclaim/data-es-cluster-master-1   Bound    pvc-ed4a704c-7b13-421e-85e1-d710e556ca4e   1Gi        RWO            standard       29m

```

- `StatefulSet` - 3 StatefulSets are created for 3 types Elasticsearch nodes. The StatefulSets are named after the Elasticsearch instance with given suffix: `{Elasticsearch-Name}-{Sufix}`.
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
es-cluster             ClusterIP   10.96.67.225    <none>        9200/TCP   11m
es-cluster-master      ClusterIP   None            <none>        9300/TCP   11m
es-cluster-pods        ClusterIP   None            <none>        9200/TCP   11m
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
es-cluster-archiver-cert           kubernetes.io/tls                     3      12m
es-cluster-ca-cert                 kubernetes.io/tls                     2      12m
es-cluster-config                  Opaque                                1      12m
es-cluster-elastic-cred            kubernetes.io/basic-auth              2      12m
es-cluster-http-cert               kubernetes.io/tls                     3      12m
es-cluster-token-hx5mn             kubernetes.io/service-account-token   3      12m
es-cluster-transport-cert          kubernetes.io/tls                     3      12m
```
Now, we can connect to the database with `es-cluster-elastic-cred` which contains the admin level credentials to connect with the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

```bash
$ kubectl get secret -n demo es-cluster-elastic-cred -o jsonpath='{.data.username}' | base64 -d
elastic
$ kubectl get secret -n demo es-cluster-elastic-cred -o jsonpath='{.data.password}' | base64 -d
tS$k!2IBI.ASI7FJ
```

Now, let's check the health of our Elasticsearch cluster

```bash
# curl -XGET -k -u 'username:password' https://localhost:9200/_cluster/health?pretty"
$ curl -XGET -k --user 'elastic:tS$k!2IBI.ASI7FJ' "https://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "es-cluster",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 7,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 1,
  "active_shards" : 2,
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

## Insert Sample Data

Now, we are going to insert some data into Elasticsearch.

```bash
$ curl -XPOST -k --user 'elastic:tS$k!2IBI.ASI7FJ' "https://localhost:9200/info/_doc?pretty" -H 'Content-Type: application/json' -d'
         {
             "Company": "AppsCode Inc",
             "Product": "KubeDB"
         }
         '

```
Now, let’s verify that the index have been created successfully.

```bash
$ curl -XGET -k --user 'elastic:tS$k!2IBI.ASI7FJ' "https://localhost:9200/_cat/indices?v&s=index&pretty"
health status index            uuid                   pri rep docs.count docs.deleted store.size pri.store.size
green  open   .geoip_databases FsJlvTyRSsuRWTpX8OpkOA   1   1         40            0       76mb           38mb
green  open   info             9Z2Cl5fjQWGBAfjtF9LqBw   1   1          1            0      8.9kb          4.4kb
```
Also, let’s verify the data in the indexes:

```bash
curl -XGET -k --user 'elastic:tS$k!2IBI.ASI7FJ' "https://localhost:9200/info/_search?pretty"
{
  "took" : 79,
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
        "_index" : "info",
        "_type" : "_doc",
        "_id" : "mQCvA4ABs70-lBxlFWZD",
        "_score" : 1.0,
        "_source" : {
          "Company" : "AppsCode Inc",
          "Product" : "KubeDB"
        }
      }
    ]
  }
}

```


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