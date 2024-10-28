---
title: Combined Cluster
menu:
  docs_{{ .version }}:
    identifier: sl-combined-solr
    name: Combined Cluster
    parent: sl-clustering-solr
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Combined Cluster

An Solr combined cluster is a group of one or more Solr nodes where each node can perform as overseer, data, and coordinator nodes simultaneously.

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

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/solr/yamls) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create Standalone Solr Cluster

The KubeDB Solr runs in `solrcloud` mode. Hence, it needs a external zookeeper to distribute replicas among pods and save configurations.

We will use KubeDB ZooKeeper for this purpose.

The ZooKeeper instance used for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zoo-com
  namespace: demo
spec:
  version: 3.8.3
  replicas: 3
  deletionPolicy: Delete
  adminServerPort: 8080
  storage:
    resources:
      requests:
        storage: "100Mi"
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
```

We have to apply zookeeper first and wait till atleast pods are running to make sure that a cluster has been formed.

Here,

- `spec.version` - is the name of the ZooKeeperVersion CR. Here, a ZooKeeper of version `3.8.3` will be created.
- `spec.replicas` - specifies the number of ZooKeeper nodes.
- `spec.storageType` - specifies the type of storage that will be used for ZooKeeper database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the ZooKeeper database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the Petsets created by the KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests. If you don't specify `spec.storageType: Ephemeral`, then this field is required.
- `spec.deletionPolicy` specifies what KubeDB should do when a user try to delete ZooKeeper CR. Deletion policy `Delete` will delete the database pods, secret and PVC when the ZooKeeper CR is deleted. Checkout the [link](/docs/guides/zookeeper/concepts/zookeeper.md#specdeletionpolicy) for details.

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in the `storage.resources.requests` field. Don't specify `limits` here. PVC does not get resized automatically.

Let's create the ZooKeeper CR that is shown above:

```bash
$ $ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/solr/quickstart/overview/yamls/zookeeper/zookeeper.yaml
zooKeeper.kubedb.com/zoo-com created
```

The ZooKeeper's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the database.

```bash
$ kubectl get ZooKeeper -n demo -w
NAME       TYPE                  VERSION   STATUS   AGE
zoo-com    kubedb.com/v1alpha2   3.7.2     Ready    13m

Here, we are going to create a standalone (ie. `replicas: 1`) Solr cluster. We will use the Solr image provided by the Solr (`9.6.1`) for this demo. To learn more about Solr CR, visit [here](/docs/guides/solr/concepts/solr.md).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-combined
  namespace: demo
spec:
  version: 9.6.1
  deletionPolicy: DoNotTerminate
  replicas: 2
  enableSSL: true
  zookeeperRef:
    name: zoo-com
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
```

Let's deploy the above example by the following command:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/solr/clustering/yamls/combined-standalone.yaml
solr.kubedb.com/solr-combined created
```

Watch the bootstrap progress:

```bash
$ kubectl get sl -n demo
NAME            TYPE                  VERSION   STATUS   AGE
solr-combined   kubedb.com/v1alpha2   9.6.1     Ready    3h37m

```

Hence the cluster is ready to use.
Let's check the k8s resources created by the operator on the deployment of Elasticsearch CRO:

```bash
$ kubectl get all,secret,pvc -n demo  -l 'app.kubernetes.io/instance=solr-combined'
NAME                  READY   STATUS    RESTARTS   AGE
pod/solr-combined-0   1/1     Running   0          75s

NAME                         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/solr-combined        ClusterIP   10.96.247.33   <none>        8983/TCP   78s
service/solr-combined-pods   ClusterIP   None           <none>        8983/TCP   78s

NAME                                               TYPE              VERSION   AGE
appbinding.appcatalog.appscode.com/solr-combined   kubedb.com/solr   9.6.1     78s

NAME                                      TYPE                       DATA   AGE
secret/solr-combined-auth                 kubernetes.io/basic-auth   2      78s
secret/solr-combined-auth-config          Opaque                     1      78s
secret/solr-combined-client-cert          kubernetes.io/tls          5      78s
secret/solr-combined-config               Opaque                     1      78s
secret/solr-combined-keystore-cred        Opaque                     1      78s
secret/solr-combined-server-cert          kubernetes.io/tls          5      78s
secret/solr-combined-zk-digest            kubernetes.io/basic-auth   2      78s
secret/solr-combined-zk-digest-readonly   kubernetes.io/basic-auth   2      78s

NAME                                                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/solr-combined-data-solr-combined-0   Bound    pvc-c073b8b8-9005-41c5-ac21-bd060a5214a1   1Gi        RWO            standard       75s
```

- `PetSet` - a PetSet(Appscode manages customized petset) named after the Solr instance. In topology mode, the operator creates 3 PetSets with name `{Solr-Name}-{Sufix}`.
- `Services` -  2 services are generated for each Solr database.
    - `{Solr-Name}` - the client service which is used to connect to the database. It points to the `overseer` nodes.
    - `{Solr-Name}-pods` - the node discovery service which is used by the Solr nodes to communicate each other. It is a headless service.
