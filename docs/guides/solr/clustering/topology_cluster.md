---
title: Topology Cluster
menu:
  docs_{{ .version }}:
    identifier: sl-topology-solr
    name: Topology Cluster
    parent: sl-clustering-solr
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Solr Simple Dedicated Cluster

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

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/solr/clustering/yamls) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will have to provide `StorageClass` in Solr CR specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  1h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Create Solr Topology Cluster

We are going to create a Solr Cluster in topology mode. Our cluster will be composed of 1 overseer nodes, 2 data nodes, 1 coordinator nodes. Here, we are using Solr version ( `9.4.1` ). To learn more about the Solr CR, visit [here](/docs/guides/solr/concepts/solr.md).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: self-signed-issuer
      kind: ClusterIssuer
    certificates:
    - alias: server
      subject:
        organizations:
         - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  deletionPolicy: DoNotTerminate
  version: 9.4.1
  zookeeperRef:
    name: zoo-com
    namespace: demo
  topology:
    overseer:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    data:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    coordinator:
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

Here,

- `spec.version` - is the name of the SolrVersion CR. Here, we are using Solr version `9.4.1`.
- `spec.enableSSL` - specifies whether the HTTP layer is secured with certificates or not.
- `spec.storageType` - specifies the type of storage that will be used for Solr database. It can be `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create the Solr database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.topology` - specifies the node-specific properties for the Solr cluster.
    - `topology.overseer` - specifies the properties of overseer nodes.
        - `overseer.replicas` - specifies the number of overseer nodes.
        - `overseer.storage` - specifies the overseer node storage information that passed to the PetSet.
    - `topology.data` - specifies the properties of data nodes.
        - `data.replicas` - specifies the number of data nodes.
        - `data.storage` - specifies the data node storage information that passed to the PetSet.
    - `topology.coordinator` - specifies the properties of coordinator nodes.
        - `coordinator.replicas` - specifies the number of coordinator nodes.
        - `coordinator.storage` - specifies the coordinator node storage information that passed to the PetSet.

