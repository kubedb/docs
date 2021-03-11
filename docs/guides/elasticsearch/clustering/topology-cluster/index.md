---
title: Elasticsearch Topology Cluster
menu:
  docs_{{ .version }}:
    identifier: es-topology-cluster
    name: Topology Cluster
    parent: es-clustering-elasticsearch
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Topology Cluster

An Elasticsearch  topology cluster is a group of Elasticsearch nodes ( `>= 3` ) where each node is assigned with a dedicated role such as master, data, and ingest. In a topology cluster, there has to be at least one master node, one data node, and one ingest node.

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

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/elasticsearch/clustering/topology-cluster/yamls
) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will have to provide `StorageClass` in Elasticsearch CR specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  1h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Create Dedicated Elasticsearch Cluster

We are going to create a dedicated Elasticsearch cluster in topology mode. Our cluster will be consist of 2 master nodes, 3 data nodes, and 2 ingest nodes. We will the Elasticsearch image provided by the [SearchGurad](https://hub.docker.com/u/floragunncom) ( `searchguard-7.9.3` ) for this demo. To learn more about the Elasticsearch CR, visit [here](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-topology
  namespace: demo
spec:
  enableSSL: true 
  version: searchguard-7.9.3
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

- `spec.version` - is the name of the ElasticsearchVersion CR. Here, an Elasticsearch of version `7.9.3` will be created with `SearchGuard` security plugin.
- `spec.enableSSL` - specifies whether the HTTP layer is secured with certificates or not.
- `spec.storageType` - specifies the type of storage that will be used for Elasticsearch database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the Elasticsearch database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.topology` - specifies the node-specific properties for the Elasticsearch cluster.
  - `topology.master` - specifies the properties of master nodes.
    - `master.replicas` - specifies the number of master nodes.
    - `master.storage` - specifies the master node storage information that passed to the StatefulSet.
  - `topology.data` - specifies the properties of data nodes.
    - `master.replicas` - specifies the number of data nodes.
    - `master.storage` - specifies the data node storage information that passed to the StatefulSet.
  - `topology.ingest` - specifies the properties of ingest nodes.
    - `master.replicas` - specifies the number of ingest nodes.
    - `master.storage` - specifies the ingest node storage information that passed to the StatefulSet.

Let's deploy the above example by the following command:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/clustering/topology-cluster/yamls/es-topology.yaml
elasticsearch.kubedb.com/es-topology created
```

Watch the bootstrap progress:

```bash
$ kubectl get elasticsearch -n demo -w
NAME          VERSION             STATUS         AGE
es-topology   searchguard-7.9.3   Provisioning   15s
es-topology   searchguard-7.9.3   Provisioning   25s
es-topology   searchguard-7.9.3   Provisioning   33s
es-topology   searchguard-7.9.3   Provisioning   2m9s 
es-topology   searchguard-7.9.3   Ready          2m11s
```

Hence the cluster is **ready** to use.

Describe the Elasticsearch object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe elasticsearch -n demo es-topology 
Name:         es-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Elasticsearch
Metadata:
  Creation Timestamp:  2021-03-02T09:26:51Z
  Finalizers:
    kubedb.com
  Generation:  3
  Resource Version:  153315
  UID:               0b1aa286-ccf0-4ac8-bc4f-593eb9df4955
Spec:
  Auth Secret:
    Name:      es-topology-admin-cred
  Enable SSL:  true
  Internal Users:
    Admin:
      Backend Roles:
        admin
      Reserved:  true
    Kibanaro:
    Kibanaserver:
      Reserved:  true
    Logstash:
    Readall:
    Snapshotrestore:
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
                  app.kubernetes.io/instance:    es-topology
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
                  app.kubernetes.io/instance:    es-topology
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        elasticsearches.kubedb.com
              Namespaces:
                demo
              Topology Key:  failure-domain.beta.kubernetes.io/zone
            Weight:          50
      Resources:
      Service Account Name:  es-topology
  Storage Type:              Durable
  Termination Policy:        Delete
  Tls:
    Certificates:
      Alias:  ca
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-topology-ca-cert
      Subject:
        Organizations:
          kubedb
      Alias:  transport
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-topology-transport-cert
      Subject:
        Organizations:
          kubedb
      Alias:  http
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-topology-http-cert
      Subject:
        Organizations:
          kubedb
      Alias:  admin
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-topology-admin-cert
      Subject:
        Organizations:
          kubedb
      Alias:  archiver
      Private Key:
        Encoding:   PKCS8
      Secret Name:  es-topology-archiver-cert
      Subject:
        Organizations:
          kubedb
  Topology:
    Data:
      Replicas:  3
      Resources:
        Limits:
          Cpu:     500m
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
          Cpu:     500m
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
          Cpu:     500m
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
  Version:                   searchguard-7.9.3
Status:
  Conditions:
    Last Transition Time:  2021-03-02T09:26:51Z
    Message:               The KubeDB operator has started the provisioning of Elasticsearch: demo/es-topology
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2021-03-02T09:27:24Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2021-03-02T09:28:59Z
    Message:               The Elasticsearch: demo/es-topology is accepting client requests.
    Observed Generation:   3
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2021-03-02T09:29:00Z
    Message:               The Elasticsearch: demo/es-topology is ready.
    Observed Generation:   3
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2021-03-02T09:29:02Z
    Message:               The Elasticsearch: demo/es-topology is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     3
  Phase:                   Ready
Events:
  Type    Reason      Age   From                    Message
  ----    ------      ----  ----                    -------
  Normal  Successful   9m   Elasticsearch operator  Successfully  governing service
  Normal  Successful   9m   Elasticsearch operator  Successfully patched Elasticsearch

```

### KubeDB Operator Generated Resources

Let's check the k8s resources created by the operator on the deployment of Elasticsearch CRO:

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=es-topology'
NAME                       READY   STATUS    RESTARTS   AGE
pod/es-topology-data-0     1/1     Running   0          5m55s
pod/es-topology-data-1     1/1     Running   0          5m42s
pod/es-topology-data-2     1/1     Running   0          5m31s
pod/es-topology-ingest-0   1/1     Running   0          5m55s
pod/es-topology-ingest-1   1/1     Running   0          5m45s
pod/es-topology-master-0   1/1     Running   0          5m55s
pod/es-topology-master-1   1/1     Running   0          5m42s

NAME                         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/es-topology          ClusterIP   10.96.143.113   <none>        9200/TCP   5m57s
service/es-topology-master   ClusterIP   None            <none>        9300/TCP   5m57s
service/es-topology-pods     ClusterIP   None            <none>        9200/TCP   5m57s

NAME                                  READY   AGE
statefulset.apps/es-topology-data     3/3     5m55s
statefulset.apps/es-topology-ingest   2/2     5m55s
statefulset.apps/es-topology-master   2/2     5m55s

NAME                                             TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-topology   kubedb.com/elasticsearch   7.9.3     5m55s

NAME                                      TYPE                       DATA   AGE
secret/es-topology-admin-cert             kubernetes.io/tls          3      5m57s
secret/es-topology-admin-cred             kubernetes.io/basic-auth   2      5m56s
secret/es-topology-archiver-cert          kubernetes.io/tls          3      5m56s
secret/es-topology-ca-cert                kubernetes.io/tls          2      5m57s
secret/es-topology-config                 Opaque                     3      5m55s
secret/es-topology-http-cert              kubernetes.io/tls          3      5m57s
secret/es-topology-kibanaro-cred          kubernetes.io/basic-auth   2      5m56s
secret/es-topology-kibanaserver-cred      kubernetes.io/basic-auth   2      5m56s
secret/es-topology-logstash-cred          kubernetes.io/basic-auth   2      5m56s
secret/es-topology-readall-cred           kubernetes.io/basic-auth   2      5m56s
secret/es-topology-snapshotrestore-cred   kubernetes.io/basic-auth   2      5m56s
secret/es-topology-transport-cert         kubernetes.io/tls          3      5m57s

NAME                                              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-es-topology-data-0     Bound    pvc-7cad281b-6b1d-4474-ba76-4347b49cd647   1Gi        RWO            standard       5m55s
persistentvolumeclaim/data-es-topology-data-1     Bound    pvc-64637ae8-48b8-40c8-b0f6-16f52d375b8f   1Gi        RWO            standard       5m42s
persistentvolumeclaim/data-es-topology-data-2     Bound    pvc-2b21c196-e029-4ef6-a515-8b8c2570d00b   1Gi        RWO            standard       5m31s
persistentvolumeclaim/data-es-topology-ingest-0   Bound    pvc-ffc0a0f2-420d-44f2-a075-749f9abaa4d0   1Gi        RWO            standard       5m55s
persistentvolumeclaim/data-es-topology-ingest-1   Bound    pvc-fd9b4f65-00b6-4add-aae9-0f792c1cd620   1Gi        RWO            standard       5m45s
persistentvolumeclaim/data-es-topology-master-0   Bound    pvc-d6f2e28d-92d6-4ea0-b764-a8dc38c013f2   1Gi        RWO            standard       5m55s
persistentvolumeclaim/data-es-topology-master-1   Bound    pvc-74178260-c1d9-47e8-977f-a31ddd97b31d   1Gi        RWO            standard       5m42s
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

Let's port-forward the port `9200` to local machine:

```bash
$ $ kubectl port-forward -n demo svc/es-topology 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, our Elasticsearch cluster is accessible at `localhost:9200`.

**Connection information:**

- Address: `localhost:9200`
- Username:

  ```bash
  $ kubectl get secret -n demo es-topology-admin-cred -o jsonpath='{.data.username}' | base64 -d
  admin
  ```

- Password:

  ```bash
  $ kubectl get secret -n demo es-topology-admin-cred -o jsonpath='{.data.password}' | base64 -d
  MaBoYQX*KAe4na(f
  ```

Now let's check the health of our Elasticsearch database.

```bash
$ curl -XGET -k -u 'admin:MaBoYQX*KAe4na(f' "https://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "es-topology",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 7,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 6,
  "active_shards" : 13,
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

## Cleaning Up

TO cleanup the k8s resources created by this tutorial, run:

```bash
$ kubectl patch -n demo elasticsearch es-topology -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"

$ kubectl delete elasticsearch -n demo es-topology 

# delete namespace
$ kubectl delete namespace demo
```

## Next Steps

- Learn about [taking backup](/docs/guides/elasticsearch/backup/stash.md) of Elasticsearch database using Stash.
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).