- `AppBinding` - an [AppBinding](/docs/guides/solr/concepts/appbinding.md) which hold to connect information for the database. It is also named after the solr instance.
- `Secrets` - 3 types of secrets are generated for each Solr database.
    - `{Solr-Name}-auth` - the auth secrets which hold the `username` and `password` for the solr users. The auth secret `solr-combined-admin-cred` holds the `username` and `password` for `admin` user which lets administrative access.
    - `{Solr-Name}-config` - the default configuration secret created by the operator.
    - `{Solr-Name}-auth-config` - the configuration secret of admin user information created by the operator.
    - `{Solr-Name}-zk-digest` - the auth secret which contains the `username` and `password` for zookeeper digest secret which is able to access zookeeper data.
    - `{Solr-Name}-zk-digest-readonly` - the auth secret which contains the `username` and `password` for zookeeper readonly digest secret which is able to read zookeeper data.


## Connect with Solr Database

We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our Solr database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our Solr database is working well.

Let's port-forward the port `8983` to local machine:

```bash
$ kubectl port-forward -n demo svc/solr-combined 8983
Forwarding from 127.0.0.1:8983 -> 8983
Forwarding from [::1]:8983 -> 8983
```

Now, our Solr cluster is accessible at `localhost:8983`.

**Connection information:**

- Address: `localhost:8983`
- Username:

  ```bash
  $ kubectl get secret -n demo solr-combined-admin-cred -o jsonpath='{.data.username}' | base64 -d
  admin
    ```

- Password:

  ```bash
  $ kubectl get secret -n demo solr-combined-admin-cred -o jsonpath='{.data.password}' | base64 -d
  Xy3ZjyU)~(9IO8_n
  ```

Now let's check the health of our Solr database.

```bash
$ curl -XGET -k -u 'admin:Xy3ZjyU)~(9IO8_n' "http://localhost:8983/solr/admin/collections?action=CLUSTERSTATUS"
{
  "responseHeader":{
    "status":0,
    "QTime":1
  },
  "cluster":{
    "collections":{
      "kubedb-collection":{
        "pullReplicas":"0",
        "configName":"kubedb-system.AUTOCREATED",
        "replicationFactor":1,
        "router":{
          "name":"compositeId"
        },
        "nrtReplicas":1,
        "tlogReplicas":"0",
        "shards":{
          "shard1":{
            "range":"80000000-7fffffff",
            "state":"active",
            "replicas":{
              "core_node2":{
                "core":"kubedb-system_shard1_replica_n1",
                "node_name":"solr-combined-2.solr-combined-pods.demo:8983_solr",
                "type":"NRT",
                "state":"active",
                "leader":"true",
                "force_set_state":"false",
                "base_url":"http://solr-combined-0.solr-combined-pods.demo:8983/solr"
              }
            },
            "health":"GREEN"
          }
        },
        "health":"GREEN",
        "znodeVersion":4
      }
    },
    "live_nodes":["solr-combined-0.solr-combined-pods.demo:8983_solr"]
  }
}
```

## Create Multi-Node Combined Solr Cluster

Here, we are going to create a multi-node (say `replicas: 2`) Solr cluster. We will use the Solr image provided by the Solr (`9.6.1`) for this demo. To learn more about Solr CR, visit [here](/docs/guides/solr/concepts/solr.md).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-combined
  namespace: demo
spec:
  version: 9.6.1
  deletionPolicy: DoNotTerminate
  replicas: 2
  enableSSL: true
  zookeeperRef:
    name: zoo
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
```

Let's deploy the above example by the following command:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/solr/clusteringyamls/combined-multinode.yaml
solr.kubedb.com/solr-combined created
```

Watch the bootstrap progress:

```bash
$ kubectl get sl -n demo
NAME            TYPE                  VERSION   STATUS   AGE
solr-combined   kubedb.com/v1alpha2   9.6.1     Ready    3h37m

```

Hence the cluster is ready to use.
Let's check the k8s resources created by the operator on the deployment of Elasticsearch CRO:

```bash
$ kubectl get all,secret,pvc -n demo  -l 'app.kubernetes.io/instance=solr-combined'
NAME                  READY   STATUS    RESTARTS   AGE
pod/solr-combined-0   1/1     Running   0          75s
pod/solr-combined-1   1/1     Running   0          66s

NAME                         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/solr-combined        ClusterIP   10.96.247.33   <none>        8983/TCP   78s
service/solr-combined-pods   ClusterIP   None           <none>        8983/TCP   78s

NAME                                               TYPE              VERSION   AGE
appbinding.appcatalog.appscode.com/solr-combined   kubedb.com/solr   9.6.1     78s

NAME                                      TYPE                       DATA   AGE
secret/solr-combined-auth                 kubernetes.io/basic-auth   2      78s
secret/solr-combined-auth-config          Opaque                     1      78s
secret/solr-combined-client-cert          kubernetes.io/tls          5      78s
secret/solr-combined-config               Opaque                     1      78s
secret/solr-combined-keystore-cred        Opaque                     1      78s
secret/solr-combined-server-cert          kubernetes.io/tls          5      78s
secret/solr-combined-zk-digest            kubernetes.io/basic-auth   2      78s
secret/solr-combined-zk-digest-readonly   kubernetes.io/basic-auth   2      78s

NAME                                                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/solr-combined-data-solr-combined-0   Bound    pvc-c073b8b8-9005-41c5-ac21-bd060a5214a1   1Gi        RWO            standard       75s
persistentvolumeclaim/solr-combined-data-solr-combined-1   Bound    pvc-69b509b1-5e42-4b7e-a64e-1b8e15b25bc7   1Gi        RWO            standard       66s
```



## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo solr solr-combined -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
solr.kubedb.com/solr-combined patched

$ kubectl delete -n demo sl/solr-combined
solr.kubedb.com "solr-combined" deleted

$  kubectl delete namespace demo
namespace "demo" deleted
```