Let's deploy the above example by the following command:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/solr/clustering/yamls/topology.yaml
solr.kubedb.com/solr-cluster created
```
KubeDB will create the necessary resources to deploy the Solr cluster according to the above specification. Let’s wait until the database to be ready to use,

```bash
$ kubectl get sl -n demo
NAME           TYPE                  VERSION   STATUS   AGE
solr-cluster   kubedb.com/v1alpha2   9.4.1     Ready    3d2h
```
Here, Solr is in `Ready` state. It means the database is ready to accept connections.

Describe the Solr object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe sl -n demo solr-cluster
Name:         solr-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Solr
Metadata:
  Creation Timestamp:  2024-10-25T10:51:15Z
  Finalizers:
    kubedb.com
  Generation:        4
  Resource Version:  439177
  UID:               caefb440-1f25-4994-98c0-11fa7afca778
Spec:
  Auth Config Secret:
    Name:  solr-cluster-auth-config
  Auth Secret:
    Name:           solr-cluster-admin-cred
  Deletion Policy:  Delete
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     20
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Pod Placement Policy:
        Name:  default
  Solr Modules:
    s3-repository
    gcs-repository
    prometheus-exporter
  Solr Opts:
    -Daws.accessKeyId=local-identity
    -Daws.secretAccessKey=local-credential
  Storage Type:  Durable
  Topology:
    Coordinator:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  solr
            Resources:
              Limits:
                Memory:  2Gi
              Requests:
                Cpu:     900m
                Memory:  2Gi
            Security Context:
              Allow Privilege Escalation:  false
              Capabilities:
                Drop:
                  ALL
              Run As Non Root:  true
              Run As User:      8983
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-solr
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
              Run As Non Root:  true
              Run As User:      8983
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  8983
      Replicas:        1
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:         1Gi
        Storage Class Name:  standard
      Suffix:                coordinator
    Data:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  solr
            Resources:
              Limits:
                Memory:  2Gi
              Requests:
                Cpu:     900m
                Memory:  2Gi
            Security Context:
              Allow Privilege Escalation:  false
              Capabilities:
                Drop:
                  ALL
              Run As Non Root:  true
              Run As User:      8983
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-solr
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
              Run As Non Root:  true
              Run As User:      8983
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  8983
      Replicas:        1
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:         1Gi
        Storage Class Name:  standard
      Suffix:                data
    Overseer:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  solr
            Resources:
              Limits:
                Memory:  2Gi
              Requests:
                Cpu:     900m
                Memory:  2Gi
            Security Context:
              Allow Privilege Escalation:  false
              Capabilities:
                Drop:
                  ALL
              Run As Non Root:  true
              Run As User:      8983
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-solr
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
              Run As Non Root:  true
              Run As User:      8983
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  8983
      Replicas:        1
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:         1Gi
        Storage Class Name:  standard
      Suffix:                overseer
  Version:                   9.4.1
  Zookeeper Digest Readonly Secret:
    Name:  solr-cluster-zk-digest-readonly
  Zookeeper Digest Secret:
    Name:  solr-cluster-zk-digest
  Zookeeper Ref:
    Name:       zoo
    Namespace:  demo
Status:
  Conditions:
    Last Transition Time:  2024-10-25T10:51:15Z
    Message:               The KubeDB operator has started the provisioning of Solr: demo/solr-cluster
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-10-25T11:05:22Z
    Message:               All desired replicas are ready.
    Observed Generation:   4
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-10-25T11:05:40Z
    Message:               The Solr: demo/solr-cluster is accepting connection
    Observed Generation:   4
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-10-25T11:05:40Z
    Message:               The Solr: demo/solr-cluster is accepting write request.
    Observed Generation:   4
    Reason:                DatabaseWriteAccessCheckSucceeded
    Status:                True
    Type:                  DatabaseWriteAccess
    Last Transition Time:  2024-10-25T11:06:00Z
    Message:               The Solr: demo/solr-cluster is accepting read request.
    Observed Generation:   4
    Reason:                DatabaseReadAccessCheckSucceeded
    Status:                True
    Type:                  DatabaseReadAccess
    Last Transition Time:  2024-10-25T11:05:40Z
    Message:               The Solr: demo/solr-cluster is ready
    Observed Generation:   4
    Reason:                AllReplicasReady,AcceptingConnection,ReadinessCheckSucceeded,DatabaseWriteAccessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-10-25T10:52:31Z
    Message:               The Solr: demo/solr-cluster is successfully provisioned.
    Observed Generation:   1
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```
- Here, in `Status.Conditions`
    - `Conditions.Status` is `True` for the `Condition.Type:ProvisioningStarted` which means database provisioning has been started successfully.
    - `Conditions.Status` is `True` for the `Condition.Type:ReplicaReady` which specifies all replicas are ready in the cluster.
    - `Conditions.Status` is `True` for the `Condition.Type:AcceptingConnection` which means database has been accepting connection request.
    - `Conditions.Status` is `True` for the `Condition.Type:Ready` which defines database is ready to use.
    - `Conditions.Status` is `True` for the `Condition.Type:Provisioned` which specifies Database has been successfully provisioned.

### KubeDB Operator Generated Resources

Let's check the Kubernetes resources created by the operator on the deployment of Solr CRO:

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=solr-cluster'
NAME                             READY   STATUS    RESTARTS   AGE
pod/solr-cluster-coordinator-0   1/1     Running   0          3d2h
pod/solr-cluster-data-0          1/1     Running   0          3d2h
pod/solr-cluster-data-1          1/1     Running   0          3d2h
pod/solr-cluster-overseer-0      1/1     Running   0          3d2h

NAME                        TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
service/solr-cluster        ClusterIP   10.43.2.22   <none>        8983/TCP   3d2h
service/solr-cluster-pods   ClusterIP   None         <none>        8983/TCP   3d2h

NAME                                              TYPE              VERSION   AGE
appbinding.appcatalog.appscode.com/solr-cluster   kubedb.com/solr   9.4.1     3d2h

NAME                                     TYPE                       DATA   AGE
secret/solr-cluster-admin-cred           kubernetes.io/basic-auth   2      10d
secret/solr-cluster-auth-config          Opaque                     1      10d
secret/solr-cluster-config               Opaque                     1      3d2h
secret/solr-cluster-zk-digest            kubernetes.io/basic-auth   2      10d
secret/solr-cluster-zk-digest-readonly   kubernetes.io/basic-auth   2      10d

NAME                                                                 STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/solr-cluster-data-solr-cluster-coordinator-0   Bound    pvc-66f8d7f3-dd6e-4347-8b06-8c4a598c096b   1Gi        RWO            standard       <unset>                 3d2h
persistentvolumeclaim/solr-cluster-data-solr-cluster-data-0          Bound    pvc-6c7c1f9d-68cd-4ed6-b151-6d1b88dccbe0   1Gi        RWO            standard       <unset>                 3d2h
persistentvolumeclaim/solr-cluster-data-solr-cluster-data-1          Bound    pvc-6c7d1f9d-68cd-4ed6-b151-6d1b88dccbe0   1Gi        RWO            standard       <unset>                 3d2h
persistentvolumeclaim/solr-cluster-data-solr-cluster-overseer-0      Bound    pvc-106da684-7414-44a7-97e1-f13b65834c36   1Gi        RWO            standard       <unset>                 3d2h
```

- `PetSet` - 3 PetSets are created for 3 types Solr nodes. The PetSets are named after the Solr instance with given suffix: `{Solr-Name}-{Sufix}`.
- `Services` -  3 services are generated for each Solr database.
    - `{Solr-Name}` - the client service which is used to connect to the database. It points to the `overseer` nodes.
    - `{Solr-Name}-pods` - the node discovery service which is used by the Solr nodes to communicate each other. It is a headless service.
- `AppBinding` - an [AppBinding](/docs/guides/solr/concepts/appbinding.md) which hold the connect information for the database. It is also named after the Elastics
- `Secrets` - 3 types of secrets are generated for each Solr database.
    - `{Solr-Name}-auth` - the auth secrets which hold the `username` and `password` for the Solr users.
    - `{Solr-Name}-config` - the default configuration secret created by the operator.

## Connect with Solr Database

We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our Solr database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our Solr database is working well.

#### Port-forward the Service

KubeDB will create few Services to connect with the database. Let’s check the Services by following command,

```bash
$ kubectl get svc -n demo -l 'app.kubernetes.io/instance=solr-cluster'
NAME                TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
solr-cluster        ClusterIP   10.43.2.22   <none>        8983/TCP   3d2h
solr-cluster-pods   ClusterIP   None         <none>        8983/TCP   3d2h
```
Here, we are going to use `solr-cluster` Service to connect with the database. Now, let’s port-forward the `es-cluster` Service to the port `9200` to local machine:

```bash
$ kubectl port-forward -n demo svc/solr-cluster 8983
Forwarding from 127.0.0.1:8983 -> 8983
```
Now, our Solr cluster is accessible at `localhost:8983`.

#### Export the Credentials

KubeDB also create some Secrets for the database. Let’s check which Secrets have been created by KubeDB for our `es-cluster`.

```bash
$ kubectl get secret -n demo
NAME                              TYPE                       DATA   AGE
solr-cluster-auth                 kubernetes.io/basic-auth   2      10d
solr-cluster-auth-config          Opaque                     1      10d
solr-cluster-config               Opaque                     1      3d2h
solr-cluster-zk-digest            kubernetes.io/basic-auth   2      10d
solr-cluster-zk-digest-readonly   kubernetes.io/basic-auth   2      10d
```
Now, we can connect to the database with `solr-cluster-auth` which contains the admin level credentials to connect with the database.

### Accessing Database Through CLI

To access the database through CLI, we have to get the credentials to access. Let’s export the credentials as environment variable to our current shell :

```bash
$ kubectl get secret -n demo solr-cluster-auth -o jsonpath='{.data.username}' | base64 -d
elastic
$ kubectl get secret -n demo solr-cluster-auth -o jsonpath='{.data.password}' | base64 -d
tS$k!2IBI.ASI7FJ
```

Now, let's check the health of our Solr cluster

```bash
# curl -XGET -k -u 'username:password' https://localhost:9200/_cluster/health?pretty"
$ curl -XGET -k --user "admin:7eONFVgU9BS50eiB" "http://localhost:8983/solr/admin/collections?action=CLUSTERSTATUS"
{
  "responseHeader":{
    "status":0,
    "QTime":1
  },
  "cluster":{
    "collections":{
      "kubedb-system":{
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
                "node_name":"solr-cluster-data-0.solr-cluster-pods.demo:8983_solr",
                "type":"NRT",
                "state":"active",
                "leader":"true",
                "force_set_state":"false",
                "base_url":"http://solr-cluster-data-0.solr-cluster-pods.demo:8983/solr"
              }
            },
            "health":"GREEN"
          }
        },
        "health":"GREEN",
        "znodeVersion":14
      }
    },
    "properties":{
      "urlScheme":"http"
    },
    "live_nodes":["solr-cluster-data-0.solr-cluster-pods.demo:8983_solr","solr-cluster-overseer-0.solr-cluster-pods.demo:8983_solr","solr-cluster-coordinator-0.solr-cluster-pods.demo:8983_solr"]
  }
}

```

## Insert Sample Data

Now, we are going to insert some data into Solr.

```bash
$ curl -XPOST -k -u "admin:7eONFVgU9BS50eiB" "http://localhost:8983/solr/admin/collections?action=CREATE&name=book&numShards=2&replicationFactor=2&wt=xml"
<?xml version="1.0" encoding="UTF-8"?>
<response>

<lst name="responseHeader">
  <int name="status">0</int>
  <int name="QTime">1721</int>
</lst>
<lst name="success">
  <lst name="solr-cluster-data-0.solr-cluster-pods.demo:8983_solr">
    <lst name="responseHeader">
      <int name="status">0</int>
      <int name="QTime">272</int>
    </lst>
    <str name="core">book_shard1_replica_n5</str>
  </lst>
  <lst name="solr-cluster-data-0.solr-cluster-pods.demo:8983_solr">
    <lst name="responseHeader">
      <int name="status">0</int>
      <int name="QTime">273</int>
    </lst>
    <str name="core">book_shard2_replica_n6</str>
  </lst>
  <lst name="solr-cluster-data-0.solr-cluster-pods.demo:8983_solr">
    <lst name="responseHeader">
      <int name="status">0</int>
      <int name="QTime">1145</int>
    </lst>
    <str name="core">book_shard2_replica_n1</str>
  </lst>
  <lst name="solr-cluster-data-0.solr-cluster-pods.demo:8983_solr">
    <lst name="responseHeader">
      <int name="status">0</int>
      <int name="QTime">1150</int>
    </lst>
    <str name="core">book_shard1_replica_n2</str>
  </lst>
</lst>
```
Now, let’s verify that the index have been created successfully.

```bash
$ curl -XGET -k --user "admin:7eONFVgU9BS50eiB" "http://localhost:8983/solr/admin/collections?action=LIST"
{
  "responseHeader":{
    "status":0,
    "QTime":2
  },
  "collections":["book","kubedb-system"]
}
```
Also, let’s verify the data in the indexes:

```bash
$ curl -X POST -u "admin:7eONFVgU9BS50eiB"  http://localhost:8983/solr/book/select -H 'Content-Type: application/json' -d '
                           {
                             "query": "*:*",
                             "limit": 10,
                           }'
{
  "responseHeader":{
    "zkConnected":true,
    "status":0,
    "QTime":1,
    "params":{
      "json":"\n                       {\n                         \"query\": \"*:*\",\n         \"limit\": 10,\n                       }\n                     ",
      "_forwardedCount":"1"
    }
  },
  "response":{
    "numFound":1,
    "start":0,
    "numFoundExact":true,
    "docs":[{
      "id":"1",
      "db":["elasticsearch"],
      "_version_":1814163798543564800
    }]
  }
}

```


## Cleaning Up

To cleanup the k8s resources created by this tutorial, run:

```bash
$ kubectl patch -n demo solr solr-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"

$ kubectl delete Solr -n demo solr-cluster 

# Delete namespace
$ kubectl delete namespace demo
```

## Next Steps

- Monitor your Solr database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md).
- Monitor your Solr database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).
- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).